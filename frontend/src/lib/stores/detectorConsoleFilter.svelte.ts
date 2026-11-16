import { EventsOn } from "../../../wailsjs/runtime/runtime";

export type ResultPayload = {
  scanFreqMHz: number;
  gainDb: number;
  model: string;
  serialNumber: string;
  freqMHz: number;
  bandMHz: number;
  droneLongitude: number;
  droneLatitude: number;
  droneRange: number;
  droneHeight: number;
  droneAltitude: number;
  homeLongitude: number;
  homeLatitude: number;
  timestamp: number;
};

class DetectorConsoleFilterState {
  entries = $state<ResultPayload[]>([]);

  detectMaxCount = 2;
  detectTimeWindowMs = 10_000;
  maxLines = 10;

  #detectCount = 0;
  #lastDetectMessageTime = 0;
  #started = false;
  #unsubResult: (() => void) | null = null;

  start() {
    if (this.#started) return;
    this.#started = true;
    this.#unsubResult = EventsOn("tcp:result", (payload: ResultPayload) => this.#handleResult(payload));
  }

  stop() {
    if (!this.#started) return;
    this.#started = false;
    this.#unsubResult?.();
    this.#unsubResult = null;
  }

  #handleResult(payload: ResultPayload) {
    const now = Date.now();

    if (now - this.#lastDetectMessageTime > this.detectTimeWindowMs) {
      this.#detectCount = 0;
    }
    this.#lastDetectMessageTime = now;

    if (this.#detectCount < 65535) {
      this.#detectCount += 1;
    }

    if (this.#detectCount >= Math.max(1, this.detectMaxCount)) {
      this.entries = [...this.entries, { ...payload, timestamp: now }];
      if (this.entries.length > this.maxLines) {
        this.entries = this.entries.slice(-this.maxLines);
      }
      this.#detectCount = 0;
    }
  }
}

export const detectorConsoleFilter = new DetectorConsoleFilterState();