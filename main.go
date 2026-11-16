package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// mapsServerAddr is a fixed local port dedicated to serving offline map
// assets (style.json, .pmtiles, sprite, fonts). In `wails dev`, the main
// AssetServer.Handler fallback is unreliable because Wails proxies
// unmatched requests to the Vite dev server instead of invoking it, so
// the frontend fetches map assets from this separate server directly
// (in both dev and packaged builds) for consistent behaviour.
const mapsServerAddr = "localhost:34117"

//go:embed all:frontend/dist
var assets embed.FS

// mapsDir resolves the directory that holds offline map assets
// (pmtiles, style.json, sprite, fonts). It looks next to the executable
// first (for packaged builds), falling back to ./maps for `wails dev`.
func mapsDir() string {
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(exeDir, "maps"),                    // Windows/Linux: alongside the binary
			filepath.Join(exeDir, "..", "Resources", "maps"), // macOS .app bundle: Contents/Resources/maps
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				return c
			}
		}
	}
	return "maps"
}

// mapsHandler serves the maps directory under /maps/, with native
// support for HTTP Range requests (required by the pmtiles protocol).
func mapsHandler() http.Handler {
	return http.StripPrefix("/maps/", http.FileServer(http.Dir(mapsDir())))
}

// mapsHandlerCORS wraps mapsHandler with permissive CORS, for the
// dedicated maps server (see mapsServerAddr): the frontend is served from
// a different origin/port by Wails and fetches these assets directly.
func mapsHandlerCORS() http.Handler {
	handler := mapsHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		handler.ServeHTTP(w, r)
	})
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Serve offline map assets on their own fixed port so the frontend can
	// reach them the same way in `wails dev` and in packaged builds.
	go func() {
		if err := http.ListenAndServe(mapsServerAddr, mapsHandlerCORS()); err != nil {
			log.Println("maps server error:", err)
		}
	}()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Drone-Blocker",
		Width:     1024,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: mapsHandler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []any{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
