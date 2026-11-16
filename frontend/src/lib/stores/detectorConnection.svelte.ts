// Shared detector connection state.
// DetectorStatus.svelte writes to this, and other components (e.g., Jammer.svelte)
// read from it to react to the detector's connection status.

class DetectorConnectionState {
  connected = $state(false);
  enabled = $state(false); // user-toggled connection switch, persisted across page navigation
}

export const detectorConnection = new DetectorConnectionState();
