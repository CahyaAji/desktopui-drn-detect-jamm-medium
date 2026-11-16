const net = require("net");
const readline = require("readline");

const PORT = 7;

// ── Configurable simulation parameters ──────────────────────────────────────
const STATUS_INTERVAL_MS = 2000; // send status every 2s
const RESULT_INTERVAL_MS = 8000; // send a drone detection result every 8s

// ── Manual mode ──────────────────────────────────────────────────────────────
// Run with `node DetectorMock.js --manual` to disable the automatic result
// interval; press Enter on the console to trigger a detection on demand.
const MANUAL_MODE = process.argv.includes("--manual");
const activeSockets = new Set();

// ── Simulated telemetry state ───────────────────────────────────────────────
let scanFreqMHz = 2400.0;
let gainDb = 30.0;
let homeLongitude = 106.8456; // Jakarta
let homeLatitude = -6.2088;

// ── Drone simulation pool ───────────────────────────────────────────────────
const drones = [
  {
    model: "DJI-Mavic3",
    serialNumber: "DRN-001",
    freqMHz: 2450.0,
    bandMHz: 20.0,
    longitude: 110.4413263249627,
    latitude: -7.777935221587237,
    range: 150.0,
    height: 80.0,
    altitude: 120.0,
  },
  {
    model: "Autel-EVO2",
    serialNumber: "DRN-002",
    freqMHz: 2480.0,
    bandMHz: 15.0,
    longitude: 110.43980262956426,
    latitude: -7.777773701137387,
    range: 300.0,
    height: 100.0,
    altitude: 150.0,
  },
  // {
  //   model: "DJI-Mini3",
  //   serialNumber: "DRN-003",
  //   freqMHz: 2420.0,
  //   bandMHz: 10.0,
  //   longitude: 110.43969901093178,
  //   latitude: -7.78338228992434,
  //   range: 80.0,
  //   height: 40.0,
  //   altitude: 60.0,
  // },
];

let droneIndex = 0;

// ── Helper: build a comma-separated line with the required field layout ─────
// The Go parser uses splitAndTrim which splits on "," and trims spaces, so
// spaces around commas are harmless but we keep it tight for clarity.

function buildStatusLine() {
  // Indices 0-10 required (len > 10). p[9]=scanFreqMHz, p[10]=gainDb.
  // Fill indices 2-8 with placeholder values so the parser is happy.
  const fields = [
    "#", // 0
    "status", // 1
    "OK", // 2
    "0", // 3
    "1", // 4
    "0", // 5
    "0", // 6
    "0", // 7
    "0", // 8
    scanFreqMHz.toFixed(1), // 9  → scanFreqMHz
    gainDb.toFixed(1), // 10 → gainDb
  ];
  return fields.join(",");
}

function buildResultLine(drone) {
  // Indices 0-30 required (len > 30).
  // p[9]=scanFreqMHz, p[10]=gainDb, p[14]=model, p[15]=serialNumber,
  // p[16]=freqMHz, p[17]=bandMHz, p[18]=droneLongitude, p[19]=droneLatitude,
  // p[20]=droneRange, p[21]=droneHeight, p[22]=droneAltitude,
  // p[29]=homeLongitude, p[30]=homeLatitude.
  const fields = [
    "#", // 0
    "result", // 1
    "DETECT", // 2
    "0", // 3
    "1", // 4
    "0", // 5
    "0", // 6
    "0", // 7
    "0", // 8
    scanFreqMHz.toFixed(1), // 9  → scanFreqMHz
    gainDb.toFixed(1), // 10 → gainDb
    "0", // 11
    "0", // 12
    "0", // 13
    drone.model, // 14 → model
    drone.serialNumber, // 15 → serialNumber
    drone.freqMHz.toFixed(1), // 16 → freqMHz
    drone.bandMHz.toFixed(1), // 17 → bandMHz
    drone.longitude.toFixed(6), // 18 → droneLongitude
    drone.latitude.toFixed(6), // 19 → droneLatitude
    drone.range.toFixed(1), // 20 → droneRange
    drone.height.toFixed(1), // 21 → droneHeight
    drone.altitude.toFixed(1), // 22 → droneAltitude
    "0", // 23
    "0", // 24
    "0", // 25
    "0", // 26
    "0", // 27
    "0", // 28
    homeLongitude.toFixed(6), // 29 → homeLongitude
    homeLatitude.toFixed(6), // 30 → homeLatitude
  ];
  return fields.join(",");
}

