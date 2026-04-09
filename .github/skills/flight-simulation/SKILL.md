---
name: flight-simulator
description: >
  Design and build flight simulators — from browser-based canvas/WebGL
  implementations to architecture documents for full sim systems. Use this
  skill whenever the user asks to build, design, or reason about a flight
  simulator, aircraft simulation, flight dynamics model, avionics system,
  ATC simulator, autopilot system, or any software that models aircraft
  behavior in 3D space. Also trigger for: "make a flying game", "simulate
  aircraft physics", "build a cockpit UI", "design a flight model", "how
  does X work in a flight sim", "build an autopilot", "ATC radar display",
  or any request involving aerodynamics, flight instruments, navigation
  systems, or aircraft performance modeling. Covers both educational
  browser-based sims and serious simulation architecture.
---

# Flight Simulator Design Skill

Covers the full spectrum: quick browser-based flying demos up to serious
flight dynamics architecture. Read the relevant section for your task.

---

## Scope Selection

Identify which tier the user needs before building anything:

| Tier | Description | Output |
|---|---|---|
| **Toy** | Simple flying game, arcade physics | Single HTML file, canvas 2D |
| **Casual** | Realistic-ish 3D sim, basic instruments | HTML + Three.js, WebGL |
| **Serious** | Real flight dynamics model, full avionics | Architecture doc + modular JS/TS |
| **Training** | FAA-creditable fidelity concepts | Architecture + reference docs |

When the user says "flight simulator" without qualification, default to **Casual**.
When they say "flight game" or "make a plane fly", default to **Toy**.
When they say "realistic", "accurate physics", or "flight model", go **Serious**.

---

## Physics: Flight Dynamics Model (FDM)

The heart of any serious sim. All forces computed per-frame in body frame,
then integrated to update world-frame position and attitude.

### The Six Degrees of Freedom (6DoF)

Aircraft state vector:
```
Position:    [x, y, z]          world frame (meters, NED or ENU)
Velocity:    [u, v, w]          body frame (forward, right, down)
Attitude:    [φ, θ, ψ]          Euler angles (roll, pitch, yaw) in radians
Ang. rates:  [p, q, r]          body frame (roll rate, pitch rate, yaw rate)
```

### Four Forces

```
Lift     L = 0.5 * ρ * V² * S * CL(α, δe, δf)
Drag     D = 0.5 * ρ * V² * S * CD(α, δf)
Thrust   T = f(throttle, V, altitude, engine_model)
Weight   W = mass * g  (acts downward in world frame)
```

Where:
- `ρ` = air density (varies with altitude, use ISA model)
- `V` = airspeed magnitude
- `S` = wing reference area
- `CL`, `CD` = lift/drag coefficients (lookup tables or polynomial fits)
- `α` = angle of attack
- `δe`, `δf` = elevator and flap deflection

### Aerodynamic Coefficients

For a casual sim, polynomial approximations are fine:
```javascript
function CL(alpha, deltaFlap) {
  const CL0 = 0.2;
  const CLalpha = 5.5;   // per radian, typical for light GA
  const CLflap  = 0.8 * deltaFlap;
  const alphaStall = 0.28; // ~16 degrees
  
  if (Math.abs(alpha) > alphaStall) {
    // Post-stall: rapid CL loss
    return Math.sign(alpha) * 0.8 * Math.cos(alpha);
  }
  return CL0 + CLalpha * alpha + CLflap;
}

function CD(alpha, deltaFlap) {
  const CD0   = 0.027;   // parasite drag
  const k     = 0.045;   // induced drag factor  
  const cl    = CL(alpha, deltaFlap);
  return CD0 + k * cl * cl + 0.03 * Math.abs(deltaFlap);
}
```

For a serious sim, use actual aerodynamic tables from DATCOM, AVL, or
aircraft-specific data (NACA reports for classic aircraft are public domain).

### ISA Atmosphere Model

```javascript
function atmosphere(altitudeM) {
  const T0 = 288.15;    // sea level temp (K)
  const L  = 0.0065;    // lapse rate (K/m)
  const P0 = 101325;    // sea level pressure (Pa)
  const R  = 287.058;   // specific gas constant
  const g  = 9.80665;
  
  const T = T0 - L * altitudeM;
  const P = P0 * Math.pow(T / T0, g / (R * L));
  const rho = P / (R * T);
  return { T, P, rho };
}
```

### Equations of Motion Integration

Use RK4 for accuracy, or semi-implicit Euler for simplicity:

