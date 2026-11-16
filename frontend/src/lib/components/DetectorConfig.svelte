<script lang="ts">
  import { onMount } from 'svelte';
  import { GetConfig, SaveConfig } from '../../../wailsjs/go/main/App';
  import { main } from '../../../wailsjs/go/models';

  let config: main.AppConfig | null = null;
  let address = '';
  let port = 0;
  let detectionTimeoutSeconds = 10;
  let message = '';
  let isSaving = false;

  onMount(async () => {
    try {
      config = await GetConfig();
      address = config.detector.address;
      port = config.detector.port;
      detectionTimeoutSeconds = config.detector.detectionTimeoutSeconds || 10;
    } catch (error) {
      message = `Unable to load detector settings: ${error}`;
    }
  });

  const saveDetectorSettings = async () => {
    if (!config) {
      message = 'Detector settings are still loading.';
      return;
    }

    isSaving = true;
    message = '';
    try {
      const currentConfig = await GetConfig();
      await SaveConfig(new main.AppConfig({
        ...currentConfig,
        detector: { address, port, detectionTimeoutSeconds },
      }));
      config = await GetConfig();
      address = config.detector.address;
      port = config.detector.port;
      detectionTimeoutSeconds = config.detector.detectionTimeoutSeconds;
      message = 'Detector settings saved.';
    } catch (error) {
      message = `Unable to save detector settings: ${error}`;
    } finally {
      isSaving = false;
    }
  };
</script>

<div class="config">
  <div class="config-heading">📡 Detector Configuration</div>
  <div class="config-content">
    <label for="detector-address">Detector Address</label>
    <input id="detector-address" type="text" placeholder="192.168.1.100" bind:value={address} />
  </div>
  <div class="config-content">
    <label for="detector-port">Detector Port</label>
    <input id="detector-port" type="number" placeholder="7" bind:value={port} />
  </div>
  <div class="config-content">
    <label for="detector-timeout">Detection Timeout (seconds)</label>
    <input
      id="detector-timeout"
      type="number"
      placeholder="10"
      min="5"
      max="60"
      step="1"
      bind:value={detectionTimeoutSeconds}
    />
  </div>
  <button class="save-config-button" type="button" disabled={isSaving} on:click={saveDetectorSettings}>
    {isSaving ? 'Saving...' : 'Save Detector Settings'}
  </button>
  <div class="config-message" aria-live="polite">{message}</div>
</div>
