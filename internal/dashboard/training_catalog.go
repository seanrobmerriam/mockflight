package dashboard

import "github.com/mockflight/mockflight/internal/sim"

type ScenarioFidelity string

type ScenarioAdvanceUntil string

const (
	FidelityFullySimulated ScenarioFidelity     = "fully_simulated"
	AdvanceUntilVr         ScenarioAdvanceUntil = "vr"
	AdvanceUntilAirborne   ScenarioAdvanceUntil = "airborne"
	AdvanceUntilAltitude   ScenarioAdvanceUntil = "altitude"
)

type TrainingScenarioSetup struct {
	Failures  sim.FailureConfig               `json:"failures"`
	Simulator *TrainingScenarioSimulatorSetup `json:"simulator,omitempty"`
}

type TrainingScenarioSimulatorSetup struct {
	Mode  sim.Mode                        `json:"mode"`
	Steps []TrainingScenarioSimulatorStep `json:"steps,omitempty"`
}

type TrainingScenarioSimulatorStep struct {
	Controls TrainingScenarioControls `json:"controls"`
	Failures *sim.FailureConfig      `json:"failures,omitempty"`
	Advance  TrainingScenarioAdvance `json:"advance,omitempty"`
}

type TrainingScenarioControls struct {
	Throttle              float64 `json:"throttle,omitempty"`
	TOGA                  bool    `json:"toga,omitempty"`
	Rotate                bool    `json:"rotate,omitempty"`
	ReleaseRotate         bool    `json:"release_rotate,omitempty"`
	AutopilotEngaged      bool    `json:"autopilot_engaged,omitempty"`
	AltitudeHold          bool    `json:"altitude_hold,omitempty"`
	SelectedAltitude      float64 `json:"selected_altitude,omitempty"`
	VerticalSpeedMode     bool    `json:"vertical_speed_mode,omitempty"`
	SelectedVerticalSpeed float64 `json:"selected_vertical_speed,omitempty"`
}

type TrainingScenarioAdvance struct {
	Milliseconds   int                  `json:"milliseconds,omitempty"`
	Until          ScenarioAdvanceUntil `json:"until,omitempty"`
	TargetAltitude float64              `json:"target_altitude,omitempty"`
}

type TrainingScenario struct {
	Key              string                `json:"key"`
	Title            string                `json:"title"`
	Category         string                `json:"category"`
	TrainingPhase    string                `json:"training_phase"`
	Weather          string                `json:"weather"`
	Trigger          string                `json:"trigger"`
	InitialCues      []string              `json:"initial_cues"`
	Progression      []string              `json:"progression"`
	NavSource        string                `json:"nav_source"`
	ApproachState    string                `json:"approach_state"`
	Failures         sim.FailureConfig     `json:"failures"`
	ChecklistOptions []string              `json:"checklist_options"`
	CorrectChecklist string                `json:"correct_checklist"`
	Fidelity         ScenarioFidelity      `json:"fidelity"`
	Description      string                `json:"description"`
	Setup            TrainingScenarioSetup `json:"setup"`
}

type TrainingChecklist struct {
	Title string   `json:"title"`
	Items []string `json:"items"`
}

