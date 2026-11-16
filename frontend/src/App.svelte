<script lang="ts">
  import { onMount } from 'svelte';
  import DetectionRecord from './lib/components/DetectionRecord.svelte';
  import Maps from './lib/components/Maps.svelte';
  import RightPanel from './lib/components/RightPanel.svelte';
  import SettingsPage from './lib/components/SettingsPage.svelte';

  type TabName = 'dashboard' | 'record' | 'settings';

  let activeTab: TabName = 'dashboard';

  const setTab = (tab: TabName) => {
    activeTab = tab;
    window.location.hash = tab === 'dashboard' ? '' : `#/${tab}`;
  };

  const updatePageFromHash = () => {
    const hash = window.location.hash.replace(/^#\//, '');
    if (hash === 'settings') {
      activeTab = 'settings';
      return;
    }

    if (hash === 'record') {
      activeTab = 'record';
      return;
    }

    activeTab = 'dashboard';
  };

  onMount(() => {
    updatePageFromHash();
    window.addEventListener('hashchange', updatePageFromHash);

    return () => {
      window.removeEventListener('hashchange', updatePageFromHash);
    };
  });
</script>

<div style="display: flex; flex-direction: column; height: 100vh; width: 99vw;">
  <div class="app-title">
    <div class="app-heading">Travix S5</div>

    <div class="tab-bar" aria-label="Main navigation">
      <button
        class:active={activeTab === 'dashboard'}
        class="nav-button"
        type="button"
        on:click={() => setTab('dashboard')}
      >
        Dashboard
      </button>
      <button
        class:active={activeTab === 'record'}
        class="nav-button"
        type="button"
        on:click={() => setTab('record')}
      >
        Record
      </button>
      <button
        class:active={activeTab === 'settings'}
        class="nav-button"
        type="button"
        on:click={() => setTab('settings')}
      >
        Settings
      </button>
    </div>
  </div>

  {#if activeTab === 'settings'}
    <SettingsPage />
  {:else if activeTab === 'record'}
    <DetectionRecord />
  {:else}
    <div class="app-content">
      <div class="map-area">
        <Maps />
      </div>
      <div class="panel-area">
        <RightPanel />
      </div>
    </div>
  {/if}
</div>