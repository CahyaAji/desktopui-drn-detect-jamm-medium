// deprecated will delete this file in the future, use droneDetection.svelte.ts instead

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
};

// Shared, verified "result" state. Any component (result panel, map, etc.)
// can import `droneDetection` and read `.result` reactively - the
// verification (debounce) logic below only runs once, no matter how many
// components consume it.
//
// The "status" packet needs no verification and is not handled here - each
// component listens for "tcp:status" directly.
class DroneDetectionState {
  result = $state<ResultPayload | null>(null);

  // User-adjustable verification settings, mirroring the detector's own
  // debounce logic: a "result" packet is only confirmed once it has been
  // seen detectMaxCount times in a row, each arriving within
  // detectTimeWindowMs of the previous one.
  detectMaxCount = $state(3);
  detectTimeWindowMs = $state(3000);

  #detectCount = 0;
  #lastDetectMessageTime = 0;
  #started = false;

  // Safe to call from every consumer's onMount/effect - only subscribes once.
  start() {
    if (this.#started) return;
    this.#started = true;
    EventsOn("tcp:result", (payload: ResultPayload) => this.#handleResult(payload));
  }

  #handleResult(payload: ResultPayload) {
    const now = Date.now();
    console.log("[droneDetection] result received:", payload);

    // Reset count if the last "result" was outside the time window.
    if (now - this.#lastDetectMessageTime > this.detectTimeWindowMs) {
      this.#detectCount = 0;
    }
    this.#lastDetectMessageTime = now;

    if (this.#detectCount < 65535) {
      this.#detectCount += 1;
    }

    console.log(
      `[droneDetection] count: ${this.#detectCount} / ${this.detectMaxCount}`,
    );

    if (this.#detectCount >= Math.max(1, this.detectMaxCount)) {
      console.log("[droneDetection] confirmed:", payload);
      this.result = payload;
      this.#detectCount = 0;
    }
  }
}

export const droneDetection = new DroneDetectionState();
