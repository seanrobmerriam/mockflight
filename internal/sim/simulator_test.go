package sim

import (
	"math"
	"testing"
	"time"
)

func TestSnapshotIncludesAllSensors(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	snapshot := simulator.Snapshot()

	if snapshot.Phase == "" {
		t.Fatal("expected a flight phase")
	}

	if snapshot.Altimeter.Name != "altimeter" {
		t.Fatalf("expected altimeter sensor, got %q", snapshot.Altimeter.Name)
	}

	if snapshot.Airspeed.Name != "airspeed" {
		t.Fatalf("expected airspeed sensor, got %q", snapshot.Airspeed.Name)
	}

	if snapshot.Accelerometer.Name != "accelerometer" {
		t.Fatalf("expected accelerometer sensor, got %q", snapshot.Accelerometer.Name)
	}

	if snapshot.Gyroscope.Name != "gyroscope" {
		t.Fatalf("expected gyroscope sensor, got %q", snapshot.Gyroscope.Name)
	}

	if snapshot.Magnetometer.Name != "magnetometer" {
		t.Fatalf("expected magnetometer sensor, got %q", snapshot.Magnetometer.Name)
	}

	if snapshot.AHRS.Name != "ahrs" {
		t.Fatalf("expected AHRS sensor, got %q", snapshot.AHRS.Name)
	}

	if snapshot.Pitot.Value < snapshot.StaticAir.Value {
		t.Fatalf("expected pitot pressure %.3f to be >= static pressure %.3f", snapshot.Pitot.Value, snapshot.StaticAir.Value)
	}

	if snapshot.Timestamp.IsZero() {
		t.Fatal("expected snapshot timestamp")
	}
}

func TestAdvanceProgressesFlightState(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	initial := simulator.Snapshot()
	performTakeoff(t, simulator)

	simulator.Advance(20 * time.Second)
	after := simulator.Snapshot()

	if after.Phase == initial.Phase {
		t.Fatalf("expected phase to change from %q", initial.Phase)
	}

	if after.Altimeter.Value <= initial.Altimeter.Value {
		t.Fatalf("expected altitude to increase from %.1f to > %.1f", initial.Altimeter.Value, initial.Altimeter.Value)
	}

	if after.Airspeed.Indicated <= initial.Airspeed.Indicated {
		t.Fatalf("expected airspeed to increase from %.1f to > %.1f", initial.Airspeed.Indicated, initial.Airspeed.Indicated)
	}

	if after.AHRS.Heading < 0 || after.AHRS.Heading >= 360 {
		t.Fatalf("expected normalized heading, got %.1f", after.AHRS.Heading)
	}
}

func TestAltimeterFollowsMeasuredStaticPressure(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	performTakeoff(t, simulator)
	simulator.Advance(18 * time.Second)
	snapshot := simulator.Snapshot()

	calculatedAltitude := altitudeFromPressure(snapshot.StaticAir.Value)
	if diff := math.Abs(snapshot.Altimeter.Value - calculatedAltitude); diff > 5 {
		t.Fatalf("expected barometric altitude %.1f to match static pressure altitude %.1f within 5 ft", snapshot.Altimeter.Value, calculatedAltitude)
	}
}

func TestVerticalSpeedRespondsWithLag(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	performTakeoff(t, simulator)
	simulator.Advance(4 * time.Second)

	truth := simulator.stateAt(simulator.elapsed)
	snapshot := simulator.Snapshot()

	if snapshot.VerticalSpeed.Value <= 0 {
		t.Fatalf("expected positive VSI reading during takeoff, got %.2f", snapshot.VerticalSpeed.Value)
	}

	if snapshot.VerticalSpeed.Value >= truth.verticalSpeedMPS {
		t.Fatalf("expected lagged VSI reading %.2f to remain below true climb rate %.2f", snapshot.VerticalSpeed.Value, truth.verticalSpeedMPS)
	}
}

