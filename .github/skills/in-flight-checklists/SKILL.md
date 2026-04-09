---
name: inflight-checklist
description: >
  Create accurate, well-formatted aviation checklists for any phase of flight
  and any aircraft type. Use this skill whenever the user asks to create,
  generate, write, format, or review a flight checklist, cockpit checklist,
  emergency checklist, abnormal procedure checklist, or any document that
  organizes aircraft operating steps by phase or condition. Trigger for:
  "make a checklist for [aircraft]", "create a normal procedures checklist",
  "write an emergency checklist", "build a preflight checklist", "design a
  cockpit flow", "generate POH-style checklist", "create a checklist for
  my sim", "format these procedures as a checklist", "IFR departure
  checklist", "approach and landing checklist", or any request to turn
  aircraft procedures into a structured, actionable list. Covers GA
  piston, turboprop, jet, helicopter, and simulator/game aircraft.
---

# In-Flight Checklist Creation Skill

Produce accurate, properly structured aviation checklists for any aircraft
type and any phase of flight. Follows real-world checklist design standards
used by POH publishers and airline operations manuals.

---

## Checklist Design Philosophy

### Challenge-Response vs. Flow
Two paradigms — choose based on context:

**Challenge-Response** (printed checklist): Pilot reads item name aloud, co-pilot
or pilot responds with action/state. Formal, used in real aircraft.
```
FUEL SELECTOR ......................... BOTH
MIXTURE .............................. RICH
PRIMER ............................... IN AND LOCKED
```

**Flow + Verify**: Pilot runs a cockpit flow from memory (touch and check),
then uses the checklist to verify completion. Used by experienced pilots and
airlines. Checklist is shorter — only critical/verify items.

**For simulators and games**: use challenge-response format — it's clearer for
users who may not know the aircraft.

### Item Ordering
Items must follow cockpit ergonomics and procedural logic:
1. Safety-critical items first (fuel, mixture, controls free)
2. Left-to-right, top-to-bottom sweep within a panel
3. Items that take time (warm-up, gyro stabilize) go early
4. Doors/windows last in pre-takeoff (common omission)
5. Never put a "verify set" item before the action that sets it

### Response Format Standards
```
ITEM NAME ........................ ACTION/STATE
```
- Item name: ALL CAPS, concise, matches cockpit placard
- Dots fill to consistent column (usually 40 chars total)
- Response: brief — "ON", "RICH", "SET ___", "CHECKED", "IN THE GREEN"
- When pilot sets a value: use "SET ___" with blank for pilot to fill
- When pilot verifies a pre-set: use "CHECKED" or "VERIFIED"

---

## Standard Checklist Phases

Generate these phases unless the user specifies otherwise. Include only
phases relevant to the aircraft type (e.g., no pressurization for a C172).

### Normal Procedures
1. **Preflight Inspection** (walkaround — exterior, by station)
2. **Before Starting Engine**
3. **Engine Start**
4. **After Start**
5. **Taxi**
6. **Before Takeoff (Runup)**
7. **Lineup / Line-Up and Wait**
8. **After Takeoff / Climb**
9. **Cruise**
10. **Descent**
11. **Approach**
12. **Before Landing**
13. **After Landing**
14. **Shutdown**
15. **Securing Aircraft**

### IFR-Specific Additions
Insert between relevant normal phases:
- **IFR Departure Brief** (after Before Takeoff, before Lineup)
- **Approach Brief** (during Descent)
- **Missed Approach** (after Approach)
- **Holding** (as needed)

### Emergency / Abnormal Procedures
Format differently — emergency checklists use **bold** or **MEMORY ITEMS**
marked clearly. Memory items performed immediately without reference;
remaining items from checklist.

```
ENGINE FAILURE AFTER TAKEOFF

MEMORY ITEMS (complete without checklist):
  AIRSPEED .......................... Vg (65 KIAS)
  THROTTLE .......................... IDLE
  MIXTURE ........................... IDLE CUTOFF
  FUEL SELECTOR ..................... OFF
  IGNITION .......................... OFF
  FLAPS ............................. AS REQUIRED
  MASTER SWITCH ..................... OFF (after electrical fire confirmed)

THEN USE CHECKLIST:
  DOORS ............................. UNLATCH PRIOR TO LANDING
  LANDING AREA ...................... SELECT
  TRANSPONDER ....................... 7700
```

