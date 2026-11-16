<script lang="ts">
  import maplib from "maplibre-gl";
  import "maplibre-gl/dist/maplibre-gl.css";
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

  type MapViewMode = "normal" | "satellite";
  let viewMode = $state<MapViewMode>("normal");
  let maptilerKey = "";

  // MapTiler styles. The API key is read from the backend config
  // (MapsConfig.maptilerKey) and appended to the style URL. The satellite
  // mode uses MapTiler's "hybrid" style (satellite imagery + labels).
  const MAPTILER_STYLE_URLS: Record<MapViewMode, string> = {
    normal: "https://api.maptiler.com/maps/openstreetmap/style.json?key=",
    satellite: "https://api.maptiler.com/maps/hybrid/style.json?key=",
  };

  function buildStyle(
    mode: MapViewMode = viewMode,
    key: string = maptilerKey,
  ): maplib.StyleSpecification | string {
    if (key) return MAPTILER_STYLE_URLS[mode] + encodeURIComponent(key);

    // Fallback: keep the map instance (and the HTML drone/location markers)
    // functional when no MapTiler key has been configured yet.
    return { version: 8, sources: {}, layers: [] };
  }

  function setViewMode(mode: MapViewMode) {
    if (viewMode === mode) return;
    viewMode = mode;
    map?.setStyle(buildStyle(mode));
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
      GetConfig()
        .catch(() => null)
        .then((config) => {
          if (!mapContainer || map) return;
          const mapSettings = config?.maps;
          maptilerKey = mapSettings?.maptilerKey ?? "";
          const center: [number, number] = mapSettings
            ? [mapSettings.longitude, mapSettings.latitude]
            : [110.44058375436, -7.777392115];
          map = new maplib.Map({
            container: mapContainer,
            style: buildStyle(viewMode, maptilerKey),
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
        });
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

  let markerTimeoutMs = 10_000;
  const droneMarkers = new Map<string, DroneMarker>();

  async function refreshMarkerTimeout() {
    try {
      const config = await GetConfig();
      const seconds = Math.min(
        60,
        Math.max(5, config.detector.detectionTimeoutSeconds || 10),
      );
      markerTimeoutMs = seconds * 1000;
    } catch {
      // Keep the previous value if the config cannot be read.
    }
  }

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
      if (now - entry.lastSeen > markerTimeoutMs) {
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
    // Periodic stale pruning, using the configured detection timeout.
    refreshMarkerTimeout();
    const pruneInterval = setInterval(pruneStaleMarkers, 1000);
    const refreshInterval = setInterval(refreshMarkerTimeout, 5000);
    return () => {
      clearInterval(pruneInterval);
      clearInterval(refreshInterval);
    };
  });

  // --- Jammer signal animation ---
  // Any active jammer band: continuous animation
  // No active jammer bands: no animation

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
    const bandsActive = jammerConnection.bandsActive;
    const connected = jammerConnection.connected;

    if (!connected || bandsActive === 0) {
      // No active jammer bands or disconnected — stop animation
      stopJammerAnimation();
      if (jammerAnimTimer) {
        clearTimeout(jammerAnimTimer);
        jammerAnimTimer = null;
      }
      return;
    }

    // At least one jammer band is active — continuous animation
    startJammerAnimation();
    if (jammerAnimTimer) {
      clearTimeout(jammerAnimTimer);
      jammerAnimTimer = null;
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
  <div bind:this={mapContainer} class="map-container">
    <div class="map-view-toggle" role="group" aria-label="Map view mode">
      <button
        class:active={viewMode === "normal"}
        type="button"
        onclick={() => setViewMode("normal")}
      >
        Map
      </button>
      <button
        class:active={viewMode === "satellite"}
        type="button"
        onclick={() => setViewMode("satellite")}
      >
        Satellite
      </button>
    </div>
  </div>
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

  .map-view-toggle {
    position: absolute;
    top: 10px;
    left: 10px;
    z-index: 10;
    display: flex;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    overflow: hidden;
    background: rgba(19, 24, 37, 0.9);
  }

  .map-view-toggle button {
    border: 0;
    background: transparent;
    color: var(--text-secondary);
    padding: 0.45rem 0.8rem;
    cursor: pointer;
    font: inherit;
    font-size: 0.85rem;
    transition:
      background-color 0.2s ease,
      color 0.2s ease;
  }

  .map-view-toggle button:hover {
    background: var(--bg-panel-hover);
  }

  .map-view-toggle button.active {
    background: var(--primary-color);
    color: var(--bg-dark);
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
