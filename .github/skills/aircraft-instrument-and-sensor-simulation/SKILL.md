---
name: aircraft-instruments
description: >
  Simulate, render, and model aircraft flight instruments and sensors in
  software — including real-time display widgets, physics-accurate instrument
  behavior, sensor error modeling, and failure simulation. Use this skill
  whenever the user asks to build, design, or model any aircraft instrument
  or avionics display, including: attitude indicator, airspeed indicator,
  altimeter, VSI, turn coordinator, heading indicator, HSI, CDI, NAV/COM
  radios, transponder, EGT/CHT gauges, fuel quantity, oil pressure, ammeter,
  vacuum gauge, glass cockpit displays (PFD/MFD/EFIS), autopilot annunciators,
  engine instruments, or any cockpit sensor. Also trigger for: "simulate
  instrument lag", "model pitot-static errors", "build a six-pack widget",
  "render a gyro horizon", "instrument failure injection", "simulate icing
  on pitot tube", "gyroscopic precession model", "VOR needle simulation",
  "ILS CDI/glideslope", "build an EFIS", "model gyro tumble", "ADC
  (air data computer) simulation", or any request to render instruments
  as canvas/SVG/WebGL graphics in a browser or sim environment.
---

# Aircraft Instruments & Sensor Simulation Skill

Build accurate, visually faithful instrument simulations — from individual
canvas-rendered gauges to full glass cockpit EFIS displays. Covers physics
modeling, error behavior, failure modes, and rendering technique.

---

## Architecture Overview

Every instrument simulation has three layers:

```
┌─────────────────────────────────────────┐
│  SENSOR / PHYSICS MODEL                 │  ← computes "true" value + errors
│  (atmosphere, gyro dynamics, radio nav) │
├─────────────────────────────────────────┤
│  INSTRUMENT MODEL                       │  ← applies lag, precession, limits
│  (lag filter, error accumulation,       │
│   failure state machine)                │
├─────────────────────────────────────────┤
│  DISPLAY / RENDERER                     │  ← canvas/SVG drawing, animations
│  (needle positions, tape scrolling,     │
│   annunciators, color bands)            │
└─────────────────────────────────────────┘
```

Never skip the instrument model layer — raw physics values fed directly to
the display produce unrealistically responsive instruments.

---

## §1 — Pitot-Static Instruments

### Shared Data Source
```javascript
// Air Data Computer (ADC) — feeds all pitot-static instruments
function computeADC(state, environment) {
  const { altitude, verticalSpeed, trueAirspeed, pitch, roll } = state;
  const { rho, rho0 } = atmosphere(altitude);

  // Indicated Airspeed (IAS) — based on sea-level density
  const dynamicPressure = 0.5 * rho * trueAirspeed ** 2;
  const ias = Math.sqrt(2 * dynamicPressure / rho0);

  // Calibrated Airspeed (CAS) — IAS corrected for position/instrument error
  // For sim purposes, treat as equal to IAS
  const cas = ias;

  return { ias, cas, altitude, verticalSpeed };
}
```

### Airspeed Indicator (ASI)

**Physics**: reads dynamic pressure (pitot minus static). Display in knots or MPH.

```javascript
class AirspeedIndicator {
  constructor(config) {
    this.config = {
      Vso: config.Vso || 44,   // stall, flaps full
      Vs1: config.Vs1 || 50,   // stall, clean
      Vfe: config.Vfe || 85,   // max flap extension
      Vno: config.Vno || 129,  // normal operating
      Vne: config.Vne || 163,  // never exceed
      maxDisplay: config.maxDisplay || 200,
      ...config
    };
    this.indicated = 0;
    this.lag = 0.15; // seconds — ASI responds quickly
  }

  update(ias, dt) {
    // First-order lag filter
    const tau = this.lag;
    this.indicated += (ias - this.indicated) * (1 - Math.exp(-dt / tau));
  }
}
```

