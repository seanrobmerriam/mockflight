package dashboard

import "github.com/mockflight/mockflight/internal/sim"

type ScenarioFidelity string

type ScenarioAdvanceUntil string

const (
	FidelityFullySimulated ScenarioFidelity     = "fully_simulated"
	AdvanceUntilVr         ScenarioAdvanceUntil = "vr"
	AdvanceUntilAirborne   ScenarioAdvanceUntil = "airborne"
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
	Advance  TrainingScenarioAdvance  `json:"advance,omitempty"`
}

type TrainingScenarioControls struct {
	Throttle float64 `json:"throttle,omitempty"`
	TOGA     bool    `json:"toga,omitempty"`
	Rotate   bool    `json:"rotate,omitempty"`
}

type TrainingScenarioAdvance struct {
	Milliseconds int                  `json:"milliseconds,omitempty"`
	Until        ScenarioAdvanceUntil `json:"until,omitempty"`
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
