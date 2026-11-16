<script lang="ts">
  import DetectedDroneList from "./DetectedDroneList.svelte";
  import DetectorConsole from "./DetectorConsole.svelte";
  import DetectorStatus from "./DetectorStatus.svelte";
  import Jammer from "./Jammer.svelte";
  import SimpleDroneResult from "./SimpleDroneResult.svelte";
</script>

<aside class="right-panel">
  <div class="card-list">
    <DetectorStatus />
    <Jammer />
    <DetectedDroneList />
    <!-- <SimpleDroneResult /> -->
    <DetectorConsole />
  </div>
</aside>

<style>
  .right-panel {
    background: var(--bg-dark);
    border-left: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .card-list {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    gap: 0.3rem;
  }

  .right-panel::-webkit-scrollbar {
    width: 6px;
  }

  .right-panel::-webkit-scrollbar-track {
    background: var(--bg-dark);
  }

  .right-panel::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 3px;
  }

  .right-panel::-webkit-scrollbar-thumb:hover {
    background: var(--primary-color);
  }
</style>