**Rendering** — color arcs are the signature feature:
```javascript
function renderASI(ctx, instrument, cx, cy, radius) {
  const { Vso, Vs1, Vfe, Vno, Vne, maxDisplay } = instrument.config;
  const toAngle = v => (v / maxDisplay) * 270 - 135; // 270° sweep

  // White arc: Vso to Vfe (flap operating range)
  drawArc(ctx, cx, cy, radius * 0.88, toAngle(Vso), toAngle(Vfe), '#ffffff', 6);
  // Green arc: Vs1 to Vno (normal operating)
  drawArc(ctx, cx, cy, radius * 0.88, toAngle(Vs1), toAngle(Vno), '#00cc44', 6);
  // Yellow arc: Vno to Vne (caution)
  drawArc(ctx, cx, cy, radius * 0.88, toAngle(Vno), toAngle(Vne), '#ffcc00', 6);
  // Red line: Vne
  drawRadialMark(ctx, cx, cy, radius * 0.82, radius * 0.94, toAngle(Vne), '#ff2200', 3);

  // Needle
  const needleAngle = toAngle(instrument.indicated);
  drawNeedle(ctx, cx, cy, radius * 0.75, needleAngle, '#ffffff');
}
```

### Altimeter

**Physics**: reads static pressure, converts to altitude via ISA model.

```javascript
class Altimeter {
  constructor() {
    this.indicated = 0;
    this.kollsman = 29.92; // inches Hg (standard)
    this.lag = 1.5;        // altimeter lags more than ASI
    this.error = 0;        // barometric setting error: (kollsman - actual) * 1000 ft/inHg
  }

  update(pressureAltitude, dt) {
    // Apply barometric correction
    const actualAltitude = pressureAltitude + (this.kollsman - 29.92) * 1000;
    this.indicated += (actualAltitude - this.indicated) * (1 - Math.exp(-dt / this.lag));
  }

  setKollsman(setting) {
    this.kollsman = Math.max(28.0, Math.min(31.0, setting));
  }
}
```

**Rendering** — three concentric needles:
```javascript
function renderAltimeter(ctx, instrument, cx, cy, radius) {
  const alt = instrument.indicated;
  // 10,000ft hand: full rotation = 100,000 ft
  // 1,000ft hand: full rotation = 10,000 ft
  // 100ft hand: full rotation = 1,000 ft
  const hand100   = ((alt % 1000)  / 1000)  * 360 - 90;
  const hand1000  = ((alt % 10000) / 10000) * 360 - 90;
  const hand10000 = ((alt % 100000)/ 100000)* 360 - 90;

  drawNeedle(ctx, cx, cy, radius * 0.65, hand100,   '#ffffff', 2.5); // short wide
  drawNeedle(ctx, cx, cy, radius * 0.75, hand1000,  '#ffffff', 2);
  drawNeedle(ctx, cx, cy, radius * 0.55, hand10000, '#ffffff', 1.5); // stubby
}
```

### Vertical Speed Indicator (VSI)

**Physics**: measures rate of static pressure change. Significant lag (~6–9 seconds).

```javascript
class VSI {
  constructor() {
    this.indicated = 0;
    this.lag = 7.0; // seconds — VSI is the laggiest pitot-static instrument
    this.range = 2000; // fpm, each side
  }

  update(trueVS, dt) {
    this.indicated += (trueVS - this.indicated) * (1 - Math.exp(-dt / this.lag));
    this.indicated = Math.max(-this.range, Math.min(this.range, this.indicated));
  }
}
```

---

## §2 — Gyroscopic Instruments

### Attitude Indicator (AI / Artificial Horizon)

**Physics**: vacuum-driven gyro. Key behaviors to model:
- **Precession**: gyro axis drifts ~3°/min, more during maneuvers
- **Erection**: vacuum system slowly re-erects gyro (~3 min to stabilize)
- **Tumble limits**: older instruments tumble beyond ±60° pitch / ±100° bank
- **Power loss**: gyro spools down over ~15 min, instrument gradually topples

```javascript
class AttitudeIndicator {
  constructor() {
    this.displayPitch = 0;
    this.displayRoll  = 0;
    this.gyroPitch    = 0; // gyro-stabilized values
    this.gyroRoll     = 0;
    this.precessionRate = (3 * Math.PI / 180) / 60; // 3°/min in radians/s
    this.erectionRate   = (3 * Math.PI / 180) / 60; // erection toward gravity
    this.failed = false;
    this.vacuum = 1.0; // 0–1, normal = 1
  }

  update(truePitch, trueRoll, dt) {
    if (this.failed) return;

    // Precession: random walk proportional to angular rate
    const precessNoise = this.precessionRate * dt;
    this.gyroPitch += (Math.random() - 0.5) * precessNoise;
    this.gyroRoll  += (Math.random() - 0.5) * precessNoise;

    // Erection: vacuum system pulls gyro toward true attitude
    const erectionK = this.erectionRate * this.vacuum * dt;
    this.gyroPitch += (truePitch - this.gyroPitch) * erectionK;
    this.gyroRoll  += (trueRoll  - this.gyroRoll)  * erectionK;

    // Lag: display follows gyro with slight smoothing
    this.displayPitch += (this.gyroPitch - this.displayPitch) * (1 - Math.exp(-dt / 0.1));
    this.displayRoll  += (this.gyroRoll  - this.displayRoll)  * (1 - Math.exp(-dt / 0.1));
  }

  fail()   { this.failed = true; }
  setVacuum(v) { this.vacuum = Math.max(0, Math.min(1, v)); }
}
```

