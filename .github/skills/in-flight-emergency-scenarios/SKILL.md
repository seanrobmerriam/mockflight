---
name: inflight-emergency
description: >
  Create detailed, realistic in-flight emergency scenarios for pilot training,
  simulator sessions, instructional design, and sim game scripting. Use this
  skill whenever the user asks to create, generate, write, or design any
  in-flight emergency, abnormal procedure scenario, or pilot training exercise
  involving system failures, weather encounters, medical emergencies, or
  aircraft malfunctions. Trigger for: "create an engine failure scenario",
  "write an emergency training scenario", "design a sim emergency", "build a
  partial panel scenario", "make an IFR emergency exercise", "generate a CFII
  training scenario", "write an emergency for my flight sim game", "create a
  progressive failure sequence", "design an inadvertent IMC scenario",
  "make a hypoxia scenario", "engine fire", "electrical failure scenario",
  "fuel exhaustion scenario", "structural damage scenario", "bird strike",
  "gear malfunction", "pressurization failure", or any request to describe
  realistic aircraft emergencies with symptoms, decision points, and outcomes.
---

# In-Flight Emergency Scenario Creation Skill

Generate detailed, realistic in-flight emergency scenarios for pilot training,
CFII lesson planning, simulator sessions, and flight sim game scripting.
Scenarios are calibrated for educational value — symptoms are accurate,
decision trees are realistic, and outcomes reflect real-world consequences.

> **Disclaimer**: For training and simulation purposes only. Always train
> with a qualified CFI/CFII. Emergency procedures must be verified against
> your aircraft's current AFM/POH.

---

## Scenario Anatomy

Every scenario produced by this skill follows this structure:

```
1. Setup          — aircraft, environment, phase of flight
2. Trigger        — what causes the emergency, when it occurs
3. Initial Cues   — what the pilot sees/hears/feels first
4. Progression    — how the situation develops if no action taken
5. Decision Points — key moments where pilot choices diverge outcomes
6. Correct Actions — memory items + checklist + ATC coordination
7. Common Errors   — what pilots typically get wrong in this scenario
8. Outcomes        — consequence tree (best case → worst case)
9. Debrief Points  — what the instructor should discuss after
10. Sim Inject Notes — if used in a sim, how to inject the failure
```

---

## Scenario Complexity Tiers

Choose based on user's stated purpose:

| Tier | Use Case | Complexity |
|---|---|---|
| **Introductory** | Student pilot first exposure, low-stress familiarization | Single failure, clear cues, forgiving timeline |
| **Proficiency** | Instrument rating / commercial training, recurrent | Single or dual failure, moderate time pressure |
| **Advanced** | ATP, CFII, upset recovery, CRM | Cascading failures, ambiguous cues, high workload |
| **Sim Game** | Browser/desktop flight sim entertainment | Dramatic cues, forgiving outcomes, narrative flavor |

---

## Emergency Category Library

### Category A — Powerplant Failures

**Engine Failure — Takeoff Roll**
- Trigger: abort point before or after V1 (jets) / past point of no return (GA)
- Cues: sudden RPM drop, right yaw (single), loss of acceleration feel
- Decision point: abort vs. continue (GA: always abort if runway remaining)
- Time pressure: 3–5 seconds to decide and act

**Engine Failure — Climb Out (below 1,000 AGL)**
- Most critical phase — no time for troubleshooting
- Cues: sudden silence, prop windmilling, yaw, nose wants to rise
- Memory items ONLY — no checklist reference at low altitude
- Landing area selection is the primary task
- Common error: attempting to return to runway ("impossible turn" — 180° below 1,000 AGL is nearly always fatal)

**Engine Failure — Cruise**
- Time available: altitude ÷ 500 fpm (rough glide estimate)
- Cue sequence: roughness → RPM decay → silence → prop windmilling
- Restart attempt sequence: rich mixture, boost pump, switch tanks, carb heat
- Mayday call, squawk 7700, ATC assistance

**Engine Roughness / Partial Power Loss**
- Carb ice: gradual roughness, RPM loss, cured by carb heat (temp drop then rise)
- Fuel contamination: intermittent roughness, cured by switching tanks
- Magneto failure: rough on both, smooth on one — identify on runup
- Vapor lock (fuel-injected): restart procedure differs from carbureted

**Engine Fire**
- In flight vs. on ground — very different procedures
- Cues: smoke/flames from cowling, smell, fire warning light
- Fuel flow OFF, mixture cutoff, throttle closed — starve the fire
- Best glide, Mayday, land ASAP — no restart attempt after confirmed fire
- Common error: opening cowl flaps (fans the fire)

---

### Category B — Electrical Failures

