package dashboard

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mockflight/mockflight/internal/sim"
)

type stubProvider struct {
	snapshot sim.Snapshot
	failures sim.FailureConfig
	controls sim.ControlState
}

type snapshotOnlyProvider struct {
	snapshot sim.Snapshot
}

func (s snapshotOnlyProvider) Snapshot() sim.Snapshot {
	return s.snapshot
}

func (s stubProvider) Snapshot() sim.Snapshot {
	return s.snapshot
}

func (s *stubProvider) SetFailures(failures sim.FailureConfig) {
	s.failures = failures
}

func (s *stubProvider) Failures() sim.FailureConfig {
	return s.failures
}

func (s *stubProvider) Controls() sim.ControlState {
	return s.controls
}

func (s *stubProvider) ApplyControls(command sim.ControlCommand) sim.ControlState {
	if command.Mode != nil {
		s.controls.Mode = *command.Mode
	}
	if command.Throttle != nil {
		s.controls.Throttle = *command.Throttle
	}
	if command.TOGA != nil {
		s.controls.TOGA = *command.TOGA
	}
	if command.Rotate != nil {
		s.controls.RotateCommanded = *command.Rotate
	}
	if command.AutopilotEngaged != nil {
		s.controls.Autopilot.Engaged = *command.AutopilotEngaged
	}
	if command.HeadingHold != nil {
		s.controls.Autopilot.HeadingHold = *command.HeadingHold
	}
	if command.AltitudeHold != nil {
		s.controls.Autopilot.AltitudeHold = *command.AltitudeHold
	}
	if command.VerticalSpeedMode != nil {
		s.controls.Autopilot.VerticalSpeedMode = *command.VerticalSpeedMode
	}
	if command.SelectedHeading != nil {
		s.controls.SelectedHeading = *command.SelectedHeading
	}
	if command.SelectedAltitude != nil {
		s.controls.SelectedAltitude = *command.SelectedAltitude
	}
	if command.SelectedVerticalSpeed != nil {
		s.controls.SelectedVerticalSpeed = *command.SelectedVerticalSpeed
	}
	return s.controls
}

func TestAPIEndpointReturnsSnapshotJSON(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{
		Phase:          "cruise",
		Timestamp:      time.Unix(1712400000, 0).UTC(),
		ActiveFailures: []string{"pitot_blocked"},
		Altimeter:      sim.ScalarReading{Name: "altimeter", Value: 5120, Unit: "ft"},
		StaticAir:      sim.ScalarReading{Name: "static_air", Value: 24.91, Unit: "inHg"},
		Pitot:          sim.ScalarReading{Name: "pitot", Value: 25.13, Unit: "inHg"},
		VerticalSpeed:  sim.ScalarReading{Name: "vertical_speed", Value: 0.1, Unit: "m/s"},
		Airspeed:       sim.AirspeedReading{Name: "airspeed", Indicated: 129.4, Unit: "kts", PitotPressure: 25.13, StaticPressure: 24.91},
		Accelerometer:  sim.VectorReading{Name: "accelerometer", X: -0.1, Y: 0.0, Z: -9.8, Unit: "m/s^2"},
		Gyroscope:      sim.VectorReading{Name: "gyroscope", X: 0.0, Y: 0.1, Z: 0.2, Unit: "deg/s"},
		Magnetometer:   sim.VectorReading{Name: "magnetometer", X: 18, Y: -2, Z: 42, Unit: "uT"},
		AHRS:           sim.AttitudeReading{Name: "ahrs", Roll: 1.1, Pitch: 2.2, Yaw: 270.3, Heading: 270.3, Unit: "deg"},
	}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/snapshot", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var snapshot sim.Snapshot
	if err := json.NewDecoder(recorder.Body).Decode(&snapshot); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if snapshot.Phase != "cruise" {
		t.Fatalf("expected cruise phase, got %q", snapshot.Phase)
	}

	if snapshot.Airspeed.Indicated != 129.4 {
		t.Fatalf("expected airspeed 129.4, got %.1f", snapshot.Airspeed.Indicated)
	}

	if len(snapshot.ActiveFailures) != 1 || snapshot.ActiveFailures[0] != "pitot_blocked" {
		t.Fatalf("expected active failures to round-trip, got %#v", snapshot.ActiveFailures)
	}
}

