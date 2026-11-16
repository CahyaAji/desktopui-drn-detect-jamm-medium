package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Tune these to change how quickly a dead TCP connection (e.g. unplugged
// cable, powered-off detector) is detected:
//   - keepAliveIdle: how long the connection must be idle before the OS
//     starts sending keepalive probes.
//   - keepAliveInterval: how long to wait between probes.
//   - keepAliveCount: how many unanswered probes before the connection is
//     considered dead.
//
// Worst-case detection time is roughly keepAliveIdle + (keepAliveInterval * keepAliveCount).
const (
	keepAliveIdle     = 2 * time.Second
	keepAliveInterval = 2 * time.Second
	keepAliveCount    = 2
)

// dataIdleTimeout guards against a silently abandoned connection: the
// detector only serves one client at a time, and if another device takes
// over, the detector may simply stop sending us data without ever closing
// our TCP socket (no FIN/RST, so keepalive alone won't notice). If no line
// is received within this window, the connection is treated as dead.
// The detector normally sends a line every 1-3s while connected, so 6s
// (2x the slowest normal interval) leaves margin for jitter/latency.
// Tune this up if you see false disconnects, or down for faster takeover
// detection.
const dataIdleTimeout = 6 * time.Second

// App struct
type App struct {
	ctx        context.Context
	conn       net.Conn
	connGen    uint64
	mu         sync.Mutex
	configMu   sync.RWMutex
	configPath string
	config     AppConfig
	db         *DetectionDB
}

type StatusPayload struct {
	ScanFreqMHz float64 `json:"scanFreqMHz"`
	GainDb      float64 `json:"gainDb"`
}

type ResultPayload struct {
	ScanFreqMHz    float64 `json:"scanFreqMHz"`
	GainDb         float64 `json:"gainDb"`
	Model          string  `json:"model"`
	SerialNumber   string  `json:"serialNumber"`
	FreqMHz        float64 `json:"freqMHz"`
	BandMHz        float64 `json:"bandMHz"`
	DroneLongitude float64 `json:"droneLongitude"`
	DroneLatitude  float64 `json:"droneLatitude"`
	DroneRange     float64 `json:"droneRange"`
	DroneHeight    float64 `json:"droneHeight"`
	DroneAltitude  float64 `json:"droneAltitude"`
	HomeLongitude  float64 `json:"homeLongitude"`
	HomeLatitude   float64 `json:"homeLatitude"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	config, path, err := loadOrCreateConfig()
	if err != nil {
		// Keep the app usable with defaults if the config directory is unavailable.
		a.config = defaultConfig()
	} else {
		a.configMu.Lock()
		a.config = config
		a.configPath = path
		a.configMu.Unlock()
	}

	// Open the detection database alongside the config file.
	dbPath := filepath.Join(filepath.Dir(a.configPath), "detections.db")
	if a.configPath == "" {
		// Fallback: config dir was unavailable, place DB next to the binary.
		if exe, err := os.Executable(); err == nil {
			dbPath = filepath.Join(filepath.Dir(exe), "detections.db")
		} else {
			dbPath = "detections.db"
		}
	}
	db, err := OpenDetectionDB(dbPath)
	if err != nil {
		log.Printf("[app] failed to open detection database: %v", err)
	} else {
		a.db = db
		a.db.StartAutoCleanup(a.ctx)
	}
}

func (a *App) GetConfig() AppConfig {
	a.configMu.RLock()
	defer a.configMu.RUnlock()
	return a.config
}

func (a *App) SaveConfig(config AppConfig) error {
	if config.Version == 0 {
		config.Version = 1
	}
	if config.Detector.Address == "" {
		return fmt.Errorf("detector address is required")
	}
	if config.Detector.Port < 1 || config.Detector.Port > 65535 {
		return fmt.Errorf("detector port must be between 1 and 65535")
	}
	if config.Detector.DetectionTimeoutSeconds < 5 || config.Detector.DetectionTimeoutSeconds > 60 {
		return fmt.Errorf("detection timeout must be between 5 and 60 seconds")
	}
	if config.Jammer.Port < 0 || config.Jammer.Port > 65535 {
		return fmt.Errorf("jammer port must be between 0 and 65535")
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()
	if a.configPath == "" {
		path, err := configPath()
		if err != nil {
			return err
		}
		a.configPath = path
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create config directory: %w", err)
		}
	}
	if err := saveConfigFile(a.configPath, config); err != nil {
		return err
	}
	a.config = config
	return nil
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(_ context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("[app] close detection database: %v", err)
		}
	}
}

func (a *App) Connect(address string) error {
	a.mu.Lock()

	// Close any previous connection before starting a new one so its
	// reader goroutine exits and can't emit stale events for this
	// connection's generation.
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}

	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		a.mu.Unlock()
		return err
	}

	// Enable TCP keepalive so a physically severed link (e.g. unplugged
	// cable or powered-off detector) is detected by the OS - and the
	// blocking read below unblocked with an error - even when no
	// application data is flowing.
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(keepAliveIdle)
		_ = tcpConn.SetKeepAliveConfig(net.KeepAliveConfig{
			Enable:   true,
			Idle:     keepAliveIdle,
			Interval: keepAliveInterval,
			Count:    keepAliveCount,
		})
	}

	a.connGen++
	gen := a.connGen
	a.conn = conn
	a.mu.Unlock()

	// start a goroutine to handle incoming data and emit events to the frontend
	go func() {
		scanner := bufio.NewScanner(conn)
		_ = conn.SetReadDeadline(time.Now().Add(dataIdleTimeout))
		for scanner.Scan() {
			_ = conn.SetReadDeadline(time.Now().Add(dataIdleTimeout))
			line := scanner.Text()
			if payload, kind, ok := parseTCPLine(line); ok {
				runtime.EventsEmit(a.ctx, "tcp:data", payload)
				runtime.EventsEmit(a.ctx, "tcp:"+kind, payload)

				// Persist result detections to SQLite (fire-and-forget).
				if kind == "result" {
					if rp, ok := payload.(ResultPayload); ok && a.db != nil {
						a.db.InsertDetection(rp)
					}
				}
				continue
			}

			// Keep emitting raw lines that do not match known packet formats.
			runtime.EventsEmit(a.ctx, "tcp:data", line)
		}

		a.mu.Lock()
		// Only react/emit if this goroutine still owns the active
		// connection - an explicit Disconnect() or a newer Connect()
		// call already handled cleanup and notified the frontend.
		isCurrent := a.connGen == gen
		if isCurrent {
			a.conn = nil
		}
		a.mu.Unlock()
		if !isCurrent {
			return
		}

		if err := scanner.Err(); err != nil {
			runtime.EventsEmit(a.ctx, "tcp:error", err.Error())
		}
		runtime.EventsEmit(a.ctx, "tcp:disconnected", nil)
	}()
	return nil
}

// Send writes a string to the open connection
func (a *App) Send(msg string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn == nil {
		return fmt.Errorf("not connected")
	}
	_, err := fmt.Fprintln(a.conn, msg)
	return err
}

// Disconnect closes the connection
func (a *App) Disconnect() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connGen++
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
}

func parseTCPLine(line string) (any, string, bool) {
	parts := splitAndTrim(line)
	if len(parts) < 3 || parts[0] != "#" {
		return nil, "", false
	}

	switch strings.ToLower(parts[1]) {
	case "status":
		payload, ok := parseStatus(parts)
		if !ok {
			return nil, "", false
		}
		return payload, "status", true
	case "result":
		payload, ok := parseResult(parts)
		if !ok {
			return nil, "", false
		}
		return payload, "result", true
	default:
		return nil, "", false
	}
}

func parseStatus(p []string) (StatusPayload, bool) {
	if len(p) <= 10 {
		return StatusPayload{}, false
	}

	scanFreqMHz, ok := parseFloat(p[9])
	if !ok {
		return StatusPayload{}, false
	}
	gainDb, ok := parseFloat(p[10])
	if !ok {
		return StatusPayload{}, false
	}

	return StatusPayload{
		ScanFreqMHz: scanFreqMHz,
		GainDb:      gainDb,
	}, true
}

func parseResult(p []string) (ResultPayload, bool) {
	if len(p) <= 30 {
		return ResultPayload{}, false
	}

	scanFreqMHz, ok := parseFloat(p[9])
	if !ok {
		return ResultPayload{}, false
	}
	gainDb, ok := parseFloat(p[10])
	if !ok {
		return ResultPayload{}, false
	}
	freqMHz, ok := parseFloat(p[16])
	if !ok {
		return ResultPayload{}, false
	}
	bandMHz, ok := parseFloat(p[17])
	if !ok {
		return ResultPayload{}, false
	}
	droneLongitude, ok := parseFloat(p[18])
	if !ok {
		return ResultPayload{}, false
	}
	droneLatitude, ok := parseFloat(p[19])
	if !ok {
		return ResultPayload{}, false
	}
	droneRange, ok := parseFloat(p[20])
	if !ok {
		return ResultPayload{}, false
	}
	droneHeight, ok := parseFloat(p[21])
	if !ok {
		return ResultPayload{}, false
	}
	droneAltitude, ok := parseFloat(p[22])
	if !ok {
		return ResultPayload{}, false
	}
	homeLongitude, ok := parseFloat(p[29])
	if !ok {
		return ResultPayload{}, false
	}
	homeLatitude, ok := parseFloat(p[30])
	if !ok {
		return ResultPayload{}, false
	}

	return ResultPayload{
		ScanFreqMHz:    scanFreqMHz,
		GainDb:         gainDb,
		Model:          p[14],
		SerialNumber:   p[15],
		FreqMHz:        freqMHz,
		BandMHz:        bandMHz,
		DroneLongitude: droneLongitude,
		DroneLatitude:  droneLatitude,
		DroneRange:     droneRange,
		DroneHeight:    droneHeight,
		DroneAltitude:  droneAltitude,
		HomeLongitude:  homeLongitude,
		HomeLatitude:   homeLatitude,
	}, true
}

func splitAndTrim(line string) []string {
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func parseFloat(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (a *App) JammerGet(targetURL string) (string, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// JammerPost sends an HTTP POST request with a JSON body to the jammer and
// returns the raw response body.
func (a *App) JammerPost(targetURL string, body string) (string, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(a.ctx, http.MethodPost, targetURL, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(respBody), nil
}

// GetDetections returns a paginated list of detection records from the
// SQLite database, ordered newest-first. Returns an empty slice if the
// database is not available.
func (a *App) GetDetections(limit, offset int) ([]Detection, error) {
	if a.db == nil {
		return []Detection{}, nil
	}
	return a.db.GetDetections(limit, offset)
}

// GetDetectionCount returns the total number of detection records.
// Returns 0 if the database is not available.
func (a *App) GetDetectionCount() (int, error) {
	if a.db == nil {
		return 0, nil
	}
	return a.db.GetDetectionCount()
}
