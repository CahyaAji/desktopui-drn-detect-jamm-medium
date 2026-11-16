<script lang="ts">
  import { onMount } from "svelte";
  import { GetConfig, SaveConfig } from "../../../wailsjs/go/main/App";
  import { main } from "../../../wailsjs/go/models";

  let config: main.AppConfig | null = null;
  let address = "";
  let port = 0;
  let threshold = 0;
  let message = "";
  let isSaving = false;

  onMount(async () => {
    try {
      config = await GetConfig();
      address = config.jammer.address;
      port = config.jammer.port;
      threshold = config.jammer.threshold;
    } catch (error) {
      message = `Unable to load jammer settings: ${error}`;
    }
  });

  const saveJammerSettings = async () => {
    if (!config) {
      message = "Jammer settings are still loading.";
      return;
    }

    isSaving = true;
    message = "";
    try {
      const currentConfig = await GetConfig();
      await SaveConfig(
        new main.AppConfig({
          ...currentConfig,
          jammer: { address, port, threshold },
        }),
      );
      config = await GetConfig();
      address = config.jammer.address;
      port = config.jammer.port;
      threshold = config.jammer.threshold;
      message = "Jammer settings saved.";
    } catch (error) {
      message = `Unable to save jammer settings: ${error}`;
    } finally {
      isSaving = false;
    }
  };
</script>

<div class="config">
  <div class="config-heading">📡 Jammer Configuration</div>

  <div class="config-content">
    <label for="jammer-address">Jammer Address</label>
    <input
      id="jammer-address"
      type="text"
      placeholder="192.168.1.150"
      bind:value={address}
    />
  </div>

  <div class="config-content">
    <label for="jammer-port">Jammer Port</label>
    <input
      id="jammer-port"
      type="number"
      placeholder="8000"
      bind:value={port}
    />
  </div>

  <div class="config-content">
    <label for="jammer-threshold">Auto Jam Threshold (meters)</label>
    <input
      id="jammer-threshold"
      type="number"
      placeholder="1000"
      min="0"
      step="any"
      bind:value={threshold}
    />
  </div>

  <button
    class="save-config-button"
    type="button"
    disabled={isSaving}
    on:click={saveJammerSettings}
  >
    {isSaving ? "Saving..." : "Save Jammer Settings"}
  </button>
  <div class="config-message" aria-live="polite">{message}</div>
</div>
