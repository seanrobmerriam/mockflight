package dashboard

import (
	"encoding/json"
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

func TestDashboardPageRendersShell(t *testing.T) {
	handler := NewHandler(stubProvider{snapshot: sim.Snapshot{Phase: "ground", Timestamp: time.Now().UTC()}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "MockFlight Sensor Deck") {
		t.Fatalf("expected dashboard title in body: %s", body)
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