**Rendering** — rotating sky/ground horizon:
```javascript
function renderAI(ctx, instrument, cx, cy, radius) {
  ctx.save();
  ctx.beginPath();
  ctx.arc(cx, cy, radius, 0, Math.PI * 2);
  ctx.clip();

  // Rotate canvas by roll angle
  ctx.translate(cx, cy);
  ctx.rotate(-instrument.displayRoll);

  // Pitch offset: 1° pitch = ~radius/30 pixels (tunable)
  const pitchPx = instrument.displayPitch * (radius / (30 * Math.PI / 180));

  // Sky (blue)
  ctx.fillStyle = '#1a6fb5';
  ctx.fillRect(-radius, -radius * 2 + pitchPx, radius * 2, radius * 2);

  // Ground (brown)
  ctx.fillStyle = '#8B5e3c';
  ctx.fillRect(-radius, pitchPx, radius * 2, radius * 2);

  // Horizon line
  ctx.strokeStyle = '#ffffff';
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(-radius, pitchPx);
  ctx.lineTo(radius, pitchPx);
  ctx.stroke();

  ctx.restore();

  // Fixed aircraft symbol (drawn after restore, not rotated)
  drawAircraftSymbol(ctx, cx, cy);
}
```

### Heading Indicator (HI / Directional Gyro)

**Physics**: horizontal gyro. Precesses up to 3°/15 min — must be reset to compass.

```javascript
class HeadingIndicator {
  constructor() {
    this.displayHeading = 0;
    this.gyroHeading    = 0;
    this.precessionRate = (3 * Math.PI / 180) / 900; // 3°/15 min
    this.vacuum = 1.0;
  }

  update(trueHeading, dt) {
    // Gyroscopic rigidity: gyro maintains its orientation
    // Precession accumulates over time
    this.gyroHeading += this.precessionRate * dt * (this.vacuum > 0.5 ? 1 : 3);

    // Display follows gyro (no erection — DG doesn't self-correct to heading)
    this.displayHeading = this.gyroHeading;
  }

  syncToCompass(magneticHeading) {
    // Pilot action: press and hold knob to cage/reset
    this.gyroHeading = magneticHeading;
  }
}
```

### Turn Coordinator

**Physics**: electric gyro, senses roll rate and yaw rate. Separate electrical bus.

```javascript
class TurnCoordinator {
  constructor() {
    this.bankAngle  = 0; // display: -30 to +30 degrees
    this.ballOffset = 0; // slip/skid: -20 to +20 mm
    this.lag = 0.3;
    this.powered = true;
  }

  update(rollRate, yawRate, lateralAccel, dt) {
    if (!this.powered) { this.bankAngle = 0; return; }

    const sensitivityRoll = 20 / (3 * Math.PI / 180); // 20° display for 3°/s rate
    const target = Math.max(-30, Math.min(30,
      rollRate * sensitivityRoll + yawRate * 5
    ));
    this.bankAngle += (target - this.bankAngle) * (1 - Math.exp(-dt / this.lag));

    // Ball: responds to lateral (slip/skid) acceleration
    // Ball moves opposite to uncoordinated force
    const ballTarget = -lateralAccel * 40;
    this.ballOffset += (ballTarget - this.ballOffset) * (1 - Math.exp(-dt / 0.8));
    this.ballOffset = Math.max(-20, Math.min(20, this.ballOffset));
  }
}
```

---

## §3 — Navigation Instruments

### VOR / CDI (Course Deviation Indicator)

