<script lang="ts">
  // Detection/alert logic lives in droneAlertService so it keeps running
  // in the background even when this component is unmounted.
  import { droneAlertService } from "../stores/droneAlertService.svelte";

  let drones = $derived(droneAlertService.drones);

  // function formatRange(value: number | undefined): string {
  //   if (value === undefined || Number.isNaN(value)) return "-";
  //   return `${value.toFixed(1)}m`;
  // }
</script>

<div class="card">
  <div class="title-row">
    <h3 class="title">🛸 Detected Drones</h3>
    <span class="count">{drones.length} active</span>
  </div>

  <div class="list-container">
    {#if drones.length === 0}
      <div class="placeholder">No drones detected</div>
    {:else}
      <table class="drone-table" aria-label="Detected drones list">
        <thead>
          <tr>
            <th>ID</th>
            <th>Model</th>
            <!-- <th>Range</th> -->
            <th>Last Seen</th>
          </tr>
        </thead>
        <tbody>
          {#each drones as drone (drone.id)}
            <tr>
              <td class="id-cell" title={drone.serialNumber !== "-" ? `SN: ${drone.serialNumber}` : drone.model}
                >{drone.id}</td
              >
              <td>{drone.model}</td>
              <!-- <td class="num-cell">{formatRange(drone.droneRange)}</td> -->
              <td class="time-cell">{new Date(drone.lastSeen).toLocaleTimeString()}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style>
  .card {
    display: flex;
    flex-direction: column;
    /* flex: 1;
    min-height: 4rem;
    height: 6rem;
    max-height: 16rem; */
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

  .title {
    margin: 0;
    font-size: 1rem;
    line-height: 1.35;
    font-weight: 700;
    color: var(--text-primary);
  }

  .count {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--success-color);
    white-space: nowrap;
  }

  .list-container {
    margin-top: 0.3rem;
    border-radius: 4px;
    border: 1px solid var(--border-color);
    background: var(--bg-input);
    min-height: 8rem;
    max-height: 10rem;
    flex: 1;
    overflow: auto;
    font-family: "JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono",
      "Menlo", monospace;
    font-size: 0.4rem;
  }

  .list-container::-webkit-scrollbar {
    width: 2px;
    height: 2px;
  }

  .list-container::-webkit-scrollbar-track {
    background: transparent;
  }

  .list-container::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 2px;
  }

  .placeholder {
    color: var(--text-muted);
    font-style: italic;
    padding: 0.6rem;
    text-align: center;
  }

  .drone-table {
    width: 100%;
    border-collapse: collapse;
  }

  .drone-table thead {
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .drone-table th {
    text-align: left;
    padding: 0.35rem 0.5rem;
    font-size: 0.6rem;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    background: var(--bg-input);
    border-bottom: 1px solid var(--border-color);
  }

  .drone-table td {
    padding: 0.3rem 0.5rem;
    border-bottom: 1px solid var(--border-color);
    color: var(--text-primary);
    white-space: nowrap;
  }

  .drone-table tbody tr:last-child td {
    border-bottom: none;
  }

  .drone-table tbody tr {
    transition: background 0.2s;
  }

  .drone-table tbody tr:hover {
    background: var(--bg-hover, rgba(255, 255, 255, 0.04));
  }

  .id-cell {
    color: var(--success-color);
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .num-cell {
    text-align: right;
  }

  .time-cell {
    color: var(--text-muted);
    font-size: 0.68rem;
  }
</style>
