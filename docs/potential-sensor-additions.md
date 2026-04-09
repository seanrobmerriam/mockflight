# Potential Sensor Additions

## Basis For The List

This list combines the current MockFlight sensor gaps with common aircraft instrument categories and a small set of external references reviewed during this pass. In particular:

- angle of attack is directly tied to lift margin and stall behavior
- radar altimeters measure height above ground rather than barometric altitude
- outside air temperature is used in density-altitude, performance, and Mach-related calculations
- GNSS adds position, groundspeed, track, and precise timing inputs for navigation stacks

## Recommended Additions

### 1. Angle of attack sensor

- Why add it: AOA is one of the most useful missing flight-safety signals because it reflects lift margin and stall proximity more directly than raw airspeed.
- What it could output: angle in degrees, normalized stall margin, warning thresholds.
- Simulation value: unlocks approach, short-field, and stall-warning scenarios.
- Priority: high.

### 2. Radar altimeter

- Why add it: the current altimeter is barometric only. A radar altimeter adds true height above terrain, especially useful on approach and flare.
- What it could output: AGL altitude, decision-height flag, flare cue.
- Simulation value: makes landing and terrain-clearance behavior much more realistic.
- Priority: high.

### 3. GNSS receiver

- Why add it: GNSS introduces latitude, longitude, groundspeed, ground track, and accurate time, which opens the door to map views and route simulation.
- What it could output: position, track, groundspeed, UTC time, estimated accuracy.
- Simulation value: enables navigation overlays, flight path traces, and ADS-B style concepts.
- Priority: high.

### 4. Outside air temperature probe

- Why add it: OAT is a core performance input for density altitude, true airspeed estimation, climb performance, and icing logic.
- What it could output: OAT, ISA deviation, derived density altitude input.
- Simulation value: improves environmental realism without much implementation cost.
- Priority: high.

### 5. Fuel flow sensor

- Why add it: fuel flow gives the simulator a direct engine-performance and endurance dimension.
- What it could output: gallons per hour or liters per hour, remaining endurance, fuel used.
- Simulation value: supports range calculations, leaning, and engine anomaly simulation.
- Priority: medium-high.

### 6. Fuel quantity sensor

- Why add it: fuel flow without tank quantity is incomplete for mission simulation.
- What it could output: per-tank quantity, imbalance, low-fuel warnings.
- Simulation value: makes cross-country, diversion, and imbalance scenarios possible.
- Priority: medium-high.

### 7. Engine monitoring sensors

- Why add it: piston and turbine aircraft both rely on engine instrumentation that is absent from the current model.
- What it could output: RPM, manifold pressure, oil temperature, oil pressure, cylinder head temperature, exhaust gas temperature.
- Simulation value: allows engine-page dashboards and failure-mode drills.
- Priority: medium-high.

### 8. Turn coordinator / turn rate sensor

- Why add it: current gyroscope output is raw rotational data. A higher-level instrument view for standard-rate turns is still useful.
- What it could output: turn rate, slip/skid indication, standard-rate marker.
- Simulation value: good for IFR-style instrument panels and pilot training flows.
- Priority: medium.

### 9. Sideslip or yaw-string style sensor

- Why add it: it captures aerodynamic coordination more directly than raw yaw rate.
- What it could output: slip angle, coordination ball displacement, crosswind effect.
- Simulation value: improves turn quality, crosswind landing, and upset-recovery scenarios.
- Priority: medium.

### 10. Weight-on-wheels / squat switch

- Why add it: several aircraft systems change behavior depending on whether the aircraft is airborne.
- What it could output: on-ground state, transition events, touchdown detection.
- Simulation value: useful for landing logic, gear logic, and dashboard state transitions.
- Priority: medium.

### 11. Total air temperature probe

- Why add it: for faster aircraft, total air temperature and outside air temperature can diverge in useful ways.
- What it could output: TAT, computed SAT, ram-rise correction.
- Simulation value: improves air-data realism if MockFlight grows beyond small-aircraft assumptions.
- Priority: medium-low.

### 12. Radio navigation receivers

- Why add it: these are technically receivers more than pure sensors, but they are common instrumentation inputs.
- What it could output: VOR radial, localizer deviation, glideslope deviation, DME distance.
- Simulation value: enables IFR procedure training and richer panel layouts.
- Priority: medium-low.

## Suggested Build Order

If the goal is the highest value with the least implementation risk, the next four should be:

1. Angle of attack
2. Outside air temperature
3. GNSS receiver
4. Radar altimeter

That sequence adds stall margin, environment, navigation, and landing awareness without requiring engine modeling first.

## Reference Notes

- Angle of attack references describe AOA as the direct angle between wing reference and relative airflow, with stall behavior tied to critical AOA rather than a single fixed airspeed.
- Radar altimeter references distinguish AGL measurement from barometric altitude and note its importance for flare, autoland, and terrain-clearance systems.
- Outside air temperature references describe OAT as a standard flight-planning and performance input.
- GPS and GNSS references reinforce the value of accurate position, speed, heading support, and timing as foundational navigation signals.

## Candidate Grouping By Theme

### Air-data expansion

- angle of attack
- outside air temperature
- total air temperature
- sideslip

### Navigation expansion

- GNSS receiver
- radio navigation receivers
- radar altimeter

### Engine and aircraft-state expansion

- fuel flow
- fuel quantity
- engine monitoring pack
- weight-on-wheels