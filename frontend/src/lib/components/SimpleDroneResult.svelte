// deprecated will delete this file in the future, use DetectorConsole.svelte instead

<script lang="ts">
  import { droneDetection } from "../stores/droneDetection.svelte";

  droneDetection.start();

  function formatNumber(value: number | undefined, digits = 2): string {
    if (value === undefined || Number.isNaN(value)) return "-";
    return value.toFixed(digits);
  }
</script>

<div class="card">
  <div class="title-row">
    <h3 class="title">Drone Result <br>(tidak dipakai)</h3>
    <span class="status" class:offline={!droneDetection.result}
      >{droneDetection.result ? "Tracking" : "No Data"}</span
    >
  </div>

  <div class="verification-controls">
    <label>
      Confirm after
      <input
        type="number"
        min="1"
        bind:value={droneDetection.detectMaxCount}
      />
      hits
    </label>
    <label>
      within
      <input
        type="number"
        min="0"
        step="100"
        bind:value={droneDetection.detectTimeWindowMs}
      />
      ms
    </label>
  </div>

  <table class="info-table" aria-label="Drone detection result values">
    <tbody>
      <tr><th>Model</th><td>{droneDetection.result?.model ?? "-"}</td></tr>
      <tr
        ><th>Freq (MHz)</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.freqMHz, 1) : "-"}</td
        ></tr
      >
      <tr
        ><th>Band (MHz)</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.bandMHz, 1) : "-"}</td
        ></tr
      >
      <tr
        ><th>Drone Lon</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.droneLongitude, 6) : "-"}</td
        ></tr
      >
      <tr
        ><th>Drone Lat</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.droneLatitude, 6) : "-"}</td
        ></tr
      >
      <tr
        ><th>Drone Range</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.droneRange, 1) : "-"}</td
        ></tr
      >
      <tr
        ><th>Drone Height</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.droneHeight, 1) : "-"}</td
        ></tr
      >
      <tr
        ><th>Drone Altitude</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.droneAltitude, 1) : "-"}</td
        ></tr
      >
      <tr
        ><th>Home Lon</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.homeLongitude, 6) : "-"}</td
        ></tr
      >
      <tr
        ><th>Home Lat</th><td
          >{droneDetection.result ? formatNumber(droneDetection.result.homeLatitude, 6) : "-"}</td
        ></tr
      >
    </tbody>
  </table>
</div>

<style>
  .verification-controls {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    font-size: 0.8rem;
  }

  .title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
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
    font-size: 0.72rem;
    line-height: 1.2;
    font-weight: 600;
    color: var(--success-color);
    white-space: nowrap;
  }

  .status.offline {
    color: var(--text-secondary);
  }

  .verification-controls label {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    color: var(--text-secondary);
  }

  .verification-controls input {
    width: 4.5rem;
    border: 1px solid var(--border-light);
    border-radius: 6px;
    padding: 0.2rem 0.35rem;
    background: var(--bg-input);
    color: var(--text-primary);
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
    border-bottom: 1px solid var(--border-color);
  }

  .info-table tr:last-child th,
  .info-table tr:last-child td {
    border-bottom: none;
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