```javascript
// Semi-implicit Euler (good enough for dt < 20ms)
function integrate(state, forces, moments, dt) {
  const { mass, inertia } = aircraft;
  
  // Translational
  const accel = forces.map(f => f / mass);
  state.velocity = state.velocity.map((v, i) => v + accel[i] * dt);
  state.position = state.position.map((p, i) => p + state.velocity[i] * dt);
  
  // Rotational
  const angAccel = moments.map((m, i) => m / inertia[i]);
  state.angularRate = state.angularRate.map((r, i) => r + angAccel[i] * dt);
  state.attitude = state.attitude.map((a, i) => a + state.angularRate[i] * dt);
  
  // Clamp attitude
  state.attitude[0] = clamp(state.attitude[0], -Math.PI, Math.PI); // roll
  state.attitude[1] = clamp(state.attitude[1], -Math.PI/2, Math.PI/2); // pitch
}
```

---

## Aircraft Models

### Light GA (Cessna 172-class)
```javascript
const C172 = {
  mass:       1111,      // kg MTOW
  wingArea:   16.2,      // m²
  span:       11.0,      // m
  AR:         7.4,       // aspect ratio
  engineHP:   180,
  fuelCapL:   212,
  Vso:        50,        // knots stall, flaps full
  Vs1:        54,        // knots stall, clean
  Vy:         75,        // knots best rate of climb
  Vx:         70,        // knots best angle of climb
  Vne:        163,       // knots never exceed
  cruiseKtas: 122,       // knots TAS at 75% power
};
```

### Control Surfaces
```javascript
const controls = {
  aileron:  0,    // -1 to 1 (left to right bank)
  elevator: 0,    // -1 to 1 (nose down to nose up)
  rudder:   0,    // -1 to 1 (left to right yaw)
  throttle: 0,    // 0 to 1
  flaps:    0,    // 0, 0.1, 0.2, 0.3, 0.4 (0, 10, 20, 30, 40 degrees)
  brakes:   false,
  gear:     'down', // 'up' | 'down' | 'transit'
};
```

---

## Instruments & Avionics

### The Six-Pack (Classic Steam Gauges)

Render each as an HTML canvas element or SVG:

| Instrument | Data source | Key rendering detail |
|---|---|---|
| Airspeed Indicator | `V` (IAS) | Colored arcs: white (Vso-Vfe), green (Vs1-Vno), yellow (Vno-Vne), red line (Vne) |
| Attitude Indicator | `φ, θ` | Gyro horizon: sky/ground split, bank angle marks, pitch ladder |
| Altimeter | `altitude` | Three hands (100ft, 1000ft, 10000ft), Kollsman window |
| VSI | `dAlt/dt` | Vertical speed in ft/min, ±2000 fpm typical range |
| Turn Coordinator | `r`, bank | Standard rate = 3°/sec, inclinometer ball |
| Heading Indicator | `ψ` | Gyroscopic compass, set to magnetic heading |

### Navigation Instruments

- **VOR**: bearing to/from a VOR station, CDI deflection
- **ILS**: localizer + glideslope deviation bars
- **GPS moving map**: position on a 2D map tile, track line, waypoints
- **ADF**: relative bearing to NDB station

### Modern Glass Cockpit (G1000-style)

Two large displays: PFD (Primary Flight Display) + MFD (Multi-Function Display)

PFD contains: attitude tape, airspeed tape, altitude tape, VSI, HSI, engine
strip, annunciator bar.

Render as HTML elements with CSS transforms for tape scrolling — much easier
than canvas for this use case.

---

## Browser Implementation (Casual Tier)

Use Three.js for 3D rendering:

```html
<script src="https://cdnjs.cloudflare.com/ajax/libs/three.js/r128/three.min.js"></script>
```

### Scene Setup
```javascript
const scene    = new THREE.Scene();
const camera   = new THREE.PerspectiveCamera(60, w/h, 0.1, 100000);
const renderer = new THREE.WebGLRenderer({ antialias: true });

// Sky
scene.background = new THREE.Color(0x87CEEB);
scene.fog = new THREE.Fog(0xC9E8F5, 5000, 50000);

// Terrain: flat green plane for MVP, heightmap for v2
const terrain = new THREE.Mesh(
  new THREE.PlaneGeometry(100000, 100000, 256, 256),
  new THREE.MeshLambertMaterial({ color: 0x4a7c59 })
);
terrain.rotation.x = -Math.PI / 2;
scene.add(terrain);

// Lighting
scene.add(new THREE.AmbientLight(0xffffff, 0.6));
const sun = new THREE.DirectionalLight(0xffffff, 0.8);
sun.position.set(10000, 8000, 5000);
scene.add(sun);
```

### Camera Modes
Implement multiple camera views, cycle with `C`:
- **Chase cam**: behind and above aircraft
- **Cockpit cam**: from pilot eye position, no external model visible
- **Tower cam**: fixed point, tracks aircraft
- **Free cam**: orbit controls, detached

