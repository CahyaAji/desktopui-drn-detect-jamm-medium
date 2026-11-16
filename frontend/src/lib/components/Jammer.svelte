<script lang="ts">
  import { GetConfig, JammerGet, JammerPost } from "../../../wailsjs/go/main/App";
  import { jammerConnection } from "../stores/jammerConnection.svelte";

  const FALLBACK_JAMMER_BASE = "http://192.168.17.150:8018";
  const poolJammerStatusInterval = 5000; // 5 seconds

  let jammerData = $state("");
  let serverMode = $state<0 | 1 | 2>(0); // 0 = OFF, 1 = ON, 2 = AUTO
  let lastError = $state<string | null>(null);

  type BandKey =
    | "jamm_900"
    | "jamm_1500"
    | "jamm_2400"
    | "jamm_5200"
    | "jamm_5800";

  const JAMM_BANDS: { key: BandKey; label: string }[] = [
    { key: "jamm_900", label: "900 MHz" },
    { key: "jamm_1500", label: "1.5 GHz" },
    { key: "jamm_2400", label: "2.4 GHz" },
    { key: "jamm_5200", label: "5.2 GHz" },
    { key: "jamm_5800", label: "5.8 GHz" },
  ];

  // Current band state as reported by GET /jamm_state.
  let bandStatus = $state<Record<BandKey, boolean>>({
    jamm_900: false,
    jamm_1500: false,
    jamm_2400: false,
    jamm_5200: false,
    jamm_5800: false,
  });

  // Desired band state selected with the band sliders (true = ON).
  let bandSelection = $state<Record<BandKey, boolean>>({
    jamm_900: false,
    jamm_1500: false,
    jamm_2400: false,
    jamm_5200: false,
    jamm_5800: false,
  });

  function updateBandStatus(jammerJson: Record<string, unknown>) {
    let activeCount = 0;
    for (const band of JAMM_BANDS) {
      const value = jammerJson[band.key];
      // /state reports each band as 0 (OFF) or 1 (ON).
      const active = value === 1;
      bandStatus[band.key] = active;
      if (active) activeCount++;
    }
    jammerConnection.bandsActive = activeCount;
  }

  async function applyBandSelection() {
    if (serverMode !== 1) return;
    lastError = null;

    try {
      const jammerBase = await getJammerBase();
      const body = JSON.stringify({
        jamm_900: bandSelection.jamm_900,
        jamm_1500: bandSelection.jamm_1500,
        jamm_2400: bandSelection.jamm_2400,
        jamm_5200: bandSelection.jamm_5200,
        jamm_5800: bandSelection.jamm_5800,
        desktop_mode: 1,
      });
      await JammerPost(`${jammerBase}/deskcmd`, body);

      // Reflect the applied state locally for immediate feedback.
      for (const band of JAMM_BANDS) {
        bandStatus[band.key] = bandSelection[band.key];
      }
      jammerConnection.bandsActive = JAMM_BANDS.filter(
        (band) => bandSelection[band.key],
      ).length;
    } catch (error) {
      lastError = `Error applying jammer bands: ${error}`;
      console.error(lastError);
    }
  }

  async function getJammerBase(): Promise<string> {
    try {
      const config = await GetConfig();
      const { address, port } = config.jammer;
      if (!address || port < 1 || port > 65535) {
        return FALLBACK_JAMMER_BASE;
      }
      return `http://${address}:${port}`;
    } catch {
      return FALLBACK_JAMMER_BASE;
    }
  }

  async function setJammerMode(newMode: 0 | 1 | 2) {
    lastError = null;

    try {
      const jammerBase = await getJammerBase();

      // OFF / ON / AUTO all switch mode via POST /deskcmd. The individual
      // band sliders are applied separately; this request only changes
      // desktop_mode.
      const body = JSON.stringify({
        jamm_900: false,
        jamm_1500: false,
        jamm_2400: false,
        jamm_5200: false,
        jamm_5800: false,
        desktop_mode: newMode,
      });
      await JammerPost(`${jammerBase}/deskcmd`, body);
      serverMode = newMode;
      jammerConnection.mode = newMode;

      // Mode changes disable every band locally until the next /state poll.
      for (const band of JAMM_BANDS) {
        bandStatus[band.key] = false;
      }
      jammerConnection.bandsActive = 0;
    } catch (error) {
      lastError = `Error sending ${newMode} command: ${error}`;
      console.error(lastError);
    }
  }

  $effect(() => {
    let timerId: ReturnType<typeof setTimeout> | null = null;

    async function pollJammerStatus() {
      // Only poll the jammer when the toggle is enabled
      if (!jammerConnection.enabled) {
        timerId = setTimeout(pollJammerStatus, poolJammerStatusInterval);
        jammerConnection.connected = false;
        return;
      }

      try {
        const jammerBase = await getJammerBase();
        jammerData = await JammerGet(`${jammerBase}/state`);
        const jammerJson = JSON.parse(jammerData);
        if (jammerJson.success == true) {
          jammerConnection.connected = true;
          const rawMode: 0 | 1 | 2 = jammerJson.desktop_mode;
          if (rawMode === 0 || rawMode === 1 || rawMode === 2) {
            serverMode = rawMode;
            jammerConnection.mode = rawMode;
          }
          updateBandStatus(jammerJson);
        } else {
          jammerConnection.connected = false;
          console.error("Jammer response indicates failure:", jammerJson);
        }
      } catch (error) {
        console.error("Error fetching jammer status:", error);
        jammerConnection.connected = false;
      } finally {
        timerId = setTimeout(pollJammerStatus, poolJammerStatusInterval);
      }
    }
    pollJammerStatus();

    return () => {
      if (timerId !== null) {
        clearTimeout(timerId);
      }
    };
  });
