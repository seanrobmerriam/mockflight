package dashboard

import (
	"strings"
	"testing"

	"github.com/mockflight/mockflight/internal/sim"
)

func TestTrainingCatalogScenariosExposeConsistentPhaseOneDataContract(t *testing.T) {
	for key, scenario := range trainingScenarios {
		if scenario.Key != key {
			t.Fatalf("scenario map key %q does not match scenario key %q", key, scenario.Key)
		}
		if len(strings.TrimSpace(scenario.Title)) < 8 {
			t.Fatalf("scenario %q missing substantive title", key)
		}
		if len(strings.TrimSpace(scenario.Category)) < 3 {
			t.Fatalf("scenario %q missing category", key)
		}
		if len(strings.TrimSpace(scenario.TrainingPhase)) < 3 {
			t.Fatalf("scenario %q missing training phase", key)
		}
		if len(strings.TrimSpace(scenario.Weather)) < 8 {
			t.Fatalf("scenario %q missing weather context", key)
		}
		if len(strings.TrimSpace(scenario.Trigger)) < 12 {
			t.Fatalf("scenario %q missing substantive trigger", key)
		}
		if len(strings.TrimSpace(scenario.Description)) < 12 {
			t.Fatalf("scenario %q missing substantive description", key)
		}
		if len(strings.TrimSpace(scenario.NavSource)) < 3 {
			t.Fatalf("scenario %q missing nav source", key)
		}
		if len(strings.TrimSpace(scenario.ApproachState)) < 3 {
			t.Fatalf("scenario %q missing approach state", key)
		}
		if len(scenario.InitialCues) < 2 {
			t.Fatalf("scenario %q should include at least two initial cues", key)
		}
		for index, cue := range scenario.InitialCues {
			if len(strings.TrimSpace(cue)) < 12 {
				t.Fatalf("scenario %q initial cue %d is too short", key, index)
			}
		}
		if len(scenario.Progression) < 2 {
			t.Fatalf("scenario %q should include at least two progression steps", key)
		}
		for index, step := range scenario.Progression {
			if len(strings.TrimSpace(step)) < 12 {
				t.Fatalf("scenario %q progression step %d is too short", key, index)
			}
		}
		if scenario.Fidelity != FidelityFullySimulated {
			t.Fatalf("scenario %q must remain fully simulated in phase 1, got %q", key, scenario.Fidelity)
		}
		if scenario.CorrectChecklist == "" {
			t.Fatalf("scenario %q missing correct checklist", key)
		}
		if _, ok := trainingChecklists[scenario.CorrectChecklist]; !ok {
			t.Fatalf("scenario %q references unknown checklist %q", key, scenario.CorrectChecklist)
		}
		if len(scenario.ChecklistOptions) < 3 {
			t.Fatalf("scenario %q should present at least 3 checklist options", key)
		}

		found := false
		for _, option := range scenario.ChecklistOptions {
			if option == scenario.CorrectChecklist {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("scenario %q must include correct checklist %q in its options", key, scenario.CorrectChecklist)
		}

		if scenario.Category == "Takeoff Energy" {
			if scenario.Setup.Simulator == nil {
				t.Fatalf("takeoff-energy scenario %q missing simulator setup", key)
			}
			if scenario.Setup.Simulator.Mode == "" {
				t.Fatalf("takeoff-energy scenario %q missing simulator mode", key)
			}
			if len(scenario.Setup.Simulator.Steps) == 0 {
				t.Fatalf("takeoff-energy scenario %q missing simulator control steps", key)
			}

			hasMachineAction := false
			for index, step := range scenario.Setup.Simulator.Steps {
				if step.Controls.Throttle > 0 || step.Controls.TOGA || step.Controls.Rotate {
					hasMachineAction = true
				}
				if step.Advance.Milliseconds < 0 {
					t.Fatalf("takeoff-energy scenario %q step %d has negative advance duration", key, index)
				}
				if step.Advance.Until != "" && step.Advance.Until != AdvanceUntilVr && step.Advance.Until != AdvanceUntilAirborne {
					t.Fatalf("takeoff-energy scenario %q step %d has unsupported advance target %q", key, index, step.Advance.Until)
				}
			}
			if !hasMachineAction {
				t.Fatalf("takeoff-energy scenario %q missing machine-actionable control inputs", key)
			}
		}
	}
}

func TestTrainingCatalogChecklistsExposeConsistentTitlesAndItems(t *testing.T) {
	for key, checklist := range trainingChecklists {
		if checklist.Title != key {
			t.Fatalf("checklist map key %q does not match checklist title %q", key, checklist.Title)
		}
		if len(checklist.Items) == 0 {
			t.Fatalf("checklist %q must include items", key)
		}
		for index, item := range checklist.Items {
			if strings.TrimSpace(item) == "" {
				t.Fatalf("checklist %q item %d must not be empty", key, index)
			}
		}
	}
}

func TestTrainingCatalogIncludesExpandedSingleEngineIFRScenarios(t *testing.T) {
	expected := []string{
		"pitotDeparture",
		"staticApproach",
		"unreliableAirspeedIMC",
		"magnetometerApproach",
		"partialPanelMissed",
		"earlyRotate",
		"lowEnergyClimb",
		"runwayOverrun",
	}

	for _, key := range expected {
		if _, ok := trainingScenarios[key]; !ok {
			t.Fatalf("expected scenario %q in catalog", key)
		}
	}
}

func TestTrainingCatalogTakeoffEnergyScenariosStayBoundToSupportedSimulatorSetup(t *testing.T) {
	expected := map[string]struct {
		mode  sim.Mode
		steps []TrainingScenarioSimulatorStep
	}{
		"earlyRotate": {
			mode: sim.ModeFull,
			steps: []TrainingScenarioSimulatorStep{
				{
					Controls: TrainingScenarioControls{Throttle: 0.65},
					Advance:  TrainingScenarioAdvance{Milliseconds: 3000},
				},
				{
					Controls: TrainingScenarioControls{Rotate: true},
				},
			},
		},
		"lowEnergyClimb": {
			mode: sim.ModeTransitional,
			steps: []TrainingScenarioSimulatorStep{
				{
					Controls: TrainingScenarioControls{Throttle: 0.72},
					Advance:  TrainingScenarioAdvance{Milliseconds: 15000, Until: AdvanceUntilVr},
				},
				{
					Controls: TrainingScenarioControls{Rotate: true},
					Advance:  TrainingScenarioAdvance{Milliseconds: 6000, Until: AdvanceUntilAirborne},
				},
			},
		},
		"runwayOverrun": {
			mode: sim.ModeFull,
			steps: []TrainingScenarioSimulatorStep{
				{
					Controls: TrainingScenarioControls{Throttle: 0.4},
					Advance:  TrainingScenarioAdvance{Milliseconds: 12000},
				},
			},
		},
	}

	for key, want := range expected {
		scenario, ok := trainingScenarios[key]
		if !ok {
			t.Fatalf("expected takeoff-energy scenario %q in catalog", key)
		}

		if scenario.Category != "Takeoff Energy" {
			t.Fatalf("scenario %q must stay in Takeoff Energy category, got %q", key, scenario.Category)
		}
		if scenario.Fidelity != FidelityFullySimulated {
			t.Fatalf("scenario %q must remain fully simulated, got %q", key, scenario.Fidelity)
		}
		if scenario.Setup.Failures != (sim.FailureConfig{}) {
			t.Fatalf("scenario %q should not introduce failure flags, got %#v", key, scenario.Setup.Failures)
		}
		if scenario.Setup.Simulator == nil {
			t.Fatalf("scenario %q must keep a simulator setup", key)
		}
		if scenario.Setup.Simulator.Mode != want.mode {
			t.Fatalf("scenario %q expected simulator mode %q, got %q", key, want.mode, scenario.Setup.Simulator.Mode)
		}
		if len(scenario.Setup.Simulator.Steps) != len(want.steps) {
			t.Fatalf("scenario %q expected %d simulator steps, got %d", key, len(want.steps), len(scenario.Setup.Simulator.Steps))
		}

		for index, wantStep := range want.steps {
			gotStep := scenario.Setup.Simulator.Steps[index]
			if gotStep.Controls != wantStep.Controls {
				t.Fatalf("scenario %q step %d expected controls %#v, got %#v", key, index, wantStep.Controls, gotStep.Controls)
			}
			if gotStep.Advance != wantStep.Advance {
				t.Fatalf("scenario %q step %d expected advance %#v, got %#v", key, index, wantStep.Advance, gotStep.Advance)
			}
		}
	}
}

func TestTrainingCatalogPartialPanelMissedUsesExistingAttitudeHeadingFailures(t *testing.T) {
	scenario, ok := trainingScenarios["partialPanelMissed"]
	if !ok {
		t.Fatal("expected partialPanelMissed in catalog")
	}

	wantFailures := sim.FailureConfig{MagnetometerDisturbed: true, GyroSaturation: true}
	if scenario.Failures != wantFailures {
		t.Fatalf("expected partialPanelMissed failures %#v, got %#v", wantFailures, scenario.Failures)
	}
	if scenario.Setup.Failures != wantFailures {
		t.Fatalf("expected partialPanelMissed setup failures %#v, got %#v", wantFailures, scenario.Setup.Failures)
	}
	if scenario.Setup.Simulator != nil {
		t.Fatal("expected partialPanelMissed to rely on existing failure flags without scripted simulator steps")
	}
}

func TestTrainingCatalogFailureDrivenScenariosStayAlignedToSimulatorFailures(t *testing.T) {
	expected := map[string]sim.FailureConfig{
		"pitotDeparture":        {PitotBlocked: true},
		"staticApproach":        {StaticPortBlocked: true},
		"unreliableAirspeedIMC": {PitotDrainBlocked: true},
		"magnetometerApproach":  {MagnetometerDisturbed: true},
	}

	for key, wantFailures := range expected {
		scenario, ok := trainingScenarios[key]
		if !ok {
			t.Fatalf("expected scenario %q in catalog", key)
		}

		if scenario.Fidelity != FidelityFullySimulated {
			t.Fatalf("scenario %q must remain fully simulated, got %q", key, scenario.Fidelity)
		}
		if scenario.Failures != wantFailures {
			t.Fatalf("scenario %q expected failures %#v, got %#v", key, wantFailures, scenario.Failures)
		}
		if scenario.Setup.Failures != wantFailures {
			t.Fatalf("scenario %q expected setup failures %#v, got %#v", key, wantFailures, scenario.Setup.Failures)
		}
		if scenario.Setup.Simulator != nil {
			t.Fatalf("scenario %q should remain failure-driven without scripted simulator steps", key)
		}
	}
}
