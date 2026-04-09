package sim

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type Config struct {
	Seed       int64
	NoiseScale float64
	StartTime  time.Time
	Failures   FailureConfig
}

type FailureConfig struct {
	PitotBlocked          bool `json:"pitot_blocked"`
	PitotDrainBlocked     bool `json:"pitot_drain_blocked"`
	StaticPortBlocked     bool `json:"static_port_blocked"`
	StaticLeak            bool `json:"static_leak"`
	MagnetometerDisturbed bool `json:"magnetometer_disturbed"`
	GyroSaturation        bool `json:"gyro_saturation"`
}

type Mode string

const (
	ModeFull         Mode = "full"
	ModeTransitional Mode = "transitional"
)

type WarningState string

const (
	WarningNone          WarningState = ""
	WarningBelowVrRotate WarningState = "below_vr_rotate"
	WarningLowEnergy     WarningState = "low_energy"
	WarningSinkRate      WarningState = "sink_rate"
	WarningStall         WarningState = "stall"
	WarningPullUp        WarningState = "pull_up"
	WarningCrash         WarningState = "crash"
)

type AutopilotState struct {
	Engaged           bool `json:"engaged"`
	HeadingHold       bool `json:"heading_hold"`
	AltitudeHold      bool `json:"altitude_hold"`
	VerticalSpeedMode bool `json:"vertical_speed_mode"`
}

type ControlState struct {
	Mode                  Mode           `json:"mode"`
	Throttle              float64        `json:"throttle"`
	TOGA                  bool           `json:"toga"`
	RotateCommanded       bool           `json:"rotate_commanded"`
	Airborne              bool           `json:"airborne"`
	Warning               WarningState   `json:"warning"`
	Crashed               bool           `json:"crashed"`
	V1                    float64        `json:"v1"`
	Vr                    float64        `json:"vr"`
	V2                    float64        `json:"v2"`
	SelectedHeading       float64        `json:"selected_heading"`
	SelectedAltitude      float64        `json:"selected_altitude"`
	SelectedVerticalSpeed float64        `json:"selected_vertical_speed"`
	RunwayDistance        float64        `json:"runway_distance"`
	RunwayRemaining       float64        `json:"runway_remaining"`
	Autopilot             AutopilotState `json:"autopilot"`
}

type ControlCommand struct {
	Mode                  *Mode    `json:"mode,omitempty"`
	Throttle              *float64 `json:"throttle,omitempty"`
	TOGA                  *bool    `json:"toga,omitempty"`
	Rotate                *bool    `json:"rotate,omitempty"`
	AutopilotEngaged      *bool    `json:"autopilot_engaged,omitempty"`
	HeadingHold           *bool    `json:"heading_hold,omitempty"`
	AltitudeHold          *bool    `json:"altitude_hold,omitempty"`
	VerticalSpeedMode     *bool    `json:"vertical_speed_mode,omitempty"`
	SelectedHeading       *float64 `json:"selected_heading,omitempty"`
	SelectedAltitude      *float64 `json:"selected_altitude,omitempty"`
	SelectedVerticalSpeed *float64 `json:"selected_vertical_speed,omitempty"`
	Reset                 *bool    `json:"reset,omitempty"`
}

type ScalarReading struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type VectorReading struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	Unit string  `json:"unit"`
}

type AirspeedReading struct {
	Name           string  `json:"name"`
	Indicated      float64 `json:"indicated"`
	Unit           string  `json:"unit"`
	PitotPressure  float64 `json:"pitot_pressure"`
	StaticPressure float64 `json:"static_pressure"`
}

type AttitudeReading struct {
	Name    string  `json:"name"`
	Roll    float64 `json:"roll"`
	Pitch   float64 `json:"pitch"`
	Yaw     float64 `json:"yaw"`
	Heading float64 `json:"heading"`
	Unit    string  `json:"unit"`
}

type Snapshot struct {
	Phase          string          `json:"phase"`
	Timestamp      time.Time       `json:"timestamp"`
	ActiveFailures []string        `json:"active_failures,omitempty"`
	Controls       ControlState    `json:"controls"`
	Altimeter      ScalarReading   `json:"altimeter"`
	VerticalSpeed  ScalarReading   `json:"vertical_speed"`
	StaticAir      ScalarReading   `json:"static_air"`
	Pitot          ScalarReading   `json:"pitot"`
	Airspeed       AirspeedReading `json:"airspeed"`
	Accelerometer  VectorReading   `json:"accelerometer"`
	Gyroscope      VectorReading   `json:"gyroscope"`
	Magnetometer   VectorReading   `json:"magnetometer"`
	AHRS           AttitudeReading `json:"ahrs"`
}

