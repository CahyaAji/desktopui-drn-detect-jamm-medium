<script lang="ts">
  import maplib from "maplibre-gl";
  import "maplibre-gl/dist/maplibre-gl.css";
  import { Protocol } from "pmtiles";
  import {
    detectorConsoleFilter,
    type ResultPayload,
  } from "../stores/detectorConsoleFilter.svelte";
  import droneSvgRaw from "../assets/drone-marker.svg?raw";
  import { GetConfig } from "../../../wailsjs/go/main/App";
  import { jammerConnection } from "../stores/jammerConnection.svelte";
  import { droneAlertService } from "../stores/droneAlertService.svelte";

  // --- Map state ---

  let mapContainer = $state<HTMLDivElement | null>(null);
  let map: maplib.Map | null = null;
  let locationMarkerEl = $state<HTMLDivElement | null>(null);

  // Registers the "pmtiles://" URL scheme once, globally, so MapLibre can
  // read tiles directly out of a .pmtiles archive via HTTP range requests.
  const protocol = new Protocol();
  maplib.addProtocol("pmtiles", protocol.tile);

  // Offline map assets (style.json, .pmtiles, sprite, fonts) are served by
  // a dedicated Go HTTP server on a fixed port (see mapsServerAddr in
  // main.go) rather than through Wails' main AssetServer, since that one
  // proxies unmatched requests to the Vite dev server in `wails dev`
  // instead of reaching our custom Handler.
  const MAPS_ORIGIN = "http://localhost:34117";

  // MapLibre requires the sprite/glyphs/source URLs in the style to be
  // absolute, so we fetch style.json and absolutize its root-relative
  // paths against MAPS_ORIGIN at runtime.
  async function loadOfflineStyle(): Promise<maplib.StyleSpecification> {
    const res = await fetch(`${MAPS_ORIGIN}/maps/style.json`);
    const style = await res.json();

    if (style.sprite) style.sprite = MAPS_ORIGIN + style.sprite;
    if (style.glyphs) style.glyphs = MAPS_ORIGIN + style.glyphs;
    for (const source of Object.values(style.sources) as any[]) {
      if (
        typeof source.url === "string" &&
        source.url.startsWith("pmtiles:///")
      ) {
        source.url =
          "pmtiles://" + MAPS_ORIGIN + source.url.slice("pmtiles://".length);
      }
    }

    return style;
  }

  function createLocationMarker(): HTMLDivElement {
    const marker = document.createElement("div");
    marker.className = "location-marker";
    marker.setAttribute("aria-label", "Configured location");

    // Signal rings for jammer animation
    for (let i = 1; i <= 3; i++) {
      const ring = document.createElement("div");
      ring.className = `jammer-ring ring-${i}`;
      marker.appendChild(ring);
    }

    return marker;
  }

  $effect(() => {
    if (!mapContainer) return;

    if (!map) {
      Promise.all([loadOfflineStyle(), GetConfig().catch(() => null)]).then(
        ([style, config]) => {
          if (!mapContainer || map) return;
          const mapSettings = config?.maps;
          const center: [number, number] = mapSettings
            ? [mapSettings.longitude, mapSettings.latitude]
            : [110.44058375436, -7.777392115];
          map = new maplib.Map({
            container: mapContainer,
            style,
            center,
            zoom: mapSettings?.zoom ?? 8,
          });

          map.addControl(new maplib.NavigationControl(), "top-right");
          new maplib.Marker({
            element: createLocationMarker(),
            anchor: "center",
          })
            .setLngLat(center)
            .setPopup(
              new maplib.Popup({ offset: 24 }).setText("Configured location"),
            )
            .addTo(map);
        },
      );
    }

    return () => {
      if (map) {
        map.remove();
        map = null;
      }
    };
  });

  // --- Drone markers ---

  interface DroneMarker {
    id: string;
    marker: maplib.Marker;
    lastSeen: number;
  }

  const STALE_TIMEOUT_MS = 10_000;
  const droneMarkers = new Map<string, DroneMarker>();

  function droneId(payload: ResultPayload): string {
    return payload.serialNumber || payload.model || "unknown";
  }

  function createDroneElement(): HTMLDivElement {
    const el = document.createElement("div");
    el.innerHTML = droneSvgRaw;
    el.style.width = "32px";
    el.style.height = "32px";
    el.style.cursor = "pointer";
    el.style.display = "flex";
    el.style.alignItems = "center";
    el.style.justifyContent = "center";
    el.style.overflow = "hidden";
    el.style.borderRadius = "50%";
    el.style.background = "rgba(255, 0, 0, 0.16)";
    el.style.border = "3px solid #ff0000";
    el.style.boxSizing = "border-box";
    el.style.padding = "0";

    const svg = el.querySelector("svg");
    if (svg) {
      svg.setAttribute("width", "100%");
      svg.setAttribute("height", "100%");
      svg.setAttribute("preserveAspectRatio", "xMidYMid meet");
    }

    return el;
  }

  function upsertDroneMarker(payload: ResultPayload) {
    if (!map) return;

    // Gate: skip entries with uninitialized GPS
    if (payload.droneLatitude === 0 && payload.droneLongitude === 0) return;

    const id = droneId(payload);
    const lng = payload.droneLongitude;
    const lat = payload.droneLatitude;

    const existing = droneMarkers.get(id);
    if (existing) {
      existing.lastSeen = payload.timestamp;
      existing.marker.setLngLat([lng, lat]);
    } else {
      const el = createDroneElement();
      const marker = new maplib.Marker({
        element: el,
        anchor: "center",
      })
        .setLngLat([lng, lat])
        .addTo(map);
      droneMarkers.set(id, { id, marker, lastSeen: payload.timestamp });
    }
  }

  function syncMarkersFromStore() {
    const payloads = detectorConsoleFilter.entries;
    for (const p of payloads) {
      upsertDroneMarker(p);
    }
  }

  function pruneStaleMarkers() {
    const now = Date.now();
    for (const [id, entry] of droneMarkers) {
      if (now - entry.lastSeen > STALE_TIMEOUT_MS) {
        entry.marker.remove();
        droneMarkers.delete(id);
      }
    }
  }

  $effect(() => {
    // React to new filtered entries
    detectorConsoleFilter.entries;
    syncMarkersFromStore();
  });

  $effect(() => {
    // Periodic stale pruning
    const interval = setInterval(pruneStaleMarkers, 1000);
    return () => clearInterval(interval);
  });

  // --- Jammer signal animation ---
  // Mode 1 (ON): continuous animation
  // Mode 2 (AUTO): animate for 10 seconds after each alert, then stop
  // Mode 0 (OFF): no animation

  const JAMMER_ANIM_CLASS = "jammer-active";
  const AUTO_ANIM_DURATION_MS = 10_000;

  let jammerAnimTimer: ReturnType<typeof setTimeout> | null = null;

  function startJammerAnimation() {
    const el = document.querySelector(".location-marker");
    if (!el) return;
    el.classList.add(JAMMER_ANIM_CLASS);
  }

  function stopJammerAnimation() {
    const el = document.querySelector(".location-marker");
    if (!el) return;
    el.classList.remove(JAMMER_ANIM_CLASS);
  }

  $effect(() => {
    const mode = jammerConnection.mode;
    const connected = jammerConnection.connected;

    if (!connected || mode === 0) {
      // OFF or disconnected — stop animation
      stopJammerAnimation();
      if (jammerAnimTimer) {
        clearTimeout(jammerAnimTimer);
        jammerAnimTimer = null;
      }
      return;
    }

    if (mode === 1) {
      // ON — continuous animation
      startJammerAnimation();
      if (jammerAnimTimer) {
        clearTimeout(jammerAnimTimer);
        jammerAnimTimer = null;
      }
      return;
    }

    // mode === 2 (AUTO) — stop continuous animation; alerts will trigger it
    if (mode === 2) {
      if (!jammerAnimTimer) {
        stopJammerAnimation();
      }
      return;
    }
  });

  $effect(() => {
    // AUTO mode: react to alert triggers
    const alertTime = droneAlertService.lastAlertAt;
    if (!alertTime) return;

    const mode = jammerConnection.mode;
    const connected = jammerConnection.connected;
    if (!connected || mode !== 2) return;

    // Start animation
    startJammerAnimation();

    // Clear any existing timer
    if (jammerAnimTimer) clearTimeout(jammerAnimTimer);

    // Stop after 10 seconds
    jammerAnimTimer = setTimeout(() => {
      stopJammerAnimation();
      jammerAnimTimer = null;
    }, AUTO_ANIM_DURATION_MS);
  });