func TestAHRSIsEstimatedFromSensors(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	performTakeoff(t, simulator)
	simulator.ApplyControls(ControlCommand{AutopilotEngaged: ptrBool(true), HeadingHold: ptrBool(true), SelectedHeading: ptrFloat(320)})
	simulator.Advance(20*time.Second + simulator.TickInterval())

	truth := simulator.stateAt(simulator.elapsed)
	snapshot := simulator.Snapshot()

	rollDiff := math.Abs(snapshot.AHRS.Roll - truth.rollDeg)
	headingDiff := math.Abs(angleDelta(snapshot.AHRS.Heading, truth.headingDeg))

	if rollDiff < 0.05 && headingDiff < 0.05 {
		t.Fatalf("expected AHRS estimate to lag truth during maneuver, got roll diff %.3f and heading diff %.3f", rollDiff, headingDiff)
	}

	if headingDiff > 60 {
		t.Fatalf("expected AHRS heading to stay reasonably close to truth, got diff %.3f", headingDiff)
	}
}

func TestPitotBlockedFreezesPitotPressure(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{PitotBlocked: true}})
	initial := simulator.Snapshot()
	simulator.Advance(90 * time.Second)
	after := simulator.Snapshot()

	if after.Pitot.Value != initial.Pitot.Value {
		t.Fatalf("expected blocked pitot pressure %.3f to stay frozen, got %.3f", initial.Pitot.Value, after.Pitot.Value)
	}

	if len(after.ActiveFailures) == 0 {
		t.Fatal("expected active failures to be reported")
	}
}

func TestStaticPortBlockedFreezesAltimeterAndVSISettles(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{StaticPortBlocked: true}})
	initial := simulator.Snapshot()
	simulator.Advance(90 * time.Second)
	after := simulator.Snapshot()

	if after.StaticAir.Value != initial.StaticAir.Value {
		t.Fatalf("expected blocked static pressure %.3f to stay frozen, got %.3f", initial.StaticAir.Value, after.StaticAir.Value)
	}

	if after.Altimeter.Value != initial.Altimeter.Value {
		t.Fatalf("expected blocked altimeter %.1f to stay frozen, got %.1f", initial.Altimeter.Value, after.Altimeter.Value)
	}

	if math.Abs(after.VerticalSpeed.Value) > 0.05 {
		t.Fatalf("expected blocked VSI to settle near zero, got %.3f", after.VerticalSpeed.Value)
	}
}

func TestSetFailuresUpdatesActiveFailuresAtRuntime(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.Advance(40 * time.Second)
	before := simulator.Snapshot()

	simulator.SetFailures(FailureConfig{PitotBlocked: true, GyroSaturation: true})
	simulator.Advance(20 * time.Second)
	after := simulator.Snapshot()

	if len(after.ActiveFailures) != 2 {
		t.Fatalf("expected two active failures after runtime update, got %#v", after.ActiveFailures)
	}

	if after.Pitot.Value != before.Pitot.Value {
		t.Fatalf("expected pitot reading to freeze at runtime from %.3f, got %.3f", before.Pitot.Value, after.Pitot.Value)
	}
}

func TestPitotDrainBlockedBleedsOffDynamicPressure(t *testing.T) {
	normal := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	performTakeoff(t, normal)
	normal.Advance(20 * time.Second)
	normalSnapshot := normal.Snapshot()

	faulted := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{PitotDrainBlocked: true}})
	performTakeoff(t, faulted)
	faulted.Advance(20 * time.Second)
	faultedSnapshot := faulted.Snapshot()

	if faultedSnapshot.Airspeed.Indicated >= normalSnapshot.Airspeed.Indicated-18 {
		t.Fatalf("expected pitot drain blockage to materially reduce IAS, normal %.1f faulted %.1f", normalSnapshot.Airspeed.Indicated, faultedSnapshot.Airspeed.Indicated)
	}
}

func TestStaticLeakBiasesPressureAltitude(t *testing.T) {
	normal := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	performTakeoff(t, normal)
	normal.Advance(18 * time.Second)
	normalSnapshot := normal.Snapshot()

	leaky := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{StaticLeak: true}})
	performTakeoff(t, leaky)
	leaky.Advance(18 * time.Second)
	leakySnapshot := leaky.Snapshot()

	if math.Abs(leakySnapshot.Altimeter.Value-normalSnapshot.Altimeter.Value) < 150 {
		t.Fatalf("expected static leak to bias altitude materially, normal %.1f leaky %.1f", normalSnapshot.Altimeter.Value, leakySnapshot.Altimeter.Value)
	}
}

