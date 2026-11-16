package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// Detection represents a single drone detection record stored in SQLite.
// Exported fields are bound to the Wails frontend.
type Detection struct {
	ID           int64   `json:"id"`
	Timestamp    int64   `json:"timestamp"` // Unix milliseconds
	SerialNumber string  `json:"serialNumber"`
	Model        string  `json:"model"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ScanFreqMHz  float64 `json:"scanFreqMHz"`
	DroneRange   float64 `json:"droneRange"`
	Altitude     float64 `json:"altitude"`
}

const (
	// retentionDuration controls how long detection records are kept.
	// 180 days = ~6 months. Change to 365 for 1 year.
	retentionDuration = 180 * 24 * time.Hour

	// dedupWindowMs is the minimum interval between two detections of the
	// same serial number. If the same drone is seen again within this
	// window, the duplicate is skipped.
	dedupWindowMs = 30_000
)

// DetectionDB wraps a SQLite connection for drone detection persistence.
type DetectionDB struct {
	sql *sql.DB
}

// Open opens (or creates) the SQLite database at dbPath and runs schema
// migration. Returns an error if the database cannot be opened.
func OpenDetectionDB(dbPath string) (*DetectionDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Reasonable pragmas for an embedded desktop database.
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DetectionDB{sql: db}, nil
}

// Close closes the database connection.
func (d *DetectionDB) Close() error {
	return d.sql.Close()
}

// InsertDetection persists a drone detection event, applying a
// deduplication filter: if the same serial number was already recorded
// within dedupWindowMs (30s), the insert is silently skipped. Different
// serial numbers are always inserted regardless of timing.
func (d *DetectionDB) InsertDetection(payload ResultPayload) {
	now := time.Now().UnixMilli()

	// Dedup: skip if same serial number seen within the window.
	var lastTs sql.NullInt64
	err := d.sql.QueryRow(
		"SELECT timestamp FROM detections WHERE serial_number = ? ORDER BY timestamp DESC LIMIT 1",
		payload.SerialNumber,
	).Scan(&lastTs)
	if err == nil && lastTs.Valid && (now-lastTs.Int64) < dedupWindowMs {
		return // duplicate within window — skip
	}

	_, err = d.sql.Exec(
		`INSERT INTO detections (timestamp, serial_number, model, latitude, longitude, scan_freq_mhz, drone_range, altitude)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		now,
		payload.SerialNumber,
		payload.Model,
		payload.DroneLatitude,
		payload.DroneLongitude,
		payload.ScanFreqMHz,
		payload.DroneRange,
		payload.DroneAltitude,
	)
	if err != nil {
		log.Printf("[db] insert detection: %v", err)
	}
}

// GetDetections returns a paginated slice of detection records ordered by
// timestamp descending (newest first).
func (d *DetectionDB) GetDetections(limit, offset int) ([]Detection, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := d.sql.Query(
		`SELECT id, timestamp, serial_number, model, latitude, longitude, scan_freq_mhz, drone_range, altitude
		 FROM detections ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query detections: %w", err)
	}
	defer rows.Close()

	var results []Detection
	for rows.Next() {
		var det Detection
		if err := rows.Scan(
			&det.ID, &det.Timestamp, &det.SerialNumber, &det.Model,
			&det.Latitude, &det.Longitude, &det.ScanFreqMHz,
			&det.DroneRange, &det.Altitude,
		); err != nil {
			return nil, fmt.Errorf("scan detection: %w", err)
		}
		results = append(results, det)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate detections: %w", err)
	}

	return results, nil
}

// GetDetectionCount returns the total number of detection records.
func (d *DetectionDB) GetDetectionCount() (int, error) {
	var count int
	err := d.sql.QueryRow("SELECT COUNT(*) FROM detections").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count detections: %w", err)
	}
	return count, nil
}

// StartAutoCleanup launches a background goroutine that purges records
// older than retentionDuration. It runs once immediately, then every 24
// hours. The goroutine exits when ctx is cancelled.
func (d *DetectionDB) StartAutoCleanup(ctx context.Context) {
	d.purgeOld()
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-ticker.C:
				d.purgeOld()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// purgeOld deletes detection records older than retentionDuration.
func (d *DetectionDB) purgeOld() {
	cutoff := time.Now().Add(-retentionDuration).UnixMilli()
	result, err := d.sql.Exec("DELETE FROM detections WHERE timestamp < ?", cutoff)
	if err != nil {
		log.Printf("[db] purge old detections: %v", err)
		return
	}
	if n, _ := result.RowsAffected(); n > 0 {
		log.Printf("[db] purged %d detection(s) older than %s", n, retentionDuration)
	}
}

// migrate creates the schema if it does not already exist.
func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS detections (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp     INTEGER NOT NULL,
			serial_number TEXT    NOT NULL DEFAULT '',
			model         TEXT    NOT NULL DEFAULT '',
			latitude      REAL    NOT NULL DEFAULT 0.0,
			longitude     REAL    NOT NULL DEFAULT 0.0,
			scan_freq_mhz REAL    NOT NULL DEFAULT 0.0,
			drone_range   REAL    NOT NULL DEFAULT 0.0,
			altitude      REAL    NOT NULL DEFAULT 0.0
		);

		CREATE INDEX IF NOT EXISTS idx_detections_serial_ts
			ON detections(serial_number, timestamp DESC);
	`)
	return err
}
