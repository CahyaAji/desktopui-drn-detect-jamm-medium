<script lang="ts">
  import DetectorConfig from './DetectorConfig.svelte';
  import MapsConfig from './MapsConfig.svelte';
  import JammerConfig from './JammerConfig.svelte';
</script>

<div class="settings-page">
  <section class="config-card">
    <DetectorConfig />
  </section>
  <section class="config-card">
    <MapsConfig />
  </section>
  <section class="config-card">
    <JammerConfig />
  </section>
</div>

<style>
  .settings-page {
    width: 100%;
    flex: 1;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    align-content: start;
    gap: 0.75rem;
    padding: 1rem;
    overflow: auto;
    background: var(--bg-dark);
  }

  .config-card {
    min-width: 400px;
    padding: 1rem;
    border: 1px solid var(--border-color);
    border-radius: 8px;
    background: var(--bg-panel);
    box-shadow: var(--shadow);
  }
</style>
