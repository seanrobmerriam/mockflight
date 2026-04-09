# IFR Dashboard Design

## Summary

MockFlight's dashboard will be redesigned from a sensor card grid into an IFR-only training panel. The new interface should feel closer to a real training device than a generic metrics dashboard, with a dominant pilot-facing scan zone, a compact navigation and approach band, and a secondary instructor or systems area for raw sensor and failure controls.

The design targets a glass-lite IFR trainer rather than a vendor replica. It should support two primary use cases on day one:

- instrument scan and approach workflow
- partial-panel and failure training
- failure recognition and checklist selection practice

The simulator does not yet provide full navigation data. The dashboard may therefore use synthetic navigation and approach display values initially, with the expectation that simulator-backed data will replace them later.

## Goals

- Make the dashboard read as an IFR training panel instead of a technical sensor dashboard.
- Center the UI around pilot scan flow: attitude, airspeed, altitude, vertical speed, heading, and lateral guidance.
- Surface failures where they affect the pilot's scan rather than only in a separate status block.
- Support failure-template driven training scenarios tied to abnormal and emergency checklist practice.
- Preserve access to current raw sensor data for debugging and simulator validation.
- Keep the existing HTTP surface stable while allowing the cockpit display model to evolve.

## Non-Goals

- Replicating a specific avionics vendor or certified panel product.
- Building a complete FMS, radio stack, or procedure database in this pass.
- Expanding backend simulator physics as part of the visual redesign.
- Creating a pixel-perfect mobile cockpit replica.
- Creating a full scoring, grading, or instructor analytics system in this pass.

## User Experience Direction

The visual tone should be close to a real IFR trainer cockpit. The page should feel dark, dense, and operational rather than decorative. Color should communicate aviation meaning: sky and ground in the attitude area, green or magenta for nav symbology, amber and red for failures, and otherwise restrained grayscale framing.

The page should present one clear pilot-facing cockpit and one clear instructor-facing support area. The pilot-facing area is where the eye should spend almost all of its time. The instructor-facing support area exists for failures, diagnostics, and raw sensor detail, but it should not compete for attention during normal operation.

The training loop should require recognition, not just action. Scenario templates may inject the aircraft state and cockpit symptoms, but the user must still choose the correct checklist from a dropdown before opening the procedure.

## Layout And Hierarchy

### Primary Flight Display Zone

The primary flight display occupies the center and upper portion of the page. It is the dominant visual element and the core scan zone.

It should include:

- a synthetic horizon with pitch ladder and bank references
- an airspeed tape on the left
- an altitude tape on the right
- a vertical speed presentation adjacent to altitude
- a heading band or compact lower heading presentation
- a small annunciator strip near the top of the PFD

This zone is where the pilot should understand aircraft state first. It should visually suppress supporting data unless the user intentionally looks away from the scan center.

### Navigation And Approach Band

Below the PFD sits a compact navigation band that supports IFR approach workflow.

It should include:

- a compact HSI or CDI-oriented display
- selected course
- nav source
- lateral deviation
- approach state labels

These elements may be synthetic in the first implementation. The design requirement is that they occupy the correct conceptual location in the scan flow so that future simulator-backed data can slot in without a layout rewrite.

### Instructor And Systems Bay

The current raw sensor detail should move to a lower-priority systems bay. This area can include:

- pitot and static details
- accelerometer, gyroscope, and magnetometer readouts
- AHRS numeric detail
- failure injection controls
- scenario launcher controls
- checklist selector and procedure drawer
- simulator state summary

This area should feel like an instructor station or maintenance strip, not the main cockpit panel.

### Responsive Behavior

Desktop should show the full cockpit composition: PFD first, nav band second, systems bay third.

On smaller screens, the layout should stack by training priority:

1. PFD
2. navigation band
3. systems bay

Mobile should remain readable and operational, but it does not need to preserve full desktop spatial fidelity.

## Failure And Partial-Panel Behavior

Failures must present as cockpit consequences, not only as UI toggles or badges.

Each failure should visibly affect the relevant instrument region:

- pitot-related failures should degrade or freeze airspeed behavior
- static-related failures should corrupt or freeze altitude and vertical speed behavior
- magnetometer disturbance should destabilize heading and nav-related cues
- gyro saturation should degrade attitude or rate-driven cues enough to force a partial-panel scan

The interface should use direct in-place effects such as:

- instrument flags
- amber or red failure labels
- hatched or crossed unreliable regions
- frozen or drifting indicators
- annunciator strip warnings for quick confirmation

The lower systems bay should still contain the control surface for toggling failures, but the training value comes from the pilot-facing display degradation in the scan zone.

## Failure Templates And Checklist Workflow

Failure templates should be modeled as linked training scenarios. A scenario injects the relevant simulator failures and cockpit symptoms, but it does not auto-open the matching checklist.

The intended user flow is:

1. launch a scenario template
2. observe the cockpit symptoms and identify the likely failure
3. choose a checklist from a dropdown containing the correct procedure plus plausible distractors
4. open the chosen checklist in a side drawer or kneeboard-style panel
5. continue flying the scenario while working the procedure

Recognition is part of the exercise, so the correct checklist must not be revealed automatically when the scenario starts.

### Scenario Templates

Scenario templates should define:

- scenario name
- active simulator failure configuration
- cockpit symptom expectations
- optional scenario description or training context
- candidate checklist options shown in the dropdown