</script>

<div class="page-layout">
  <div bind:this={mapContainer} class="map-container"></div>
</div>

<style>
  .page-layout {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    background-color: #0d1117;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
      Helvetica, Arial, sans-serif;
  }

  .map-container {
    flex: 1; /* Dynamic expansion calculation filling the remaining window viewport */
    /* width: 100%; */
    height: 100%;
    position: relative;
  }

  /* Allow jammer rings to expand beyond the marker without clipping */
  :global(.maplibregl-marker) {
    overflow: visible !important;
  }

  :global(.location-marker) {
    width: 20px;
    height: 20px;
    border: 4px solid #ffffff;
    border-radius: 50%;
    background: #1a73e8;
    box-shadow:
      0 0 0 6px rgba(26, 115, 232, 0.28),
      0 1px 4px rgba(0, 0, 0, 0.45);
    position: relative;
    overflow: visible;
  }

  /* Jammer signal — expanding filled circles */
  :global(.jammer-ring) {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: rgba(220, 40, 20, 0.7);
    transform: translate(-50%, -50%) scale(1);
    pointer-events: none;
    opacity: 0;
  }

  :global(.jammer-active .jammer-ring) {
    animation: jammerPulse 3s ease-out infinite;
  }

  :global(.jammer-active .ring-2) {
    animation-delay: 1s;
  }

  :global(.jammer-active .ring-3) {
    animation-delay: 2s;
  }

  @keyframes jammerPulse {
    0% {
      transform: translate(-50%, -50%) scale(1);
      opacity: 0.7;
    }
    100% {
      transform: translate(-50%, -50%) scale(10);
      opacity: 0;
    }
  }
</style>
