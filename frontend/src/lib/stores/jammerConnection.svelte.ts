class JammerConnectionState {
  connected = $state(false);
  mode = $state<0 | 1 | 2>(0); // 0 = OFF, 1 = ON, 2 = AUTO
  enabled = $state(false); // user-toggled polling switch, persisted across page navigation
  bandsActive = $state(0); // number of jammer bands currently ON (from /state)
}

export const jammerConnection = new JammerConnectionState();