func TestGyroSaturationClampsMeasuredRates(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{GyroSaturation: true}})
	performTakeoff(t, simulator)
	simulator.ApplyControls(ControlCommand{AutopilotEngaged: ptrBool(true), HeadingHold: ptrBool(true), SelectedHeading: ptrFloat(335)})
	simulator.Advance(16*time.Second + simulator.TickInterval())
	snapshot := simulator.Snapshot()
	truth := simulator.stateAt(simulator.elapsed)

	if math.Abs(snapshot.Gyroscope.Z) > gyroSaturationLimitDegPerSec {
		t.Fatalf("expected saturated gyro yaw rate within %.1f deg/s, got %.2f", gyroSaturationLimitDegPerSec, snapshot.Gyroscope.Z)
	}

	if math.Abs(truth.yawRateDegPerSec) <= gyroSaturationLimitDegPerSec {
		t.Fatalf("expected truth yaw rate to exceed saturation threshold for this test, got %.2f", truth.yawRateDegPerSec)
	}
}

func TestSimulatorStartsIdleOnRunway(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	snapshot := simulator.Snapshot()

	if snapshot.Phase != "runway_idle" {
		t.Fatalf("expected runway_idle phase, got %q", snapshot.Phase)
	}

	if snapshot.Airspeed.Indicated > 1 {
		t.Fatalf("expected idle runway airspeed near zero, got %.1f", snapshot.Airspeed.Indicated)
	}

	if snapshot.Controls.Throttle != 0 {
		t.Fatalf("expected idle throttle, got %.2f", snapshot.Controls.Throttle)
	}

	if snapshot.Controls.Airborne {
		t.Fatal("expected simulator to start on the runway")
	}
}

func TestThrottleAdvancesTakeoffRoll(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.ApplyControls(ControlCommand{Throttle: ptrFloat(0.85)})
	before := simulator.Snapshot()

	simulator.Advance(12 * time.Second)
	after := simulator.Snapshot()

	if after.Airspeed.Indicated <= before.Airspeed.Indicated+20 {
		t.Fatalf("expected throttle to materially increase airspeed from %.1f to > %.1f, got %.1f", before.Airspeed.Indicated, before.Airspeed.Indicated+20, after.Airspeed.Indicated)
	}

	if after.Phase != "takeoff_roll" {
		t.Fatalf("expected takeoff_roll phase after throttle advance, got %q", after.Phase)
	}
}

func TestRotateBeforeVrTriggersTakeoffWarning(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.ApplyControls(ControlCommand{Throttle: ptrFloat(0.65)})
	simulator.Advance(3 * time.Second)
	simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})
	simulator.Advance(2 * time.Second)
	snapshot := simulator.Snapshot()

	if snapshot.Controls.Warning != WarningBelowVrRotate {
		t.Fatalf("expected below Vr rotate warning, got %q", snapshot.Controls.Warning)
	}

	if snapshot.Controls.Airborne {
		t.Fatal("expected early rotate to remain on the runway")
	}
}

func TestRotateNearVrProducesLiftoff(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.ApplyControls(ControlCommand{TOGA: ptrBool(true)})
	advanceToVr(t, simulator, 80)

	simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})
	simulator.Advance(6 * time.Second)
	snapshot := simulator.Snapshot()

	if !snapshot.Controls.Airborne {
		t.Fatal("expected liftoff after rotating at or above Vr")
	}

	if snapshot.Altimeter.Value <= fieldElevationFeet {
		t.Fatalf("expected positive climb above field elevation, got %.1f", snapshot.Altimeter.Value)
	}
}

func TestUnsafeTakeoffCanReachCrashState(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.ApplyControls(ControlCommand{Throttle: ptrFloat(0.4)})
	simulator.Advance(2 * time.Second)
	simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})

	for step := 0; step < 800; step++ {
		simulator.Advance(simulator.TickInterval())
		if simulator.Snapshot().Controls.Crashed {
			break
		}
	}

	snapshot := simulator.Snapshot()
	if !snapshot.Controls.Crashed {
		t.Fatal("expected unrecovered unsafe takeoff to reach crash state")
	}

	if snapshot.Controls.Warning != WarningCrash {
		t.Fatalf("expected crash warning state, got %q", snapshot.Controls.Warning)
	}
}

