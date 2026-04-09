package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/mockflight/mockflight/internal/sim"
)

type SnapshotProvider interface {
	Snapshot() sim.Snapshot
}

type FailureProvider interface {
	SetFailures(sim.FailureConfig)
	Failures() sim.FailureConfig
}

type ControlProvider interface {
	Controls() sim.ControlState
	ApplyControls(sim.ControlCommand) sim.ControlState
}

type ScenarioProgressProvider interface {
	Advance(time.Duration)
	TickInterval() time.Duration
}

type scenarioSelectionRequest struct {
	Key string `json:"key"`
}

type pageData struct {
	Title          string
	ScenariosJSON  template.JS
	ChecklistsJSON template.JS
}

func newPageData() pageData {
	return pageData{
		Title:          "MockFlight IFR Trainer",
		ScenariosJSON:  mustJSON(trainingScenarios),
		ChecklistsJSON: mustJSON(trainingChecklists),
	}
}

func mustJSON(value any) template.JS {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return template.JS(payload)
}

func NewHandler(provider SnapshotProvider) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pageTemplate.Execute(w, newPageData())
	})
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(provider.Snapshot())
	})
	mux.HandleFunc("/api/failures", func(w http.ResponseWriter, r *http.Request) {
		controller, ok := provider.(FailureProvider)
		if !ok {
			http.Error(w, "failure controls unavailable", http.StatusNotImplemented)
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(controller.Failures())
		case http.MethodPost:
			var failures sim.FailureConfig
			if err := json.NewDecoder(r.Body).Decode(&failures); err != nil {
				http.Error(w, "invalid failure payload", http.StatusBadRequest)
				return
			}
			controller.SetFailures(failures)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(controller.Failures())
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/controls", func(w http.ResponseWriter, r *http.Request) {
		controller, ok := provider.(ControlProvider)
		if !ok {
			http.Error(w, "simulator controls unavailable", http.StatusNotImplemented)
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(controller.Controls())
		case http.MethodPost:
			var command sim.ControlCommand
			if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
				http.Error(w, "invalid controls payload", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(controller.ApplyControls(command))
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scenarios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var selection scenarioSelectionRequest
			if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
				http.Error(w, "invalid scenario payload", http.StatusBadRequest)
				return
			}
			if status, message := applyTrainingScenario(provider, selection.Key); status != 0 {
				http.Error(w, message, status)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}

func applyTrainingScenario(provider SnapshotProvider, key string) (int, string) {
	failures, ok := provider.(FailureProvider)
	if !ok {
		return http.StatusNotImplemented, "failure controls unavailable"
	}

	if key == "" {
		failures.SetFailures(sim.FailureConfig{})
		return 0, ""
	}

	scenario, ok := trainingScenarios[key]
	if !ok {
		return http.StatusNotFound, "scenario not found"
	}

	var controls ControlProvider
	var progress ScenarioProgressProvider
	if scenario.Setup.Simulator != nil {
		var ok bool
		controls, ok = provider.(ControlProvider)
		if !ok {
			return http.StatusNotImplemented, "simulator controls unavailable"
		}
		progress, ok = provider.(ScenarioProgressProvider)
		if !ok {
			return http.StatusNotImplemented, "scenario progression unavailable"
		}
	}

	failures.SetFailures(scenario.Setup.Failures)

	if scenario.Setup.Simulator != nil {
		if err := applyScenarioSimulatorSetup(provider, controls, progress, scenario.Setup.Simulator); err != nil {
			return http.StatusConflict, err.Error()
		}
	}

	return 0, ""
}

func applyScenarioSimulatorSetup(provider SnapshotProvider, controls ControlProvider, progress ScenarioProgressProvider, setup *TrainingScenarioSimulatorSetup) error {
	mode := setup.Mode
	controls.ApplyControls(sim.ControlCommand{Mode: &mode})
	reset := true
	controls.ApplyControls(sim.ControlCommand{Reset: &reset})

	for _, step := range setup.Steps {
		command := sim.ControlCommand{}
		hasCommand := false
		if step.Controls.Throttle > 0 {
			throttle := step.Controls.Throttle
			command.Throttle = &throttle
			hasCommand = true
		}
		if step.Controls.TOGA {
			toga := true
			command.TOGA = &toga
			hasCommand = true
		}
		if step.Controls.Rotate {
			rotate := true
			command.Rotate = &rotate
			hasCommand = true
		}
		if hasCommand {
			controls.ApplyControls(command)
		}
		advanceScenarioStep(provider, progress, step.Advance)
	}

	return nil
}

func advanceScenarioStep(provider SnapshotProvider, progress ScenarioProgressProvider, advance TrainingScenarioAdvance) {
	if advance.Milliseconds <= 0 && advance.Until == "" {
		return
	}

	budget := time.Duration(advance.Milliseconds) * time.Millisecond
	if advance.Until == "" {
		progress.Advance(budget)
		return
	}

	step := progress.TickInterval()
	if step <= 0 {
		step = 250 * time.Millisecond
	}

	for elapsed := time.Duration(0); elapsed <= budget; {
		if scenarioAdvanceSatisfied(provider.Snapshot(), advance.Until) {
			return
		}

		if elapsed == budget {
			break
		}

		delta := step
		if remaining := budget - elapsed; delta > remaining {
			delta = remaining
		}
		progress.Advance(delta)
		elapsed += delta
	}
}

func scenarioAdvanceSatisfied(snapshot sim.Snapshot, until ScenarioAdvanceUntil) bool {
	switch until {
	case AdvanceUntilVr:
		return snapshot.Airspeed.Indicated >= snapshot.Controls.Vr
	case AdvanceUntilAirborne:
		return snapshot.Controls.Airborne
	default:
		return true
	}
}

var pageTemplate = template.Must(template.New("dashboard").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root {
      --bg: #05070d;
      --panel: #0f1624;
      --panel-strong: #151f31;
      --panel-soft: #1d2b40;
      --edge: rgba(203, 215, 235, 0.16);
      --text: #eef4ff;
      --muted: #92a3bf;
      --sky: #2c74b5;
      --ground: #684e3a;
      --green: #60e0a1;
      --magenta: #ff53cf;
      --amber: #ffc463;
      --red: #ff6f6f;
      --cyan: #81d8ff;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Avenir Next Condensed", "Avenir Next", "Segoe UI", sans-serif;
      color: var(--text);
      background:
        radial-gradient(circle at top, rgba(76, 120, 188, 0.18), transparent 30%),
        linear-gradient(180deg, #0a1020 0%, #05070d 100%);
    }
    body[data-connection="offline"] .connection-banner {
      opacity: 1;
      transform: translateY(0);
    }
    .connection-banner {
      position: sticky;
      top: 0;
      z-index: 20;
      padding: 10px 16px;
      text-align: center;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      background: rgba(255, 111, 111, 0.9);
      color: #2c0b0b;
      opacity: 0;
      transform: translateY(-100%);
      transition: opacity 160ms ease, transform 160ms ease;
    }
    .sim-shell {
      max-width: 1440px;
      margin: 0 auto;
      padding: 24px 18px 40px;
    }
    .topbar {
      display: flex;
      justify-content: space-between;
      gap: 18px;
      align-items: end;
      margin-bottom: 18px;
    }
    .topbar h1 {
      margin: 6px 0 0;
      font-size: clamp(2rem, 4vw, 3.4rem);
      letter-spacing: 0.02em;
      text-transform: uppercase;
    }
    .eyebrow,
    .section-kicker {
      margin: 0;
      font-size: 0.78rem;
      letter-spacing: 0.22em;
      text-transform: uppercase;
      color: var(--amber);
    }
    .intro {
      margin: 10px 0 0;
      color: var(--muted);
      max-width: 72ch;
      line-height: 1.5;
    }
    .flight-meta {
      display: grid;
      gap: 10px;
      grid-template-columns: repeat(3, minmax(120px, 1fr));
      min-width: 360px;
    }
    .meta-box,
    .panel,
    .pfd-panel,
    .nav-panel,
    .systems-card {
      background: linear-gradient(180deg, rgba(18, 27, 41, 0.98), rgba(10, 16, 27, 0.98));
      border: 1px solid var(--edge);
      box-shadow: inset 0 1px 0 rgba(255,255,255,0.05), 0 18px 32px rgba(0, 0, 0, 0.28);
    }
    .meta-box {
      border-radius: 16px;
      padding: 12px 14px;
    }
    .meta-label,
    .panel h2,
    .systems-card h2,
    .systems-card h3,
    .tape-label,
    .mini-label {
      margin: 0;
      font-size: 0.75rem;
      letter-spacing: 0.15em;
      text-transform: uppercase;
      color: var(--muted);
    }
    .meta-value {
      margin: 8px 0 0;
      font-size: 1.4rem;
      font-weight: 700;
    }
    .deck-grid {
      display: grid;
      gap: 18px;
      grid-template-columns: minmax(0, 1.9fr) minmax(340px, 0.9fr);
      align-items: start;
    }
    .flight-deck {
      display: grid;
      gap: 18px;
    }
    .pfd-panel,
    .nav-panel,
    .systems-card {
      border-radius: 22px;
      padding: 18px;
    }
    .annunciators {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      min-height: 38px;
      margin-bottom: 14px;
    }
    .annunciators span {
      padding: 7px 10px;
      border-radius: 999px;
      border: 1px solid rgba(255, 196, 99, 0.3);
      background: rgba(255, 196, 99, 0.14);
      color: #ffe4a6;
      font-size: 0.74rem;
      letter-spacing: 0.12em;
      text-transform: uppercase;
    }
    .annunciators span.nominal {
      border-color: rgba(96, 224, 161, 0.28);
      background: rgba(96, 224, 161, 0.12);
      color: #b6f2d1;
    }
    .pfd-core {
      display: grid;
      gap: 16px;
      grid-template-columns: 124px minmax(0, 1fr) 124px;
      align-items: stretch;
    }
    .tape {
      position: relative;
      border-radius: 18px;
      overflow: hidden;
      padding: 12px 10px;
      background: linear-gradient(180deg, rgba(18, 27, 41, 1), rgba(31, 46, 69, 1));
      border: 1px solid rgba(255,255,255,0.06);
    }
    .tape-window {
      position: relative;
      height: 100%;
      min-height: 320px;
      border-radius: 14px;
      background: linear-gradient(180deg, rgba(5, 7, 13, 0.96), rgba(14, 21, 32, 0.92));
      overflow: hidden;
    }
    .tape-readout {
      position: absolute;
      left: 10px;
      right: 10px;
      top: 50%;
      transform: translateY(-50%);
      padding: 10px 12px;
      border-radius: 12px;
      background: rgba(0, 0, 0, 0.72);
      border: 1px solid rgba(129, 216, 255, 0.35);
      text-align: center;
      font-size: 1.8rem;
      font-weight: 700;
      color: var(--cyan);
    }
    .tape-trend {
      position: absolute;
      left: 50%;
      bottom: 18px;
      width: 6px;
      margin-left: -3px;
      border-radius: 999px;
      background: linear-gradient(180deg, rgba(96, 224, 161, 0.08), rgba(96, 224, 161, 0.7));
      transform-origin: bottom center;
    }
    .tape-subvalue {
      margin-top: 12px;
      font-size: 0.92rem;
      color: var(--muted);
    }
    .attitude-stage {
      position: relative;
      min-height: 360px;
      overflow: hidden;
      border-radius: 24px;
      border: 1px solid rgba(255,255,255,0.08);
      background: #07101c;
    }
    .horizon {
      position: absolute;
      inset: -35%;
      background: linear-gradient(180deg, var(--sky) 0 50%, var(--ground) 50% 100%);
      transition: transform 200ms linear;
    }
    .pitch-ladder {
      position: absolute;
      inset: 0;
      pointer-events: none;
      background-image:
        linear-gradient(to bottom, transparent 0 8%, rgba(255,255,255,0.18) 8% 8.6%, transparent 8.6% 18%, rgba(255,255,255,0.12) 18% 18.4%, transparent 18.4% 28%, rgba(255,255,255,0.16) 28% 28.6%, transparent 28.6% 38%, rgba(255,255,255,0.12) 38% 38.4%, transparent 38.4% 48%, rgba(255,255,255,0.22) 48% 49%, transparent 49% 51%, rgba(255,255,255,0.22) 51% 52%, transparent 52% 62%, rgba(255,255,255,0.12) 62% 62.4%, transparent 62.4% 72%, rgba(255,255,255,0.16) 72% 72.6%, transparent 72.6% 82%, rgba(255,255,255,0.12) 82% 82.4%, transparent 82.4% 92%, rgba(255,255,255,0.18) 92% 92.6%, transparent 92.6% 100%);
      opacity: 0.9;
    }
    .bank-arc {
      position: absolute;
      left: 50%;
      top: 18px;
      width: 220px;
      height: 110px;
      margin-left: -110px;
      border: 4px solid rgba(255,255,255,0.55);
      border-bottom: none;
      border-radius: 220px 220px 0 0;
    }
    .bank-pointer {
      position: absolute;
      left: 50%;
      top: 20px;
      width: 0;
      height: 0;
      margin-left: -9px;
      border-left: 9px solid transparent;
      border-right: 9px solid transparent;
      border-top: 15px solid var(--amber);
      z-index: 3;
    }
    .aircraft-symbol {
      position: absolute;
      left: 50%;
      top: 50%;
      width: 180px;
      height: 18px;
      margin-left: -90px;
      margin-top: -9px;
      z-index: 3;
    }
    .aircraft-symbol::before,
    .aircraft-symbol::after {
      content: "";
      position: absolute;
      top: 4px;
      width: 72px;
      height: 10px;
      border-top: 4px solid #fff;
    }
    .aircraft-symbol::before { left: 0; border-left: 4px solid #fff; }
    .aircraft-symbol::after { right: 0; border-right: 4px solid #fff; }
    .aircraft-center {
      position: absolute;
      left: 50%;
      top: 50%;
      width: 18px;
      height: 18px;
      margin-left: -9px;
      margin-top: -9px;
      border: 3px solid var(--amber);
      border-radius: 50%;
      z-index: 3;
      background: rgba(0,0,0,0.32);
    }
    .failure-flag {
      position: absolute;
      left: 50%;
      top: 18px;
      transform: translateX(-50%);
      padding: 8px 14px;
      border-radius: 999px;
      background: rgba(255, 111, 111, 0.92);
      color: #2d0808;
      letter-spacing: 0.16em;
      text-transform: uppercase;
      font-weight: 700;
      opacity: 0;
      transition: opacity 140ms ease;
      z-index: 4;
    }
    .failure-flag.active { opacity: 1; }
    .attitude-footer {
      position: absolute;
      left: 18px;
      right: 18px;
      bottom: 14px;
      display: grid;
      gap: 10px;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      z-index: 3;
    }
    .attitude-metric {
      padding: 10px 12px;
      border-radius: 12px;
      background: rgba(0, 0, 0, 0.55);
      border: 1px solid rgba(255,255,255,0.08);
    }
    .attitude-metric strong {
      display: block;
      margin-top: 5px;
      font-size: 1.2rem;
      color: var(--cyan);
    }
    .nav-panel {
      display: grid;
      gap: 16px;
    }
    .nav-grid {
      display: grid;
      gap: 16px;
      grid-template-columns: 1.2fr 0.8fr;
    }
    .hsi {
      padding: 16px;
      border-radius: 18px;
      background: linear-gradient(180deg, rgba(7, 11, 18, 0.96), rgba(15, 22, 36, 0.96));
      border: 1px solid rgba(255,255,255,0.08);
    }
    .cdi-track {
      position: relative;
      height: 22px;
      margin-top: 16px;
      border-radius: 999px;
      background: rgba(255,255,255,0.08);
      overflow: hidden;
    }
    .cdi-center {
      position: absolute;
      left: 50%;
      top: 0;
      bottom: 0;
      width: 2px;
      background: rgba(255,255,255,0.6);
    }
    .cdi-bar {
      position: absolute;
      top: 2px;
      bottom: 2px;
      width: 34px;
      margin-left: -17px;
      border-radius: 999px;
      background: var(--magenta);
      transition: left 220ms linear;
    }
    .nav-metrics {
      display: grid;
      gap: 10px;
    }
    .nav-metric {
      padding: 14px;
      border-radius: 16px;
      background: rgba(255,255,255,0.04);
      border: 1px solid rgba(255,255,255,0.06);
    }
    .nav-metric strong {
      display: block;
      margin-top: 8px;
      font-size: 1.35rem;
      color: var(--green);
    }
    .systems-bay {
      display: grid;
      gap: 18px;
    }
    .systems-card h2,
    .systems-card h3 {
      margin-bottom: 10px;
    }
    .failure-list,
    .checklist-items,
    .sensor-list {
      list-style: none;
      padding: 0;
      margin: 0;
    }
    .failure-list {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-bottom: 14px;
    }
    .failure-list li {
      padding: 7px 10px;
      border-radius: 999px;
      border: 1px solid rgba(255,255,255,0.1);
      background: rgba(255,255,255,0.06);
      color: #dde7fb;
      font-size: 0.8rem;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .failure-controls,
    .selector-grid,
    .raw-grid {
      display: grid;
      gap: 10px;
    }
    .failure-controls {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .failure-controls label,
    .selector-grid label {
      display: grid;
      gap: 6px;
      font-size: 0.88rem;
      color: var(--muted);
    }
    .failure-controls span {
      display: flex;
      gap: 8px;
      align-items: center;
      padding: 11px 12px;
      border-radius: 12px;
      background: rgba(255,255,255,0.05);
      color: var(--text);
    }
    .failure-controls input {
      accent-color: var(--amber);
    }
    .control-grid,
    .autopilot-grid,
    .vspeed-grid {
      display: grid;
      gap: 10px;
    }
    .button-row {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      margin-top: 12px;
    }
    button {
      padding: 11px 14px;
      border: 1px solid rgba(255,255,255,0.12);
      border-radius: 12px;
      background: linear-gradient(180deg, rgba(33, 45, 64, 0.98), rgba(14, 20, 31, 0.98));
      color: var(--text);
      font: inherit;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      cursor: pointer;
    }
    button.primary {
      border-color: rgba(255, 196, 99, 0.42);
      color: #ffe7b7;
    }
    button.warn {
      border-color: rgba(255, 111, 111, 0.42);
      color: #ffd1d1;
    }
    button:disabled,
    select:disabled,
    input:disabled {
      opacity: 0.55;
      cursor: not-allowed;
    }
    input[type="range"] {
      width: 100%;
      accent-color: var(--amber);
    }
    input[type="number"] {
      width: 100%;
      padding: 11px 12px;
      border-radius: 12px;
      border: 1px solid rgba(255,255,255,0.12);
      background: #09111c;
      color: var(--text);
      font: inherit;
    }
    .control-readout,
    .warning-banner {
      padding: 10px 12px;
      border-radius: 12px;
      background: rgba(255,255,255,0.05);
      border: 1px solid rgba(255,255,255,0.08);
      color: var(--text);
    }
    .warning-banner {
      color: #ffe3b1;
      background: rgba(255, 196, 99, 0.12);
      border-color: rgba(255, 196, 99, 0.24);
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .warning-banner.crash {
      color: #ffd5d5;
      background: rgba(255, 111, 111, 0.16);
      border-color: rgba(255, 111, 111, 0.28);
    }
    .aural-status {
      padding: 10px 12px;
      border-radius: 12px;
      border: 1px solid rgba(116, 214, 175, 0.24);
      background: rgba(116, 214, 175, 0.1);
      color: #d6ffea;
      text-transform: uppercase;
      letter-spacing: 0.08em;
    }
    .aural-status.muted,
    .aural-status.unavailable {
      border-color: rgba(255,255,255,0.08);
      background: rgba(255,255,255,0.05);
      color: var(--muted);
    }
    .vspeed-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    select {
      width: 100%;
      padding: 11px 12px;
      border-radius: 12px;
      border: 1px solid rgba(255,255,255,0.12);
      background: #09111c;
      color: var(--text);
      font: inherit;
    }
    .feedback {
      margin-top: 10px;
      padding: 11px 12px;
      border-radius: 12px;
      background: rgba(255,255,255,0.04);
      color: var(--muted);
      border: 1px solid rgba(255,255,255,0.06);
      min-height: 44px;
    }
    .feedback.good {
      color: #b7f2d0;
      border-color: rgba(96, 224, 161, 0.22);
      background: rgba(96, 224, 161, 0.1);
    }
    .feedback.bad {
      color: #ffd7a1;
      border-color: rgba(255, 196, 99, 0.28);
      background: rgba(255, 196, 99, 0.12);
    }
    .checklist-drawer {
      margin-top: 12px;
      padding: 14px;
      border-radius: 16px;
      background: rgba(4, 8, 14, 0.64);
      border: 1px solid rgba(255,255,255,0.08);
    }
    .checklist-drawer header {
      display: flex;
      justify-content: space-between;
      gap: 10px;
      align-items: center;
      margin-bottom: 10px;
    }
    .checklist-tag {
      padding: 6px 8px;
      border-radius: 999px;
      font-size: 0.72rem;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      background: rgba(255,255,255,0.08);
      color: var(--muted);
    }
    .checklist-tag.good {
      background: rgba(96, 224, 161, 0.16);
      color: #c5f8dd;
    }
    .checklist-tag.bad {
      background: rgba(255, 196, 99, 0.18);
      color: #ffe4ab;
    }
    .checklist-items li {
      padding: 10px 0;
      border-top: 1px solid rgba(255,255,255,0.08);
      color: #e5edfb;
      line-height: 1.45;
    }
    .checklist-items li:first-child { border-top: none; }
    .raw-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .sensor-list li {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      padding: 8px 0;
      border-top: 1px solid rgba(255,255,255,0.06);
      color: var(--muted);
    }
    .sensor-list li:first-child { border-top: none; }
    .sensor-list strong { color: var(--text); }
    .footer-note {
      margin-top: 18px;
      text-align: right;
      color: var(--muted);
      font-size: 0.9rem;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }
    @media (max-width: 1120px) {
      .deck-grid {
        grid-template-columns: 1fr;
      }
      .systems-bay {
        grid-template-columns: repeat(2, minmax(0, 1fr));
      }
    }
    @media (max-width: 860px) {
      .topbar,
      .nav-grid,
      .pfd-core,
      .systems-bay,
      .flight-meta,
      .failure-controls,
      .raw-grid {
        grid-template-columns: 1fr;
        display: grid;
      }
      .topbar {
        align-items: start;
      }
      .flight-meta {
        min-width: 0;
      }
      .pfd-core {
        gap: 12px;
      }
      .tape-window {
        min-height: 150px;
      }
      .attitude-stage {
        min-height: 280px;
      }
    }
  </style>
</head>
<body data-connection="online">
  <div class="connection-banner">Simulator data link lost</div>
  <main class="sim-shell">
    <header class="topbar">
      <div>
        <p class="eyebrow">IFR Only Training Device</p>
        <h1>MockFlight IFR Trainer</h1>
        <p class="intro">A glass-lite instrument trainer focused on approach scan discipline, partial-panel recovery, and failure recognition. The cockpit view stays pilot-facing while diagnostics and scenario controls live in the instructor bay.</p>
      </div>
      <section class="flight-meta">
        <article class="meta-box">
          <p class="meta-label">Flight Phase</p>
          <p class="meta-value" id="phase">Ground</p>
        </article>
        <article class="meta-box">
          <p class="meta-label">UTC Time</p>
          <p class="meta-value" id="timestamp">--</p>
        </article>
        <article class="meta-box">
          <p class="meta-label">Scenario</p>
          <p class="meta-value" id="scenarioState">Manual</p>
        </article>
      </section>
    </header>

    <div class="deck-grid">
      <section class="flight-deck">
        <section class="pfd-panel" aria-label="Primary Flight Display">
          <p class="section-kicker">Primary Flight Display</p>
          <div class="annunciators" id="annunciators">
            <span class="nominal">Nominal</span>
          </div>
          <div class="pfd-core">
            <article class="tape">
              <p class="tape-label">Airspeed</p>
              <div class="tape-window">
                <div class="tape-readout" id="airspeedReadout">0 KT</div>
                <div class="tape-trend" id="airspeedTrend"></div>
              </div>
              <p class="tape-subvalue">Pitot <span id="pitot">--</span></p>
            </article>

            <article class="attitude-stage">
              <div class="horizon" id="horizon"></div>
              <div class="pitch-ladder"></div>
              <div class="bank-arc"></div>
              <div class="bank-pointer"></div>
              <div class="failure-flag" id="failureFlag">Partial Panel</div>
              <div class="aircraft-symbol"></div>
              <div class="aircraft-center"></div>
              <div class="attitude-footer">
                <div class="attitude-metric">
                  <p class="mini-label">Pitch</p>
                  <strong id="pitchValue">0.0 deg</strong>
                </div>
                <div class="attitude-metric">
                  <p class="mini-label">Heading</p>
                  <strong id="headingValue">000 deg</strong>
                </div>
                <div class="attitude-metric">
                  <p class="mini-label">Roll</p>
                  <strong id="rollValue">0.0 deg</strong>
                </div>
              </div>
            </article>

            <article class="tape">
              <p class="tape-label">Altitude</p>
              <div class="tape-window">
                <div class="tape-readout" id="altitudeReadout">0 FT</div>
                <div class="tape-trend" id="altitudeTrend"></div>
              </div>
              <p class="tape-subvalue">Vertical speed <span id="verticalSpeed">--</span></p>
            </article>
          </div>
        </section>

        <section class="nav-panel" aria-label="Approach Guidance">
          <p class="section-kicker">Approach Guidance</p>
          <div class="nav-grid">
            <article class="hsi panel">
              <h2>Approach Guidance</h2>
              <p class="intro">Synthetic nav cues for IFR scan practice until simulator-backed nav data is added.</p>
              <div class="cdi-track">
                <div class="cdi-center"></div>
                <div class="cdi-bar" id="cdiBar"></div>
              </div>
            </article>
            <div class="nav-metrics">
              <article class="nav-metric">
                <p class="mini-label">Nav Source</p>
                <strong id="navSource">LOC1</strong>
              </article>
              <article class="nav-metric">
                <p class="mini-label">Selected Course</p>
                <strong id="selectedCourse">000</strong>
              </article>
              <article class="nav-metric">
                <p class="mini-label">Approach State</p>
                <strong id="approachState">Vectors</strong>
              </article>
            </div>
          </div>
        </section>
      </section>

      <aside class="systems-bay">
        <section class="systems-card">
          <h2>Runway Controls</h2>
          <div class="control-grid">
            <label>
              Simulator Mode
              <select id="simMode">
                <option value="full">Full</option>
                <option value="transitional">Transitional</option>
              </select>
            </label>
            <label>
              Throttle
              <input type="range" id="throttleInput" min="0" max="1" step="0.01" value="0">
            </label>
            <div class="control-readout" id="throttleReadout">Throttle 0%</div>
            <div class="vspeed-grid">
              <div class="control-readout">V1 <strong id="v1Value">--</strong></div>
              <div class="control-readout">Vr <strong id="vrValue">--</strong></div>
              <div class="control-readout">V2 <strong id="v2Value">--</strong></div>
            </div>
            <div class="warning-banner" id="warningBanner">No takeoff warnings</div>
            <label><span><input type="checkbox" id="auralAlertsEnabled" checked> Aural Alerts</span></label>
            <div>
              <p class="mini-label">Aural Status</p>
              <div class="aural-status" id="auralAlertStatus">Aural alerts armed</div>
            </div>
            <div class="button-row">
              <button class="primary" id="togaButton">TO/GA</button>
              <button class="primary" id="rotateButton">Rotate</button>
              <button class="warn" id="resetButton">Reset</button>
            </div>
          </div>
        </section>

        <section class="systems-card">
          <h2>Autopilot</h2>
          <div class="autopilot-grid">
            <label><span><input type="checkbox" id="autopilotMaster"> Autopilot</span></label>
            <label><span><input type="checkbox" id="headingHold"> Heading Select</span></label>
            <label>
              Heading Select
              <input type="number" id="headingTarget" min="0" max="359" step="1" value="270">
            </label>
            <label><span><input type="checkbox" id="altitudeHold"> Altitude Hold</span></label>
            <label>
              Altitude Target
              <input type="number" id="altitudeTarget" min="342" max="12000" step="100" value="1800">
            </label>
            <label><span><input type="checkbox" id="verticalSpeedMode"> Vertical Speed</span></label>
            <label>
              Vertical Speed Target
              <input type="number" id="verticalSpeedTarget" min="-1500" max="2500" step="100" value="700">
            </label>
          </div>
        </section>

        <section class="systems-card">
          <h2>Failures</h2>
          <ul class="failure-list" id="failures">
            <li>Nominal</li>
          </ul>
          <div class="failure-controls">
            <label><span><input type="checkbox" id="pitotBlocked"> Pitot Blocked</span></label>
            <label><span><input type="checkbox" id="pitotDrainBlocked"> Pitot Drain Blocked</span></label>
            <label><span><input type="checkbox" id="staticBlocked"> Static Port Blocked</span></label>
            <label><span><input type="checkbox" id="staticLeak"> Static Leak</span></label>
            <label><span><input type="checkbox" id="magDisturbed"> Magnetometer Disturbed</span></label>
            <label><span><input type="checkbox" id="gyroSaturation"> Gyro Saturation</span></label>
          </div>
        </section>

        <section class="systems-card">
          <h2>Training Scenarios</h2>
          <div class="selector-grid">
            <label>
              Scenario Template
              <select id="scenarioSelect">
                <option value="">Select scenario</option>
              </select>
            </label>
          </div>
          <div class="feedback" id="scenarioFeedback">Choose a failure template to inject the cockpit state, then identify the correct checklist.</div>
          <ul class="sensor-list">
            <li><span>Category</span><strong id="scenarioCategory">Manual</strong></li>
            <li><span>Training Phase</span><strong id="scenarioTrainingPhase">Free Flight</strong></li>
            <li><span>Weather</span><strong id="scenarioWeather">Live simulator state</strong></li>
            <li><span>Fidelity</span><strong id="scenarioFidelity">Manual</strong></li>
          </ul>
        </section>

        <section class="systems-card">
          <h2>Emergency Checklists</h2>
          <div class="selector-grid">
            <label>
              Checklist Selector
              <select id="checklistSelect" disabled>
                <option value="">Choose checklist</option>
              </select>
            </label>
          </div>
          <div class="feedback" id="checklistFeedback">The correct procedure is not auto-opened. Recognition is part of the exercise.</div>
          <div class="checklist-drawer" id="checklistDrawer">
            <header>
              <div>
                <p class="mini-label">Active Procedure</p>
                <h3 id="checklistTitle">No checklist selected</h3>
              </div>
              <span class="checklist-tag" id="checklistTag">Standby</span>
            </header>
            <ul class="checklist-items" id="checklistItems">
              <li>Select a scenario, diagnose the failure, and choose a checklist to begin.</li>
            </ul>
          </div>
        </section>

        <section class="systems-card">
          <h2>Raw Sensor Bay</h2>
          <div class="raw-grid">
            <div>
              <h3>Pitot / Static</h3>
              <ul class="sensor-list">
                <li><span>Pitot</span><strong id="pitotValue">--</strong></li>
                <li><span>Static</span><strong id="staticValue">--</strong></li>
                <li><span>Altitude</span><strong id="altitudeValue">--</strong></li>
                <li><span>Vertical Speed</span><strong id="verticalValue">--</strong></li>
              </ul>
            </div>
            <div>
              <h3>Inertial / Magnetic</h3>
              <ul class="sensor-list">
                <li><span>Accel X</span><strong id="accelX">--</strong></li>
                <li><span>Accel Y</span><strong id="accelY">--</strong></li>
                <li><span>Accel Z</span><strong id="accelZ">--</strong></li>
                <li><span>Gyro Z</span><strong id="gyroZ">--</strong></li>
                <li><span>Mag X</span><strong id="magX">--</strong></li>
                <li><span>Mag Y</span><strong id="magY">--</strong></li>
              </ul>
            </div>
          </div>
        </section>
      </aside>
    </div>

    <p class="footer-note">Polling /api/snapshot every 750ms from a single in-process simulator</p>
  </main>

  <script>
    const scenarios = {{.ScenariosJSON}};

    const checklists = {{.ChecklistsJSON}};

    const state = {
      activeScenario: '',
      selectedChecklist: '',
      snapshot: null,
      savingFailures: false,
      savingControls: false,
      audio: {
        unlocked: false,
        muted: false,
        activeWarning: '',
        intervalId: null,
        supported: typeof window !== 'undefined' && 'speechSynthesis' in window && 'SpeechSynthesisUtterance' in window,
      },
    };

    const text = (id, value) => { document.getElementById(id).textContent = value; };
    const format = (value, unit, digits) => Number(value || 0).toFixed(digits) + ' ' + unit;

    function setConnectionState(online) {
      document.body.dataset.connection = online ? 'online' : 'offline';
    }

    function normalizePhase(phase) {
      return String(phase || 'ground').replaceAll('_', ' ');
    }

    function toDisplayModel(snapshot) {
      const heading = Number(snapshot.ahrs.heading || 0);
      const pitch = Number(snapshot.ahrs.pitch || 0);
      const roll = Number(snapshot.ahrs.roll || 0);
      const altitude = Number(snapshot.altimeter.value || 0);
      const airspeed = Number(snapshot.airspeed.indicated || 0);
      const verticalSpeed = Number(snapshot.vertical_speed.value || 0);
      const activeFailures = snapshot.active_failures || [];
      const controls = snapshot.controls || {
        mode: 'full',
        throttle: 0,
        toga: false,
        rotate_commanded: false,
        warning: '',
        crashed: false,
        v1: 62,
        vr: 67,
        v2: 74,
        selected_heading: heading,
        selected_altitude: altitude,
        selected_vertical_speed: 700,
        autopilot: {},
      };
      const scenario = scenarios[state.activeScenario];
      return {
        phaseLabel: normalizePhase(snapshot.phase),
        timestamp: new Date(snapshot.timestamp).toISOString().slice(11, 19),
        heading: heading,
        pitch: pitch,
        roll: roll,
        altitude: altitude,
        airspeed: airspeed,
        verticalSpeed: verticalSpeed,
        pitot: Number(snapshot.pitot.value || 0),
        staticPressure: Number(snapshot.static_air.value || 0),
        accelX: Number(snapshot.accelerometer.x || 0),
        accelY: Number(snapshot.accelerometer.y || 0),
        accelZ: Number(snapshot.accelerometer.z || 0),
        gyroZ: Number(snapshot.gyroscope.z || 0),
        magX: Number(snapshot.magnetometer.x || 0),
        magY: Number(snapshot.magnetometer.y || 0),
        activeFailures: activeFailures,
        controls: controls,
        navSource: scenario ? scenario.nav_source : 'LOC1',
        approachState: scenario ? scenario.approach_state : 'Vectors',
        selectedCourse: String(Math.round((controls.selected_heading || heading) / 10) * 10).padStart(3, '0'),
        cdiOffset: Math.max(-42, Math.min(42, Math.sin(heading * Math.PI / 180) * 42)),
      };
    }

    function formatFidelity(value) {
      return String(value || 'manual').replaceAll('_', ' ');
    }

    function renderScenarioMetadata(scenario) {
      text('scenarioCategory', scenario ? scenario.category : 'Manual');
      text('scenarioTrainingPhase', scenario ? scenario.training_phase : 'Free Flight');
      text('scenarioWeather', scenario ? scenario.weather : 'Live simulator state');
      text('scenarioFidelity', scenario ? formatFidelity(scenario.fidelity) : 'Manual');
    }

    function failureTokens(activeFailures) {
      if (!activeFailures.length) {
        return [{ label: 'Nominal', className: 'nominal' }];
      }
      const labels = {
        pitot_blocked: 'PITOT',
        pitot_drain_blocked: 'DRAIN',
        static_port_blocked: 'STATIC',
        static_leak: 'STATIC LEAK',
        magnetometer_disturbed: 'MAG',
        gyro_saturation: 'GYRO',
      };
      return activeFailures.map((failure) => ({ label: labels[failure] || failure.replaceAll('_', ' '), className: '' }));
    }

    function renderAnnunciators(activeFailures) {
      const tokens = failureTokens(activeFailures);
      const container = document.getElementById('annunciators');
      container.innerHTML = '';
      tokens.forEach((token) => {
        const node = document.createElement('span');
        if (token.className) {
          node.className = token.className;
        }
        node.textContent = token.label;
        container.appendChild(node);
      });
    }

    function renderFailures(activeFailures) {
      const list = document.getElementById('failures');
      list.innerHTML = '';
      const failures = activeFailures.length ? activeFailures : ['nominal'];
      failures.forEach((failure) => {
        const item = document.createElement('li');
        item.textContent = failure.replaceAll('_', ' ');
        list.appendChild(item);
      });
      const unreliable = activeFailures.includes('gyro_saturation') || activeFailures.includes('magnetometer_disturbed');
      const flag = document.getElementById('failureFlag');
      flag.classList.toggle('active', unreliable);
      flag.textContent = unreliable ? 'Partial Panel' : 'Nominal';
    }

    function syncFailureControls(activeFailures) {
      const set = new Set(activeFailures || []);
      document.getElementById('pitotBlocked').checked = set.has('pitot_blocked');
      document.getElementById('pitotDrainBlocked').checked = set.has('pitot_drain_blocked');
      document.getElementById('staticBlocked').checked = set.has('static_port_blocked');
      document.getElementById('staticLeak').checked = set.has('static_leak');
      document.getElementById('magDisturbed').checked = set.has('magnetometer_disturbed');
      document.getElementById('gyroSaturation').checked = set.has('gyro_saturation');
    }

    function failurePayload() {
      return {
        pitot_blocked: document.getElementById('pitotBlocked').checked,
        pitot_drain_blocked: document.getElementById('pitotDrainBlocked').checked,
        static_port_blocked: document.getElementById('staticBlocked').checked,
        static_leak: document.getElementById('staticLeak').checked,
        magnetometer_disturbed: document.getElementById('magDisturbed').checked,
        gyro_saturation: document.getElementById('gyroSaturation').checked,
      };
    }

    async function saveFailures(payload) {
      if (state.savingFailures) {
        return;
      }
      state.savingFailures = true;
      try {
        const response = await fetch('/api/failures', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload || failurePayload()),
        });
        if (!response.ok) {
          throw new Error('failed to save failures');
        }
      } finally {
        state.savingFailures = false;
      }
    }

    async function saveScenarioSelection(key) {
      const response = await fetch('/api/scenarios', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key }),
      });
      if (!response.ok) {
        throw new Error('failed to apply scenario');
      }
    }

    async function postControls(payload) {
      if (state.savingControls) {
        return;
      }
      state.savingControls = true;
      try {
        const response = await fetch('/api/controls', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        });
        if (!response.ok) {
          throw new Error('failed to update controls');
        }
        const controls = await response.json();
        if (state.snapshot) {
          state.snapshot.controls = controls;
        }
      } finally {
        state.savingControls = false;
      }
    }

    function populateScenarioSelect() {
      const select = document.getElementById('scenarioSelect');
      Object.entries(scenarios).forEach(([key, scenario]) => {
        const option = document.createElement('option');
        option.value = key;
        option.textContent = scenario.title;
        select.appendChild(option);
      });
    }

    function populateChecklistSelect(scenarioKey) {
      const select = document.getElementById('checklistSelect');
      select.innerHTML = '<option value="">Choose checklist</option>';
      const scenario = scenarios[scenarioKey];
      if (!scenario) {
        select.disabled = true;
        return;
      }
      scenario.checklist_options.forEach((name) => {
        const option = document.createElement('option');
        option.value = name;
        option.textContent = name;
        select.appendChild(option);
      });
      select.disabled = false;
    }

    function renderChecklist(name, correct) {
      const checklist = checklists[name];
      const items = checklist ? checklist.items : ['Checklist content unavailable.'];
      text('checklistTitle', checklist ? checklist.title : (name || 'No checklist selected'));
      const tag = document.getElementById('checklistTag');
      tag.className = 'checklist-tag';
      tag.textContent = correct == null ? 'Standby' : (correct ? 'Matched' : 'Mismatch');
      if (correct === true) {
        tag.classList.add('good');
      }
      if (correct === false) {
        tag.classList.add('bad');
      }
      const list = document.getElementById('checklistItems');
      list.innerHTML = '';
      items.forEach((item) => {
        const node = document.createElement('li');
        node.textContent = item;
        list.appendChild(node);
      });
    }

    function setFeedback(id, message, className) {
      const node = document.getElementById(id);
      node.className = 'feedback';
      if (className) {
        node.classList.add(className);
      }
      node.textContent = message;
    }

    async function applyScenario(key) {
      state.activeScenario = key;
      state.selectedChecklist = '';
      text('scenarioState', key ? scenarios[key].title : 'Manual');
      document.getElementById('checklistSelect').value = '';
      populateChecklistSelect(key);
      if (!key) {
        renderScenarioMetadata(null);
        setFeedback('scenarioFeedback', 'Manual failure configuration active.', '');
        setFeedback('checklistFeedback', 'Choose a scenario to start checklist recognition practice.', '');
        renderChecklist('', null);
	      await saveScenarioSelection('');
        return;
      }

      const scenario = scenarios[key];
      renderScenarioMetadata(scenario);
      setFeedback('scenarioFeedback', scenario.description, 'good');
      renderChecklist('', null);
      setFeedback('checklistFeedback', 'Identify the failure from the panel, then choose the correct checklist.', '');
      syncFailureControls(Object.keys(scenario.failures).filter((name) => scenario.failures[name]));
      await saveScenarioSelection(key);
    }

    function renderChecklistSelection() {
      const scenario = scenarios[state.activeScenario];
      const selected = state.selectedChecklist;
      if (!selected) {
        renderChecklist('', null);
        setFeedback('checklistFeedback', 'The correct procedure is not auto-opened. Recognition is part of the exercise.', '');
        return;
      }
      const correct = scenario && selected === scenario.correct_checklist;
      renderChecklist(selected, correct);
      if (correct) {
        setFeedback('checklistFeedback', 'Correct checklist selected. Keep flying the scenario while you work the procedure.', 'good');
        return;
      }
      setFeedback('checklistFeedback', 'Selected checklist does not match the observed failure. Re-evaluate the panel and choose again.', 'bad');
    }

    function warningAudioConfig(warning) {
      const config = {
        sink_rate: { text: 'SINK RATE', intervalMs: 1800, rate: 0.95 },
        pull_up: { text: 'PULL UP', intervalMs: 1100, rate: 1.05 },
      };
      return config[warning || ''] || null;
    }

    function setAuralAlertsEnabled(enabled) {
      state.audio.muted = !enabled;
      document.getElementById('auralAlertsEnabled').checked = enabled;
      if (typeof window !== 'undefined' && window.localStorage) {
        window.localStorage.setItem('mockflight.auralAlertsEnabled', enabled ? 'true' : 'false');
      }
      updateAuralAlertStatus();
      if (!enabled) {
        stopAuralAlerts();
        return;
      }
      const currentWarning = state.snapshot && state.snapshot.controls ? state.snapshot.controls.warning : '';
      syncAuralAlerts(currentWarning || '');
    }

    function updateAuralAlertStatus() {
      const node = document.getElementById('auralAlertStatus');
      node.className = 'aural-status';
      if (!state.audio.supported) {
        node.classList.add('unavailable');
        node.textContent = 'Aural alerts unavailable';
        return;
      }
      if (state.audio.muted) {
        node.classList.add('muted');
        node.textContent = 'Aural alerts muted';
        return;
      }
      node.textContent = state.audio.unlocked ? 'Aural alerts armed' : 'Aural alerts standby';
    }

    function hydrateAuralAlertControls() {
      let enabled = true;
      if (typeof window !== 'undefined' && window.localStorage) {
        enabled = window.localStorage.getItem('mockflight.auralAlertsEnabled') !== 'false';
      }
      setAuralAlertsEnabled(enabled);
    }

    function stopAuralAlerts() {
      if (state.audio.intervalId !== null) {
        window.clearInterval(state.audio.intervalId);
        state.audio.intervalId = null;
      }
      state.audio.activeWarning = '';
      if (state.audio.supported) {
        window.speechSynthesis.cancel();
      }
    }

    function speakAuralAlert(config) {
      if (!state.audio.unlocked || !state.audio.supported || state.audio.muted) {
        return;
      }
      window.speechSynthesis.cancel();
      const utterance = new SpeechSynthesisUtterance(config.text);
      utterance.rate = config.rate;
      utterance.pitch = 0.82;
      utterance.volume = 1;
      window.speechSynthesis.speak(utterance);
    }

    function syncAuralAlerts(warning) {
      const config = warningAudioConfig(warning);
      if (!config) {
        stopAuralAlerts();
        return;
      }
      if (state.audio.muted) {
        stopAuralAlerts();
        return;
      }
      if (state.audio.activeWarning === warning && state.audio.intervalId !== null) {
        return;
      }
      stopAuralAlerts();
      state.audio.activeWarning = warning;
      if (!state.audio.unlocked || !state.audio.supported) {
        return;
      }
      speakAuralAlert(config);
      state.audio.intervalId = window.setInterval(() => {
        if (state.audio.activeWarning !== warning) {
          return;
        }
        speakAuralAlert(config);
      }, config.intervalMs);
    }

    function unlockAuralAlerts() {
      if (state.audio.unlocked) {
        return;
      }
      state.audio.unlocked = true;
      updateAuralAlertStatus();
      const currentWarning = state.snapshot && state.snapshot.controls ? state.snapshot.controls.warning : '';
      syncAuralAlerts(currentWarning || '');
    }

    function warningText(warning) {
      const labels = {
        '': 'No takeoff warnings',
        below_vr_rotate: 'Below Vr rotate',
        low_energy: 'Low energy',
        sink_rate: 'Sink rate',
        stall: 'Stall warning',
        pull_up: 'Pull up',
        crash: 'Crash state',
      };
      return labels[warning || ''] || warning.replaceAll('_', ' ');
    }

    function syncControlPanel(controls) {
      document.getElementById('simMode').value = controls.mode || 'full';
      document.getElementById('throttleInput').value = Number(controls.throttle || 0);
      text('throttleReadout', 'Throttle ' + Math.round(Number(controls.throttle || 0) * 100) + '%');
      text('v1Value', Math.round(Number(controls.v1 || 0)) + ' kt');
      text('vrValue', Math.round(Number(controls.vr || 0)) + ' kt');
      text('v2Value', Math.round(Number(controls.v2 || 0)) + ' kt');
      const warningNode = document.getElementById('warningBanner');
      warningNode.className = 'warning-banner';
      if (controls.warning === 'crash') {
        warningNode.classList.add('crash');
      }
      warningNode.textContent = warningText(controls.warning || '');
      document.getElementById('rotateButton').disabled = !!controls.crashed;
      document.getElementById('togaButton').disabled = !!controls.crashed;
      document.getElementById('autopilotMaster').checked = !!(controls.autopilot && controls.autopilot.engaged);
      document.getElementById('headingHold').checked = !!(controls.autopilot && controls.autopilot.heading_hold);
      document.getElementById('altitudeHold').checked = !!(controls.autopilot && controls.autopilot.altitude_hold);
      document.getElementById('verticalSpeedMode').checked = !!(controls.autopilot && controls.autopilot.vertical_speed_mode);
      document.getElementById('headingTarget').value = Math.round(Number(controls.selected_heading || 270));
      document.getElementById('altitudeTarget').value = Math.round(Number(controls.selected_altitude || 1800));
      document.getElementById('verticalSpeedTarget').value = Math.round(Number(controls.selected_vertical_speed || 700));
    }

    function renderPFD(model) {
      text('phase', model.phaseLabel);
      text('timestamp', model.timestamp);
      text('airspeedReadout', Math.round(model.airspeed) + ' KT');
      text('altitudeReadout', Math.round(model.altitude) + ' FT');
      text('verticalSpeed', format(model.verticalSpeed, 'm/s', 2));
      text('pitot', format(model.pitot, 'inHg', 3));
      text('pitchValue', format(model.pitch, 'deg', 1));
      text('headingValue', String(Math.round(model.heading)).padStart(3, '0') + ' deg');
      text('rollValue', format(model.roll, 'deg', 1));
      document.getElementById('horizon').style.transform = 'translateY(' + (model.pitch * 3) + 'px) rotate(' + (-model.roll) + 'deg)';
      document.getElementById('airspeedTrend').style.height = Math.max(12, Math.min(180, Math.abs(model.airspeed - 90) * 1.4)) + 'px';
      document.getElementById('altitudeTrend').style.height = Math.max(12, Math.min(200, Math.abs(model.verticalSpeed) * 28)) + 'px';
    }

    function renderNav(model) {
      text('navSource', model.navSource);
      text('selectedCourse', model.selectedCourse);
      text('approachState', model.approachState);
      document.getElementById('cdiBar').style.left = 'calc(50% + ' + model.cdiOffset + 'px)';
    }

    function renderRawSensors(snapshot) {
      text('pitotValue', format(snapshot.pitot.value, snapshot.pitot.unit, 3));
      text('staticValue', format(snapshot.static_air.value, snapshot.static_air.unit, 3));
      text('altitudeValue', format(snapshot.altimeter.value, snapshot.altimeter.unit, 0));
      text('verticalValue', format(snapshot.vertical_speed.value, snapshot.vertical_speed.unit, 2));
      text('accelX', format(snapshot.accelerometer.x, snapshot.accelerometer.unit, 2));
      text('accelY', format(snapshot.accelerometer.y, snapshot.accelerometer.unit, 2));
      text('accelZ', format(snapshot.accelerometer.z, snapshot.accelerometer.unit, 2));
      text('gyroZ', format(snapshot.gyroscope.z, snapshot.gyroscope.unit, 2));
      text('magX', format(snapshot.magnetometer.x, snapshot.magnetometer.unit, 2));
      text('magY', format(snapshot.magnetometer.y, snapshot.magnetometer.unit, 2));
    }

    function renderTakeoffControls(model) {
      syncControlPanel(model.controls);
      text('scenarioState', model.controls.crashed ? 'Crash' : (state.activeScenario ? scenarios[state.activeScenario].title : 'Manual'));
    }

    async function refresh() {
      try {
        const response = await fetch('/api/snapshot', { cache: 'no-store' });
        if (!response.ok) {
          throw new Error('snapshot unavailable');
        }
        const data = await response.json();
        state.snapshot = data;
        setConnectionState(true);
        const model = toDisplayModel(data);
        renderAnnunciators(model.activeFailures);
        renderFailures(model.activeFailures);
        if (!state.savingFailures) {
          syncFailureControls(model.activeFailures);
        }
        renderPFD(model);
        renderNav(model);
        renderRawSensors(data);
        renderTakeoffControls(model);
        syncAuralAlerts(model.controls.warning || '');
      } catch (error) {
        setConnectionState(false);
        stopAuralAlerts();
      }
    }

    populateScenarioSelect();
    renderScenarioMetadata(null);
    hydrateAuralAlertControls();
    document.addEventListener('pointerdown', unlockAuralAlerts, { once: true });
    document.addEventListener('keydown', unlockAuralAlerts, { once: true });
    refresh();
    setInterval(refresh, 750);

    document.getElementById('auralAlertsEnabled').addEventListener('change', (event) => {
      setAuralAlertsEnabled(event.target.checked);
    });

    document.getElementById('scenarioSelect').addEventListener('change', async (event) => {
      await applyScenario(event.target.value);
    });

    document.getElementById('checklistSelect').addEventListener('change', (event) => {
      state.selectedChecklist = event.target.value;
      renderChecklistSelection();
    });

    document.getElementById('simMode').addEventListener('change', async (event) => {
      await postControls({ mode: event.target.value });
      await postControls({ reset: true });
      await refresh();
    });

    document.getElementById('throttleInput').addEventListener('input', (event) => {
      text('throttleReadout', 'Throttle ' + Math.round(Number(event.target.value) * 100) + '%');
    });

    document.getElementById('throttleInput').addEventListener('change', async (event) => {
      await postControls({ throttle: Number(event.target.value) });
      await refresh();
    });

    document.getElementById('togaButton').addEventListener('click', async () => {
      await postControls({ toga: true });
      await refresh();
    });

    document.getElementById('rotateButton').addEventListener('click', async () => {
      await postControls({ rotate: true });
      await refresh();
    });

    document.getElementById('resetButton').addEventListener('click', async () => {
      await postControls({ reset: true });
      await refresh();
    });

    document.getElementById('autopilotMaster').addEventListener('change', async (event) => {
      await postControls({ autopilot_engaged: event.target.checked });
      await refresh();
    });

    document.getElementById('headingHold').addEventListener('change', async (event) => {
      await postControls({ heading_hold: event.target.checked });
      await refresh();
    });

    document.getElementById('altitudeHold').addEventListener('change', async (event) => {
      await postControls({ altitude_hold: event.target.checked });
      await refresh();
    });

    document.getElementById('verticalSpeedMode').addEventListener('change', async (event) => {
      await postControls({ vertical_speed_mode: event.target.checked });
      await refresh();
    });

    document.getElementById('headingTarget').addEventListener('change', async (event) => {
      await postControls({ selected_heading: Number(event.target.value) });
      await refresh();
    });

    document.getElementById('altitudeTarget').addEventListener('change', async (event) => {
      await postControls({ selected_altitude: Number(event.target.value) });
      await refresh();
    });

    document.getElementById('verticalSpeedTarget').addEventListener('change', async (event) => {
      await postControls({ selected_vertical_speed: Number(event.target.value) });
      await refresh();
    });

    document.querySelectorAll('.failure-controls input').forEach((node) => {
      node.addEventListener('change', async () => {
        state.activeScenario = '';
        document.getElementById('scenarioSelect').value = '';
        text('scenarioState', 'Manual');
        populateChecklistSelect('');
        renderScenarioMetadata(null);
        setFeedback('scenarioFeedback', 'Manual failure configuration active.', '');
        await saveFailures();
      });
    });
  </script>
</body>
</html>`))
