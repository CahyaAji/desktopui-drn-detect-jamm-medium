// Tracks detected drones and fires jammer alerts app-wide, independent of
// whether DetectedDroneList (or any other component) is currently mounted.
// Started once below at module load, so it keeps receiving "tcp:result"
// events and sending alerts even while the user is on another page.

import { GetConfig, JammerGet } from "../../../wailsjs/go/main/App";
import { detectorConsoleFilter, type ResultPayload } from "./detectorConsoleFilter.svelte";
import { jammerConnection } from "./jammerConnection.svelte";

const FALLBACK_JAMMER_BASE = "http://192.168.17.101:8017";
const STALE_TIMEOUT_MS = 10_000;

export interface DroneEntry {
  id: string;
  model: string;
  serialNumber: string;
  lastSeen: number;
  droneRange: number;
}

async function getJammerAlertUrl(): Promise<string> {
  try {
    const config = await GetConfig();
    const { address, port } = config.jammer;
    if (!address || port < 1 || port > 65535) {
      return `${FALLBACK_JAMMER_BASE}/alert`;
    }
    return `http://${address}:${port}/alert`;
  } catch {
    return `${FALLBACK_JAMMER_BASE}/alert`;
  }
}

function droneId(payload: ResultPayload): string {
  return payload.serialNumber || payload.model || "unknown";
}

class DroneAlertService {
  drones = $state<DroneEntry[]>([]);
  lastAlertAt = $state<number>(0);

  #droneMap = new Map<string, DroneEntry>();
  #audioCtx: AudioContext | null = null;
  #started = false;

  start() {
    if (this.#started) return;
    this.#started = true;

    detectorConsoleFilter.start();

    $effect.root(() => {
      $effect(() => {
        // React to new filtered entries
        detectorConsoleFilter.entries;
        this.#syncFromStore();
      });
    });

    setInterval(() => this.#pruneStale(), 1000);
  }

  #getAudioContext(): AudioContext {
    if (!this.#audioCtx) {
      this.#audioCtx = new AudioContext();
    }
    return this.#audioCtx;
  }

  #playWarningSound() {
    try {
      const ctx = this.#getAudioContext();
      const now = ctx.currentTime;

      // Two-tone alert: 880Hz → 660Hz, 100ms each
      const osc1 = ctx.createOscillator();
      const gain1 = ctx.createGain();
      osc1.type = "square";
      osc1.frequency.value = 880;
      gain1.gain.setValueAtTime(0.3, now);
      gain1.gain.exponentialRampToValueAtTime(0.01, now + 0.12);
      osc1.connect(gain1);
      gain1.connect(ctx.destination);
      osc1.start(now);
      osc1.stop(now + 0.12);

      const osc2 = ctx.createOscillator();
      const gain2 = ctx.createGain();
      osc2.type = "square";
      osc2.frequency.value = 660;
      gain2.gain.setValueAtTime(0.3, now + 0.12);
      gain2.gain.exponentialRampToValueAtTime(0.01, now + 0.24);
      osc2.connect(gain2);
      gain2.connect(ctx.destination);
      osc2.start(now + 0.12);
      osc2.stop(now + 0.24);
    } catch {
      // Audio not available — silently ignore
    }
  }

  async #sendAlert() {
    if (!jammerConnection.connected) return;
    if (jammerConnection.mode !== 2) return; // only in AUTO mode

    try {
      const alertUrl = await getJammerAlertUrl();
      await JammerGet(alertUrl);
      this.lastAlertAt = Date.now();
    } catch (error) {
      console.error("[droneAlert] failed to send alert:", error);
    }
  }

  #upsertDrone(payload: ResultPayload) {
    const id = droneId(payload);

    const existing = this.#droneMap.get(id);
    if (existing) {
      existing.lastSeen = payload.timestamp;
      existing.droneRange = payload.droneRange;
    } else {
      this.#droneMap.set(id, {
        id,
        model: payload.model ?? "-",
        serialNumber: payload.serialNumber ?? "-",
        lastSeen: payload.timestamp,
        droneRange: payload.droneRange,
      });
      this.#playWarningSound();
      this.#sendAlert();
    }
  }

  #syncFromStore() {
    const payloads = detectorConsoleFilter.entries;
    for (const p of payloads) {
      this.#upsertDrone(p);
    }
    this.drones = Array.from(this.#droneMap.values());
  }

  #pruneStale() {
    const now = Date.now();
    let changed = false;
    for (const [id, entry] of this.#droneMap) {
      if (now - entry.lastSeen > STALE_TIMEOUT_MS) {
        this.#droneMap.delete(id);
        changed = true;
      }
    }
    if (changed) {
      this.drones = Array.from(this.#droneMap.values());
    }
  }
}

export const droneAlertService = new DroneAlertService();
droneAlertService.start();