type Simulator struct {
	mu         sync.Mutex
	rand       *rand.Rand
	noiseScale float64
	startTime  time.Time
	elapsed    time.Duration
	failures   FailureConfig
	controls   ControlState

	initialized         bool
	truth               flightState
	staticPressure      float64
	pitotPressure       float64
	vsiChamberPressure  float64
	altimeterFeet       float64
	verticalSpeedMPS    float64
	indicatedAirspeed   float64
	gyroBias            [3]float64
	gyroMeasured        [3]float64
	accelMeasured       [3]float64
	magMeasured         [3]float64
	magBias             [3]float64
	magScale            [3]float64
	declinationDeg      float64
	rollEstimate        float64
	pitchEstimate       float64
	headingEstimate     float64
	ahrsInitialized     bool
	latestSnapshotCache Snapshot
}

type flightState struct {
	phase              string
	altitudeFeet       float64
	verticalSpeedMPS   float64
	airspeedKTS        float64
	staticPressure     float64
	pitotPressure      float64
	pitchDeg           float64
	rollDeg            float64
	headingDeg         float64
	rollRateDegPerSec  float64
	pitchRateDegPerSec float64
	yawRateDegPerSec   float64
}

const (
	fieldElevationFeet           = 342.0
	runwayLengthFeet             = 8200.0
	tickInterval                 = 250 * time.Millisecond
	vsiTimeConstant              = 6.0
	gyroSaturationLimitDegPerSec = 1.5
	stallSpeedKTS                = 53.0
	runwayHeadingDeg             = 270.0
)

func NewSimulator(cfg Config) *Simulator {
	seed := cfg.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	startTime := cfg.StartTime.UTC()
	if startTime.IsZero() {
		startTime = time.Now().UTC()
	}

	s := &Simulator{
		rand:       rand.New(rand.NewSource(seed)),
		noiseScale: cfg.NoiseScale,
		startTime:  startTime,
		failures:   cfg.Failures,
	}
	s.initializeLocked()
	return s
}

func (s *Simulator) TickInterval() time.Duration {
	return tickInterval
}

func (s *Simulator) Advance(delta time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if delta <= 0 {
		return
	}

	remaining := delta
	for remaining > 0 {
		step := tickInterval
		if remaining < step {
			step = remaining
		}
		s.advanceStepLocked(step)
		remaining -= step
	}
}

func (s *Simulator) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		s.initializeLocked()
	}

	return s.latestSnapshotCache
}

func (s *Simulator) Controls() ControlState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.controls
}

func (s *Simulator) ApplyControls(command ControlCommand) ControlState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if command.Reset != nil && *command.Reset {
		s.resetLocked(s.controls.Mode)
		return s.controls
	}

	if command.Mode != nil {
		s.controls.Mode = *command.Mode
	}
	if command.Throttle != nil {
		s.controls.Throttle = clamp(*command.Throttle, 0, 1)
		if s.controls.Throttle < 0.995 {
			s.controls.TOGA = false
		}
	}
	if command.TOGA != nil {
		s.controls.TOGA = *command.TOGA
		if s.controls.TOGA {
			s.controls.Throttle = 1
		}
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
		s.controls.SelectedHeading = normalizeHeading(*command.SelectedHeading)
	}
	if command.SelectedAltitude != nil {
		s.controls.SelectedAltitude = math.Max(fieldElevationFeet, *command.SelectedAltitude)
	}
	if command.SelectedVerticalSpeed != nil {
		s.controls.SelectedVerticalSpeed = clamp(*command.SelectedVerticalSpeed, -1500, 2500)
	}

	s.controls.RunwayRemaining = math.Max(0, runwayLengthFeet-s.controls.RunwayDistance)
	s.latestSnapshotCache = s.buildSnapshotLocked()
	return s.controls
}

func (s *Simulator) SetFailures(failures FailureConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failures = failures
	s.latestSnapshotCache = s.buildSnapshotLocked()
}

func (s *Simulator) Failures() FailureConfig {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.failures
}

func (s *Simulator) initializeLocked() {
	if s.initialized {
		return
	}

	s.resetLocked(ModeFull)
	s.initialized = true
}