func TestTakeoffEnergyCatalogScenariosRemainSupportedBySimulator(t *testing.T) {
	t.Run("earlyRotate", func(t *testing.T) {
		simulator := newScenarioSimulator(ModeFull, FailureConfig{})
		simulator.ApplyControls(ControlCommand{Throttle: ptrFloat(0.65)})
		simulator.Advance(3 * time.Second)
		simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})
		simulator.Advance(2 * time.Second)

		snapshot := simulator.Snapshot()
		if snapshot.Controls.Warning != WarningBelowVrRotate {
			t.Fatalf("expected earlyRotate sequence to trigger %q, got %q", WarningBelowVrRotate, snapshot.Controls.Warning)
		}
		if snapshot.Controls.Airborne {
			t.Fatal("expected earlyRotate sequence to remain on the runway")
		}
	})

	t.Run("lowEnergyClimb", func(t *testing.T) {
		simulator := newScenarioSimulator(ModeTransitional, FailureConfig{})
		simulator.ApplyControls(ControlCommand{Throttle: ptrFloat(0.72)})
		simulator.Advance(15 * time.Second)

		setupSnapshot := simulator.Snapshot()
		if setupSnapshot.Controls.Crashed {
			t.Fatal("expected lowEnergyClimb setup to remain recoverable before rotation")
		}

		simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})
		airborne := advanceUntil(t, simulator, 12*time.Second, func(snapshot Snapshot) bool {
			return snapshot.Controls.Airborne
		})

		if !airborne.Controls.Airborne {
			t.Fatal("expected lowEnergyClimb sequence to produce a simulated climb state after rotation")
		}
		if airborne.Controls.Crashed {
			t.Fatal("expected lowEnergyClimb sequence to avoid immediate crash during setup")
		}
	})

	t.Run("runwayOverrun", func(t *testing.T) {
		simulator := newScenarioSimulator(ModeFull, FailureConfig{})
		simulator.ApplyControls(ControlCommand{Throttle: ptrFloat(0.4)})
		simulator.Advance(12 * time.Second)

		setupSnapshot := simulator.Snapshot()
		if setupSnapshot.Controls.Airborne {
			t.Fatal("expected runwayOverrun setup to remain on the runway")
		}
		if setupSnapshot.Controls.RunwayRemaining >= runwayLengthFeet {
			t.Fatalf("expected runwayOverrun setup to consume runway, remaining %.1f", setupSnapshot.Controls.RunwayRemaining)
		}

		simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})
		crashed := advanceUntil(t, simulator, 200*time.Second, func(snapshot Snapshot) bool {
			return snapshot.Controls.Crashed
		})

		if !crashed.Controls.Crashed {
			t.Fatal("expected runwayOverrun sequence to be able to reach crash state if the takeoff continues")
		}
		if crashed.Controls.Warning != WarningCrash {
			t.Fatalf("expected runwayOverrun crash sequence to end in %q, got %q", WarningCrash, crashed.Controls.Warning)
		}
	})
}

func TestScenarioPartialPanelMissedUsesExistingGyroAndMagnetometerFailures(t *testing.T) {
	normal := newScenarioSimulator(ModeFull, FailureConfig{})
	faulted := newScenarioSimulator(ModeFull, FailureConfig{MagnetometerDisturbed: true, GyroSaturation: true})

	configurePartialPanelMissedManeuver(t, normal)
	configurePartialPanelMissedManeuver(t, faulted)

	normalSnapshot := normal.Snapshot()
	faultedSnapshot := faulted.Snapshot()

	if !containsFailure(faultedSnapshot.ActiveFailures, "magnetometer_disturbed") {
		t.Fatalf("expected partialPanelMissed sequence to report magnetometer_disturbed, got %#v", faultedSnapshot.ActiveFailures)
	}
	if !containsFailure(faultedSnapshot.ActiveFailures, "gyro_saturation") {
		t.Fatalf("expected partialPanelMissed sequence to report gyro_saturation, got %#v", faultedSnapshot.ActiveFailures)
	}
	if math.Abs(faultedSnapshot.Gyroscope.Z) > gyroSaturationLimitDegPerSec {
		t.Fatalf("expected gyro saturation to clamp yaw rate within %.1f deg/s, got %.2f", gyroSaturationLimitDegPerSec, faultedSnapshot.Gyroscope.Z)
	}
	if magnetometerDelta(normalSnapshot, faultedSnapshot) < 12 {
		t.Fatalf("expected magnetometer disturbance to materially change the measured field, normal=%#v faulted=%#v", normalSnapshot.Magnetometer, faultedSnapshot.Magnetometer)
	}
}