```javascript
class VORReceiver {
  constructor() {
    this.obsSetting   = 0;    // degrees (pilot-set)
    this.radialFrom   = null; // current radial from station
    this.cdiDeflection = 0;   // -1 to +1 (full scale = 10°)
    this.toFrom       = null; // 'TO' | 'FROM' | 'OFF'
    this.signalValid  = false;
    this.flagShown    = true;
  }

  update(aircraftPos, stations) {
    const station = this.findStation(aircraftPos, stations);
    if (!station || station.distanceNm > 130) {
      this.flagShown = true; this.signalValid = false; return;
    }

    this.signalValid = true;
    this.flagShown = false;
    this.radialFrom = bearing(station.pos, aircraftPos);

    const courseError = normalizeAngle(this.obsSetting - this.radialFrom);
    const inboundError = normalizeAngle(courseError);

    // TO/FROM
    this.toFrom = Math.abs(inboundError) < 90 ? 'TO' : 'FROM';

    // CDI: 10° = full scale (1.0)
    const deviation = normalizeAngle(this.radialFrom - this.obsSetting);
    this.cdiDeflection = Math.max(-1, Math.min(1, deviation / 10));
  }
}
```

### ILS (Localizer + Glideslope)

```javascript
class ILSReceiver {
  constructor() {
    this.locDeflection = 0;  // -1 to +1 (full scale = 2.5°)
    this.gsDeflection  = 0;  // -1 to +1 (full scale = 0.7°)
    this.locValid = false;
    this.gsValid  = false;
  }

  update(aircraftPos, runwayThreshold, runwayCourse, gsAngleDeg) {
    const bearingToThreshold = bearing(aircraftPos, runwayThreshold);
    const distNm = distance(aircraftPos, runwayThreshold);

    // Localizer: angular deviation from extended centerline
    const locErrorDeg = normalizeAngle(bearingToThreshold - runwayCourse);
    this.locDeflection = Math.max(-1, Math.min(1, locErrorDeg / 2.5));
    this.locValid = distNm < 18;

    // Glideslope: compare actual elevation angle to nominal
    const altAgl = aircraftPos.altMsl - runwayThreshold.elevFt;
    const actualGsAngle = Math.atan2(altAgl, distNm * 6076) * 180 / Math.PI;
    const gsErrorDeg = actualGsAngle - gsAngleDeg;
    this.gsDeflection = Math.max(-1, Math.min(1, gsErrorDeg / 0.7));
    this.gsValid = distNm < 10 && altAgl < 5000;
  }
}
```

---

## §4 — Engine Instruments

```javascript
class EngineInstruments {
  constructor(config) {
    this.rpm     = 0;
    this.mp      = 29.92; // manifold pressure, in Hg
    this.egt     = 0;     // exhaust gas temp, °F
    this.cht     = 0;     // cylinder head temp, °F
    this.oilTemp = 0;     // °F
    this.oilPres = 0;     // PSI
    this.fuelFlow = 0;    // gph
    this.fuelQty  = { left: config.fuelLeft || 25, right: config.fuelRight || 25 }; // gallons

    // Lag times (seconds) — engine temps are very slow
    this.lags = { rpm: 0.5, mp: 0.3, egt: 15, cht: 45, oilTemp: 60, oilPres: 2 };
    this.raw = { ...this };
  }

  update(trueRPM, throttle, mixture, dt) {
    this.raw.rpm = trueRPM;
    this.raw.egt = 1200 + (throttle * 200) + ((mixture - 0.5) * -400); // simplified
    this.raw.cht = 300  + (throttle * 150);
    this.raw.oilTemp = 180 + (throttle * 40);
    this.raw.oilPres = trueRPM > 500 ? 60 - (trueRPM / 100) : 0;
    this.raw.fuelFlow = throttle * mixture * 10; // gph

    // Apply lag filters
    for (const key of ['rpm', 'egt', 'cht', 'oilTemp', 'oilPres']) {
      const tau = this.lags[key];
      this[key] += (this.raw[key] - this[key]) * (1 - Math.exp(-dt / tau));
    }

    // Fuel consumption
    this.fuelQty.left  -= (this.fuelFlow / 2) * (dt / 3600);
    this.fuelQty.right -= (this.fuelFlow / 2) * (dt / 3600);
  }
}
```

---

## §5 — Failure Injection System

Every instrument should connect to a central failure manager:

```javascript
const FailureManager = {
  failures: new Set(),

  inject(systemId) {
    this.failures.add(systemId);
    this.applyFailure(systemId);
  },

  clear(systemId) {
    this.failures.delete(systemId);
  },

  applyFailure(id) {
    switch (id) {
      case 'pitot_ice':
        // ASI reads as altimeter — pitot blocked, static open
        instruments.asi.pitotBlocked = true;
        break;
      case 'static_blocked':
        instruments.asi.staticBlocked = true;
        instruments.altimeter.staticBlocked = true;
        instruments.vsi.staticBlocked = true;
        break;
      case 'vacuum_failure':
        instruments.ai.setVacuum(0);
        instruments.hi.vacuum = 0;
        // TC unaffected (electric)
        break;
      case 'ai_tumble':
        instruments.ai.fail();
        break;
      case 'electrical_failure':
        instruments.tc.powered = false;
        instruments.ilsReceiver.powered = false;
        break;
      case 'vor_flag':
        instruments.vor.flagShown = true;
        instruments.vor.signalValid = false;
        break;
    }
  },
};
```

---

## §6 — Glass Cockpit / EFIS Display

### PFD Layout (HTML + CSS approach)

For a G1000-style PFD, use HTML elements with CSS transforms — easier
than canvas for tape-style displays:

```html
<div class="pfd">
  <!-- Attitude display (canvas) -->
  <canvas id="adi-canvas" class="pfd-adi"></canvas>

  <!-- Airspeed tape (scrolling div) -->
  <div class="airspeed-tape-window">
    <div id="airspeed-tape" class="airspeed-tape">
      <!-- Generated tick marks for 0–300 kts -->
    </div>
    <div class="airspeed-bug"></div> <!-- current speed bug -->
  </div>

  <!-- Altitude tape (similar structure) -->
  <div class="altitude-tape-window">
    <div id="altitude-tape" class="altitude-tape"></div>
  </div>

  <!-- VSI (small vertical strip, right of altitude tape) -->
  <canvas id="vsi-canvas" class="pfd-vsi"></canvas>

  <!-- HSI (bottom center) -->
  <canvas id="hsi-canvas" class="pfd-hsi"></canvas>

  <!-- Annunciator strip (top) -->
  <div id="annunciators" class="pfd-annunciators"></div>
</div>
```

**Tape scrolling** via CSS transform:
```javascript
function updateAirspeedTape(ias) {
  // Each knot = 6px on tape
  const PX_PER_KT = 6;
  const offset = ias * PX_PER_KT;
  document.getElementById('airspeed-tape').style.transform
    = `translateY(${offset}px)`;
}
```

### Annunciator Colors
```javascript
const ANNUNCIATOR_LEVELS = {
  warning:  { color: '#ff2200', flash: true  }, // red — immediate action
  caution:  { color: '#ffcc00', flash: false }, // amber — monitor/correct
  advisory: { color: '#00aaff', flash: false }, // blue — information
  normal:   { color: '#00cc44', flash: false }, // green — normal operation
};
```

---

## §7 — Rendering Utilities

Common drawing helpers used across all instruments:

```javascript
// Draw arc (for color bands)
function drawArc(ctx, cx, cy, r, startDeg, endDeg, color, width) {
  ctx.beginPath();
  ctx.arc(cx, cy, r, startDeg * Math.PI/180, endDeg * Math.PI/180);
  ctx.strokeStyle = color;
  ctx.lineWidth = width;
  ctx.stroke();
}

// Draw needle from center
function drawNeedle(ctx, cx, cy, length, angleDeg, color, width = 2) {
  const rad = angleDeg * Math.PI / 180;
  ctx.beginPath();
  ctx.moveTo(cx, cy);
  ctx.lineTo(cx + length * Math.cos(rad), cy + length * Math.sin(rad));
  ctx.strokeStyle = color;
  ctx.lineWidth = width;
  ctx.stroke();
}

// Normalize angle to [-180, 180]
function normalizeAngle(deg) {
  let a = deg % 360;
  if (a > 180) a -= 360;
  if (a < -180) a += 360;
  return a;
}

// First-order lag (use everywhere)
function lagFilter(current, target, tau, dt) {
  return current + (target - current) * (1 - Math.exp(-dt / tau));
}
```

---

## Output Notes

When producing instrument simulations:
- Default output: single HTML file with all instruments, demo flight state,
  and controls to manually input values
- For integration into an existing sim: export as a JS module with a clean
  `update(state, dt)` / `render(ctx)` interface per instrument
- Always include a failure injection panel in demo outputs — it's the most
  educational feature
- Label every instrument with its name, and show the raw vs. indicated value
  during development (toggle-able in production)