func (s *Simulator) resetLocked(mode Mode) {
	if mode == "" {
		mode = ModeFull
	}
	s.elapsed = 0
	s.controls = ControlState{
		Mode:                  mode,
		Throttle:              0,
		TOGA:                  false,
		RotateCommanded:       false,
		Airborne:              false,
		Warning:               WarningNone,
		Crashed:               false,
		V1:                    62,
		Vr:                    67,
		V2:                    74,
		SelectedHeading:       runwayHeadingDeg,
		SelectedAltitude:      fieldElevationFeet + 1800,
		SelectedVerticalSpeed: 700,
		RunwayDistance:        0,
		RunwayRemaining:       runwayLengthFeet,
		Autopilot: AutopilotState{
			Engaged:           false,
			HeadingHold:       false,
			AltitudeHold:      false,
			VerticalSpeedMode: false,
		},
	}
	s.truth = flightState{
		phase:            "runway_idle",
		altitudeFeet:     fieldElevationFeet,
		verticalSpeedMPS: 0,
		airspeedKTS:      0,
		staticPressure:   pressureAtFeet(fieldElevationFeet),
		pitchDeg:         0,
		rollDeg:          0,
		headingDeg:       runwayHeadingDeg,
	}
	s.staticPressure = s.truth.staticPressure
	s.pitotPressure = s.truth.staticPressure
	s.vsiChamberPressure = s.staticPressure
	s.altimeterFeet = fieldElevationFeet
	s.indicatedAirspeed = 0
	s.verticalSpeedMPS = 0
	s.declinationDeg = 11.0
	s.rollEstimate = 0
	s.pitchEstimate = 0
	s.headingEstimate = runwayHeadingDeg
	s.ahrsInitialized = false

	for index := range s.gyroBias {
		s.gyroBias[index] = s.noisy(0, 0.015)
		s.magBias[index] = s.noisy(0, 0.6)
		s.magScale[index] = 1 + s.noisy(0, 0.025)
		s.gyroMeasured[index] = 0
		s.accelMeasured[index] = 0
		s.magMeasured[index] = 0
	}

	s.updateIMUMeasurements(s.truth, s.truth, tickInterval.Seconds())
	s.updateMagnetometerMeasurement(s.truth)
	s.updateAHRS(tickInterval.Seconds())
	s.latestSnapshotCache = s.buildSnapshotLocked()
}

func (s *Simulator) advanceStepLocked(step time.Duration) {
	if !s.initialized {
		s.initializeLocked()
	}

	currentTruth := s.truth
	s.elapsed += step
	s.updateControlledFlight(step.Seconds())
	nextTruth := s.truth

	s.updatePressureMeasurements(nextTruth, step.Seconds())
	s.updateAirspeedMeasurement(step.Seconds())
	s.updateVSIMeasurement(step.Seconds())
	s.updateIMUMeasurements(currentTruth, nextTruth, step.Seconds())
	s.updateMagnetometerMeasurement(nextTruth)
	s.updateAHRS(step.Seconds())
	s.latestSnapshotCache = s.buildSnapshotLocked()
}

func (s *Simulator) updateControlledFlight(dt float64) {
	previous := s.truth
	if s.controls.TOGA {
		s.controls.Throttle = 1
	}

	if s.controls.Crashed {
		s.truth.phase = "crash"
		s.truth.airspeedKTS = firstOrderResponse(s.truth.airspeedKTS, 0, dt, 0.8)
		s.truth.verticalSpeedMPS = 0
		s.truth.pitchDeg = firstOrderResponse(s.truth.pitchDeg, 0, dt, 0.6)
		s.truth.rollDeg = firstOrderResponse(s.truth.rollDeg, 0, dt, 0.6)
		s.truth.altitudeFeet = fieldElevationFeet
		s.updateRates(previous, dt)
		return
	}

	if s.controls.Airborne {
		s.updateAirborneState(dt)
	} else {
		s.updateGroundState(dt)
	}

	s.controls.Warning = s.computeWarningState()
	if s.shouldCrashLocked() {
		s.controls.Warning = WarningCrash
		s.controls.Crashed = true
		s.controls.Airborne = false
		s.truth.phase = "crash"
		s.truth.altitudeFeet = fieldElevationFeet
		s.truth.verticalSpeedMPS = 0
		s.truth.pitchDeg = 0
		s.truth.rollDeg = 0
	}

	s.controls.RunwayRemaining = math.Max(0, runwayLengthFeet-s.controls.RunwayDistance)
	s.truth.staticPressure = pressureAtFeet(s.truth.altitudeFeet)
	s.truth.pitotPressure = s.truth.staticPressure + dynamicPressureAtKnots(s.truth.airspeedKTS)
	s.updateRates(previous, dt)
}