func TestDashboardPageRendersIFRTrainingShell(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	checks := []string{
		"MockFlight IFR Trainer",
		"Primary Flight Display",
		"Approach Guidance",
		"Training Scenarios",
		"Emergency Checklists",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected dashboard shell to contain %q", check)
		}
	}

	if strings.Contains(body, "MockFlight Sensor Deck") {
		t.Fatalf("expected legacy dashboard title to be removed: %s", body)
	}
}

func TestDashboardPageIncludesScenarioChecklistControls(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	checks := []string{
		"scenarioSelect",
		"checklistSelect",
		"checklistDrawer",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected dashboard shell to contain %q", check)
		}
	}
}

func TestDashboardPageIncludesScenarioOptions(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	checks := []string{
		"Pitot Blockage In Climb",
		"Rejected Takeoff And Runway Overrun Risk",
		`"category":"Takeoff Energy"`,
		`"training_phase":"Missed Approach"`,
		`"weather":"Night IMC on vectors"`,
		`"fidelity":"fully_simulated"`,
		`"title":"Static System Unreliable"`,
		"PITCH / POWER ...................... SET KNOWN IFR VALUES",
		"THROTTLE .......................... IDLE",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected scenario shell to contain %q", check)
		}
	}

	legacyChecks := []string{
		"Pitot Blockage on Approach",
		"Altimeter / Static System Failure",
		"Pitch and power: set known values for the current phase of flight.",
	}

	for _, check := range legacyChecks {
		if strings.Contains(body, check) {
			t.Fatalf("expected legacy scenario content to be removed: %q", check)
		}
	}
}

func TestDashboardPageIncludesScenarioMetadataShell(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	checks := []string{
		"scenarioCategory",
		"scenarioTrainingPhase",
		"scenarioWeather",
		"scenarioFidelity",
		`"title":"Low Energy Takeoff","items":["PITCH ............................. REDUCE TO RECOVER ENERGY"`,
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected metadata shell to contain %q", check)
		}
	}
}

