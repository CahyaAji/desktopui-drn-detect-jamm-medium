<script lang="ts">
  import { onMount } from "svelte";
  import { GetConfig, SaveConfig } from "../../../wailsjs/go/main/App";
  import { main } from "../../../wailsjs/go/models";

  let config: main.AppConfig | null = null;
  let latitude = 0;
  let longitude = 0;
  let zoom = 0;
  let message = "";
  let isSaving = false;

  onMount(async () => {
    try {
      config = await GetConfig();
      latitude = config.maps.latitude;
      longitude = config.maps.longitude;
      zoom = config.maps.zoom;
    } catch (error) {
      message = `Unable to load map settings: ${error}`;
    }
  });

  const saveMapsSettings = async () => {
    if (!config) {
      message = "Map settings are still loading.";
      return;
    }

    isSaving = true;
    message = "";
    try {
      const currentConfig = await GetConfig();
      await SaveConfig(
        new main.AppConfig({
          ...currentConfig,
          maps: { latitude, longitude, zoom },
        }),
      );
      config = await GetConfig();
      latitude = config.maps.latitude;
      longitude = config.maps.longitude;
      zoom = config.maps.zoom;
      message = "Map settings saved.";
    } catch (error) {
      message = `Unable to save map settings: ${error}`;
    } finally {
      isSaving = false;
    }
  };
</script>

<div class="config">
  <div class="config-heading">️ Maps Configuration</div>
  <div class="config-content">
    <label for="latitude">Latitude</label>
    <input
      id="latitude"
      type="number"
      placeholder="-6.200000"
      step="any"
      bind:value={latitude}
    />
  </div>
  <div class="config-content">
    <label for="longitude">Longitude</label>
    <input
      id="longitude"
      type="number"
      placeholder="106.816666"
      step="any"
      bind:value={longitude}
    />
  </div>
  <div class="config-content">
    <label for="zoom">Zoom Level</label>
    <input
      id="zoom"
      type="number"
      placeholder="14"
      step="any"
      bind:value={zoom}
    />
  </div>
  <button
    class="save-config-button"
    type="button"
    disabled={isSaving}
    on:click={saveMapsSettings}
  >
    {isSaving ? "Saving..." : "Save Maps Settings"}
  </button>
  <div class="config-message" aria-live="polite">{message}</div>
</div>