func (s *Simulator) updateGroundState(dt float64) {
	accel := s.groundAcceleration()
	s.truth.airspeedKTS = math.Max(0, s.truth.airspeedKTS+accel*dt)
	s.controls.RunwayDistance += s.truth.airspeedKTS * 1.68781 * dt
	s.truth.headingDeg = runwayHeadingDeg
	s.truth.rollDeg = firstOrderResponse(s.truth.rollDeg, 0, dt, 0.4)

	targetPitch := 0.0
	if s.controls.RotateCommanded {
		targetPitch = 11.0
	}
	s.truth.pitchDeg = firstOrderResponse(s.truth.pitchDeg, targetPitch, dt, 0.5)
	s.truth.verticalSpeedMPS = 0
	s.truth.altitudeFeet = fieldElevationFeet

	if s.controls.RotateCommanded && s.truth.airspeedKTS >= s.controls.Vr-1 && s.controls.Throttle >= 0.5 {
		liftFactor := (s.truth.airspeedKTS - s.controls.Vr) / 10
		energyBonus := s.controls.Throttle * 0.75
		pitchBonus := s.truth.pitchDeg * 0.04
		threshold := 0.8
		if s.controls.Mode == ModeTransitional {
			threshold = 0.55
		}
		if liftFactor+energyBonus+pitchBonus >= threshold {
			s.controls.Airborne = true
			s.truth.phase = "initial_climb"
			s.truth.altitudeFeet += 5
			s.truth.verticalSpeedMPS = 1.2
			return
		}
	}

	if s.controls.Throttle < 0.05 && s.truth.airspeedKTS < 1 {
		s.truth.phase = "runway_idle"
		return
	}

	if s.controls.RotateCommanded {
		s.truth.phase = "rotation_attempt"
		return
	}

	s.truth.phase = "takeoff_roll"
}

func (s *Simulator) updateAirborneState(dt float64) {
	targetRoll := 0.0
	if s.controls.Autopilot.Engaged && s.controls.Autopilot.HeadingHold {
		targetRoll = clamp(angleDelta(s.truth.headingDeg, s.controls.SelectedHeading)*0.45, -18, 18)
	}
	s.truth.rollDeg = firstOrderResponse(s.truth.rollDeg, targetRoll, dt, 1.2)

	turnRate := s.truth.rollDeg * 0.12
	if s.controls.Mode == ModeTransitional {
		turnRate *= 0.9
	}
	s.truth.headingDeg = normalizeHeading(s.truth.headingDeg + turnRate*dt)

	desiredVS := s.commandedVerticalSpeedMPS()
	if s.controls.Mode == ModeTransitional && s.controls.Throttle >= 0.7 && desiredVS < 2.5 {
		desiredVS = 2.5
	}

	energyMargin := s.controls.Throttle*92 - s.truth.airspeedKTS - math.Max(0, desiredVS*8) - math.Max(0, s.truth.pitchDeg*1.8)
	airspeedAccel := energyMargin*0.03 - 0.35
	if s.controls.Mode == ModeTransitional {
		airspeedAccel += 0.35
	}
	s.truth.airspeedKTS = math.Max(0, s.truth.airspeedKTS+airspeedAccel*dt)

	if s.truth.airspeedKTS < stallSpeedKTS {
		desiredVS = math.Min(desiredVS, -2.8)
	}

	s.truth.verticalSpeedMPS = firstOrderResponse(s.truth.verticalSpeedMPS, desiredVS, dt, 1.4)
	s.truth.altitudeFeet += s.truth.verticalSpeedMPS * dt * 3.28084

	targetPitch := clamp(2.0+s.truth.verticalSpeedMPS*1.7, -4, 16)
	if !s.controls.Autopilot.Engaged && s.controls.RotateCommanded {
		targetPitch = math.Max(targetPitch, 10)
	}
	s.truth.pitchDeg = firstOrderResponse(s.truth.pitchDeg, targetPitch, dt, 1.1)

	if s.truth.altitudeFeet <= fieldElevationFeet {
		s.truth.altitudeFeet = fieldElevationFeet
	}

	if s.truth.altitudeFeet-fieldElevationFeet < 1500 {
		s.truth.phase = "initial_climb"
	} else {
		s.truth.phase = "airborne"
	}
}