func TestDashboardPageEmbedsTrainingCatalogJSONShape(t *testing.T) {
	body := renderDashboardPage(t, stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	scenarios := extractEmbeddedJSONMap(t, body, "const scenarios = ", ";\n\n    const checklists =")
	checklists := extractEmbeddedJSONMap(t, body, "const checklists = ", ";\n\n    const state =")

	lowEnergyClimb := requireMapValue(t, scenarios, "lowEnergyClimb")
	requireStringValue(t, lowEnergyClimb, "title", "Low Energy Initial Climb")
	requireStringValue(t, lowEnergyClimb, "category", "Takeoff Energy")
	requireStringValue(t, lowEnergyClimb, "training_phase", "Initial Climb")
	requireStringValue(t, lowEnergyClimb, "weather", "Low ceiling after departure")
	requireStringValue(t, lowEnergyClimb, "nav_source", "DEP / HDG")
	requireStringValue(t, lowEnergyClimb, "approach_state", "Initial Climb")
	requireStringValue(t, lowEnergyClimb, "correct_checklist", "Low Energy Takeoff")
	requireStringValue(t, lowEnergyClimb, "fidelity", "fully_simulated")

	checklistOptions := requireSliceValue(t, lowEnergyClimb, "checklist_options")
	if len(checklistOptions) != 3 {
		t.Fatalf("expected three checklist options, got %d", len(checklistOptions))
	}
	if option, ok := checklistOptions[0].(string); !ok || option != "Low Energy Takeoff" {
		t.Fatalf("expected first checklist option to be Low Energy Takeoff, got %#v", checklistOptions[0])
	}

	setup := requireMapValue(t, lowEnergyClimb, "setup")
	failures := requireMapValue(t, setup, "failures")
	if len(failures) == 0 {
		t.Fatal("expected setup.failures object to be embedded even when all values are false")
	}
	simulatorSetup := requireMapValue(t, setup, "simulator")
	requireStringValue(t, simulatorSetup, "mode", string(sim.ModeTransitional))
	steps := requireSliceValue(t, simulatorSetup, "steps")
	if len(steps) != 2 {
		t.Fatalf("expected two simulator setup steps, got %d", len(steps))
	}
	firstStep := requireMap(t, steps[0], "lowEnergyClimb.setup.simulator.steps[0]")
	firstControls := requireMapValue(t, firstStep, "controls")
	requireNumberValue(t, firstControls, "throttle", 0.72)
	firstAdvance := requireMapValue(t, firstStep, "advance")
	requireStringValue(t, firstAdvance, "until", "vr")
	requireNumberValue(t, firstAdvance, "milliseconds", 15000)
	secondStep := requireMap(t, steps[1], "lowEnergyClimb.setup.simulator.steps[1]")
	secondControls := requireMapValue(t, secondStep, "controls")
	requireBoolValue(t, secondControls, "rotate", true)
	secondAdvance := requireMapValue(t, secondStep, "advance")
	requireStringValue(t, secondAdvance, "until", "airborne")
	requireNumberValue(t, secondAdvance, "milliseconds", 6000)

	pitotDeparture := requireMapValue(t, scenarios, "pitotDeparture")
	pitotSetup := requireMapValue(t, pitotDeparture, "setup")
	if _, ok := pitotSetup["simulator"]; ok {
		t.Fatal("expected pitotDeparture setup to omit simulator block")
	}
	pitotFailures := requireMapValue(t, pitotSetup, "failures")
	requireBoolValue(t, pitotFailures, "pitot_blocked", true)

	rejectedTakeoff := requireMapValue(t, checklists, "Rejected Takeoff")
	requireStringValue(t, rejectedTakeoff, "title", "Rejected Takeoff")
	items := requireSliceValue(t, rejectedTakeoff, "items")
	if len(items) != 4 {
		t.Fatalf("expected four rejected takeoff checklist items, got %d", len(items))
	}
	if item, ok := items[0].(string); !ok || item != "THROTTLE .......................... IDLE" {
		t.Fatalf("expected first rejected takeoff item to round-trip, got %#v", items[0])
	}
}

func TestDashboardPageModeChangePostsModeBeforeReset(t *testing.T) {
	body := renderDashboardPage(t, stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	if !strings.Contains(body, "await postControls({ mode: event.target.value });") {
		t.Fatal("expected mode selector to send mode update before reset")
	}

	if !strings.Contains(body, "await postControls({ reset: true });") {
		t.Fatal("expected mode selector to send a follow-up reset request")
	}

	if strings.Contains(body, "await postControls({ mode: event.target.value, reset: true });") {
		t.Fatal("expected combined mode/reset payload to be removed from mode selector handler")
	}
}

func TestDashboardPageIncludesAuralWarningController(t *testing.T) {
	body := renderDashboardPage(t, stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	checks := []string{
		"function unlockAuralAlerts()",
		"function syncAuralAlerts(warning)",
		"SpeechSynthesisUtterance",
		"SINK RATE",
		"PULL UP",
		"sink_rate",
		"pull_up",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected dashboard shell to contain %q", check)
		}
	}

	if !strings.Contains(body, "document.addEventListener('pointerdown', unlockAuralAlerts, { once: true })") {
		t.Fatal("expected aural alerts to unlock on first user interaction")
	}
}

func TestDashboardPageIncludesAuralAlertMuteToggle(t *testing.T) {
	body := renderDashboardPage(t, stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	checks := []string{
		`id="auralAlertsEnabled"`,
		"Aural Alerts",
		"function setAuralAlertsEnabled(enabled)",
		"function hydrateAuralAlertControls()",
		"mockflight.auralAlertsEnabled",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected dashboard shell to contain %q", check)
		}
	}
}

func TestDashboardPageIncludesAuralAlertStatusIndicator(t *testing.T) {
	body := renderDashboardPage(t, stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	checks := []string{
		`id="auralAlertStatus"`,
		"Aural Status",
		"Aural alerts armed",
		"function updateAuralAlertStatus()",
		"aural-status",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected dashboard shell to contain %q", check)
		}
	}
}

func TestFailureEndpointUpdatesProviderState(t *testing.T) {
	provider := &stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}}
	handler := NewHandler(provider)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/failures", strings.NewReader(`{"pitot_blocked":true,"gyro_saturation":true}`))
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !provider.failures.PitotBlocked || !provider.failures.GyroSaturation {
		t.Fatalf("expected provider failures to update, got %#v", provider.failures)
	}
}

func TestFailureEndpointRejectsMalformedPayload(t *testing.T) {
	handler := NewHandler(&stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/failures", strings.NewReader(`{"pitot_blocked":`))
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "invalid failure payload") {
		t.Fatalf("expected invalid failure payload response, got %q", recorder.Body.String())
	}
}