---

## Aircraft-Type Templates

### Single-Engine GA Piston (C172, PA-28, DA40-class)

**BEFORE STARTING ENGINE**
```
PREFLIGHT INSPECTION ................. COMPLETE
SEATS AND BELTS ...................... ADJUSTED AND LOCKED
FUEL SELECTOR ........................ BOTH
AVIONICS MASTER ...................... OFF
ELECTRICAL EQUIPMENT ................. OFF
BRAKES ............................... TEST AND SET
CIRCUIT BREAKERS ..................... IN
```

**ENGINE START (Fuel-Injected)**
```
THROTTLE ............................. OPEN 1/4 INCH
MIXTURE .............................. RICH
FUEL PUMP ............................ ON (check pressure)
FUEL PUMP ............................ OFF
MIXTURE .............................. IDLE CUTOFF
THROTTLE ............................. OPEN 1/2 INCH
PROPELLER AREA ....................... CLEAR
MASTER SWITCH ........................ ON (BAT)
IGNITION ............................. START
[when engine fires]
MIXTURE .............................. ADVANCE TO RICH
OIL PRESSURE ......................... CHECK GREEN
ALTERNATOR ........................... ON
```

**BEFORE TAKEOFF (RUNUP)**
```
PARKING BRAKE ........................ SET
FLIGHT CONTROLS ...................... FREE AND CORRECT
TRIM ................................. SET FOR TAKEOFF
MIXTURE .............................. RICH (or per POH for field elevation)
FUEL SELECTOR ........................ BOTH
THROTTLE ............................. 1800 RPM
  MAGNETOS ........................... CHECK (max 125 RPM drop, 50 RPM differential)
  CARB HEAT .......................... ON (check RPM drop), OFF
  ENGINE INSTRUMENTS ................. CHECK GREEN
  ALTERNATOR ......................... CHECK (ammeter positive)
THROTTLE ............................. 1000 RPM
PRIMER ............................... IN AND LOCKED
AVIONICS ............................. ON
RADIOS ............................... SET
NAV / GPS ............................ SET AND VERIFIED
TRANSPONDER .......................... ALT
LIGHTS ............................... AS REQUIRED
DOORS AND WINDOWS .................... CLOSED AND LATCHED
DEPARTURE BRIEF ...................... COMPLETE
```

**APPROACH AND LANDING**
```
FUEL SELECTOR ........................ BOTH
MIXTURE .............................. RICH (or per POH)
CARB HEAT ............................ ON
BRAKES ............................... TEST
SEATBELTS ............................ SECURE
FLAPS ................................ SET ___
AIRSPEED ............................. Vref ___
```

---

### Multi-Engine Piston (PA-44, BE-76-class)

Add engine-specific items:
```
BEFORE TAKEOFF (additional):
  PROPELLERS ......................... HIGH RPM
  COWL FLAPS ......................... OPEN
  ENGINE INSTRUMENTS (both) ........... CHECK GREEN
  CROSSFEED .......................... CHECK AND OFF
  FUEL PUMPS (both) .................. BOOST ON
```

**Engine Failure MEMORY ITEMS:**
```
AIRSPEED ............................. Vyse (BLUE LINE) — 88 KIAS
MIXTURES ............................. RICH
PROPELLERS ........................... HIGH RPM
THROTTLES ............................ FULL
GEAR .................................. UP
FLAPS ................................ RETRACT
IDENTIFY ............................. Dead foot = dead engine
VERIFY ............................... Reduce throttle on dead engine
FEATHER .............................. Dead engine propeller
```

---

### Turboprop / Single Engine Turbine (TBM, PC-12-class)