func (s *Simulator) groundAcceleration() float64 {
	base := s.controls.Throttle*7.4 - 0.7 - s.truth.airspeedKTS*0.045
	if s.controls.TOGA {
		base += 0.4
	}
	if s.controls.Throttle < 0.05 && s.truth.airspeedKTS < 0.5 {
		return -s.truth.airspeedKTS * 3
	}
	if base < -2 {
		return -2
	}
	return base
}

func (s *Simulator) commandedVerticalSpeedMPS() float64 {
	if s.controls.Autopilot.Engaged {
		if s.controls.Autopilot.AltitudeHold {
			altitudeError := s.controls.SelectedAltitude - s.truth.altitudeFeet
			target := clamp(altitudeError*0.01, -4.5, 5.2)
			if math.Abs(altitudeError) < 25 {
				return 0
			}
			return target
		}
		if s.controls.Autopilot.VerticalSpeedMode {
			return clamp(s.controls.SelectedVerticalSpeed*0.00508, -5.5, 7.5)
		}
	}

	if s.controls.RotateCommanded || s.controls.TOGA {
		return clamp((s.truth.airspeedKTS-s.controls.V2)*0.08+4.6, -4.0, 7.8)
	}

	return clamp((s.truth.airspeedKTS-s.controls.V1)*0.02, -1.0, 1.5)
}

func (s *Simulator) computeWarningState() WarningState {
	if s.controls.Crashed {
		return WarningCrash
	}

	agl := s.truth.altitudeFeet - fieldElevationFeet
	if s.controls.Airborne {
		if agl < 120 && s.truth.verticalSpeedMPS < -4.0 {
			return WarningPullUp
		}
		if agl < 500 && s.truth.verticalSpeedMPS < -2.5 {
			return WarningSinkRate
		}
		if s.truth.airspeedKTS < stallSpeedKTS || (s.truth.pitchDeg > 14 && s.truth.airspeedKTS < s.controls.V2-10) {
			return WarningStall
		}
		if s.truth.airspeedKTS < s.controls.V2-8 && s.truth.pitchDeg > 8 {
			return WarningLowEnergy
		}
		return WarningNone
	}

	if s.controls.RotateCommanded && s.truth.airspeedKTS < s.controls.Vr {
		return WarningBelowVrRotate
	}
	if s.controls.Throttle > 0.7 && s.controls.RunwayDistance > runwayLengthFeet*0.65 && s.truth.airspeedKTS < s.controls.Vr-8 {
		return WarningLowEnergy
	}
	return WarningNone
}

func (s *Simulator) shouldCrashLocked() bool {
	if s.controls.RunwayDistance >= runwayLengthFeet && !s.controls.Airborne {
		return true
	}
	if !s.controls.Airborne {
		return false
	}
	agl := s.truth.altitudeFeet - fieldElevationFeet
	if agl <= 0 && (s.truth.verticalSpeedMPS < -0.8 || s.controls.Warning == WarningPullUp || s.controls.Warning == WarningStall) {
		return true
	}
	return false
}

func (s *Simulator) updateRates(previous flightState, dt float64) {
	if dt <= 0 {
		return
	}
	s.truth.rollRateDegPerSec = angleDelta(previous.rollDeg, s.truth.rollDeg) / dt
	s.truth.pitchRateDegPerSec = angleDelta(previous.pitchDeg, s.truth.pitchDeg) / dt
	s.truth.yawRateDegPerSec = angleDelta(previous.headingDeg, s.truth.headingDeg) / dt
}

