<script lang="ts">
  import {
    detectorConsoleFilter,
    type ResultPayload,
  } from "../stores/detectorConsoleFilter.svelte";

  interface ConsoleEntry {
    id: string;
    time: string;
    text: string;
  }

  detectorConsoleFilter.start();

  function formatNumber(value: number | undefined, digits = 2): string {
    if (value === undefined || Number.isNaN(value)) return "-";
    return value.toFixed(digits);
  }

  function formatResult(payload: ResultPayload): string {
    const parts = [
      `${payload.model ?? "-"},`,
      `SN=${payload.serialNumber ?? "-"},`,
      `${formatNumber(payload.droneRange, 1)}m,`,
      `lon=${formatNumber(payload.droneLongitude, 6)}`,
      `lat=${formatNumber(payload.droneLatitude, 6)}`,
    ];
    return parts.join("  ");
  }

  function payloadId(payload: ResultPayload, index: number): string {
    return `${payload.serialNumber ?? "?"}:${payload.freqMHz ?? 0}:${index}`;
  }

  let consoleEntries = $derived.by<ConsoleEntry[]>(() => {
    return [...detectorConsoleFilter.entries].reverse().map((p, i) => ({
      id: payloadId(p, i),
      time: new Date(p.timestamp).toLocaleTimeString(),
      text: formatResult(p),
    }));
  });
</script>

<div class="card">
  <div class="title-row">
    <h3 class="title">📋 Alert</h3>
  </div>

  <div class="console">
    {#if consoleEntries.length === 0}
      <div class="placeholder"></div>
    {:else}
      {#each consoleEntries as entry (entry.id)}
        <div class="console-line">
          <span class="timestamp">[{entry.time}]</span>
          <span class="message">{entry.text}</span>
        </div>
      {/each}
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
    max-height: 10rem; */
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

  .console {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    margin-top: 0.3rem;
    padding: 0.6rem 0.1rem;
    border-radius: 4px;
    border: 1px solid var(--border-color);
    background: var(--bg-input);
    min-height: 7rem;
    max-height: 10rem;
    flex: 1;
    overflow: auto;
    font-family: "JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono",
      "Menlo", monospace;
    font-size: 0.72rem;
  }

  .console::-webkit-scrollbar {
    width: 4px;
    height: 4px;
  }

  .console::-webkit-scrollbar-track {
    background: transparent;
  }

  .console::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 2px;
  }

  .placeholder {
    color: var(--text-muted);
    font-style: italic;
    padding: 0.3rem 0;
  }

  .console-line {
    display: flex;
    gap: 0.2rem;
    white-space: nowrap;
    padding-bottom: 0.2rem;
  }

  .timestamp {
    color: var(--text-muted);
    white-space: nowrap;
    flex-shrink: 0;
    font-size: 0.72rem;
  }

  .message {
    color: var(--success-color);
    white-space: nowrap;
    font-size: 0.72rem;
  }
</style>