### Aircraft Model
For MVP, build from Three.js primitives:
```javascript
function createAircraftModel() {
  const group = new THREE.Group();
  
  // Fuselage
  group.add(new THREE.Mesh(
    new THREE.CylinderGeometry(0.5, 0.3, 8, 8),
    new THREE.MeshLambertMaterial({ color: 0xdddddd })
  ));
  
  // Wings (BoxGeometry, flat)
  const wingGeo = new THREE.BoxGeometry(12, 0.15, 1.8);
  group.add(new THREE.Mesh(wingGeo, mat));
  
  // Tail surfaces...
  return group;
}
```

---

## Controls & Input

### Keyboard Mapping
```javascript
const KEY_MAP = {
  'ArrowUp':   () => controls.elevator -= 0.05,
  'ArrowDown': () => controls.elevator += 0.05,
  'ArrowLeft': () => controls.aileron  -= 0.05,
  'ArrowRight':() => controls.aileron  += 0.05,
  'z':         () => controls.rudder   -= 0.05,
  'x':         () => controls.rudder   += 0.05,
  'PageUp':    () => controls.throttle  = Math.min(1, controls.throttle + 0.05),
  'PageDown':  () => controls.throttle  = Math.max(0, controls.throttle - 0.05),
  'f':         () => cycleFlaps(),
  'g':         () => toggleGear(),
  'b':         () => controls.brakes = true,
};
```

Apply control surface decay each frame (auto-center when key released):
```javascript
controls.elevator *= 0.92;
controls.aileron  *= 0.92;
controls.rudder   *= 0.92;
```

### Joystick / Gamepad
```javascript
window.addEventListener('gamepadconnected', (e) => {
  gamepad = navigator.getGamepads()[e.gamepad.index];
});

function readGamepad() {
  const gp = navigator.getGamepads()[0];
  if (!gp) return;
  controls.aileron  = gp.axes[0];  // left stick X
  controls.elevator = gp.axes[1];  // left stick Y
  controls.rudder   = gp.axes[2];  // right stick X
  controls.throttle = (1 - gp.axes[3]) / 2; // right stick Y, 0-1
}
```

---

## Autopilot

Basic autopilot modes:

```javascript
const AP = {
  active: false,
  modes: {
    HDG: false,   // Heading hold
    ALT: false,   // Altitude hold
    VS:  false,   // Vertical speed hold
    NAV: false,   // VOR/GPS tracking
    APR: false,   // ILS approach
  },
  targets: {
    heading: 0,   // degrees
    altitude: 0,  // feet
    vs: 0,        // ft/min
  },
};

function updateAutopilot(state, dt) {
  if (!AP.active) return;

  if (AP.modes.HDG) {
    const hdgError = normalizeAngle(AP.targets.heading - state.heading);
    controls.aileron = clamp(hdgError * 0.02, -0.5, 0.5);
  }

  if (AP.modes.ALT) {
    const altError = AP.targets.altitude - state.altitude;
    const vsTarget = clamp(altError * 2, -1500, 1500); // ft/min
    const vsError  = vsTarget - state.vs;
    controls.elevator = clamp(-vsError * 0.001, -0.3, 0.3);
  }
}
```

---

## Terrain & World

### Heightmap Terrain
```javascript
// Load a 512×512 greyscale PNG as heightmap
// Each pixel brightness maps to elevation
function buildTerrain(heightmapData, worldSize, maxHeight) {
  const geo = new THREE.PlaneGeometry(worldSize, worldSize, 511, 511);
  const vertices = geo.attributes.position.array;
  
  for (let i = 0; i < 512 * 512; i++) {
    const brightness = heightmapData[i * 4] / 255;
    vertices[i * 3 + 2] = brightness * maxHeight;
  }
  geo.computeVertexNormals();
  return geo;
}
```

### Airports
Represent as data objects:
```javascript
const AIRPORTS = [
  {
    icao: 'KSFO',
    name: 'San Francisco International',
    lat: 37.619,
    lon: -122.374,
    elevation: 13,   // feet MSL
    runways: [
      { id: '28L', heading: 280, length: 11870, width: 200, ils: { freq: 108.9, gs: 3.0 } },
      { id: '28R', heading: 280, length: 10602, width: 200, ils: { freq: 109.55, gs: 3.0 } },
    ]
  }
];
```

---

## Output Format

For **Toy** tier: single HTML file, same format as browser-sim-game skill.

For **Casual** tier: single HTML file with Three.js CDN import. Structure:
```
1. HTML/CSS (canvas, instrument panel overlay)
2. Aircraft data constants
3. FDM (atmosphere, forces, integration)
4. Avionics (instrument rendering functions)
5. Three.js scene setup
6. Game loop
7. Input handling
8. Autopilot
9. Init
```

For **Serious** tier: produce an architecture document with:
- Module breakdown and interfaces
- FDM specification with equations
- Avionics system design
- Data flow diagram
- Implementation roadmap

Always include a **Quick Start** section at the top of any produced file:
what keys do what, how to take off, what to watch on instruments.