func (s *Simulator) buildSnapshotLocked() Snapshot {
	timestamp := s.startTime.Add(s.elapsed)
	controls := s.controls
	controls.RunwayRemaining = math.Max(0, runwayLengthFeet-controls.RunwayDistance)

	return Snapshot{
		Phase:          s.truth.phase,
		Timestamp:      timestamp,
		ActiveFailures: s.activeFailures(),
		Controls:       controls,
		Altimeter:      ScalarReading{Name: "altimeter", Value: s.round(s.altimeterFeet, 0.1), Unit: "ft"},
		VerticalSpeed:  ScalarReading{Name: "vertical_speed", Value: s.round(s.verticalSpeedMPS, 0.01), Unit: "m/s"},
		StaticAir:      ScalarReading{Name: "static_air", Value: s.round(s.staticPressure, 0.001), Unit: "inHg"},
		Pitot:          ScalarReading{Name: "pitot", Value: s.round(s.pitotPressure, 0.001), Unit: "inHg"},
		Airspeed: AirspeedReading{
			Name:           "airspeed",
			Indicated:      s.round(s.indicatedAirspeed, 0.1),
			Unit:           "kts",
			PitotPressure:  s.round(s.pitotPressure, 0.001),
			StaticPressure: s.round(s.staticPressure, 0.001),
		},
		Accelerometer: VectorReading{
			Name: "accelerometer",
			X:    s.round(s.accelMeasured[0], 0.01),
			Y:    s.round(s.accelMeasured[1], 0.01),
			Z:    s.round(s.accelMeasured[2], 0.01),
			Unit: "m/s^2",
		},
		Gyroscope: VectorReading{
			Name: "gyroscope",
			X:    s.round(s.gyroMeasured[0], 0.01),
			Y:    s.round(s.gyroMeasured[1], 0.01),
			Z:    s.round(s.gyroMeasured[2], 0.01),
			Unit: "deg/s",
		},
		Magnetometer: VectorReading{
			Name: "magnetometer",
			X:    s.round(s.magMeasured[0], 0.01),
			Y:    s.round(s.magMeasured[1], 0.01),
			Z:    s.round(s.magMeasured[2], 0.01),
			Unit: "uT",
		},
		AHRS: AttitudeReading{
			Name:    "ahrs",
			Roll:    s.round(s.rollEstimate, 0.1),
			Pitch:   s.round(s.pitchEstimate, 0.1),
			Yaw:     s.round(s.headingEstimate, 0.1),
			Heading: s.round(s.headingEstimate, 0.1),
			Unit:    "deg",
		},
	}
}

func (s *Simulator) updatePressureMeasurements(truth flightState, dt float64) {
	q := dynamicPressureAtKnots(truth.airspeedKTS)
	pitchRad := truth.pitchDeg * math.Pi / 180
	rollRad := truth.rollDeg * math.Pi / 180

	staticPortError := q*0.015*math.Sin(pitchRad) + q*0.008*math.Sin(rollRad)
	staticTarget := truth.staticPressure + staticPortError + s.noisy(0, 0.0012)
	pitotTarget := truth.staticPressure + q + q*0.01*math.Sin(pitchRad) + s.noisy(0, 0.0015)
	if s.failures.StaticLeak {
		leakReference := pressureAtFeet(fieldElevationFeet)
		staticTarget += 0.78 * (leakReference - staticTarget)
	}
	if s.failures.PitotDrainBlocked {
		pitotTarget = truth.staticPressure + s.noisy(0, 0.001)
	}

	if !s.failures.StaticPortBlocked {
		s.staticPressure = firstOrderResponse(s.staticPressure, staticTarget, dt, 0.35)
	}
	if !s.failures.PitotBlocked {
		timeConstant := 0.2
		if s.failures.PitotDrainBlocked {
			timeConstant = 1.6
		}
		s.pitotPressure = firstOrderResponse(s.pitotPressure, pitotTarget, dt, timeConstant)
	}
	if s.pitotPressure < s.staticPressure {
		s.pitotPressure = s.staticPressure
	}

	s.altimeterFeet = altitudeFromPressure(s.staticPressure) + s.noisy(0, 1.5)
}

func (s *Simulator) updateAirspeedMeasurement(dt float64) {
	dynamicPressure := math.Max(0, s.pitotPressure-s.staticPressure)
	target := indicatedAirspeedFromPressure(dynamicPressure)
	s.indicatedAirspeed = firstOrderResponse(s.indicatedAirspeed, target, dt, 0.35)
}

func (s *Simulator) updateVSIMeasurement(dt float64) {
	s.vsiChamberPressure = firstOrderResponse(s.vsiChamberPressure, s.staticPressure, dt, vsiTimeConstant)
	staticAltitude := altitudeFromPressure(s.staticPressure)
	chamberAltitude := altitudeFromPressure(s.vsiChamberPressure)
	s.verticalSpeedMPS = (staticAltitude - chamberAltitude) / vsiTimeConstant / 3.28084
	if math.Abs(s.verticalSpeedMPS) < 0.01 {
		s.verticalSpeedMPS = 0
	}
}

