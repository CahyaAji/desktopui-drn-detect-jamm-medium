<script lang="ts">
  import { onMount } from "svelte";
  import { GetDetections } from "../../../wailsjs/go/main/App";
  import { main } from "../../../wailsjs/go/models";

  const PAGE_SIZE = 50;

  let records = $state<main.Detection[]>([]);
  let currentPage = $state(0);
  let hasMore = $state(false);
  let loading = $state(false);
  let error = $state("");

  async function loadPage(page: number) {
    loading = true;
    error = "";
    try {
      // Fetch one extra row to detect whether more pages exist.
      const detections = await GetDetections(PAGE_SIZE + 1, page * PAGE_SIZE);
      if (detections && detections.length > PAGE_SIZE) {
        records = detections.slice(0, PAGE_SIZE);
        hasMore = true;
      } else {
        records = detections ?? [];
        hasMore = false;
      }
      currentPage = page;
    } catch (err) {
      error = `Failed to load records: ${err}`;
      console.error("[record]", err);
    } finally {
      loading = false;
    }
  }

  function prevPage() {
    if (currentPage > 0) loadPage(currentPage - 1);
  }

  function nextPage() {
    if (hasMore) loadPage(currentPage + 1);
  }

  function formatDate(ts: number): string {
    return new Date(ts).toLocaleString();
  }

  function formatCoord(value: number): string {
    if (value === 0 || value === undefined) return "-";
    return value.toFixed(6);
  }

  function formatFreq(value: number): string {
    if (value === 0 || value === undefined) return "-";
    return `${value.toFixed(1)} MHz`;
  }

  function formatRange(value: number): string {
    if (value === 0 || value === undefined) return "-";
    return `${value.toFixed(1)} m`;
  }

  function formatAlt(value: number): string {
    if (value === 0 || value === undefined) return "-";
    return `${value.toFixed(1)} m`;
  }

  onMount(() => {
    loadPage(0);
  });
</script>

<div class="record-page">
  <div class="record-header">
    <h2 class="record-title">📋 Detection Records</h2>
    <div class="record-meta">
      <button class="refresh-btn" onclick={() => loadPage(currentPage)} disabled={loading}>
        {loading ? "Loading..." : "🔄 Refresh"}
      </button>
    </div>
  </div>

  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  <div class="table-container">
    {#if loading && records.length === 0}
      <div class="placeholder">Loading records...</div>
    {:else if records.length === 0}
      <div class="placeholder">No detection records found.</div>
    {:else}
      <table class="record-table" aria-label="Detection history records">
        <thead>
          <tr>
            <th>#</th>
            <th>Timestamp</th>
            <th>Serial Number</th>
            <th>Model</th>
            <th>Latitude</th>
            <th>Longitude</th>
            <th>Freq</th>
            <th>Range</th>
            <th>Altitude</th>
          </tr>
        </thead>
        <tbody>
          {#each records as rec, i}
            <tr>
              <td class="num-cell">{currentPage * PAGE_SIZE + i + 1}</td>
              <td class="time-cell">{formatDate(rec.timestamp)}</td>
              <td class="sn-cell" title={rec.serialNumber}>{rec.serialNumber || "-"}</td>
              <td>{rec.model || "-"}</td>
              <td class="num-cell">{formatCoord(rec.latitude)}</td>
              <td class="num-cell">{formatCoord(rec.longitude)}</td>
              <td class="num-cell">{formatFreq(rec.scanFreqMHz)}</td>
              <td class="num-cell">{formatRange(rec.droneRange)}</td>
              <td class="num-cell">{formatAlt(rec.altitude)}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  <div class="pagination">
    <button class="page-btn" onclick={prevPage} disabled={currentPage === 0 || loading}>
      ← Previous
    </button>
    <span class="page-info">Page {currentPage + 1}</span>
    <button class="page-btn" onclick={nextPage} disabled={!hasMore || loading}>
      Next →
    </button>
  </div>
</div>

<style>
  .record-page {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding: 1rem 1.5rem;
    gap: 0.75rem;
    background: var(--bg-dark);
    overflow: hidden;
  }

  .record-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .record-title {
    font-size: 1.1rem;
    font-weight: 700;
    color: var(--primary-color);
    margin: 0;
  }

  .record-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .refresh-btn {
    padding: 0.3rem 0.75rem;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-panel);
    color: var(--text-primary);
    cursor: pointer;
    font-size: 0.8rem;
    transition: border-color 0.2s;
  }

  .refresh-btn:hover:not(:disabled) {
    border-color: var(--primary-color);
  }

  .refresh-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .error-banner {
    padding: 0.5rem 0.75rem;
    border-radius: 4px;
    background: rgba(255, 68, 68, 0.12);
    border: 1px solid var(--danger-color);
    color: var(--danger-color);
    font-size: 0.82rem;
  }

  .table-container {
    flex: 1;
    overflow: auto;
    border: 1px solid var(--border-color);
    border-radius: 6px;
    background: var(--bg-panel);
  }

  .table-container::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  .table-container::-webkit-scrollbar-track {
    background: transparent;
  }

  .table-container::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 3px;
  }

  .placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 12rem;
    color: var(--text-muted);
    font-style: italic;
    font-size: 0.9rem;
  }

  .record-table {
    width: 100%;
    border-collapse: collapse;
    font-family: "JetBrains Mono", "Fira Code", "Cascadia Code", "SF Mono",
      "Menlo", monospace;
    font-size: 0.72rem;
  }

  .record-table thead {
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .record-table th {
    text-align: left;
    padding: 0.45rem 0.55rem;
    font-size: 0.68rem;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    background: var(--bg-panel);
    border-bottom: 2px solid var(--border-color);
    white-space: nowrap;
  }

  .record-table td {
    padding: 0.35rem 0.55rem;
    border-bottom: 1px solid var(--border-color);
    color: var(--text-primary);
    white-space: nowrap;
  }

  .record-table tbody tr:last-child td {
    border-bottom: none;
  }

  .record-table tbody tr {
    transition: background 0.15s;
  }

  .record-table tbody tr:hover {
    background: var(--bg-panel-hover);
  }

  .num-cell {
    text-align: right;
    font-variant-numeric: tabular-nums;
  }

  .time-cell {
    color: var(--text-secondary);
    font-size: 0.7rem;
  }

  .sn-cell {
    color: var(--success-color);
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    padding-top: 0.25rem;
  }

  .page-btn {
    padding: 0.35rem 0.9rem;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-panel);
    color: var(--text-primary);
    cursor: pointer;
    font-size: 0.8rem;
    transition: border-color 0.2s, background 0.2s;
  }

  .page-btn:hover:not(:disabled) {
    border-color: var(--primary-color);
    background: var(--bg-panel-hover);
  }

  .page-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .page-info {
    font-size: 0.82rem;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
  }
</style>