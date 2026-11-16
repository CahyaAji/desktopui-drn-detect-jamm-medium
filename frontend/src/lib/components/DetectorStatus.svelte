<script lang="ts">
  import { Connect, Disconnect, GetConfig } from "../../../wailsjs/go/main/App";
  import { EventsOn } from "../../../wailsjs/runtime/runtime";
  import { detectorConnection } from "../stores/detectorConnection.svelte";

  const FALLBACK_DETECTOR_ADDR = "127.0.0.1:8018";
  const RECONNECT_INTERVAL_MS = 5_000;

  let errorMessage = $state("");

  type StatusPayload = {
    scanFreqMHz: number;
    gainDb: number;
  };

  let busy = $state(false);
  let status = $state<StatusPayload | null>(null);
  let reconnectTimer: ReturnType<typeof setInterval> | null = null;

  function stopReconnect() {
    if (reconnectTimer !== null) {
      clearInterval(reconnectTimer);
      reconnectTimer = null;
    }
  }

  async function getDetectorAddress(): Promise<string> {
    try {
      const config = await GetConfig();
      const { address, port } = config.detector;
      if (!address || port < 1 || port > 65535) {
        return FALLBACK_DETECTOR_ADDR;
      }
      return `${address}:${port}`;
    } catch {
      return FALLBACK_DETECTOR_ADDR;
    }
  }

  async function attemptConnect() {
    try {
      const detectorAddress = await getDetectorAddress();
      await Connect(detectorAddress);
      detectorConnection.connected = true;
      stopReconnect();
    } catch (err) {
      console.error("Reconnect attempt failed:", err);
      detectorConnection.connected = false;
    }
  }

  function startReconnect() {
    stopReconnect();
    attemptConnect();
    reconnectTimer = setInterval(() => {
      if (!detectorConnection.enabled) {
        stopReconnect();
        return;
      }
      attemptConnect();
    }, RECONNECT_INTERVAL_MS);
  }

  async function handleConnectionToggle() {
    if (busy) return;
    busy = true;

    if (detectorConnection.enabled) {
      startReconnect();
    } else {
      stopReconnect();
      try {
        await Disconnect();
      } catch (err) {
        console.error("Failed to disconnect detector:", err);
      }
      detectorConnection.connected = false;
    }

    busy = false;
  }

  $effect(() => {
    const unsubStatus = EventsOn("tcp:status", (payload: StatusPayload) => {
      status = payload;
      detectorConnection.connected = true;
      stopReconnect();
    });

    const unsubDisconnected = EventsOn("tcp:disconnected", () => {
      detectorConnection.connected = false;
      if (detectorConnection.enabled) {
        startReconnect();
      }
    });

    const unsubError = EventsOn("tcp:error", (msg: string) => {
      errorMessage = msg;
      console.error("Detector error:", msg);
      detectorConnection.connected = false;
      if (detectorConnection.enabled) {
        startReconnect();
      }
    });

    return () => {
      unsubStatus();
      unsubDisconnected();
      unsubError();
      stopReconnect();
    };
  });
</script>

<div class="card">
  <div class="title-row">
    <h3 class="title">📡 Detector</h3>
    <div class="connection-controls">
      <span class="status" class:offline={!detectorConnection.connected}
        >{detectorConnection.connected ? "Connected" : "Disconnected"}</span
      >
      <label class="slider-toggle">
        <input
          type="checkbox"
          bind:checked={detectorConnection.enabled}
          onchange={handleConnectionToggle}
          disabled={busy}
        />
        <span class="slider-track"></span>
      </label>
    </div>
  </div>
  <table class="info-table" aria-label="Detector status values">
    <tbody>
      <tr>
        <th scope="row">Freq (Mhz)</th>
        <td>{status ? status.scanFreqMHz : "-"}</td>
      </tr>
      <tr>
        <th scope="row">Gain (dB)</th>
        <td>{status ? status.gainDb : "-"}</td>
      </tr>
    </tbody>
  </table>
</div>

<style>
  .title-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.65rem;
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
    border-top: 1px solid var(--border-color);
    border-bottom: 1px solid var(--border-color);
  }

  .title {
    margin: 0;
    font-size: 1rem;
    line-height: 1.35;
    font-weight: 700;
    color: var(--text-primary);
  }

  .status {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.72rem;
    line-height: 1.2;
    font-weight: 600;
    color: color-mix(in srgb, var(--success-color) 80%, white);
    white-space: nowrap;
    justify-self: end;
  }

  .status::before {
    content: "";
    width: 0.4rem;
    height: 0.4rem;
    border-radius: 999px;
    background-color: var(--success-color);
  }

  .status.offline {
    color: color-mix(in srgb, var(--danger-color) 78%, white);
  }

  .status.offline::before {
    background-color: var(--danger-color);
  }

  .connection-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .slider-toggle {
    position: relative;
    display: inline-flex;
    align-items: center;
    cursor: pointer;
    margin-left: 0.5rem;
  }

  .slider-toggle input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider-track {
    width: 44px;
    height: 24px;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    border-radius: 999px;
    position: relative;
    transition:
      background 150ms ease,
      border-color 150ms ease;
  }

  .slider-track::after {
    content: "";
    position: absolute;
    top: 2px;
    left: 2px;
    width: 18px;
    height: 18px;
    border-radius: 999px;
    background: var(--text-secondary);
    transition:
      transform 150ms ease,
      background 150ms ease;
  }

  .slider-toggle input:checked + .slider-track {
    background: var(--primary-color);
    border-color: var(--primary-color);
  }

  .slider-toggle input:checked + .slider-track::after {
    transform: translateX(20px);
    background: var(--bg-dark);
  }

  .slider-toggle input:focus-visible + .slider-track {
    outline: 2px solid var(--primary-color);
    outline-offset: 2px;
  }

  .info-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 0.5rem;
  }

  .info-table th,
  .info-table td {
    padding: 0.4rem 0.55rem;
    font-size: 0.84rem;
  }

  .info-table th {
    text-align: left;
    width: 42%;
    color: var(--text-secondary);
    font-weight: 600;
  }

  .info-table td {
    text-align: right;
    color: var(--text-primary);
    font-weight: 500;
  }

</style>
