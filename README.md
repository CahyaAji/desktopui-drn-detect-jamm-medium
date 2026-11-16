# Drone-Blocker Desktop UI

Desktop UI for the drone detection and jamming system.

## Jammer HTTP flow

This section documents how the app talks to the jammer device, so the URL and
payload can be changed in one place.

### Transport

All jammer requests go through a single Go function:

- [`JammerGet(targetURL string)`](app.go:412) — performs a **GET** request with
  a **3 second timeout**, **no request body**, and **no custom headers**. It
  returns the raw response body as a string and returns an error on network
  failure or any HTTP status `>= 400`.

The frontend binding is [`JammerGet(arg1)`](frontend/wailsjs/go/main/App.js:25),
typed as `JammerGet(arg1: string): Promise<string>`.

### Base URL

- Config comes from [`JammerConfig`](config.go:24): `address`, `port`, and
  `threshold` (stored in `config.json`).
- [`getJammerBase()`](frontend/src/lib/components/Jammer.svelte:12) builds:
  `http://${address}:${port}`
- Fallback when `address`/`port` is invalid:
  `http://192.168.17.101:8017` (defined at
  [`Jammer.svelte`](frontend/src/lib/components/Jammer.svelte:5) and duplicated
  at [`droneAlertService.svelte.ts`](frontend/src/lib/stores/droneAlertService.svelte.ts:10)).

### Requests and expected responses

#### 1. `GET {base}/jamm_state` — status polling

Called every 5 seconds when the jammer toggle is enabled:
[`Jammer.svelte`](frontend/src/lib/components/Jammer.svelte:67).

Expected response:

```json
{
  "success": true,
  "desktop_mode": 0,
  "jamm_900": true,
  "jamm_1500": true,
  "jamm_2400": true,
  "jamm_5200": false,
  "jamm_5800": true
}
```

- `success` must be `true`, otherwise the app marks the jammer as disconnected.
- `desktop_mode` must be `0` (OFF), `1` (ON), or `2` (AUTO). The `jamm_*`
  fields are currently informational and ignored by the frontend.

#### 2. `GET {base}/manualoff` — OFF button

Built in [`setJammerMode()`](frontend/src/lib/components/Jammer.svelte:25) and
sent at [`Jammer.svelte`](frontend/src/lib/components/Jammer.svelte:38).

Expected response:

```json
{ "success": true, "mode": 0 }
```

#### 3. `GET {base}/manualon` — ON button

Same function, endpoint `/manualon`.

Expected response:

```json
{ "success": true, "mode": 1 }
```

#### 4. `GET {base}/auto` — AUTO button

Same function, endpoint `/auto`.

Expected response:

```json
{ "success": true, "mode": 2 }
```

#### 5. `GET {base}/alert` — automatic drone alert

Sent from [`#sendAlert()`](frontend/src/lib/stores/droneAlertService.svelte.ts:102)
when a new drone is confirmed and the jammer is in AUTO mode. The URL is built
in [`getJammerAlertUrl()`](frontend/src/lib/stores/droneAlertService.svelte.ts:21).

The response is **not parsed** — it only needs a non-error HTTP status. The mock
returns:

```json
{ "success": true, "message": "Alert triggered" }
```

### Note on `threshold`

`jammer.threshold` is saved in
[`JammerConfig.svelte`](frontend/src/lib/components/JammerConfig.svelte:36) but
is **not used** in the current jammer HTTP flow. It is already available in the
config if it needs to be included in a future request payload.