var trainingScenarios = map[string]TrainingScenario{
	"normalTakeoff": {
		Key:           "normalTakeoff",
		Title:         "Normal Takeoff",
		Category:      "Normal Procedures",
		TrainingPhase: "Takeoff Roll",
		Weather:       "VMC, calm winds",
		Trigger:       "Standard takeoff roll, rotation, and liftoff with no system failures.",
		InitialCues: []string{
			"Airspeed builds normally through V1 and Vr.",
			"All sensor indications remain consistent with pitch, power, and runway track.",
		},
		Progression: []string{
			"Rotation at Vr produces a normal initial climb.",
			"No abnormal indication develops to require a checklist.",
		},
		NavSource:        "RWY HDG",
		ApproachState:    "Takeoff",
		Failures:         sim.FailureConfig{},
		ChecklistOptions: []string{"Normal Operations", "Low Energy Takeoff", "Rejected Takeoff"},
		CorrectChecklist: "Normal Operations",
		Fidelity:         FidelityFullySimulated,
		Description:      "Fly a standard takeoff and rotation with normal energy and no system failures; confirm nothing warrants an abnormal checklist.",
		Setup: TrainingScenarioSetup{
			Failures: sim.FailureConfig{},
			Simulator: &TrainingScenarioSimulatorSetup{
				Mode: sim.ModeFull,
				Steps: []TrainingScenarioSimulatorStep{
					{
						Controls: TrainingScenarioControls{Throttle: 0.95},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilVr, Milliseconds: 20000},
					},
					{
						Controls: TrainingScenarioControls{Rotate: true},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilAirborne, Milliseconds: 6000},
					},
				},
			},
		},
	},
	"normalClimb": {
		Key:           "normalClimb",
		Title:         "Normal Climb",
		Category:      "Normal Procedures",
		TrainingPhase: "Climb",
		Weather:       "VMC, smooth air",
		Trigger:       "Standard climb to an assigned altitude with the autopilot managing the profile.",
		InitialCues: []string{
			"Autopilot altitude hold tracks the selected altitude smoothly.",
			"All sensor indications remain consistent through the climb.",
		},
		Progression: []string{
			"The aircraft levels at the selected altitude with no abnormal indications.",
			"No checklist is required; continue normal cruise flow.",
		},
		NavSource:        "DEP / HDG",
		ApproachState:    "Climb",
		Failures:         sim.FailureConfig{},
		ChecklistOptions: []string{"Normal Operations", "Airspeed Unreliable", "Low Energy Takeoff"},
		CorrectChecklist: "Normal Operations",
		Fidelity:         FidelityFullySimulated,
		Description:      "Fly a standard climb to an assigned altitude under autopilot altitude hold with no system failures.",
		Setup: TrainingScenarioSetup{
			Failures: sim.FailureConfig{},
			Simulator: &TrainingScenarioSimulatorSetup{
				Mode: sim.ModeFull,
				Steps: []TrainingScenarioSimulatorStep{
					{
						Controls: TrainingScenarioControls{Throttle: 1.0},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilVr, Milliseconds: 20000},
					},
					{
						Controls: TrainingScenarioControls{Rotate: true},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilAirborne, Milliseconds: 6000},
					},
					{
						Controls: TrainingScenarioControls{ReleaseRotate: true, AutopilotEngaged: true, VerticalSpeedMode: true, SelectedVerticalSpeed: 300},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilAltitude, TargetAltitude: 5000, Milliseconds: 950000},
					},
				},
			},
		},
	},
	"pitotIcingCruise": {
		Key:           "pitotIcingCruise",
		Title:         "Pitot Icing / Unreliable Airspeed At FL300",
		Category:      "Pitot-Static",
		TrainingPhase: "Cruise",
		Weather:       "Cruise in icing conditions at FL300",
		Trigger:       "Pitot heat fails to prevent ice accretion once established in cruise at 30,000 ft, and indicated airspeed becomes unreliable.",
		InitialCues: []string{
			"Climb to FL300 and level-off are normal under autopilot altitude hold.",
			"Once established in cruise, indicated airspeed stops behaving credibly as pitch and power remain stable.",
		},
		Progression: []string{
			"Chasing the indicated airspeed destabilizes the cruise profile.",
			"Known pitch-power settings and cross-check of altitude and attitude keep the aircraft controllable.",
		},
		NavSource:        "GPS DIRECT",
		ApproachState:    "Cruise",
		Failures:         sim.FailureConfig{},
		ChecklistOptions: []string{"Airspeed Unreliable", "Static System Unreliable", "Partial Panel Attitude Recovery"},
		CorrectChecklist: "Airspeed Unreliable",
		Fidelity:         FidelityFullySimulated,
		Description:      "Climb to FL300, establish cruise, and diagnose a pitot icing event that renders indicated airspeed unreliable.",
		Setup: TrainingScenarioSetup{
			Failures: sim.FailureConfig{},
			Simulator: &TrainingScenarioSimulatorSetup{
				Mode: sim.ModeFull,
				Steps: []TrainingScenarioSimulatorStep{
					{
						Controls: TrainingScenarioControls{Throttle: 1.0},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilVr, Milliseconds: 20000},
					},
					{
						Controls: TrainingScenarioControls{Rotate: true},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilAirborne, Milliseconds: 6000},
					},
					{
						Controls: TrainingScenarioControls{ReleaseRotate: true, AutopilotEngaged: true, VerticalSpeedMode: true, SelectedVerticalSpeed: 300},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilAltitude, TargetAltitude: 30000, Milliseconds: 6200000},
					},
					{
						Failures: &sim.FailureConfig{PitotBlocked: true},
						Advance:  TrainingScenarioAdvance{Milliseconds: 5000},
					},
				},
			},
		},
	},
	"pitotDeparture": {
		Key:           "pitotDeparture",
		Title:         "Pitot Blockage In Climb",
		Category:      "Pitot-Static",
		TrainingPhase: "Departure",
		Weather:       "IMC after departure",
		Trigger:       "Pitot indication becomes unreliable during climb through the clouds.",
		InitialCues: []string{
			"Airspeed indication stops behaving credibly as pitch and power remain stable.",
			"Altimeter and VSI continue to trend normally.",
		},
		Progression: []string{
			"Chasing the indicated airspeed destabilizes climb performance.",
			"Correct pitch-power control keeps the climb predictable.",
		},
		NavSource:        "DEP / HDG",
		ApproachState:    "Initial Climb",
		Failures:         sim.FailureConfig{PitotBlocked: true},
		ChecklistOptions: []string{"Airspeed Unreliable", "Static System Unreliable", "Partial Panel Attitude Recovery"},
		CorrectChecklist: "Airspeed Unreliable",
		Fidelity:         FidelityFullySimulated,
		Description:      "Diagnose unreliable airspeed in IMC and hold the climb with pitch and power discipline.",
		Setup:            TrainingScenarioSetup{Failures: sim.FailureConfig{PitotBlocked: true}},
	},
	"staticApproach": {
		Key:           "staticApproach",
		Title:         "Static Port Blockage On Approach",
		Category:      "Pitot-Static",
		TrainingPhase: "Approach",
		Weather:       "IMC on arrival",
		Trigger:       "Static pressure indications freeze while descending toward the final approach course.",
		InitialCues: []string{
			"Altimeter stops responding as expected.",
			"VSI trends become suspect or settle toward zero.",
		},
		Progression: []string{
			"Chasing frozen altitude information destabilizes the approach.",
			"Pitch-power references and missed-approach discipline keep the aircraft controllable.",
		},
		NavSource:        "LOC 24",
		ApproachState:    "Intermediate Approach",
		Failures:         sim.FailureConfig{StaticPortBlocked: true},
		ChecklistOptions: []string{"Static System Unreliable", "Airspeed Unreliable", "Missed Approach Stabilization"},
		CorrectChecklist: "Static System Unreliable",
		Fidelity:         FidelityFullySimulated,
		Description:      "Recognize frozen static-system indications and fly the approach by cross-check rather than fixation.",
		Setup:            TrainingScenarioSetup{Failures: sim.FailureConfig{StaticPortBlocked: true}},
	},
	"unreliableAirspeedIMC": {
		Key:           "unreliableAirspeedIMC",
		Title:         "Unreliable Airspeed In IMC",
		Category:      "Pitot-Static",
		TrainingPhase: "Arrival",
		Weather:       "Solid IMC with light turbulence",
		Trigger:       "Pitot drain blockage causes materially degraded indicated airspeed during instrument flight.",
		InitialCues: []string{
			"Indicated airspeed is inconsistent with pitch attitude and power.",
			"Altitude and heading remain usable enough for cross-checking.",
		},
		Progression: []string{
			"Following the bad airspeed number produces low-energy handling.",
			"Known pitch-power settings preserve a safe IFR profile.",
		},
		NavSource:        "GPS LPV",
		ApproachState:    "Vectors To Final",
		Failures:         sim.FailureConfig{PitotDrainBlocked: true},
		ChecklistOptions: []string{"Airspeed Unreliable", "Heading Reference Unreliable", "Partial Panel Attitude Recovery"},
		CorrectChecklist: "Airspeed Unreliable",
		Fidelity:         FidelityFullySimulated,
		Description:      "Hold an instrument profile with unreliable IAS and avoid chasing a degraded airspeed indication.",
		Setup:            TrainingScenarioSetup{Failures: sim.FailureConfig{PitotDrainBlocked: true}},
	},
	"magnetometerApproach": {
		Key:           "magnetometerApproach",
		Title:         "Magnetometer Disturbance On Final",
		Category:      "Attitude / Heading",
		TrainingPhase: "Approach",
		Weather:       "Night IMC on vectors",
		Trigger:       "Heading reference becomes unreliable during final intercept.",
		InitialCues: []string{
			"Displayed heading no longer agrees with turn behavior and course guidance.",
			"Lateral tracking becomes inconsistent near final approach.",
		},
		Progression: []string{
			"Over-controlling with bad heading information increases localizer instability.",
			"Supporting instruments and workload reduction restore usable control.",
		},
		NavSource:        "ILS 31",
		ApproachState:    "Final Intercept",
		Failures:         sim.FailureConfig{MagnetometerDisturbed: true},
		ChecklistOptions: []string{"Heading Reference Unreliable", "Missed Approach Stabilization", "Airspeed Unreliable"},
		CorrectChecklist: "Heading Reference Unreliable",
		Fidelity:         FidelityFullySimulated,
		Description:      "Manage a destabilized final intercept with degraded heading information and disciplined cross-checking.",
		Setup:            TrainingScenarioSetup{Failures: sim.FailureConfig{MagnetometerDisturbed: true}},
	},
	"partialPanelMissed": {
		Key:           "partialPanelMissed",
		Title:         "Partial Panel Missed Approach",
		Category:      "Attitude / Heading",
		TrainingPhase: "Missed Approach",
		Weather:       "Low IMC at DA",
		Trigger:       "Gyro saturation and heading disturbance appear during the missed-approach transition.",
		InitialCues: []string{
			"Attitude and heading cues become unreliable just as workload increases.",
			"Supporting instruments remain available for partial-panel control.",
		},
		Progression: []string{
			"Attempting to keep using the failed references compounds disorientation.",
			"A prompt partial-panel scan supports a stabilized missed approach.",
		},
		NavSource:        "VOR-A",
		ApproachState:    "Missed Approach",
		Failures:         sim.FailureConfig{MagnetometerDisturbed: true, GyroSaturation: true},
		ChecklistOptions: []string{"Partial Panel Attitude Recovery", "Heading Reference Unreliable", "Missed Approach Stabilization"},
		CorrectChecklist: "Partial Panel Attitude Recovery",
		Fidelity:         FidelityFullySimulated,
		Description:      "Transition to a partial-panel missed approach without fixating on unreliable primary references.",
		Setup:            TrainingScenarioSetup{Failures: sim.FailureConfig{MagnetometerDisturbed: true, GyroSaturation: true}},
	},
	"earlyRotate": {
		Key:           "earlyRotate",
		Title:         "Early Rotate Below Vr",
		Category:      "Takeoff Energy",
		TrainingPhase: "Takeoff Roll",
		Weather:       "Dry runway, day IMC departure",
		Trigger:       "Pilot commands rotation before Vr with inadequate energy.",
		InitialCues: []string{
			"Below-Vr warning appears during the takeoff roll.",
			"Aircraft remains on the runway despite pitch-up command.",
		},
		Progression: []string{
			"Holding premature rotation degrades acceleration and margin.",
			"Prompt reject-or-stabilize judgment prevents escalation.",
		},
		NavSource:        "RWY HDG",
		ApproachState:    "Takeoff",
		Failures:         sim.FailureConfig{},
		ChecklistOptions: []string{"Low Energy Takeoff", "Rejected Takeoff", "Missed Approach Stabilization"},
		CorrectChecklist: "Low Energy Takeoff",
		Fidelity:         FidelityFullySimulated,
		Description:      "Recognize an unsafe rotate command below Vr and avoid turning it into a loss-of-control event.",
		Setup: TrainingScenarioSetup{
			Failures: sim.FailureConfig{},
			Simulator: &TrainingScenarioSimulatorSetup{
				Mode: sim.ModeFull,
				Steps: []TrainingScenarioSimulatorStep{
					{
						Controls: TrainingScenarioControls{Throttle: 0.65},
						Advance:  TrainingScenarioAdvance{Milliseconds: 3000},
					},
					{
						Controls: TrainingScenarioControls{Rotate: true},
					},
				},
			},
		},
	},
	"lowEnergyClimb": {
		Key:           "lowEnergyClimb",
		Title:         "Low Energy Initial Climb",
		Category:      "Takeoff Energy",
		TrainingPhase: "Initial Climb",
		Weather:       "Low ceiling after departure",
		Trigger:       "Aircraft rotates and lifts off with inadequate energy margin.",
		InitialCues: []string{
			"Low-energy warning develops after liftoff.",
			"Pitch attitude and climb performance disagree with expected V2 profile.",
		},
		Progression: []string{
			"Excessive pitch leads toward sink-rate and stall cues.",
			"Lowering pitch and restoring energy stabilizes the departure.",
		},
		NavSource:        "DEP / HDG",
		ApproachState:    "Initial Climb",
		Failures:         sim.FailureConfig{},
		ChecklistOptions: []string{"Low Energy Takeoff", "Airspeed Unreliable", "Rejected Takeoff"},
		CorrectChecklist: "Low Energy Takeoff",
		Fidelity:         FidelityFullySimulated,
		Description:      "Recover from an over-pitched, low-energy climb using immediate takeoff performance discipline.",
		Setup: TrainingScenarioSetup{
			Failures: sim.FailureConfig{},
			Simulator: &TrainingScenarioSimulatorSetup{
				Mode: sim.ModeTransitional,
				Steps: []TrainingScenarioSimulatorStep{
					{
						Controls: TrainingScenarioControls{Throttle: 0.72},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilVr, Milliseconds: 15000},
					},
					{
						Controls: TrainingScenarioControls{Rotate: true},
						Advance:  TrainingScenarioAdvance{Until: AdvanceUntilAirborne, Milliseconds: 6000},
					},
				},
			},
		},
	},
	"runwayOverrun": {
		Key:           "runwayOverrun",
		Title:         "Rejected Takeoff And Runway Overrun Risk",
		Category:      "Takeoff Energy",
		TrainingPhase: "Takeoff Roll",
		Weather:       "Reduced visibility on departure",
		Trigger:       "Acceleration is inadequate and the takeoff becomes unrecoverable if not rejected promptly.",
		InitialCues: []string{
			"Acceleration is weak relative to runway remaining.",
			"Warning state escalates if the pilot continues the takeoff attempt.",
		},
		Progression: []string{
			"Continuing the takeoff can end in crash state at runway end.",
			"A timely reject preserves runway margin and aircraft control.",
		},
		NavSource:        "RWY HDG",
		ApproachState:    "Takeoff",
		Failures:         sim.FailureConfig{},
		ChecklistOptions: []string{"Rejected Takeoff", "Low Energy Takeoff", "Missed Approach Stabilization"},
		CorrectChecklist: "Rejected Takeoff",
		Fidelity:         FidelityFullySimulated,
		Description:      "Use runway remaining and acceleration cues to decide for a reject before the event becomes a crash sequence.",
		Setup: TrainingScenarioSetup{
			Failures: sim.FailureConfig{},
			Simulator: &TrainingScenarioSimulatorSetup{
				Mode: sim.ModeFull,
				Steps: []TrainingScenarioSimulatorStep{
					{
						Controls: TrainingScenarioControls{Throttle: 0.4},
						Advance:  TrainingScenarioAdvance{Milliseconds: 12000},
					},
				},
			},
		},
	},
}