func (s *Simulator) updateIMUMeasurements(current flightState, next flightState, dt float64) {
	for index := range s.gyroBias {
		s.gyroBias[index] += s.noisy(0, 0.01) * dt
	}

	s.gyroMeasured[0] = next.rollRateDegPerSec + s.gyroBias[0] + s.noisy(0, 0.05)
	s.gyroMeasured[1] = next.pitchRateDegPerSec + s.gyroBias[1] + s.noisy(0, 0.04)
	s.gyroMeasured[2] = next.yawRateDegPerSec + s.gyroBias[2] + s.noisy(0, 0.06)
	if s.failures.GyroSaturation {
		s.gyroMeasured[0] = clamp(s.gyroMeasured[0], -gyroSaturationLimitDegPerSec, gyroSaturationLimitDegPerSec)
		s.gyroMeasured[1] = clamp(s.gyroMeasured[1], -gyroSaturationLimitDegPerSec, gyroSaturationLimitDegPerSec)
		s.gyroMeasured[2] = clamp(s.gyroMeasured[2], -gyroSaturationLimitDegPerSec, gyroSaturationLimitDegPerSec)
	}

	ax, ay, az := s.accelerometer(current, next, dt)
	vibration := 0.03 + next.airspeedKTS*0.0008
	s.accelMeasured[0] = ax + s.noisy(0, vibration)
	s.accelMeasured[1] = ay + s.noisy(0, vibration)
	s.accelMeasured[2] = az + s.noisy(0, vibration)
}

func (s *Simulator) updateMagnetometerMeasurement(truth flightState) {
	x, y, z := magneticFieldBody(truth, s.declinationDeg)
	if s.failures.MagnetometerDisturbed {
		disturbance := 18.0 + 4.0*math.Sin(s.elapsed.Seconds()/3.0)
		x += disturbance
		y -= disturbance * 0.6
		z += disturbance * 0.35
	}
	x = x*s.magScale[0] + s.magBias[0] + 0.02*y + s.noisy(0, 0.12)
	y = y*s.magScale[1] + s.magBias[1] - 0.015*x + s.noisy(0, 0.12)
	z = z*s.magScale[2] + s.magBias[2] + s.noisy(0, 0.12)
	if s.failures.MagnetometerDisturbed {
		x += s.noisy(0, 1.8)
		y += s.noisy(0, 1.8)
		z += s.noisy(0, 1.8)
	}

	s.magMeasured[0] = x
	s.magMeasured[1] = y
	s.magMeasured[2] = z
}

func (s *Simulator) activeFailures() []string {
	active := make([]string, 0, 6)
	if s.failures.PitotBlocked {
		active = append(active, "pitot_blocked")
	}
	if s.failures.PitotDrainBlocked {
		active = append(active, "pitot_drain_blocked")
	}
	if s.failures.StaticPortBlocked {
		active = append(active, "static_port_blocked")
	}
	if s.failures.StaticLeak {
		active = append(active, "static_leak")
	}
	if s.failures.MagnetometerDisturbed {
		active = append(active, "magnetometer_disturbed")
	}
	if s.failures.GyroSaturation {
		active = append(active, "gyro_saturation")
	}
	return active
}

func (s *Simulator) updateAHRS(dt float64) {
	accelRoll, accelPitch := calculateAccelAttitude(s.accelMeasured[0], s.accelMeasured[1], s.accelMeasured[2])
	magHeading := calculateMagHeading(s.magMeasured[0], s.magMeasured[1], s.magMeasured[2], s.rollEstimate, s.pitchEstimate)

	if !s.ahrsInitialized {
		s.rollEstimate = accelRoll
		s.pitchEstimate = accelPitch
		s.headingEstimate = magHeading
		s.ahrsInitialized = true
		return
	}

	gyroRoll := s.rollEstimate + s.gyroMeasured[0]*dt
	gyroPitch := s.pitchEstimate + s.gyroMeasured[1]*dt
	gyroHeading := normalizeHeading(s.headingEstimate + s.gyroMeasured[2]*dt)

	const alpha = 0.985
	s.rollEstimate = normalizeSignedAngle(alpha*gyroRoll + (1-alpha)*accelRoll)
	s.pitchEstimate = normalizeSignedAngle(alpha*gyroPitch + (1-alpha)*accelPitch)
	s.headingEstimate = normalizeHeading(gyroHeading + (1-alpha)*angleDelta(gyroHeading, magHeading))
}

func (s *Simulator) stateAt(_ time.Duration) flightState {
	return s.truth
}

func (s *Simulator) accelerometer(current flightState, next flightState, dt float64) (float64, float64, float64) {
	const gravity = 9.81

	pitchRad := next.pitchDeg * math.Pi / 180
	rollRad := next.rollDeg * math.Pi / 180
	longitudinalAccel := ((next.airspeedKTS - current.airspeedKTS) * 0.514444) / dt
	verticalAccel := (next.verticalSpeedMPS - current.verticalSpeedMPS) / dt
	lateralAccel := 0.0
	if math.Abs(next.rollDeg) > 0.1 {
		lateralAccel = gravity * math.Tan(rollRad)
	}

	ax := longitudinalAccel - gravity*math.Sin(pitchRad)
	ay := lateralAccel + gravity*math.Sin(rollRad)*math.Cos(pitchRad)
	az := verticalAccel - gravity*math.Cos(rollRad)*math.Cos(pitchRad)

	return ax, ay, az
}