</script>

<div class="card">
  <div class="title-row">
    <h3 class="title">⚡ Jammer</h3>
    <div class="connection-controls">
      <span class="status" class:offline={!jammerConnection.connected}>
        {jammerConnection.connected ? "Connected" : "Disconnected"}
      </span>
      <label class="slider-toggle">
        <input type="checkbox" bind:checked={jammerConnection.enabled} />
        <span class="slider-track"></span>
      </label>
    </div>
  </div>
  <div class="jammer-grid">
    <div class="jammer-col">
      <span class="led" class:led-active={serverMode === 0}></span>
      <button
        class="btn-toggle"
        disabled={!jammerConnection.connected}
        onclick={() => setJammerMode(0)}>OFF</button
      >
    </div>
    <div class="jammer-col">
      <span class="led" class:led-active={serverMode === 1}></span>
      <button
        class="btn-toggle"
        disabled={!jammerConnection.connected}
        onclick={() => setJammerMode(1)}>ON</button
      >
    </div>
    <div class="jammer-col">
      <span class="led" class:led-active={serverMode === 2}></span>
      <button
        class="btn-toggle"
        disabled={!jammerConnection.connected}
        onclick={() => setJammerMode(2)}>AUTO</button
      >
    </div>
  </div>
  <div class="bands-grid">
    <div class="band-column">
      {#each JAMM_BANDS.slice(0, 3) as band}
        <div class="jam-band-row">
          <span
            class="band-led"
            class:led-active={bandStatus[band.key]}
          ></span>
          <span class="jam-band-label">{band.label}</span>
          <label
            class="band-slider-toggle"
            aria-label={`${band.label} state`}
          >
            <input type="checkbox" bind:checked={bandSelection[band.key]} />
            <span class="band-slider-track"></span>
          </label>
        </div>
      {/each}
    </div>
    <div class="band-column">
      {#each JAMM_BANDS.slice(3) as band}
        <div class="jam-band-row">
          <span
            class="band-led"
            class:led-active={bandStatus[band.key]}
          ></span>
          <span class="jam-band-label">{band.label}</span>
          <label
            class="band-slider-toggle"
            aria-label={`${band.label} state`}
          >
            <input type="checkbox" bind:checked={bandSelection[band.key]} />
            <span class="band-slider-track"></span>
          </label>
        </div>
      {/each}
      <button
        class="apply-bands-button"
        type="button"
        disabled={serverMode !== 1}
        onclick={applyBandSelection}
      >
        Apply
      </button>
    </div>
  </div>
  {#if lastError}
    <p class="jammer-error">{lastError}</p>
  {/if}
</div>

<style>
  .title {
    margin: 0;
    font-size: 1rem;
    line-height: 1.35;
    font-weight: 700;
    color: var(--text-primary);
  }
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

  @keyframes pulse {
    0% {
      opacity: 0.6;
    }
    50% {
      opacity: 1;
    }
    100% {
      opacity: 0.6;
    }
  }

  .jammer-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.5rem;
    margin: 0.5rem;
  }

  .jammer-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.35rem;
  }

  .bands-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    margin: 0.5rem;
  }

  .band-column {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .led {
    width: 100%;
    height: 0.35rem;
    border-radius: 2px;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    transition: all 0.3s ease;
  }

  .led.led-active {
    background: var(--success-color);
    border-color: var(--success-color);
    box-shadow:
      0 0 6px color-mix(in srgb, var(--success-color) 60%, transparent),
      0 0 12px color-mix(in srgb, var(--success-color) 30%, transparent);
    animation: pulse 1.5s ease-in-out infinite;
  }

  .btn-toggle {
    width: 100%;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-toggle:hover {
    background: var(--bg-panel-hover);
    border-color: var(--primary-color);
  }

  .btn-toggle:active:not(:disabled) {
    transform: scale(0.95);
    background: var(--primary-color);
    color: var(--bg-dark);
  }

  .btn-toggle:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .jam-band-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    height: 2.25rem;
  }

  .jam-band-label {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .band-led {
    width: 0.6rem;
    height: 0.6rem;
    flex: 0 0 auto;
    border-radius: 999px;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    transition: all 0.3s ease;
  }

  .band-led.led-active {
    background: var(--success-color);
    border-color: var(--success-color);
    box-shadow:
      0 0 6px color-mix(in srgb, var(--success-color) 60%, transparent),
      0 0 12px color-mix(in srgb, var(--success-color) 30%, transparent);
  }

  .band-slider-toggle {
    position: relative;
    display: inline-flex;
    align-items: center;
    cursor: pointer;
  }

  .band-slider-toggle input {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .band-slider-track {
    width: 40px;
    height: 20px;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    border-radius: 999px;
    position: relative;
    transition:
      background 150ms ease,
      border-color 150ms ease;
  }

  .band-slider-track::after {
    content: "";
    position: absolute;
    top: 2px;
    left: 2px;
    width: 14px;
    height: 14px;
    border-radius: 999px;
    background: var(--text-muted);
    transition:
      transform 150ms ease,
      background 150ms ease;
  }

  .band-slider-toggle input:checked + .band-slider-track {
    background: var(--bg-panel-hover);
    border-color: var(--border-light);
  }

  .band-slider-toggle input:checked + .band-slider-track::after {
    transform: translateX(20px);
    background: var(--text-secondary);
  }

  .band-slider-toggle input:focus-visible + .band-slider-track {
    outline: 2px solid var(--border-light);
    outline-offset: 2px;
  }

  .apply-bands-button {
    width: 100%;
    height: 2.25rem;
    padding: 0 0.4rem;
    background: var(--bg-input);
    border: 1px solid var(--primary-color);
    border-radius: 4px;
    color: var(--primary-color);
    font: inherit;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition:
      background-color 0.2s ease,
      color 0.2s ease;
  }

  .apply-bands-button:hover:not(:disabled) {
    background: var(--primary-color);
    color: var(--bg-dark);
  }

  .apply-bands-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    border-color: var(--border-color);
    color: var(--text-muted);
  }

  .jammer-error {
    margin: 0.4rem 0.5rem 0.5rem;
    padding: 0.4rem 0.5rem;
    font-size: 0.7rem;
    color: var(--danger-color);
    background: color-mix(in srgb, var(--danger-color) 12%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger-color) 25%, transparent);
    border-radius: 4px;
  }
</style>
