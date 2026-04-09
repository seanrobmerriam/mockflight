# Current State Of MockFlight

## Summary

MockFlight started as a collection of standalone Go sensor prototypes focused on aviation instrumentation simulation. The repository now has a working main application: a single-process web dashboard that serves synchronized readings for the full current sensor set.

## What Was In The Repository Before This Pass

- A README that described the project as individually built sensor binaries.
- Nine sensor source files under `sensors/`.
- No `go.mod` file.
- No package-level tests.
- No main application that exposed all sensors in one place.

## Observed Issues In The Original Layout

- The repository was not organized as a normal Go module, which made repeatable builds and tests harder.
- The legacy sensor files were not in a reusable package structure.
- Some files were inconsistent at the package level, so the repo was not ready for `go test ./...` as-is.
- The README drifted away from the executable reality of the repository.
- There was no shared state layer to guarantee that all sensor outputs represented the same point in a flight profile.

## What Exists Now

### Working application entrypoint

- `main.go` starts a local HTTP server.
- The server hosts a live dashboard and a JSON snapshot API.

### Reusable simulator package

- `internal/sim` provides a shared flight model.
- It generates synchronized readings for:
  - altimeter
  - vertical speed
  - static air pressure
  - pitot pressure
  - airspeed
  - accelerometer
  - gyroscope
  - magnetometer
  - AHRS

### Dashboard package

- `internal/dashboard` serves the web UI and `/api/snapshot`.
- The dashboard is self-contained and uses no frontend build tooling.

### Tests

- Simulator tests confirm the snapshot shape and flight-state progression.
- Dashboard tests confirm the API and HTML shell.

### Documentation

- README now documents the real entrypoint and current architecture.
- This status document records the baseline.
- `docs/potential-sensor-additions.md` captures researched next candidates.

## Current Product Position

The project is now at an early but usable application stage:

- It is no longer just a loose set of experiments.
- It has a module definition, a verified build, and a single user-facing app.
- It still carries legacy prototype files that should eventually be either revived cleanly or retired.
- The sensor model is intentionally lightweight and useful for UI, integration, and early instrumentation work rather than high-fidelity avionics validation.

## Verified Commands

The following now succeed:

```bash
go test ./...
```

## Practical Next Steps

1. Decide whether the legacy `sensors/` files should be revived as `cmd/*` binaries or archived.
2. Expand the simulator with additional air-data, navigation, and engine sensors.
3. Add time-series history or streaming transport if the dashboard needs traces rather than point-in-time snapshots.
4. Split the simulator into configurable aircraft profiles if the project needs more than one airframe behavior.