func TestScenarioEndpointAppliesFailureDrivenScenario(t *testing.T) {
	provider := &stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}}
	handler := NewHandler(provider)

	applyScenarioSelection(t, handler, `{"key":"pitotDeparture"}`)

	if !provider.failures.PitotBlocked {
		t.Fatalf("expected pitotDeparture to enable pitot_blocked, got %#v", provider.failures)
	}

	if provider.failures.StaticPortBlocked || provider.failures.PitotDrainBlocked || provider.failures.MagnetometerDisturbed || provider.failures.GyroSaturation || provider.failures.StaticLeak {
		t.Fatalf("expected only pitot_blocked failure to be active, got %#v", provider.failures)
	}
}

func TestScenarioEndpointAppliesTakeoffEnergySetup(t *testing.T) {
	simulator := sim.NewSimulator(sim.Config{Seed: 7, NoiseScale: 0})
	handler := NewHandler(simulator)

	applyScenarioSelection(t, handler, `{"key":"lowEnergyClimb"}`)

	snapshot := snapshotFromAPI(t, handler)
	if snapshot.Controls.Mode != sim.ModeTransitional {
		t.Fatalf("expected lowEnergyClimb to select transitional mode, got %q", snapshot.Controls.Mode)
	}

	if snapshot.Controls.Throttle != 0.72 {
		t.Fatalf("expected lowEnergyClimb to apply throttle 0.72, got %.2f", snapshot.Controls.Throttle)
	}

	if !snapshot.Controls.RotateCommanded {
		t.Fatal("expected lowEnergyClimb to command rotation")
	}

	if snapshot.Controls.RunwayDistance <= 0 {
		t.Fatalf("expected lowEnergyClimb to advance down the runway, got %.1f ft", snapshot.Controls.RunwayDistance)
	}

	if !snapshot.Controls.Airborne && snapshot.Controls.Warning != sim.WarningBelowVrRotate {
		t.Fatalf("expected lowEnergyClimb setup to produce a concrete takeoff-energy state, got %#v", snapshot.Controls)
	}
}

func TestScenarioEndpointTakeoffEnergySetupIgnoresPriorPitotStaticFailures(t *testing.T) {
	cleanSimulator := sim.NewSimulator(sim.Config{Seed: 7, NoiseScale: 0})
	cleanHandler := NewHandler(cleanSimulator)
	applyScenarioSelection(t, cleanHandler, `{"key":"lowEnergyClimb"}`)
	expected := snapshotFromAPI(t, cleanHandler)

	dirtySimulator := sim.NewSimulator(sim.Config{
		Seed:       7,
		NoiseScale: 0,
		Failures: sim.FailureConfig{
			PitotBlocked:      true,
			StaticPortBlocked: true,
		},
	})
	dirtyHandler := NewHandler(dirtySimulator)
	applyScenarioSelection(t, dirtyHandler, `{"key":"lowEnergyClimb"}`)
	actual := snapshotFromAPI(t, dirtyHandler)

	if dirtySimulator.Failures() != (sim.FailureConfig{}) {
		t.Fatalf("expected lowEnergyClimb to clear prior failures before launch, got %#v", dirtySimulator.Failures())
	}

	if actual.Phase != expected.Phase {
		t.Fatalf("expected stale failures to not change phase, clean %q dirty %q", expected.Phase, actual.Phase)
	}

	if actual.Controls.Airborne != expected.Controls.Airborne {
		t.Fatalf("expected stale failures to not change airborne state, clean %#v dirty %#v", expected.Controls, actual.Controls)
	}

	if actual.Controls.Warning != expected.Controls.Warning {
		t.Fatalf("expected stale failures to not change warning state, clean %q dirty %q", expected.Controls.Warning, actual.Controls.Warning)
	}

	if diff := math.Abs(actual.Controls.RunwayDistance - expected.Controls.RunwayDistance); diff > 0.5 {
		t.Fatalf("expected stale failures to not change runway distance, clean %.1f dirty %.1f", expected.Controls.RunwayDistance, actual.Controls.RunwayDistance)
	}

	if diff := math.Abs(actual.Airspeed.Indicated - expected.Airspeed.Indicated); diff > 0.2 {
		t.Fatalf("expected stale failures to not change indicated airspeed, clean %.1f dirty %.1f", expected.Airspeed.Indicated, actual.Airspeed.Indicated)
	}
}

