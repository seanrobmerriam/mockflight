---
name: ifr-flight
description: >
  Expert reference and instructional resource for Instrument Flight Rules (IFR)
  operations, procedures, training, and systems. Use this skill whenever the
  user asks about IFR flying, instrument flying, instrument rating training,
  IFR procedures, approach plates, departure procedures, en route navigation,
  hold entries, airspace, ATC communication scripts, weather minimums, filing
  flight plans, partial panel flying, unusual attitudes, instrument currency,
  CFII instruction, or any topic related to flying in IMC (Instrument
  Meteorological Conditions). Also trigger for: "how do I fly an ILS",
  "explain a hold", "what is a missed approach", "VOR/DME approach",
  "RNAV/GPS approach", "IFR en route", "approach briefing", "instrument
  scan", "pitot-static system", "gyroscopic instruments", "IFR weather
  minimums", "alternate airport requirements", "IFR currency", "6 month
  IFR proficiency", "partial panel", "unusual attitude recovery", or any
  request to generate IFR training materials, lesson plans, or scenario briefs.
---

# IFR Flight Skills

Comprehensive reference for Instrument Flight Rules procedures, training,
systems, and operations. Covers FAA regulations (14 CFR), AIM guidance,
and practical technique for single-pilot IFR in Part 91 GA operations
unless otherwise specified.

> **Disclaimer**: This skill is for educational and simulation purposes.
> Always verify procedures against current official sources (FAA, Jeppesen,
> ForeFlight) and applicable regulations before real-world application.

---

## Skill Sections — Read What You Need

| Topic | Section |
|---|---|
| Instrument systems & failure modes | §1 |
| Instrument scan technique | §2 |
| IFR departure procedures | §3 |
| En route navigation & airways | §4 |
| Holding patterns | §5 |
| Approach procedures (ILS, VOR, RNAV, NDB) | §6 |
| Missed approach & alternate | §7 |
| IFR weather & minimums | §8 |
| ATC communication scripts | §9 |
| Partial panel & unusual attitude | §10 |
| IFR flight planning | §11 |
| Currency & training scenarios | §12 |

---

## §1 — Instrument Systems

### Pitot-Static System
Powers: Airspeed Indicator (ASI), Altimeter, Vertical Speed Indicator (VSI)

**Failure modes:**
| Failure | ASI | Altimeter | VSI |
|---|---|---|---|
| Pitot blocked (ice) | Reads zero (or acts like altimeter) | Normal | Normal |
| Static blocked | Freezes at altitude of blockage | Freezes | Reads zero |
| Both blocked | Reads zero | Freezes | Reads zero |
| Alternate static | Slightly high (cabin pressure lower) | Slightly high | Momentary deflection |

**Alternate static**: If available, use it on static blockage. In aircraft without it, break the VSI glass as last resort (cabin air source).

### Gyroscopic Instruments
**Attitude Indicator (AI)**: vacuum or electric gyro. Precesses ~3°/min. Tumbles beyond ±60° pitch / ±100° bank (older models). Errors on prolonged turns (gimbal error).

**Heading Indicator (HI/DI)**: vacuum gyro, precesses up to 3°/15 min — reset to compass every 15 minutes in straight and level.

**Turn Coordinator**: electric gyro (usually on separate bus). Senses roll rate and yaw. Ball = inclinometer (slip/skid indicator).

**Failure prioritization**: If AI fails, cross-check VSI + altimeter for pitch, TC + HI for bank. This is partial panel.

### Magnetic Compass Errors (ANDS / OSSC)
- **ANDS**: Accelerate North, Decelerate South (in northern hemisphere)
- **OSSC**: Overshoot South, Undershoot North (turning to cardinal headings)
- Most accurate: straight and level, wings level, constant speed

---

## §2 — Instrument Scan

### Primary & Supporting Method (recommended)
For each flight parameter, one instrument is **primary** (most direct), others are **supporting**:

