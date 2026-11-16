# Drone-Blocker — Desktop UI (Wails)

Desktop application for drone detection and jamming, built with [Wails v2](https://wails.io/) (Go backend + Svelte 5 frontend). Receives real-time drone telemetry over TCP from a hardware detector, displays drones on an offline map, and sends jammer alerts.

---

## Quick Start (Development)

```sh
# Install frontend dependencies
cd frontend && npm install && cd ..

# Run in development mode (hot-reload for both Go and frontend)
wails dev
```

---

## Building an Executable

```sh
wails build
```

This produces a platform-native executable:

| Platform | Output |
|----------|--------|
| **macOS** | `build/bin/desktopui-drn-detect-jamm-lite.app` |
| **Windows** | `build/bin/desktopui-drn-detect-jamm-lite.exe` |
| **Linux** | `build/bin/desktopui-drn-detect-jamm-lite` |

### After Building — Deploy the Offline Map Assets

The app serves offline map tiles from a local `maps/` folder. After building, you must place this folder in the correct location relative to the executable:

| Platform | Where to put `maps/` |
|----------|---------------------|
| **macOS** | Inside the `.app` bundle: `YourApp.app/Contents/Resources/maps/` |
| **Windows** | Next to the `.exe`: same directory as the `.exe` |
| **Linux** | Next to the binary: same directory as the binary |

The `maps/` folder must contain these runtime files:

```
maps/
├── java-island.pmtiles     # vector tile archive (~500 MB)
├── style.json              # MapLibre map style
├── fonts/
│   ├── Noto Sans Bold/     # glyph .pbf files for map labels
│   ├── Noto Sans Italic/
│   └── Noto Sans Regular/
└── sprite/
    ├── sprite.json
    ├── sprite.png
    ├── sprite@2x.json
    └── sprite@2x.png
```

> `data/` and `*.osm.pbf` files are build-time artefacts and are **not** needed at runtime.

The Go backend auto-discovers the `maps/` folder at startup — it checks next to the executable first, then falls back to a relative `maps/` path for `wails dev`. See [`main.go`](main.go) for the resolution logic.

---

## Changing the Offline Maps

There are two levels of change. Full details in [`docs/OFFLINE-MAP-GUIDE.md`](docs/OFFLINE-MAP-GUIDE.md).

### Change how the map *looks* (colors, labels, visibility)

Edit [`maps/style.json`](maps/style.json) directly. **No rebuild needed** — just reload the app window.

- **`layers` order matters** — drawn first-to-last (later layers appear on top)
- **`paint`** controls colors and widths (`fill-color`, `line-width`, etc.)
- **`layout`** controls visibility and text (`visibility: "none"` hides a layer)
- **`source-layer`** names come from the OpenMapTiles schema baked into the `.pmtiles` file (e.g. `water`, `building`, `transportation`, `place`, `poi`)

Example — change water color:
```jsonc
{ "id": "water", "type": "fill", "source": "openmaptiles",
  "source-layer": "water", "paint": { "fill-color": "#3a86ff" } }
```

Example — hide a layer:
```jsonc
{ "id": "poi", "type": "symbol", "source": "openmaptiles",
  "source-layer": "poi", "layout": { "visibility": "none" } }
```

### Change the map *region* (different area, or refresh data)

This is a slow, one-off process using **Planetiler** (requires Java 21+).

1. **Download a `.osm.pbf` extract** from [Geofabrik](https://download.geofabrik.de/) and place it in `maps/` (e.g. `maps/sumatra.osm.pbf`).

2. **Run Planetiler** to convert it to `.pmtiles`:
   ```sh
   /opt/homebrew/opt/openjdk@21/bin/java -Xmx4g -jar \
     maps/planetiler.jar \
     --download \
     --osm-path=/absolute/path/to/maps/sumatra.osm.pbf \
     --output=/absolute/path/to/maps/sumatra.pmtiles
   ```
   - `--download` fetches auxiliary datasets (~few hundred MB) into `maps/data/sources/` — cached and reused for future conversions.
   - `-Xmx4g` gives Planetiler 4 GB of heap. Increase for larger regions.
   - Use **absolute paths** for all arguments.
   - This can take minutes to hours depending on region size. Let it run to completion.

3. **Update [`maps/style.json`](maps/style.json)** — change the source URL to point at the new file:
   ```jsonc
   "sources": {
     "openmaptiles": {
       "type": "vector",
       "url": "pmtiles:///maps/sumatra.pmtiles"
     }
   }
   ```

4. **Update center/zoom** in [`Maps.svelte`](frontend/src/lib/components/Maps.svelte) if the new region is elsewhere on the globe.

---

## Project Structure

```
├── app.go                  # Go backend: TCP detector, jammer, Wails bindings
├── config.go               # App configuration (detector, jammer, maps settings)
├── database.go             # SQLite persistence for drone detection records
├── main.go                 # Entry point + offline map HTTP server
├── wails.json              # Wails project config
├── maps/                   # Offline map assets (pmtiles, style, fonts, sprites)
├── frontend/               # Svelte 5 frontend
│   └── src/
│       ├── App.svelte      # Root component with tab navigation
│       └── lib/
│           ├── components/ # UI components (Maps, DetectorConsole, Jammer, etc.)
│           └── stores/     # Reactive state (drone alerts, detector, jammer)
├── test/                   # Mock detector/jammer for testing
├── docs/                   # Additional documentation
└── plans/                  # Architecture plans
```

---

## Database

Drone detection events are persisted to a SQLite database at `{UserConfigDir}/Anti Drone System/detections.db`. The database is managed by [`database.go`](database.go) and includes:

- **Auto-insert** on every `tcp:result` event from the detector
- **Deduplication**: same serial number within 30 seconds is skipped
- **Auto-cleanup**: records older than 6 months are purged daily
- **Record tab**: view paginated detection history in the app UI

---

## Test Checklist

### Kurang
- draw drones on the maps ✅
- Menu settings ✅
- edit default coordinate ✅
- edit detector ip ✅
- edit jammer ip ✅
- background run for detector receiver ✅
- recorded data ✅
- test long run

### Test run
- Mode Standalone
   - connect detector
   - buzzer menyala jika terdeteksi drone
   - output menyala jika dalam mode jammer ON dan terdeteksi drone
   - manual on dengan tombol jammer ON ditekan 3 detik
   - bisa reconnect jika LAN terputus atau ganti2 mode tetap lancar
- Mode Laptop
   - connect detector
   - buzzer menyala jika terdeteksi drone dam mode jammer auto
   - output menyala jika dalam mode jammer Auto dan terdeteksi drone
   - hanya notif di aplikasi jika jammer OFF
   - nyalakan jammer jika jammer ON
   - bisa reconnect jika LAN terputus atau ganti2 mode tetap lancar
   - detector receiver tetap bekerja jika user unmount component