**Complete Electrical Failure**
- Cues: radios go silent, nav lights out, avionics blank, ammeter zero
- Impacts: transponder, COM radios, electric gyros (TC), electric flaps (some aircraft)
- Remaining: vacuum instruments (AI, HI), non-electric magnetos, mechanical trim
- Procedure: load shed (turn off everything), verify main bus, check alternator CB
- Squawk 7600 if transponder still has battery

**Alternator Failure (partial — battery still on)**
- Cues: ammeter negative / zero, low voltage annunciator, battery discharging
- Timeline: ~30 min of battery remaining (varies widely — know your aircraft)
- Action: load shed immediately, declare emergency, land soonest
- Common error: continuing normal flight unaware until avionics die

**Partial Panel (vacuum failure)**
- AI and HI fail, TC and ASI/ALT/VSI unaffected
- Cues: AI shows unusual bank that doesn't match TC, HI spins erratically
- Primary failure indicator: AI shows bank but TC ball is centered — trust TC
- Technique: standard rate turns using TC, timed legs, magnetic compass

---

### Category C — Pitot-Static Failures

**Pitot Blockage (ice)**
- Cues: ASI starts reading incorrectly — freezes or reads like altimeter
  - Climbs: ASI reads high (static escapes, pitot pressure trapped and expands)
  - Descends: ASI reads low (reverse of above)
- Action: pitot heat ON (if installed), use alternate airspeed sources
  (power + pitch attitude)
- Common error: chasing the incorrect ASI reading into unusual attitude

**Static Port Blockage**
- Cues: ASI slightly off, altimeter freezes at altitude of blockage, VSI reads zero
- Action: alternate static source (if installed); break VSI glass as last resort
- Note: alternate static causes slight over-read on altimeter (~50 ft typical)

**Pitot-Static System Icing (both)**
- Both blocked: ASI reads zero; altimeter + VSI freeze
- Fly by attitude and power — memorize power settings for each phase

---

### Category D — Flight Control Failures

**Elevator Trim Runaway**
- Cues: sudden unexpected pitch change, trim wheel spinning
- Action: counter with elevator pressure, pull circuit breaker, fly with elevator
- Common error: fighting trim with elevator until fatigued, not pulling CB

**Jammed Controls**
- Full or partial jam — assess range of motion immediately
- Asymmetric flaps: do not attempt to retract — land with asymmetric configuration, higher speed
- Jammed rudder: differential power for directional control (twins), land with crosswind technique

**Gear Malfunction**
- Unsafe indication: visual check via mirror/wing walkaround, fly-by tower
- Manual extension procedure: check POH sequence — hydraulic, pneumatic, gravity/freefall
- One gear up: consider landing with all gear up vs. asymmetric

---

### Category E — Weather Encounters

**Inadvertent IMC (VFR into IMC)**
- Most lethal scenario in GA — 178-second survival time statistic
- Cues: gradual visibility reduction, horizon loss, spatial disorientation
- Immediate action: wings level (TC), gentle standard rate turn to reverse course
- Do NOT descend — terrain unknown; do NOT accelerate; do NOT panic-pull
- Common error: attempting to navigate rather than immediately reversing course

**Thunderstorm Penetration**
- Never intentional — but if trapped: reduce to Va, maintain wings level, accept altitude deviation
- Hail, severe turbulence, lightning, windshear, icing — all simultaneously
- Mayday, request vectors to clear air, declare unable to maintain altitude

**Severe Icing Encounter**
- Non-FIKI aircraft: exit icing conditions immediately — descend, reverse course, climb above
- FIKI aircraft: activate all deice/anti-ice, exit if accumulating faster than equipment sheds
- Tail plane icing: flap deployment may cause pitch-down — known as tailplane stall
  Cues: elevator buffet, pitch-down tendency at flap deployment
  Recovery: retract flaps (counter-intuitive — opposite of wing stall)

---

### Category F — Cabin/Medical Emergencies

**Hypoxia (pressurized aircraft or high-altitude ops)**
- Cues: euphoria, tingling, blue lips (cyanosis), then loss of consciousness
- Insidious — pilot may not self-identify until incapacitated
- Action: 100% oxygen immediately, emergency descent, squawk 7700
- Time of useful consciousness at altitude:
  - FL250: ~3–5 min; FL350: ~30–60 sec; FL410: ~9–15 sec

**Smoke / Fumes in Cockpit**
- Source identification: electrical (acrid), oil (burning smell), CO (odorless — use detector)
- Electrical smoke: isolate by load shedding, pull CBs
- CO: open fresh air vents, oxygen if available, land immediately
- Oxygen fire: do NOT use oxygen with electrical smoke/fire