| Phase | Primary Pitch | Primary Bank | Primary Power |
|---|---|---|---|
| Straight & level | Altimeter | HI | Airspeed |
| Climb (constant airspeed) | Airspeed | HI | Tachometer |
| Climb (constant rate) | VSI | HI | Tachometer |
| Turn (standard rate) | Altimeter | Turn Coordinator | Airspeed |
| Descent | VSI or Altimeter | HI | Airspeed |

### Cross-Check Pattern
Never fixate. Scan continuously. Common errors:
- **Fixation**: staring at one instrument
- **Omission**: consistently skipping VSI or ball
- **Emphasis**: overweighting one instrument (AI) and ignoring supporting

### Scan During Approach (ILS)
Priority during final: **ADI → HSI/CDI → Airspeed → Altimeter → VSI → repeat**
Check ball every scan cycle. Call out deviations: "dot high, correcting."

---

## §3 — IFR Departure

### Clearance Readback (CRAFT)
```
C — Clearance limit (destination or fix)
R — Route
A — Altitude (initial, expect ___ in ___ minutes)
F — Frequency (departure)
T — Transponder (squawk code)
```

Always read back: assigned altitude, heading, squawk code, any "climb via" SID instructions.

### Departure Procedures
**ODP (Obstacle Departure Procedure)**: published in text or graphically. Provides obstacle clearance if unable to accept SID. File/fly if terrain/obstacles present.

**SID (Standard Instrument Departure)**: ATC-preferred routing out of busy airports. Reduces radio traffic. May include crossing restrictions.

**Diverse Departure**: If no ODP/SID, maintain runway heading to 400ft AGL before turning. Requires 200ft/NM climb gradient to 400ft AGL minimum.

**Takeoff minimums**: Part 91 = no legal minimum, but published takeoff minimums indicate obstacles present. Non-standard minimums marked ▲ or T on approach plates.

### IFR Departure Sequence
1. Obtain clearance on ground (or void time if remote)
2. Copy, read back CRAFT
3. Before takeoff: set altimeter, set heading indicator, brief departure
4. On contact with departure: report altitude, heading, any deviations

---

## §4 — En Route

### Victor Airways (VOR)
- **Width**: 4 NM each side of centerline (within 51 NM of VOR); expands beyond
- **MEA**: Minimum En Route Altitude — obstacle clearance + navaid reception
- **MOCA**: Minimum Obstruction Clearance Altitude — obstacle clearance only; navaid reception guaranteed only within 22 NM of VOR
- **MRA**: Minimum Reception Altitude — needed to receive a specific fix

### RNAV/Q-Routes
High altitude RNAV routes. Requires IFR-approved GPS. MEA applies.

### Position Reporting (non-radar)
Required at: compulsory reporting points (solid triangle on chart), and when requested.
Report: **PTA-TEN** — Position, Time, Altitude, Type of flight plan, ETA next fix, Name of fix after that

### Altitude Rules
- Below 18,000 ft MSL: odd thousands eastbound (0–179°), even thousands westbound (180–359°)
- At or above FL180: FL odd eastbound, FL even westbound (RVSM above FL290)

---

## §5 — Holding Patterns

### Standard Hold
- **Right turns**, 1-minute legs (above 14,000 MSL: 1.5 min), inbound on published course

### Entry Procedures
Determine entry based on heading relative to holding course:

```
  Parallel entry zone: ~70° on non-holding side of inbound course
  Teardrop entry zone: ~110° on non-holding side
  Direct entry zone: remaining ~180°
```

**Parallel**: Fly outbound on holding side, turn left (for standard hold) to intercept inbound
**Teardrop**: Fly 30° into holding pattern (toward non-holding side), then turn to intercept inbound
**Direct**: Turn immediately to fly the holding pattern

### Timing
- Start timing when: wings level on outbound OR abeam the fix (whichever comes later)
- Adjust outbound leg to achieve 1-minute inbound leg
- Wind correction: apply double the inbound correction on outbound (triple method)

### Speed Limits in Holds
- At/below 6,000 ft MSL: 200 KIAS max
- 6,001–14,000 ft: 230 KIAS max
- Above 14,000 ft: 265 KIAS max

### EFC (Expect Further Clearance)
Always note EFC time. If comms lost in hold, depart at EFC time.