func TestScenarioEndpointClearingToManualClearsScenarioFailures(t *testing.T) {
	provider := &stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}}
	handler := NewHandler(provider)

	applyScenarioSelection(t, handler, `{"key":"partialPanelMissed"}`)
	if !provider.failures.MagnetometerDisturbed || !provider.failures.GyroSaturation {
		t.Fatalf("expected scenario failures to be active before clearing, got %#v", provider.failures)
	}

	applyScenarioSelection(t, handler, `{"key":""}`)

	if provider.failures != (sim.FailureConfig{}) {
		t.Fatalf("expected manual scenario clear to remove scenario-applied failures, got %#v", provider.failures)
	}
}

func TestControlsEndpointUpdatesSimulatorState(t *testing.T) {
	provider := &stubProvider{snapshot: sim.Snapshot{Phase: "runway_idle", Timestamp: time.Now().UTC()}}
	handler := NewHandler(provider)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/controls", strings.NewReader(`{"throttle":0.92,"rotate":true,"autopilot_engaged":true,"heading_hold":true,"selected_heading":315}`))
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if provider.controls.Throttle != 0.92 {
		t.Fatalf("expected throttle update, got %.2f", provider.controls.Throttle)
	}

	if !provider.controls.RotateCommanded {
		t.Fatal("expected rotate command to update")
	}

	if !provider.controls.Autopilot.Engaged || !provider.controls.Autopilot.HeadingHold {
		t.Fatalf("expected autopilot state to update, got %#v", provider.controls.Autopilot)
	}

	if provider.controls.SelectedHeading != 315 {
		t.Fatalf("expected selected heading to update, got %.1f", provider.controls.SelectedHeading)
	}
}

func TestControlsEndpointRejectsMalformedPayload(t *testing.T) {
	handler := NewHandler(&stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/controls", strings.NewReader(`{"mode":`))
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "invalid controls payload") {
		t.Fatalf("expected invalid controls payload response, got %q", recorder.Body.String())
	}
}

func TestControlEndpointsMethodNotAllowed(t *testing.T) {
	handler := NewHandler(&stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	tests := []struct {
		name string
		path string
	}{
		{name: "failures", path: "/api/failures"},
		{name: "controls", path: "/api/controls"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPut, tt.path, nil)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", recorder.Code)
			}

			if allow := recorder.Header().Get("Allow"); allow != "GET, POST" {
				t.Fatalf("expected Allow header GET, POST, got %q", allow)
			}

			if !strings.Contains(recorder.Body.String(), "method not allowed") {
				t.Fatalf("expected method not allowed response, got %q", recorder.Body.String())
			}
		})
	}
}

func TestControlEndpointsReportUnsupportedProvider(t *testing.T) {
	handler := NewHandler(snapshotOnlyProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	tests := []struct {
		name        string
		path        string
		messagePart string
	}{
		{name: "failures", path: "/api/failures", messagePart: "failure controls unavailable"},
		{name: "controls", path: "/api/controls", messagePart: "simulator controls unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)

			handler.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNotImplemented {
				t.Fatalf("expected 501, got %d", recorder.Code)
			}

			if !strings.Contains(recorder.Body.String(), tt.messagePart) {
				t.Fatalf("expected response to contain %q, got %q", tt.messagePart, recorder.Body.String())
			}
		})
	}
}

func TestDashboardPageIncludesRunwayControlShell(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{Phase: "runway_idle", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	checks := []string{
		"Throttle",
		"TO/GA",
		"Rotate",
		"Autopilot",
		"Heading Select",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected runway control shell to contain %q", check)
		}
	}
}

