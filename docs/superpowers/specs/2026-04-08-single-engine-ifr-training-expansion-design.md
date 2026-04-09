# Single-Engine IFR Training Expansion Design

## Summary

MockFlight will expand from a small IFR sensor trainer into a more credible single-engine piston IFR training device with a larger scenario library, more realistic abnormal and emergency checklist content, and a phased roadmap for adding cockpit instruments and control options.

The expansion will be intentionally staged:

- phase 1 adds richer, realistic training scenarios and POH-style checklist content for failures the simulator can already express honestly
- phase 2 improves cockpit fidelity with additional IFR instruments and clearer instrument-failure presentation
- phase 3 adds more control options and aircraft-system interactions, but only when the simulator has real state behind them

The baseline aircraft assumption for this design is a single-engine piston IFR trainer with a hybrid panel feel rather than a replica of a specific aircraft or avionics vendor.

## Goals

- Expand the training library with realistic single-engine piston IFR scenarios.
- Make abnormal and emergency checklist content read like real-world trainer material rather than generic guidance text.
- Keep scenario fidelity aligned with real simulator behavior so the trainer stays credible.
- Evaluate the current simulator honestly and use that evaluation to drive the next instrument and control additions.
- Improve cockpit scan realism with additional IFR instruments that support actual training workflows.
- Provide a roadmap for future user-selectable aircraft profiles without requiring that abstraction now.

## Non-Goals

- Adding engine, electrical, or vacuum failures that are not yet modeled in the simulator for phase 1.
- Replicating a specific POH or certified avionics product line item-for-item.
- Turning the current trainer into a full aircraft systems simulator in one pass.
- Introducing UI elements that imply system depth the backend does not yet support.
- Creating a grading or instructor analytics system in this design.

## Baseline Training Identity

The trainer should be treated as a single-engine piston IFR platform first. Scenario language, checklist tone, and cockpit additions should assume a light GA aircraft flying single-pilot IFR with conventional piston-aircraft procedures.

That means:

- checklists should use concise challenge-response style entries appropriate for a training aircraft
- scenarios should emphasize workload management, scan discipline, pitch-power cross-checking, and approach control under instrument conditions
- takeoff, climb, missed approach, and partial-panel work should feel like single-pilot IFR training, not airline or turbine training

This baseline is a design choice, not a permanent limit. A future aircraft-choice feature may allow alternate aircraft profiles, but the first content expansion should stay anchored to one defensible training identity.

## Phase 1 Scope: Training Content Expansion

Phase 1 defines MockFlight as a scenario-driven IFR training device with a deeper catalog of failures and flight events that the simulator can already show through current instrument behavior.

The core content set should grow beyond the current small collection of sensor exercises into a structured library covering four realistic training buckets:

### 1. Pitot-Static Failures

- pitot blockage in climb or departure
- static blockage on approach or descent
- unreliable airspeed in IMC
- combined instrument cross-check events where airspeed, altitude, and VSI behavior disagree

### 2. Attitude And Heading Reference Failures

- gyro degradation during vectors or intercept
- magnetometer disturbance while maneuvering or tracking
- partial-panel recovery after unreliable attitude or heading indications
- degraded heading-reference events during approach or missed-approach transition

### 3. Takeoff And Climb Energy Events

- early rotation below Vr
- low-energy climb after liftoff
- runway overrun risk during delayed or weak acceleration
- unstable initial climb requiring pitch-power correction

### 4. IFR Workload Events Using Current Navigation Abstractions

- unstable intercept while coping with degraded heading information
- missed-approach transition under partial-panel conditions
- high-workload approach continuation where the correct action is to stabilize or go missed

Phase 1 should not include engine, vacuum, or electrical scenarios unless the simulator can actually degrade the displayed aircraft state in a defensible way. Those belong to later phases when the backend model supports them.

## Scenario Model

The scenario system should evolve from a small inline list into a structured training catalog. Each scenario should define:

- title
- category
- training phase
- weather or context
- trigger
- initial cockpit cues
- expected progression if mishandled
- correct checklist
- distractor checklist options
- fidelity label

The fidelity label is required because it tells the user how much of the scenario is actually modeled by the simulator. For this phase, only backend-backed scenarios should appear in the main training rotation, which means the operational fidelity tag for shipped scenarios should effectively be fully simulated within the current trainer envelope.

The scenario presentation should continue to support recognition rather than auto-answering. Launching a scenario should change aircraft state and panel behavior, but the user must still observe symptoms, decide what is wrong, and choose the matching checklist.

## Checklist Model

Checklists should move toward single-engine piston abnormal and emergency formatting rather than tutorial prose. The content should feel like trainer checklist material derived from POH-style logic, while remaining generic enough to avoid pretending to be a certified aircraft manual.

### Checklist Style

- concise challenge-response formatting
- action-first phrasing
- minimal explanation inside the checklist itself
- memory-item style only when the action really benefits from urgency
- no long educational paragraphs inside the checklist body

Examples of the target tone:

- PITOT HEAT ........................ ON
- AIRSPEED .......................... USE PITCH / POWER
- ATTITUDE INDICATION ............... CROSS-CHECK SUPPORTING INSTRUMENTS
- MISSED APPROACH ................... EXECUTE IF UNSTABLE

### Checklist Library Structure

The checklist library should be modeled separately from scenarios so that:

- multiple scenarios can map to the same checklist
- future aircraft variants can substitute checklist wording without changing scenario structure
- the same checklist can be reused across climb, cruise, approach, and missed-approach contexts

### Initial Checklist Groups

The first realistic single-engine piston checklist set should include:

- airspeed unreliable
- pitot blockage
- static source or altimeter/VSI unreliable
- heading or directional reference unreliable
- partial-panel attitude recovery
- low-energy takeoff or rejected takeoff guidance where applicable to the current modeled controls
- missed approach and go-around stabilization checklist content tied to IFR workflow

## User Experience For Phase 1

The scenario and checklist experience should remain scenario-based rather than turning into a static checklist browser.

The intended loop is:

1. choose a scenario
2. observe cues on the panel
3. identify the likely failure or unsafe state
4. choose the correct checklist from realistic options
5. continue flying the aircraft while using the procedure

Wrong-checklist selection should remain a training signal, not a hard fail state. The user should be able to re-evaluate and choose again without resetting the scenario.

## Simulator Evaluation

The current simulator is strongest in these areas:

- air-data and pitot-static behavior
- attitude and heading-reference degradation
- simplified takeoff energy and warning logic
- basic autopilot stabilization workflow

The current simulator is weakest in these areas:

- engine-system modeling
- electrical-system failures
- vacuum-system-specific failures distinct from general gyro degradation
- flap, trim, mixture, fuel, and propulsion system depth
- navigation-source realism beyond synthetic training abstractions

This evaluation should directly constrain feature work. The UI should not imply engine, electrical, or avionics-system depth that the backend does not yet have. Training realism comes from coherence between panel behavior and simulator behavior, not from the number of labels on the screen.

## Phase 2 Scope: Instrument Expansion

Phase 2 should improve cockpit fidelity where it directly supports IFR scan flow and failure recognition.

### Recommended Instrument Additions

- turn coordinator or compact rate-of-turn and slip-skid presentation
- clearer heading bug and altitude preselect bug presentation
- glideslope indication integrated with the approach guidance band
- stronger autopilot annunciation for AP, heading, altitude, vertical speed, and TO/GA states
- better unreliable-instrument flagging inside the scan zone

### Conditional Instrument Additions

These should only be added once real simulator state exists:

- engine strip with RPM, oil pressure, oil temperature, and fuel state
- alternator or bus status indications
- vacuum or suction indication if vacuum behavior becomes distinct in the simulator
- trim indication when trim state is modeled

The design rule for phase 2 is simple: if an added instrument improves real IFR scan behavior immediately, it can be justified now. If it exists only to decorate the cockpit without meaningful state behind it, it should wait.

## Phase 3 Scope: Control Expansion

Phase 3 should expand pilot controls while staying inside a single-engine piston IFR trainer identity.

### Recommended Future Control Additions

- flap setting control
- trim commands and trim indication
- course selector or nav-source selection
- more cockpit-like heading bug, altitude preselect, and vertical speed interactions

### Deferred Control Additions

These should wait for real backend system modeling:

- mixture
- fuel selector
- carb heat or alternate air
- engine-start or shutdown systems beyond the current runway-start abstraction

The trainer should only expose controls when using them changes aircraft behavior in a believable way. Otherwise the control becomes noise and undermines the overall credibility of the sim.

## Data And Code Organization Direction

The current inline scenario and checklist structures inside the dashboard handler are sufficient for the first implementation pass but should be treated as transitional.

The medium-term target is:

- a structured scenario catalog with explicit metadata and fidelity
- a separate checklist library keyed by checklist identifier
- a display-model layer that maps snapshot state plus scenario state into panel presentation

This organization keeps the trainer extensible for future aircraft selection without forcing premature abstraction now.

## Testing Strategy

Phase 1 testing should validate both training content integrity and trainer behavior.

### Scenario And Checklist Tests

- every scenario maps to exactly one correct checklist
- every scenario presents at least one plausible distractor checklist
- every checklist referenced by a scenario exists
- fidelity labels remain consistent with simulator-backed behavior

### Dashboard Tests

- the scenario shell renders the expanded catalog
- checklist selection and wrong-checklist feedback continue to work
- newly added scenario metadata does not break rendering

### Simulator Alignment Tests

- each fully simulated scenario corresponds to a real failure or unsafe state the simulator can produce
- takeoff-energy scenarios still align with the control model and warning ladder
- partial-panel scenarios match the existing degraded instrument behavior rather than invented symptoms

## Rollout Plan

### Phase 1

- expand realistic single-engine piston IFR scenarios backed by current simulator behavior
- replace generic checklist prose with more realistic abnormal and emergency checklist content
- keep the scenario/checklist recognition workflow intact

### Phase 2

- add IFR-relevant cockpit instruments and clearer annunciation
- strengthen failure presentation inside the scan zone

### Phase 3

- add more pilot controls and matching simulator behaviors
- keep new controls gated by actual simulator state depth

## Risks And Guardrails

### Risk: Content Outruns Simulator Fidelity

If the scenario library grows faster than the backend model, the trainer will look richer than it is. Guardrail: only backend-backed scenarios belong in the core training catalog.

### Risk: UI Complexity Dilutes Scan Training

If too many new instruments or controls arrive at once, the cockpit stops teaching scan discipline and turns into a feature wall. Guardrail: only add instruments that support immediate IFR scan or current failure recognition.

### Risk: Checklist Tone Drifts Into Tutorial Copy

If checklists become explanatory rather than procedural, they stop feeling like real training material. Guardrail: keep checklist entries concise, action-oriented, and aircraft-procedure flavored.

## Success Criteria

This expansion is successful when:

- the trainer clearly reads as a single-engine piston IFR training device
- the scenario library covers a broader set of realistic panel and takeoff events without overstating backend fidelity
- checklist content feels operational rather than generic
- additional instruments and controls are added in a way that improves training realism instead of visual complexity alone
- future aircraft selection remains possible without forcing that abstraction in the first pass