---

## §6 — Approach Procedures

### Approach Briefing (5 T's + MARTHA)
```
T — Time (estimated time on approach)
T — Turn (direction after takeoff on missed)
T — Time/Distance (missed approach point)
T — Twist (final approach course — set in CDI/HSI)
T — Throttle (power setting for approach)

M — Minimums (DA or MDA — circle vs. straight-in)
A — Airport (location, field elevation)
R — Route (full procedure or vectors to final)
T — Terrain (obstacles on approach, missed)
H — Highest obstacle (on missed approach climb)
A — Alternate (filed alternate, minimums)
```

### ILS Approach
Components: **Localizer** (lateral, 108.1–111.95 MHz odd tenths) + **Glideslope** (lateral 329–335 MHz, paired with LOC)

- **Full scale deflection**: LOC = 2.5° each side; GS = 0.7° each side (1 dot = ~0.35°)
- **Decision Altitude (DA)**: Height above TDZE where missed approach must be initiated if runway not in sight
- Typical DA: 200 ft HAT / ½ sm visibility (CAT I)
- Glideslope angle: typically 3.0° (some 2.5–3.5°)
- GS intercept: from below (avoid false glideslope ~6° above)

**ILS Technique**: intercept localizer at least 2 NM before GS intercept. Configure for approach before GS intercept. Gear down and final flaps at GS intercept. Fly small corrections. Announce: "Localizer alive… captured. Glideslope alive… captured."

### RNAV (GPS) Approach
Three lines of minima:
- **LPV** (Localizer Performance with Vertical Guidance): DA, like ILS. Requires WAAS GPS.
- **LNAV/VNAV**: DA, barometric vertical guidance. Requires baro-VNAV or WAAS.
- **LNAV**: MDA, lateral only. All IFR-approved GPS.

CDI scaling: switches from ±2 NM (en route) → ±1 NM (terminal, 30 NM out) → ±0.3 NM (approach, 2 NM FAF)

### VOR Approach
- Set OBS to final approach course
- FAF identified by Maltese cross (profile view)
- Time from FAF to MAP at approach speed → determines when to execute missed
- CDI full scale = 10° each side

### Circling Approaches
- Used when: straight-in not possible or runway not aligned within 30°
- Higher minimums than straight-in
- Circling radius (CAT A–D) measured from each runway threshold
- Stay within circling radius. If lost sight of runway: climbing turn toward landing runway, then fly missed approach

---

## §7 — Missed Approach & Alternate

### Executing Missed Approach
Trigger: reaching DA/MDA without required visual references, OR anytime prior to DA.
Sequence: **Pitch up → Power up → Clean up (flaps/gear in sequence)**
- Announce "Missed approach" on frequency
- Begin published procedure immediately
- Contact ATC

### Lost Comms (14 CFR 91.185)
Route: **AVEF**
```
A — Assigned (last assigned route)
V — Vectored (route being vectored to)
E — Expected (route ATC told you to expect)
F — Filed (route in flight plan)
```
Altitude: **MEA** or **Assigned** or **Expected** — highest of the three.
Approach: begin approach at EFC, or if no EFC, at ETA from flight plan.

### Alternate Requirements (1-2-3 Rule)
File alternate if, from 1 hr before to 1 hr after ETA, destination forecast is below 2,000 ft ceiling or 3 sm visibility.

**Alternate minimums** (standard): Precision approach: 600-2; Non-precision: 800-2. Some airports have non-standard (NA for alternate).

---

## §8 — IFR Weather & Minimums

### Weather Products for IFR
- **TAF**: Terminal Aerodrome Forecast — 24/30 hr, ICAO format
- **METAR**: Current conditions, updated hourly (SPECI = special)
- **PIREPs**: Pilot Reports — icing, turbulence, cloud tops
- **SIGMETs**: Significant meteorological conditions — convective, icing, turbulence (WS/WA/WST)
- **AIRMETs**: Less severe — Sierra (IFR/mtn obscuration), Tango (turbulence), Zulu (icing)
- **Prog Charts**: 12/24/36/48 hr surface/upper analysis
- **Icing forecasts**: CIP (current) and FIP (forecast)