var trainingChecklists = map[string]TrainingChecklist{
	"Normal Operations": {
		Title: "Normal Operations",
		Items: []string{
			"INSTRUMENTS ........................ CROSS-CHECKED, NO DISCREPANCIES",
			"FLIGHT PATH ......................... CONTINUE AS BRIEFED",
			"ABNORMAL PROCEDURES ................. NOT REQUIRED",
			"MONITORING .......................... MAINTAIN NORMAL SCAN",
		},
	},
	"Airspeed Unreliable": {
		Title: "Airspeed Unreliable",
		Items: []string{
			"PITCH / POWER ...................... SET KNOWN IFR VALUES",
			"PITOT HEAT ........................ ON",
			"CONFIGURATION ..................... MINIMIZE CHANGES UNTIL STABLE",
			"APPROACH .......................... CONTINUE ONLY IF PERFORMANCE IS PREDICTABLE",
		},
	},
	"Static System Unreliable": {
		Title: "Static System Unreliable",
		Items: []string{
			"ALTIMETER / VSI ................... TREAT AS UNRELIABLE",
			"PITCH / POWER ..................... SET KNOWN IFR VALUES",
			"APPROACH .......................... GO MISSED IF PROFILE CANNOT BE VERIFIED",
			"WORKLOAD .......................... REDUCE AND REQUEST HELP IF NEEDED",
		},
	},
	"Heading Reference Unreliable": {
		Title: "Heading Reference Unreliable",
		Items: []string{
			"SUPPORTING INSTRUMENTS ............ CROSS-CHECK IMMEDIATELY",
			"MODE CHANGES ...................... MINIMIZE",
			"TRACK ............................. USE RELIABLE NAV AND TURN PERFORMANCE CUES",
			"APPROACH .......................... DISCONTINUE IF DIRECTIONAL CONTROL IS NOT STABLE",
		},
	},
	"Partial Panel Attitude Recovery": {
		Title: "Partial Panel Attitude Recovery",
		Items: []string{
			"SUPPORTING INSTRUMENTS ............ TRANSITION IMMEDIATELY",
			"PITCH / BANK ...................... RE-ESTABLISH KNOWN IFR ATTITUDE",
			"POWER ............................. SET FOR STABLE FLIGHT",
			"WORKLOAD .......................... REQUEST VECTORS OR LOWER-WORKLOAD OPTION",
		},
	},
	"Low Energy Takeoff": {
		Title: "Low Energy Takeoff",
		Items: []string{
			"PITCH ............................. REDUCE TO RECOVER ENERGY",
			"THRUST ............................ VERIFY MAX AVAILABLE",
			"V2 ................................ RE-ESTABLISH OR DISCONTINUE CLIMB",
			"FLIGHT PATH ....................... PRIORITIZE CONTROL OVER PROCEDURE FLOW",
		},
	},
	"Rejected Takeoff": {
		Title: "Rejected Takeoff",
		Items: []string{
			"THROTTLE .......................... IDLE",
			"DIRECTIONAL CONTROL ............... MAINTAIN RUNWAY CENTERLINE",
			"ROTATE INPUT ...................... RELEASE",
			"STOPPING DECISION ................. CONTINUE BRAKING TO A SAFE STOP",
		},
	},
	"Missed Approach Stabilization": {
		Title: "Missed Approach Stabilization",
		Items: []string{
			"POWER ............................. SET GO-AROUND POWER",
			"PITCH ............................. ESTABLISH MISSED-APPROACH ATTITUDE",
			"CONFIGURATION ..................... CLEAN UP IN STAGES",
			"NAVIGATION ........................ TRACK PUBLISHED MISSED APPROACH",
		},
	},
}
