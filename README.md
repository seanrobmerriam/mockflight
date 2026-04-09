# MockFlight

MockFlight is a local Go flight-instrument simulator. It runs a single in-process flight model, serves a browser dashboard, exposes JSON APIs for the current sensor snapshot, and lets you inject common pitot-static and IMU failure modes at runtime.

The shared simulator keeps these readings synchronized:

- altimeter
- vertical speed
- static air pressure
- pitot pressure
- airspeed
- accelerometer
- gyroscope
- magnetometer
- AHRS

## Quick Start

```bash
git clone https://github.com/seanrobmerriam/mockflight
cd mockflight
go run .
```

Open `http://localhost:8080`.

## Runtime Flags

The dashboard app supports these flags:

```bash
go run . -addr :9090 -noise 0.5
go run . -pitot-blocked -static-leak
```

- `-addr`: HTTP listen address. Default `:8080`.
- `-noise`: global sensor noise multiplier. Default `1.0`.
- `-pitot-blocked`: freeze pitot pressure at its current reading.
- `-pitot-drain-blocked`: bleed off dynamic pressure from the pitot system.
- `-static-blocked`: freeze static pressure at its current reading.
- `-static-leak`: bias static pressure toward a leaking reference pressure.
- `-mag-disturbance`: inject magnetic disturbance into the magnetometer.
- `-gyro-saturation`: clamp gyro rates at the sensor saturation threshold.

## Dashboard

The root page renders a live IFR trainer deck backed by the shared simulator state. The UI polls the snapshot API and exposes:

- runway-start controls for throttle, TO/GA, rotate, reset, and basic autopilot targets
- failure injection controls for pitot-static and heading-reference degradations
- scenario launchers and emergency checklist selection for phase-1 single-engine IFR training
- repeating aural warnings for `SINK RATE` and `PULL UP`, with a browser-side `Aural Alerts` mute toggle

Altitude, airspeed, pitch, roll, heading, pitot-static pressure, AHRS values, and takeoff warning state all move together and react to injected faults as one flight profile.

## HTTP API

### `GET /api/snapshot`

Returns the current synchronized sensor snapshot.

Example response:

```json
{
  "phase": "initial_climb",
  "timestamp": "2026-04-06T12:00:00Z",
  "active_failures": ["pitot_blocked"],
  "controls": {
    "mode": "full",
    "throttle": 1,
    "toga": true,
    "rotate_commanded": true,
    "airborne": true,
    "warning": "",
    "crashed": false,
    "v1": 62,
    "vr": 67,
    "v2": 74,
    "selected_heading": 270,
    "selected_altitude": 2142,
    "selected_vertical_speed": 700,
    "runway_distance": 1650,
    "runway_remaining": 6550,
    "autopilot": {
      "engaged": false,
      "heading_hold": false,
      "altitude_hold": false,
      "vertical_speed_mode": false
    }
  },
  "altimeter": {"name": "altimeter", "value": 5342, "unit": "ft"},
  "vertical_speed": {"name": "vertical_speed", "value": 0.08, "unit": "m/s"},
  "static_air": {"name": "static_air", "value": 24.387, "unit": "inHg"},
  "pitot": {"name": "pitot", "value": 24.569, "unit": "inHg"},
  "airspeed": {
    "name": "airspeed",
    "indicated": 128.6,
    "unit": "kts",
    "pitot_pressure": 24.569,
    "static_pressure": 24.387
  },
  "accelerometer": {"name": "accelerometer", "x": -0.12, "y": 0.21, "z": -9.72, "unit": "m/s^2"},
  "gyroscope": {"name": "gyroscope", "x": 0.42, "y": -0.13, "z": 2.91, "unit": "deg/s"},
  "magnetometer": {"name": "magnetometer", "x": 20.14, "y": -1.44, "z": 42.35, "unit": "uT"},
  "ahrs": {"name": "ahrs", "roll": 4.8, "pitch": 2.4, "yaw": 273.1, "heading": 273.1, "unit": "deg"}
}
```

### `GET /api/failures`

Returns the currently active failure configuration.

### `POST /api/failures`

Updates the active failure configuration at runtime.

Example request:

```bash
curl -X POST http://localhost:8080/api/failures \
  -H 'Content-Type: application/json' \
  -d '{"pitot_blocked":true,"gyro_saturation":true}'
```

Supported fields:

- `pitot_blocked`
- `pitot_drain_blocked`
- `static_port_blocked`
- `static_leak`
- `magnetometer_disturbed`
- `gyro_saturation`

### `GET /api/controls`

Returns the current simulator control and autopilot state.

### `POST /api/controls`

Updates control state at runtime.

Example request:

```bash
curl -X POST http://localhost:8080/api/controls \
  -H 'Content-Type: application/json' \
  -d '{"throttle":0.85,"rotate":true,"heading_hold":true,"selected_heading":315}'
```

Common fields:

- `mode`
- `throttle`
- `toga`
- `rotate`
- `autopilot_engaged`
- `heading_hold`
- `altitude_hold`
- `vertical_speed_mode`
- `selected_heading`
- `selected_altitude`
- `selected_vertical_speed`
- `reset`

### `POST /api/scenarios`

Applies a training scenario from the phase-1 IFR catalog.

Example request:

```bash
curl -X POST http://localhost:8080/api/scenarios \
  -H 'Content-Type: application/json' \
  -d '{"key":"lowEnergyClimb"}'
```

Use an empty key to return to manual mode:

```bash
curl -X POST http://localhost:8080/api/scenarios \
  -H 'Content-Type: application/json' \
  -d '{"key":""}'
```

### `GET /healthz`

Returns `200 OK` with body `ok`.

## Legacy Sensor CLIs

The original single-sensor programs are still present in `sensors/`, but they are excluded from normal builds behind the `legacycli` build tag. You can still run them directly when needed:

```bash
go run -tags legacycli ./sensors/altimeter.go -format simple
go run -tags legacycli ./sensors/airspeed.go -hz 5 -noise 0 -pitot-blocked
```

Shared legacy CLI flags include:

- `-format`: `json`, `csv`, or `simple`.
- `-hz`: update frequency in Hz.
- `-noise`: global sensor noise multiplier.
- `-seed`: deterministic random seed. `0` uses the current time.
- failure flags matching the dashboard app: `-pitot-blocked`, `-pitot-drain-blocked`, `-static-blocked`, `-static-leak`, `-mag-disturbance`, `-gyro-saturation`.

## Project Layout

```text
.
├── internal/
│   ├── dashboard/    # HTTP handler and embedded dashboard UI
│   ├── legacycli/    # Shared runner for tagged legacy sensor binaries
│   └── sim/          # Shared flight model and sensor snapshot generation
├── sensors/          # Legacy standalone sensor entrypoints behind the legacycli build tag
├── docs/
│   ├── current-state.md
│   └── potential-sensor-additions.md
├── go.mod
└── main.go           # Dashboard entrypoint
```

## Development

```bash
go test ./...
go build ./...
gofmt -w .
```

## Status

See:

- `docs/current-state.md`
- `docs/potential-sensor-additions.md`