The initial set can focus on currently supported sensor failures, while leaving room for future engine-related scenarios once the simulator model grows.

### Checklist Library

Checklists should be modeled separately from scenarios. That allows:

- multiple scenarios to map to one checklist
- future aircraft-specific checklist variants
- checklist reuse without changing scenario behavior

Checklist content should read like concise abnormal or emergency procedures rather than tutorial prose.

### Wrong Checklist Behavior

If the user chooses the wrong checklist, the UI should indicate that clearly without ending the scenario or shifting into a game-like scoring flow. The scenario remains active and the user can re-evaluate and choose again.

## Visual Language

The current warm paper-and-glass treatment should be replaced entirely.

The new visual language should use:

- dark panel surfaces and restrained bezels
- high-contrast numeric and tape symbology
- aviation-like geometry instead of dashboard cards
- tight, condensed typography for instrument labels and values
- limited animation focused on instrument motion rather than decorative transitions

Instrument motion should feel stable and mechanical. Tapes should slide, bug markers should move, and deviation indicators should update smoothly, but generic easing and consumer-style animation should be avoided.

## Architecture And Frontend Boundaries

The backend API surface should remain unchanged in this pass:

- `GET /api/snapshot`
- `GET /api/failures`
- `POST /api/failures`
- `GET /healthz`

The frontend should add a display-model translation layer between raw snapshot data and cockpit rendering. That layer converts simulator output into pilot-facing display state such as:

- PFD tape values and trends
- synthetic horizon offsets and pitch ladder positioning
- heading presentation
- nav source and approach labels
- annunciator states
- failure-driven instrument flags
- scenario state and checklist-selection state

Even if the page remains embedded in a single template, the JavaScript should be partitioned into focused rendering units:

- snapshot-to-display-model mapping
- PFD rendering
- navigation band rendering
- annunciator rendering
- systems bay rendering
- failure control synchronization
- scenario launcher rendering
- checklist drawer rendering

This separation is necessary to keep the single page maintainable once synthetic navigation and degraded-mode behavior are introduced.

Scenario definitions and checklist definitions should exist as explicit data structures rather than being embedded ad hoc inside rendering code.

## Data Model Expectations

### Existing Data Used Directly

The redesign should continue to use existing simulator fields for:

- airspeed
- altitude
- vertical speed
- pitch
- roll
- heading
- pitot and static values
- accelerometer, gyroscope, and magnetometer detail
- active failures

### Synthetic Display Data Allowed Initially

The first pass may derive or stub the following training-oriented values in the frontend display model:

- selected course
- CDI deviation
- nav source
- approach state
- bug markers and advisory labels
- scenario definitions and checklist option sets
- checklist content displayed in the procedure drawer

These synthetic values must be visually and structurally isolated so they can be replaced later by simulator-provided fields without reworking the cockpit layout.

## Rollout Plan

### Phase 1: Structural Cockpit Rewrite

- replace the existing card grid with the IFR cockpit layout
- introduce the primary PFD, nav band, and systems bay structure
- remap current simulator data into the new hierarchy
- establish the dark IFR visual language

### Phase 2: Failure-Centric Behavior

- make failures degrade instruments in place
- add annunciator strip behavior
- refine partial-panel presentation
- preserve lower systems bay control over active failures

### Phase 3: Synthetic Navigation Layer

- add compact HSI or CDI presentation
- introduce course, nav source, and approach labels
- support training-oriented approach workflow cues pending simulator support

### Phase 4: Scenario And Checklist Layer

- add failure-template scenario launcher
- add checklist dropdown with plausible distractors
- add checklist drawer or kneeboard panel
- keep the active scenario running while the user identifies and opens the correct procedure

## Error Handling

If snapshot fetches fail or data is incomplete, the cockpit should fail clearly rather than silently degrading into stale values.

Expected behavior:

- show a visible data-loss or simulator-disconnected state in the cockpit
- retain the last known values only if clearly marked stale
- keep failure controls from implying success if the POST update fails
- avoid mixing stale pilot-facing values with fresh systems-bay values

## Testing Strategy

Testing should focus on behavioral correctness and rendering structure rather than pixel snapshots.

Tests should verify:

- the IFR shell renders expected cockpit regions
- snapshot data maps into the correct display zones
- active failures surface in the annunciator area and affected instruments
- failure controls still synchronize with the existing API
- fallback and disconnected states are visible and coherent
- scenario launch injects the expected active failures
- checklist dropdown contains both the correct option and distractors
- wrong checklist selection is clearly indicated without resetting the scenario
- correct checklist selection opens the intended procedure content

The simulator tests do not need major expansion for this design pass unless new backend display fields are added.

## Risks And Tradeoffs

- A near-real cockpit feel increases frontend complexity, especially inside one embedded HTML template.
- Synthetic nav data improves training feel quickly, but care is required so users can distinguish simulated instrument behavior from later true navigation behavior.
- Scenario templates and checklist logic can sprawl unless scenario state, symptoms, and procedure content stay as separate data models.
- Too much visual realism can turn into an accidental avionics clone. The design should stay generic and training-oriented.
- Mobile usability will require prioritization rather than faithful miniaturization of the desktop cockpit.

## Recommendation

Proceed with the IFR trainer panel centered on a dominant PFD and compact navigation band, while keeping raw sensor detail and failure controls in a secondary instructor or systems bay. This gives MockFlight an IFR-first training identity now, preserves the current simulator value, and creates a clean path for future simulator-backed navigation data.