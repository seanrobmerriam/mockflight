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
	cycleDuration                = 240 * time.Second
	tickInterval                 = 250 * time.Millisecond
	vsiTimeConstant              = 6.0
	gyroSaturationLimitDegPerSec = 1.5
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

	s.truth = s.stateAt(0)
	s.staticPressure = s.truth.staticPressure
	s.pitotPressure = s.truth.pitotPressure
	s.vsiChamberPressure = s.staticPressure
	s.altimeterFeet = altitudeFromPressure(s.staticPressure)
	s.indicatedAirspeed = indicatedAirspeedFromPressure(s.pitotPressure - s.staticPressure)
	s.verticalSpeedMPS = 0
	s.declinationDeg = 11.0

	for index := range s.gyroBias {
		s.gyroBias[index] = s.noisy(0, 0.015)
		s.magBias[index] = s.noisy(0, 0.6)
		s.magScale[index] = 1 + s.noisy(0, 0.025)
	}

	next := s.stateAt(tickInterval)
	s.updateIMUMeasurements(s.truth, next, tickInterval.Seconds())
	s.updateMagnetometerMeasurement(s.truth)
	s.updateAHRS(tickInterval.Seconds())
	s.latestSnapshotCache = s.buildSnapshotLocked()
	s.initialized = true
}

func (s *Simulator) advanceStepLocked(step time.Duration) {
	if !s.initialized {
		s.initializeLocked()
	}

	currentTruth := s.stateAt(s.elapsed)
	s.elapsed += step
	nextTruth := s.stateAt(s.elapsed)
	dt := step.Seconds()

	s.truth = nextTruth
	s.updatePressureMeasurements(nextTruth, dt)
	s.updateAirspeedMeasurement(dt)
	s.updateVSIMeasurement(dt)
	s.updateIMUMeasurements(currentTruth, nextTruth, dt)
	s.updateMagnetometerMeasurement(nextTruth)
	s.updateAHRS(dt)
	s.latestSnapshotCache = s.buildSnapshotLocked()
}

func (s *Simulator) buildSnapshotLocked() Snapshot {
	timestamp := s.startTime.Add(s.elapsed)

	return Snapshot{
		Phase:          s.truth.phase,
		Timestamp:      timestamp,
		ActiveFailures: s.activeFailures(),
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
		staticTarget += 0.18 * (leakReference - staticTarget)
	}
	if s.failures.PitotDrainBlocked {
		pitotTarget = truth.staticPressure + q*0.12 + s.noisy(0, 0.001)
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

func (s *Simulator) stateAt(elapsed time.Duration) flightState {
	seconds := math.Mod(elapsed.Seconds(), cycleDuration.Seconds())
	base := s.baseState(seconds)
	derived := s.baseState(math.Mod(seconds+0.2, cycleDuration.Seconds()))

	base.rollRateDegPerSec = angleDelta(base.rollDeg, derived.rollDeg) / 0.2
	base.pitchRateDegPerSec = angleDelta(base.pitchDeg, derived.pitchDeg) / 0.2
	base.yawRateDegPerSec = angleDelta(base.headingDeg, derived.headingDeg) / 0.2
	base.staticPressure = pressureAtFeet(base.altitudeFeet)
	base.pitotPressure = base.staticPressure + dynamicPressureAtKnots(base.airspeedKTS)

	return base
}

func (s *Simulator) baseState(seconds float64) flightState {
	switch {
	case seconds < 20:
		progress := seconds / 20
		return flightState{
			phase:            "ground",
			altitudeFeet:     fieldElevationFeet,
			verticalSpeedMPS: 0,
			airspeedKTS:      5 + progress*15,
			pitchDeg:         0.2 * math.Sin(progress*math.Pi),
			rollDeg:          0.5 * math.Sin(progress*2*math.Pi),
			headingDeg:       268,
		}
	case seconds < 35:
		progress := (seconds - 20) / 15
		return flightState{
			phase:            "takeoff",
			altitudeFeet:     fieldElevationFeet + progress*500,
			verticalSpeedMPS: 5.1,
			airspeedKTS:      20 + progress*55,
			pitchDeg:         2 + progress*10,
			rollDeg:          1.2 * math.Sin(progress*math.Pi),
			headingDeg:       270,
		}
	case seconds < 80:
		progress := (seconds - 35) / 45
		return flightState{
			phase:            "climb",
			altitudeFeet:     fieldElevationFeet + 500 + progress*4500,
			verticalSpeedMPS: 5.0 + 0.35*math.Sin(progress*4*math.Pi),
			airspeedKTS:      78 + progress*47,
			pitchDeg:         8.5 + 1.2*math.Sin(progress*3*math.Pi),
			rollDeg:          2.5 * math.Sin(progress*2*math.Pi),
			headingDeg:       normalizeHeading(270 + 4*math.Sin(progress*2*math.Pi)),
		}
	case seconds < 170:
		progress := (seconds - 80) / 90
		return flightState{
			phase:            "cruise",
			altitudeFeet:     fieldElevationFeet + 5000 + 40*math.Sin(progress*6*math.Pi),
			verticalSpeedMPS: 0.15 * math.Sin(progress*6*math.Pi),
			airspeedKTS:      126 + 4*math.Sin(progress*4*math.Pi),
			pitchDeg:         2.5 + 0.7*math.Sin(progress*4*math.Pi),
			rollDeg:          12 * math.Sin(progress*2*math.Pi),
			headingDeg:       normalizeHeading(270 + 40*math.Sin(progress*2*math.Pi)),
		}
	case seconds < 210:
		progress := (seconds - 170) / 40
		return flightState{
			phase:            "descent",
			altitudeFeet:     fieldElevationFeet + 5000 - progress*4300,
			verticalSpeedMPS: -4.0 + 0.25*math.Sin(progress*4*math.Pi),
			airspeedKTS:      120 - progress*45,
			pitchDeg:         -1.5 + 0.5*math.Sin(progress*3*math.Pi),
			rollDeg:          4 * math.Sin(progress*2*math.Pi),
			headingDeg:       normalizeHeading(252 + 15*math.Sin(progress*math.Pi)),
		}
	case seconds < 230:
		progress := (seconds - 210) / 20
		pitch := -2.0
		if progress > 0.75 {
			pitch = -2 + ((progress-0.75)/0.25)*6
		}
		return flightState{
			phase:            "landing",
			altitudeFeet:     fieldElevationFeet + 700 - progress*700,
			verticalSpeedMPS: -2.3 + progress*2.1,
			airspeedKTS:      75 - progress*50,
			pitchDeg:         pitch,
			rollDeg:          3 * math.Sin(progress*4*math.Pi),
			headingDeg:       270,
		}
	default:
		progress := (seconds - 230) / 10
		return flightState{
			phase:            "ground_rollout",
			altitudeFeet:     fieldElevationFeet,
			verticalSpeedMPS: 0,
			airspeedKTS:      22 - progress*20,
			pitchDeg:         0.4 * (1 - progress),
			rollDeg:          0.6 * math.Sin(progress*2*math.Pi),
			headingDeg:       270,
		}
	}
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