func TestDashboardControlsCrashAndResetFlow(t *testing.T) {
	simulator := sim.NewSimulator(sim.Config{Seed: 7, NoiseScale: 0})
	handler := NewHandler(simulator)

	postControls(t, handler, `{"throttle":0.4}`)
	simulator.Advance(2 * time.Second)
	postControls(t, handler, `{"rotate":true}`)

	for step := 0; step < 800; step++ {
		simulator.Advance(simulator.TickInterval())
		if simulator.Snapshot().Controls.Crashed {
			break
		}
	}

	crashed := snapshotFromAPI(t, handler)
	if !crashed.Controls.Crashed {
		t.Fatal("expected crash state after unsafe takeoff through dashboard controls")
	}

	if crashed.Controls.Warning != sim.WarningCrash {
		t.Fatalf("expected crash warning, got %q", crashed.Controls.Warning)
	}

	if crashed.Phase != "crash" {
		t.Fatalf("expected crash phase, got %q", crashed.Phase)
	}

	postControls(t, handler, `{"reset":true}`)
	reset := snapshotFromAPI(t, handler)

	if reset.Controls.Crashed {
		t.Fatal("expected reset to clear crash state")
	}

	if reset.Controls.Warning != sim.WarningNone {
		t.Fatalf("expected reset to clear warnings, got %q", reset.Controls.Warning)
	}

	if reset.Phase != "runway_idle" {
		t.Fatalf("expected reset to return runway_idle phase, got %q", reset.Phase)
	}

	if reset.Controls.Throttle != 0 {
		t.Fatalf("expected reset throttle to return to idle, got %.2f", reset.Controls.Throttle)
	}

	if reset.Controls.RunwayDistance != 0 {
		t.Fatalf("expected reset runway distance to return to zero, got %.1f", reset.Controls.RunwayDistance)
	}
}

func TestDashboardControlsResetUsesPreviouslySelectedMode(t *testing.T) {
	simulator := sim.NewSimulator(sim.Config{Seed: 7, NoiseScale: 0})
	handler := NewHandler(simulator)

	combined := postControls(t, handler, `{"mode":"transitional","reset":true}`)
	if combined.Mode != sim.ModeFull {
		t.Fatalf("expected combined mode/reset payload to reset using prior mode, got %q", combined.Mode)
	}

	updated := postControls(t, handler, `{"mode":"transitional"}`)
	if updated.Mode != sim.ModeTransitional {
		t.Fatalf("expected mode-only update to persist transitional mode, got %q", updated.Mode)
	}

	reset := postControls(t, handler, `{"reset":true}`)
	if reset.Mode != sim.ModeTransitional {
		t.Fatalf("expected follow-up reset to preserve selected mode, got %q", reset.Mode)
	}

	snapshot := snapshotFromAPI(t, handler)
	if snapshot.Controls.Mode != sim.ModeTransitional {
		t.Fatalf("expected snapshot controls to remain transitional after reset, got %q", snapshot.Controls.Mode)
	}

	if snapshot.Phase != "runway_idle" {
		t.Fatalf("expected reset to return to runway_idle phase, got %q", snapshot.Phase)
	}
}

func TestDashboardControlsAutopilotHeadingInteraction(t *testing.T) {
	simulator := sim.NewSimulator(sim.Config{Seed: 7, NoiseScale: 0})
	handler := NewHandler(simulator)

	postControls(t, handler, `{"toga":true}`)

	for step := 0; step < 120; step++ {
		snapshot := simulator.Snapshot()
		if snapshot.Airspeed.Indicated >= snapshot.Controls.Vr {
			postControls(t, handler, `{"rotate":true}`)
			break
		}
		simulator.Advance(simulator.TickInterval())
	}

	simulator.Advance(10 * time.Second)
	before := snapshotFromAPI(t, handler)

	controls := controlsFromAPI(t, handler)
	if !controls.TOGA {
		t.Fatal("expected TO/GA command to round-trip through controls API")
	}

	postControls(t, handler, `{"autopilot_engaged":true,"heading_hold":true,"selected_heading":315}`)
	simulator.Advance(20 * time.Second)
	after := snapshotFromAPI(t, handler)

	if !after.Controls.Autopilot.Engaged || !after.Controls.Autopilot.HeadingHold {
		t.Fatalf("expected autopilot heading mode to be active, got %#v", after.Controls.Autopilot)
	}

	if after.Controls.SelectedHeading != 315 {
		t.Fatalf("expected selected heading 315, got %.1f", after.Controls.SelectedHeading)
	}

	if math.Abs(angleDelta(before.AHRS.Heading, after.AHRS.Heading)) < 10 {
		t.Fatalf("expected autopilot to materially change heading, before %.1f after %.1f", before.AHRS.Heading, after.AHRS.Heading)
	}

	if math.Abs(angleDelta(after.AHRS.Heading, 315)) > 35 {
		t.Fatalf("expected heading to trend toward 315, got %.1f", after.AHRS.Heading)
	}
}