### Icing
Structural icing: visible moisture + OAT between -20°C and +2°C
Types: **Rime** (rough, milky — freezing drizzle), **Clear** (smooth, heavy — large droplets), **Mixed**
Avoid: known icing in non-FIKI aircraft (14 CFR 91.9 — no careless/reckless operation)

---

## §9 — ATC Communication Scripts

### IFR Clearance
```
"[Facility], [Callsign], ready to copy IFR clearance"
ATC delivers clearance → read back CRAFT items
"[Callsign] reads back: cleared to [dest] via [route], 
 initial altitude [alt], expect [alt] in [time], 
 departure [freq], squawk [code]"
```

### Departure Check-In
```
"[TRACON], [Callsign], [aircraft type], [altitude] climbing [assigned alt], 
 heading [assigned heading], out of [airport]"
```

### Approach Request
```
"[Approach], [Callsign], [position], [altitude], request [approach type] 
 approach [runway] [airport]"
```

### Reporting Leaving/Reaching Altitudes (non-radar)
```
"[Center], [Callsign], leaving [alt], reaching [alt] at [time]"
```

### Declaring Emergency
```
"MAYDAY MAYDAY MAYDAY, [Callsign], [nature of emergency], 
 [souls on board], [fuel remaining], [intentions]"
```

---

## §10 — Partial Panel & Unusual Attitude

### Partial Panel (AI and HI failed)
**Magnetic compass turns**: UNOS (Undershoot North, Overshoot South) for leading/lagging
**Standard rate turn**: 3°/sec — use turn coordinator
**Pitch control**: altimeter + VSI
**Bank**: turn coordinator + airspeed (increasing = nose down or banking)

### Unusual Attitude Recovery
**Nose high** (increasing altitude, decreasing airspeed):
1. Power — full
2. Bank — level wings with TC
3. Pitch — lower to horizon

**Nose low** (decreasing altitude, increasing airspeed):
1. Power — reduce
2. Pitch — level wings FIRST (to prevent overstress)
3. Bank — pull to horizon

**Never pull through a nose-low unusual attitude with wings banked** — exceeds Va and structural limits.

---

## §11 — IFR Flight Planning

### Route Selection
1. Check TFRs, NOTAMs, SIGMETs
2. Select routing: airways (V/J/Q), direct (if GPS-equipped + ATC agrees), preferred routes
3. Check MEAs along route vs. aircraft service ceiling
4. Identify alternates along route (fuel diversions)
5. Calculate fuel: trip + reserve + alternate + 45 min (Part 91 IFR)

### Fuel Reserve (14 CFR 91.167)
IFR fuel: enough to reach destination, fly to alternate, then fly 45 minutes at normal cruise.

### Standard IFR Flight Plan Fields
Departure, route, destination, alternate, TAS, altitude, estimated time en route, fuel on board, souls on board, color of aircraft, emergency equipment.

---

## §12 — Currency & Training Scenarios

### IFR Currency (61.57c)
Within preceding 6 calendar months:
- 6 instrument approaches
- Holding procedures and tasks
- Intercepting and tracking courses via navigational systems

If expired (6–12 months): requires Instrument Proficiency Check (IPC) with CFII or examiner.
After 12 months: IPC required.

### Training Scenario Template
When generating IFR training scenarios, include:
```
Scenario: [name]
Aircraft: [type + avionics]
Departure: [ICAO + conditions]
Route: [airways/fixes]
Destination: [ICAO]
Alternate: [ICAO]
Weather: [ceiling/vis/wind/icing/turbulence]
Primary training objectives:
  1. [skill 1]
  2. [skill 2]
Abnormal/emergency inject: [failure at what point]
Debrief focus: [what to discuss]
```

### Common IFR Training Maneuvers
- Timed turns to headings (partial panel)
- Unusual attitude recoveries
- Intercepting and tracking VOR radials
- Procedure turns and holding entries
- Coupled and raw data ILS to minimums
- Circling approach
- Lost comms scenario
- LAHSO decline and go-around on approach