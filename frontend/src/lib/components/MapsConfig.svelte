<script lang="ts">
  import { onMount } from "svelte";
  import { GetConfig, SaveConfig } from "../../../wailsjs/go/main/App";
  import { main } from "../../../wailsjs/go/models";

  const DEFAULT_MAPTILER_KEY = "QzjxMrj9AyKbV3tkoZoN";

  let config: main.AppConfig | null = null;
  let latitude = 0;
  let longitude = 0;
  let zoom = 0;
  let maptilerKey = DEFAULT_MAPTILER_KEY;
  let message = "";
  let isSaving = false;
  let showMaptilerKey = false;

  onMount(async () => {
    try {
      config = await GetConfig();
      latitude = config.maps.latitude;
      longitude = config.maps.longitude;
      zoom = config.maps.zoom;
      maptilerKey = config.maps.maptilerKey || DEFAULT_MAPTILER_KEY;
    } catch (error) {
      message = `Unable to load map settings: ${error}`;
    }
  });

  function handleShortcut(event: KeyboardEvent) {
    if (event.ctrlKey && event.shiftKey && event.key.toLowerCase() === "m") {
      event.preventDefault();
      showMaptilerKey = !showMaptilerKey;
    }
  }

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
          maps: { latitude, longitude, zoom, maptilerKey: maptilerKey.trim() || DEFAULT_MAPTILER_KEY },
        }),
      );
      config = await GetConfig();
      latitude = config.maps.latitude;
      longitude = config.maps.longitude;
      zoom = config.maps.zoom;
      maptilerKey = config.maps.maptilerKey || DEFAULT_MAPTILER_KEY;
      message = "Map settings saved.";
    } catch (error) {
      message = `Unable to save map settings: ${error}`;
    } finally {
      isSaving = false;
    }
  };
</script>

<svelte:window on:keydown={handleShortcut} />

<div class="config">
  <div class="config-heading">️ Maps Configuration</div>
  <div class="config-content">
    <label for="latitude">Latitude</label>
    <input
      id="latitude"
      type="number"
      placeholder="-6.175404983314564"
      step="any"
      bind:value={latitude}
    />
  </div>
  <div class="config-content">
    <label for="longitude">Longitude</label>
    <input
      id="longitude"
      type="number"
      placeholder="106.82706340345868"
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
  {#if showMaptilerKey}
    <div class="config-content">
      <label for="maptiler-key">MapTiler API Key</label>
      <input
        id="maptiler-key"
        type="password"
        placeholder="Enter your MapTiler API key"
        autocomplete="off"
        bind:value={maptilerKey}
      />
    </div>
  {:else}
    <div class="config-message" aria-live="polite">
      Press Ctrl + Shift + M to show the MapTiler API key field.
    </div>
  {/if}
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