Additional items vs. piston:
```
BEFORE START (additions):
  EMERGENCY POWER LEVER .............. GROUND
  IGNITION SELECTOR .................. NORM
  FUEL CONDITION LEVER ............... CUTOFF
  POWER LEVER ........................ GROUND IDLE

ENGINE START:
  FUEL PUMP .......................... ON
  IGNITION ........................... ON
  STARTER ............................. ENGAGE (at start ITT stabilized)
  FUEL CONDITION LEVER ............... LO IDLE at N1 12%
  ITT .................................. MONITOR (max starting ITT: per AFM)
  N1, N2, ITT, OIL PRESSURE .......... CHECK IN GREEN
  GENERATOR .......................... ON
  BLEED AIR .......................... AS REQUIRED

BEFORE TAKEOFF (additional):
  TORQUE LIMITER ...................... TEST
  AUTOFEATHER ......................... ARM AND TEST
  PRESSURIZATION ...................... SET
  BLEED AIR .......................... AS REQUIRED
  ANTI-ICE ........................... AS REQUIRED
```

---

### IFR Checklist Additions

**IFR DEPARTURE BRIEF**
```
ATIS / D-ATIS ........................ CURRENT [info __]
CLEARANCE ............................ COPIED AND READ BACK
DEPARTURE FREQUENCY .................. SET ___
SQUAWK ............................... SET ___
DEPARTURE PROCEDURE .................. REVIEWED [_______]
INITIAL ALTITUDE ..................... SET ___
INITIAL HEADING / COURSE ............. SET ___
ENGINE FAILURE PROCEDURE ............. BRIEFED
TAKEOFF MINIMUMS ..................... REVIEWED
```

**APPROACH BRIEF**
```
ATIS .................................. CURRENT [info __]
APPROACH ............................. [type + runway]
IAF / VECTORS ........................ [from ATC]
FINAL APPROACH COURSE ................ SET ___ IN CDI/HSI
STEP-DOWN FIXES ...................... NOTED
FAF .................................. [fix name + altitude]
MINIMUM ALTITUDE ..................... DA/MDA ___ ft
VISIBILITY REQUIRED .................. ___ sm / RVR ___
MISSED APPROACH ...................... REVIEWED
  INITIAL ACTION ..................... [climb + heading]
  HOLDING FIX ........................ [if applicable]
ALTERNATE ............................ [ICAO + mins]
```

---

## Simulator / Game Aircraft

When creating checklists for fictional or game aircraft (Microsoft Flight Simulator,
X-Plane mods, DCS, or custom sim games), adapt real-world structure to the
modeled systems. Ask the user:

1. Aircraft type / inspiration (real-world analog)?
2. What systems are modeled (fuel injection, turbo, pressurization, autopilot)?
3. Complexity level desired (full POH-style vs. quick-reference card)?
4. Format: printable document, in-sim kneeboard, HTML overlay, or plain text?

If no real-world analog exists, base the checklist on the closest real aircraft
class and note assumptions.

---

## Output Formats

### Standard Plain Text (default)
Use fixed-width formatting with dot leaders. Works universally.

### Kneeboard Card (compact)
Two-column layout, abbreviated item names, fits on 5×8 card. Use when user
says "kneeboard", "quick reference", or "QRC".

### HTML Overlay (for sim games)
Formatted as a styled HTML div, suitable for injecting into a browser-based
simulator UI. Sections collapsible, items with checkbox input.

```html
<div class="checklist">
  <h3>BEFORE TAKEOFF</h3>
  <div class="checklist-item">
    <span class="item-name">MIXTURE</span>
    <span class="item-action">RICH</span>
    <input type="checkbox">
  </div>
  ...
</div>
```

### Markdown (for documentation / repos)
```markdown
## Before Takeoff

| Item | Action |
|---|---|
| FUEL SELECTOR | BOTH |
| MIXTURE | RICH |
```

---

## Quality Rules

Every checklist produced must:

1. **Never omit safety-critical items**: fuel selector, mixture, controls free, seats locked, brakes
2. **Mark memory items clearly** in emergency checklists
3. **Match item names to cockpit placards** (or stated aircraft)
4. **Include Vref/Vg speeds** where relevant, with blanks for pilot to fill
5. **Sequence correctly** — no item depends on a later item being done first
6. **Note AFM/POH reference** where values are aircraft-specific: *(per AFM, Section 3)*
7. **Include a disclaimer** on any checklist intended for real-world use:
   > *For training/simulation reference only. Always use the current approved
   > AFM/POH for your specific aircraft registration.*