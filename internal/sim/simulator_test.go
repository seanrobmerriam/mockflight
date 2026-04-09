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

	simulator.Advance(75 * time.Second)
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
	simulator.Advance(95 * time.Second)
	snapshot := simulator.Snapshot()

	calculatedAltitude := altitudeFromPressure(snapshot.StaticAir.Value)
	if diff := math.Abs(snapshot.Altimeter.Value - calculatedAltitude); diff > 5 {
		t.Fatalf("expected barometric altitude %.1f to match static pressure altitude %.1f within 5 ft", snapshot.Altimeter.Value, calculatedAltitude)
	}
}

func TestVerticalSpeedRespondsWithLag(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	simulator.Advance(21 * time.Second)

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
	simulator.Advance(80*time.Second + simulator.TickInterval())

	truth := simulator.stateAt(simulator.elapsed)
	snapshot := simulator.Snapshot()

	rollDiff := math.Abs(snapshot.AHRS.Roll - truth.rollDeg)
	headingDiff := math.Abs(angleDelta(snapshot.AHRS.Heading, truth.headingDeg))

	if rollDiff < 0.05 && headingDiff < 0.05 {
		t.Fatalf("expected AHRS estimate to lag truth during maneuver, got roll diff %.3f and heading diff %.3f", rollDiff, headingDiff)
	}

	if headingDiff > 20 {
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
	normal.Advance(100 * time.Second)
	normalSnapshot := normal.Snapshot()

	faulted := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{PitotDrainBlocked: true}})
	faulted.Advance(100 * time.Second)
	faultedSnapshot := faulted.Snapshot()

	if faultedSnapshot.Airspeed.Indicated >= normalSnapshot.Airspeed.Indicated-20 {
		t.Fatalf("expected pitot drain blockage to materially reduce IAS, normal %.1f faulted %.1f", normalSnapshot.Airspeed.Indicated, faultedSnapshot.Airspeed.Indicated)
	}
}

func TestStaticLeakBiasesPressureAltitude(t *testing.T) {
	normal := NewSimulator(Config{Seed: 7, NoiseScale: 0})
	normal.Advance(90 * time.Second)
	normalSnapshot := normal.Snapshot()

	leaky := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{StaticLeak: true}})
	leaky.Advance(90 * time.Second)
	leakySnapshot := leaky.Snapshot()

	if math.Abs(leakySnapshot.Altimeter.Value-normalSnapshot.Altimeter.Value) < 150 {
		t.Fatalf("expected static leak to bias altitude materially, normal %.1f leaky %.1f", normalSnapshot.Altimeter.Value, leakySnapshot.Altimeter.Value)
	}
}

func TestGyroSaturationClampsMeasuredRates(t *testing.T) {
	simulator := NewSimulator(Config{Seed: 7, NoiseScale: 0, Failures: FailureConfig{GyroSaturation: true}})
	simulator.Advance(80*time.Second + simulator.TickInterval())
	snapshot := simulator.Snapshot()
	truth := simulator.stateAt(simulator.elapsed)

	if math.Abs(snapshot.Gyroscope.Z) > gyroSaturationLimitDegPerSec {
		t.Fatalf("expected saturated gyro yaw rate within %.1f deg/s, got %.2f", gyroSaturationLimitDegPerSec, snapshot.Gyroscope.Z)
	}

	if math.Abs(truth.yawRateDegPerSec) <= gyroSaturationLimitDegPerSec {
		t.Fatalf("expected truth yaw rate to exceed saturation threshold for this test, got %.2f", truth.yawRateDegPerSec)
	}
}