// ── Sends a single "detection" result to a given socket ─────────────────────
function sendResult(socket) {
  if (socket.destroyed) return;

  const drone = drones[droneIndex % drones.length];
  droneIndex++;

  // Add slight jitter to drone position
  const jitteredDrone = { ...drone };
  jitteredDrone.longitude += (Math.random() - 0.5) * 0.002;
  jitteredDrone.latitude += (Math.random() - 0.5) * 0.002;
  jitteredDrone.range += (Math.random() - 0.5) * 20;
  jitteredDrone.height += (Math.random() - 0.5) * 10;
  jitteredDrone.altitude += (Math.random() - 0.5) * 10;

  const line = buildResultLine(jitteredDrone);
  socket.write(line + "\n");
  console.log(`[detector →] RESULT  ${drone.model} @ ${jitteredDrone.freqMHz.toFixed(1)} MHz`);
}

// ── TCP Server ──────────────────────────────────────────────────────────────
const server = net.createServer((socket) => {
  const clientAddr = `${socket.remoteAddress}:${socket.remotePort}`;
  console.log(`[detector] Client connected: ${clientAddr}`);
  activeSockets.add(socket);

  // ── Status interval: sends telemetry heartbeat ──────────────────────────
  const statusTimer = setInterval(() => {
    if (socket.destroyed) return;

    // Slightly vary the simulated values for realism
    scanFreqMHz = 2400.0 + Math.random() * 100.0;
    gainDb = 25.0 + Math.random() * 15.0;

    const line = buildStatusLine();
    socket.write(line + "\n");
    console.log(`[detector →] STATUS  ${line}`);
  }, STATUS_INTERVAL_MS);

  // ── Result interval: periodically "detects" a drone (disabled in manual mode) ──
  const resultTimer = MANUAL_MODE ? null : setInterval(() => sendResult(socket), RESULT_INTERVAL_MS);

  // ── Cleanup on disconnect ───────────────────────────────────────────────
  socket.on("close", () => {
    console.log(`[detector] Client disconnected: ${clientAddr}`);
    activeSockets.delete(socket);
    clearInterval(statusTimer);
    if (resultTimer) clearInterval(resultTimer);
  });

  socket.on("error", (err) => {
    console.error(`[detector] Socket error (${clientAddr}):`, err.message);
    activeSockets.delete(socket);
    clearInterval(statusTimer);
    if (resultTimer) clearInterval(resultTimer);
  });

  // ── Echo any data received from the client ──────────────────────────────
  socket.on("data", (data) => {
    console.log(`[detector ←] ${data.toString().trim()}`);
  });
});

// ── Manual trigger: pressing Enter sends a detection to all connected clients ──
if (MANUAL_MODE) {
  const rl = readline.createInterface({ input: process.stdin });
  rl.on("line", () => {
    if (activeSockets.size === 0) {
      console.log("[detector] No clients connected, ignoring Enter press.");
      return;
    }
    for (const socket of activeSockets) {
      sendResult(socket);
    }
  });
}

// ── Start ───────────────────────────────────────────────────────────────────
server.listen(PORT, () => {
  console.log(`============================================`);
  console.log(` Drone Detector Simulator (TCP)`);
  console.log(` Listening on tcp://localhost:${PORT}`);
  console.log(` Status interval : ${STATUS_INTERVAL_MS}ms`);
  if (MANUAL_MODE) {
    console.log(` Result mode     : MANUAL (press Enter to detect)`);
  } else {
    console.log(` Result interval : ${RESULT_INTERVAL_MS}ms`);
  }
  console.log(`============================================`);
});