func TestHeadingModeTurnsTowardSelectedHeading(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.ApplyControls(ControlCommand{TOGA: ptrBool(true)})
	advanceToVr(t, simulator, 120)
	simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})

	simulator.Advance(10 * time.Second)
	before := simulator.Snapshot()
	simulator.ApplyControls(ControlCommand{
		AutopilotEngaged: ptrBool(true),
		HeadingHold:      ptrBool(true),
		SelectedHeading:  ptrFloat(315),
	})
	simulator.Advance(20 * time.Second)
	after := simulator.Snapshot()

	if math.Abs(angleDelta(before.AHRS.Heading, after.AHRS.Heading)) < 10 {
		t.Fatalf("expected heading mode to turn aircraft materially, before %.1f after %.1f", before.AHRS.Heading, after.AHRS.Heading)
	}

	if math.Abs(angleDelta(after.AHRS.Heading, 315)) > 35 {
		t.Fatalf("expected heading to trend toward selected heading, got %.1f", after.AHRS.Heading)
	}
}

func ptrBool(value bool) *bool {
	return &value
}

func ptrFloat(value float64) *float64 {
	return &value
}

func ptrMode(value Mode) *Mode {
	return &value
}

func newScenarioSimulator(mode Mode, failures FailureConfig) *Simulator {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: failures})
	if mode == ModeFull {
		return simulator
	}

	simulator.ApplyControls(ControlCommand{Mode: ptrMode(mode)})
	simulator.ApplyControls(ControlCommand{Reset: ptrBool(true)})
	return simulator
}

func advanceUntil(t *testing.T, simulator *Simulator, limit time.Duration, condition func(Snapshot) bool) Snapshot {
	t.Helper()

	remaining := limit
	for remaining > 0 {
		snapshot := simulator.Snapshot()
		if condition(snapshot) {
			return snapshot
		}

		step := simulator.TickInterval()
		if remaining < step {
			step = remaining
		}
		simulator.Advance(step)
		remaining -= step
	}

	snapshot := simulator.Snapshot()
	if condition(snapshot) {
		return snapshot
	}

	return snapshot
}

func advanceToVr(t *testing.T, simulator *Simulator, maxSteps int) Snapshot {
	t.Helper()

	for step := 0; step < maxSteps; step++ {
		snapshot := simulator.Snapshot()
		truth := simulator.stateAt(simulator.elapsed)
		if truth.airspeedKTS >= snapshot.Controls.Vr {
			return snapshot
		}
		simulator.Advance(simulator.TickInterval())
	}

	snapshot := simulator.Snapshot()
	truth := simulator.stateAt(simulator.elapsed)
	t.Fatalf("expected to reach Vr before rotate, true airspeed %.1f IAS %.1f Vr %.1f phase %q", truth.airspeedKTS, snapshot.Airspeed.Indicated, snapshot.Controls.Vr, snapshot.Phase)
	return Snapshot{}
}

func configurePartialPanelMissedManeuver(t *testing.T, simulator *Simulator) {
	t.Helper()
	performTakeoff(t, simulator)
	simulator.ApplyControls(ControlCommand{AutopilotEngaged: ptrBool(true), HeadingHold: ptrBool(true), SelectedHeading: ptrFloat(335)})
	simulator.Advance(16*time.Second + simulator.TickInterval())
}

func containsFailure(active []string, target string) bool {
	for _, failure := range active {
		if failure == target {
			return true
		}
	}

	return false
}

func magnetometerDelta(normal Snapshot, faulted Snapshot) float64 {
	return math.Abs(normal.Magnetometer.X-faulted.Magnetometer.X) +
		math.Abs(normal.Magnetometer.Y-faulted.Magnetometer.Y) +
		math.Abs(normal.Magnetometer.Z-faulted.Magnetometer.Z)
}

func performTakeoff(t *testing.T, simulator *Simulator) {
	t.Helper()
	simulator.ApplyControls(ControlCommand{TOGA: ptrBool(true)})
	advanceToVr(t, simulator, 160)
	simulator.ApplyControls(ControlCommand{Rotate: ptrBool(true)})

	for step := 0; step < 120; step++ {
		simulator.Advance(simulator.TickInterval())
		if simulator.Snapshot().Controls.Airborne {
			return
		}
	}

	t.Fatal("expected takeoff helper to reach airborne state")
}