func magneticFieldBody(state flightState, declinationDeg float64) (float64, float64, float64) {
	const fieldStrength = 50.0
	const inclinationDeg = 60.0

	declinationRad := declinationDeg * math.Pi / 180
	inclinationRad := inclinationDeg * math.Pi / 180
	headingRad := state.headingDeg * math.Pi / 180
	pitchRad := state.pitchDeg * math.Pi / 180
	rollRad := state.rollDeg * math.Pi / 180

	north := fieldStrength * math.Cos(inclinationRad) * math.Cos(declinationRad)
	east := fieldStrength * math.Cos(inclinationRad) * math.Sin(declinationRad)
	down := fieldStrength * math.Sin(inclinationRad)

	xh := north*math.Cos(headingRad) + east*math.Sin(headingRad)
	yh := -north*math.Sin(headingRad) + east*math.Cos(headingRad)

	x := xh*math.Cos(pitchRad) + down*math.Sin(pitchRad)
	y := xh*math.Sin(rollRad)*math.Sin(pitchRad) + yh*math.Cos(rollRad) - down*math.Sin(rollRad)*math.Cos(pitchRad)
	z := -xh*math.Cos(rollRad)*math.Sin(pitchRad) + yh*math.Sin(rollRad) + down*math.Cos(rollRad)*math.Cos(pitchRad)

	return x, y, z
}

func calculateAccelAttitude(ax float64, ay float64, az float64) (float64, float64) {
	roll := math.Atan2(ay, -az) * 180 / math.Pi
	pitch := math.Atan2(ax, math.Sqrt(ay*ay+az*az)) * 180 / math.Pi
	return roll, pitch
}

func calculateMagHeading(mx float64, my float64, mz float64, rollDeg float64, pitchDeg float64) float64 {
	rollRad := rollDeg * math.Pi / 180
	pitchRad := pitchDeg * math.Pi / 180

	mxHorizontal := mx*math.Cos(pitchRad) + mz*math.Sin(pitchRad)
	myHorizontal := mx*math.Sin(rollRad)*math.Sin(pitchRad) + my*math.Cos(rollRad) - mz*math.Sin(rollRad)*math.Cos(pitchRad)
	heading := math.Atan2(-myHorizontal, mxHorizontal) * 180 / math.Pi
	return normalizeHeading(heading)
}

func (s *Simulator) noisy(value float64, amplitude float64) float64 {
	if s.noiseScale == 0 || amplitude == 0 {
		return value
	}

	return value + ((s.rand.Float64()*2)-1)*amplitude*s.noiseScale
}

func (s *Simulator) round(value float64, precision float64) float64 {
	return math.Round(value/precision) * precision
}

func pressureAtFeet(altitudeFeet float64) float64 {
	return 29.92 * math.Pow(1.0-0.0000068756*altitudeFeet, 5.2559)
}

func altitudeFromPressure(pressure float64) float64 {
	if pressure <= 0 {
		return 0
	}

	return (1 - math.Pow(pressure/29.92, 1.0/5.2559)) / 0.0000068756
}

func dynamicPressureAtKnots(airspeed float64) float64 {
	return 0.00002378 * airspeed * airspeed / 2.0
}

func indicatedAirspeedFromPressure(dynamicPressure float64) float64 {
	if dynamicPressure <= 0 {
		return 0
	}

	return math.Sqrt((2 * dynamicPressure) / 0.00002378)
}

func firstOrderResponse(current float64, target float64, dt float64, timeConstant float64) float64 {
	if timeConstant <= 0 {
		return target
	}

	alpha := 1 - math.Exp(-dt/timeConstant)
	return current + alpha*(target-current)
}

func angleDelta(from float64, to float64) float64 {
	delta := to - from
	for delta > 180 {
		delta -= 360
	}
	for delta < -180 {
		delta += 360
	}
	return delta
}

func normalizeHeading(heading float64) float64 {
	for heading < 0 {
		heading += 360
	}
	for heading >= 360 {
		heading -= 360
	}
	return heading
}

func normalizeSignedAngle(angle float64) float64 {
	for angle > 180 {
		angle -= 360
	}
	for angle < -180 {
		angle += 360
	}
	return angle
}

func clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