**Pilot Incapacitation (multi-crew or with passenger)**
- Passenger taking controls: "wings level, horizon in center, don't touch rudder"
- ATC assistance: "PAN PAN, passenger on controls, request talking down to landing"
- Declare emergency, request radar vectors to long runway, simple approach

---

## Progressive / Cascading Failure Sequences

For advanced training, chain failures with realistic causal relationships:

**Scenario: Vacuum + Weather**
1. [cruise] Vacuum pump fails silently → AI and HI begin precessing
2. [30 min later] Pilot enters IMC — AI now showing 15° error
3. Pilot trusts AI over TC → enters graveyard spiral
4. Airspeed builds, altimeter unwinds — now also a VNE/structural scenario

**Scenario: Fuel Mismanagement**
1. Preflight: fuel sump shows clear but pilot accepts incorrect fuel quantity
2. [en route, 2 hrs] Engine roughness begins — one tank running dry
3. Pilot switches tanks late — engine stops, requires air restart
4. Air restart succeeds but now pilot is uncertain of total fuel remaining
5. Decision: divert vs. continue; destination is 45 min away

**Scenario: Night + Electrical**
1. [night, over mountains] Alternator fails
2. Pilot notices dim panel lights, doesn't immediately identify cause
3. [15 min later] COM radio dies — now lost comm + partial panel at night
4. Battery-backed transponder still operational — squawk 7600
5. Must locate airport using dead reckoning, execute lights-out approach

---

## Scenario Output Format

### Training Scenario (CFII / Sim Session)

```
═══════════════════════════════════════════════════════
EMERGENCY SCENARIO: [Title]
Category: [A–F]  |  Tier: [Intro/Proficiency/Advanced]
═══════════════════════════════════════════════════════

SETUP
  Aircraft:     [type + avionics + configuration]
  Phase:        [phase of flight + time of day]
  Environment:  [weather, visibility, terrain]
  Fuel state:   [quantity + time remaining]

TRIGGER
  [What fails, when, and why — include realism detail]

INITIAL CUES  (T+0 to T+30 sec)
  Visual:   [what pilot sees on instruments/outside]
  Auditory: [sounds — silence, alarms, changes]
  Physical: [forces, vibration, control feel changes]

PROGRESSION IF NO ACTION TAKEN
  T+30 sec:  [state of aircraft]
  T+2 min:   [state of aircraft]
  T+5 min:   [state of aircraft / point of no return]

DECISION POINTS
  DP1: [first key decision — options and consequences]
  DP2: [second key decision]
  DP3: [final decision before outcome locks in]

CORRECT ACTIONS
  Memory items:
    1. [item]
    2. [item]
  Checklist reference: [section + page]
  ATC coordination: [what to say, when]
  Target landing: [nearest suitable airport or off-field]

COMMON ERRORS
  • [error 1 — and why it's dangerous]
  • [error 2]
  • [error 3]

OUTCOMES
  Best case:   [correct actions, timely execution]
  Acceptable:  [delayed but correct actions]
  Poor:        [incorrect prioritization]
  Worst case:  [no action / wrong action]

DEBRIEF POINTS
  1. [key learning objective]
  2. [key learning objective]
  3. [key learning objective]

SIM INJECT NOTES
  [Specific instructions for injecting this failure in
   X-Plane / MSFS / custom sim — what system to fail,
   at what state variable value, any scripting notes]
```

---

## Sim Game Scenario Format

For flight sim games, add narrative flavor:

```
SCENARIO: [Dramatic title]
FLAVOR TEXT: [1–2 sentences of story context for the player]

TRIGGER: [At what sim event — e.g., "when player reaches 5,000 AGL"]
FAILURE: [System ID to inject]

PLAYER PROMPTS:
  - [HUD message or ATC call that alerts player]
  - [Follow-up message if no action taken in 30 sec]

SUCCESS CONDITION: [What player must do to resolve]
FAILURE CONDITION: [What ends the scenario badly]

REWARD: [Score multiplier, achievement, story unlock]
```

---

## Realism Guidelines

- **Cue accuracy**: always describe the actual physical/instrument manifestation —
  never just "the engine fails." Describe RPM decay rate, sounds, forces.
- **Timeline realism**: low-altitude failures have seconds, not minutes.
  High-altitude failures may allow many minutes of analysis.
- **No magic solutions**: if an engine fire is not extinguished in time, the
  wing fails. Outcomes must be plausible.
- **Human factors**: include startle response, confirmation bias traps,
  and task saturation as scenario elements in advanced tiers.
- **ATC realism**: include realistic ATC callsigns, frequencies, and response
  scripts — don't just say "call ATC," write the actual exchange.