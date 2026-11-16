# Drone-Blocker Desktop UI

Desktop UI for the drone detection and jamming system.

## Jammer HTTP flow

This section documents how the app talks to the jammer device, so the URL and
payload can be changed in one place.

### Transport

Jammer requests go through Go helper functions:

- [`JammerGet(targetURL string)`](app.go:412) — **GET** request, 3 second
  timeout, no request body, no custom headers.
- [`JammerPost(targetURL string, body string)`](app.go:441) — **POST** request
  with a JSON body (`Content-Type: application/json`), 3 second timeout.

Both return the raw response body as a string and return an error on network
failure or any HTTP status `>= 400`.

Frontend bindings: [`JammerGet(arg1)`](frontend/wailsjs/go/main/App.js:25) and
[`JammerPost(arg1, arg2)`](frontend/wailsjs/go/main/App.js:29).

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

#### 1. `GET {base}/state` — status polling

Called every 5 seconds when the jammer toggle is enabled:
[`Jammer.svelte`](frontend/src/lib/components/Jammer.svelte:67).

Expected response:

```json
{
  "success": true,
  "desktop_mode": 0,
  "jamm_900": 1,
  "jamm_1500": 1,
  "jamm_2400": 1,
  "jamm_5200": 0,
  "jamm_5800": 1
}
```

- `success` must be `true`, otherwise the app marks the jammer as disconnected.
- `desktop_mode` must be `0` (OFF), `1` (ON), or `2` (AUTO). The `jamm_*`
  fields are `0` (OFF) or `1` (ON) and drive the per-band status LEDs.

#### 2. `POST {base}/deskcmd` — OFF button

Sent from [`setJammerMode()`](frontend/src/lib/components/Jammer.svelte:25) when
the OFF button is pressed.

Request body:

```json
{
  "jamm_900": false,
  "jamm_1500": false,
  "jamm_2400": false,
  "jamm_5200": false,
  "jamm_5800": false,
  "desktop_mode": 0
}
```

The response is not strictly parsed; the app updates the local mode to OFF after
a successful response. The mock returns:

```json
{ "success": true, "desktop_mode": 0 }
```

#### 3. `POST {base}/deskcmd` — ON button

Sent from [`setJammerMode()`](frontend/src/lib/components/Jammer.svelte:71) when
the ON button is pressed. Same as OFF, but with `desktop_mode: 1`.

Request body:

```json
{
  "jamm_900": false,
  "jamm_1500": false,
  "jamm_2400": false,
  "jamm_5200": false,
  "jamm_5800": false,
  "desktop_mode": 1
}
```

The response is not strictly parsed; the app updates the local mode to ON after
a successful response.

#### 4. `POST {base}/deskcmd` — AUTO button

Sent from [`setJammerMode()`](frontend/src/lib/components/Jammer.svelte:99) when
the AUTO button is pressed. Same as ON, but with `desktop_mode: 2`.

Request body:

```json
{
  "jamm_900": false,
  "jamm_1500": false,
  "jamm_2400": false,
  "jamm_5200": false,
  "jamm_5800": false,
  "desktop_mode": 2
}
```

The response is not strictly parsed; the app updates the local mode to AUTO after
a successful response.

#### 5. `POST {base}/deskcmd` — automatic drone alert

Sent from [`#sendAlert()`](frontend/src/lib/stores/droneAlertService.svelte.ts:102)
when a new drone is confirmed and the jammer is in AUTO mode. It turns on all
jammer bands.

Request body:

```json
{
  "jamm_900": true,
  "jamm_1500": true,
  "jamm_2400": true,
  "jamm_5200": true,
  "jamm_5800": true,
  "desktop_mode": 2
}
```

The response is **not parsed** — it only needs a non-error HTTP status.