func postControls(t *testing.T, handler http.Handler, body string) sim.ControlState {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/controls", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from controls endpoint, got %d", recorder.Code)
	}

	var controls sim.ControlState
	if err := json.NewDecoder(recorder.Body).Decode(&controls); err != nil {
		t.Fatalf("failed to decode controls response: %v", err)
	}

	return controls
}

func applyScenarioSelection(t *testing.T, handler http.Handler, body string) {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/scenarios", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from scenario endpoint, got %d with body %q", recorder.Code, recorder.Body.String())
	}
}

func controlsFromAPI(t *testing.T, handler http.Handler) sim.ControlState {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/controls", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from controls endpoint, got %d", recorder.Code)
	}

	var controls sim.ControlState
	if err := json.NewDecoder(recorder.Body).Decode(&controls); err != nil {
		t.Fatalf("failed to decode controls payload: %v", err)
	}

	return controls
}

func snapshotFromAPI(t *testing.T, handler http.Handler) sim.Snapshot {
	t.Helper()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/snapshot", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from snapshot endpoint, got %d", recorder.Code)
	}

	var snapshot sim.Snapshot
	if err := json.NewDecoder(recorder.Body).Decode(&snapshot); err != nil {
		t.Fatalf("failed to decode snapshot payload: %v", err)
	}

	return snapshot
}

func angleDelta(a, b float64) float64 {
	delta := math.Mod(b-a+540, 360) - 180
	if delta < -180 {
		delta += 360
	}
	return delta
}

func renderDashboardPage(t *testing.T, provider SnapshotProvider) string {
	t.Helper()

	handler := NewHandler(provider)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from dashboard page, got %d", recorder.Code)
	}

	return recorder.Body.String()
}

func extractEmbeddedJSONMap(t *testing.T, body, startMarker, endMarker string) map[string]any {
	t.Helper()

	start := strings.Index(body, startMarker)
	if start == -1 {
		t.Fatalf("expected page to contain %q", startMarker)
	}
	start += len(startMarker)

	end := strings.Index(body[start:], endMarker)
	if end == -1 {
		t.Fatalf("expected page to contain %q after %q", endMarker, startMarker)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(body[start:start+end]), &payload); err != nil {
		t.Fatalf("failed to decode embedded JSON for %q: %v", startMarker, err)
	}

	return payload
}

func requireMapValue(t *testing.T, values map[string]any, key string) map[string]any {
	t.Helper()

	value, ok := values[key]
	if !ok {
		t.Fatalf("expected key %q to be present", key)
	}

	return requireMap(t, value, key)
}

func requireMap(t *testing.T, value any, label string) map[string]any {
	t.Helper()

	mapValue, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected %s to be an object, got %#v", label, value)
	}

	return mapValue
}

func requireSliceValue(t *testing.T, values map[string]any, key string) []any {
	t.Helper()

	value, ok := values[key]
	if !ok {
		t.Fatalf("expected key %q to be present", key)
	}

	sliceValue, ok := value.([]any)
	if !ok {
		t.Fatalf("expected %q to be an array, got %#v", key, value)
	}

	return sliceValue
}

func requireStringValue(t *testing.T, values map[string]any, key, want string) {
	t.Helper()

	value, ok := values[key]
	if !ok {
		t.Fatalf("expected key %q to be present", key)
	}

	stringValue, ok := value.(string)
	if !ok {
		t.Fatalf("expected %q to be a string, got %#v", key, value)
	}

	if stringValue != want {
		t.Fatalf("expected %q to be %q, got %q", key, want, stringValue)
	}
}

func requireNumberValue(t *testing.T, values map[string]any, key string, want float64) {
	t.Helper()

	value, ok := values[key]
	if !ok {
		t.Fatalf("expected key %q to be present", key)
	}

	numberValue, ok := value.(float64)
	if !ok {
		t.Fatalf("expected %q to be numeric, got %#v", key, value)
	}

	if numberValue != want {
		t.Fatalf("expected %q to be %v, got %v", key, want, numberValue)
	}
}

func requireBoolValue(t *testing.T, values map[string]any, key string, want bool) {
	t.Helper()

	value, ok := values[key]
	if !ok {
		t.Fatalf("expected key %q to be present", key)
	}

	boolValue, ok := value.(bool)
	if !ok {
		t.Fatalf("expected %q to be a bool, got %#v", key, value)
	}

	if boolValue != want {
		t.Fatalf("expected %q to be %t, got %t", key, want, boolValue)
	}
}
