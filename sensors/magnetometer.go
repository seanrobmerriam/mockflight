package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type MagnetometerReading struct {
	Sensor    string    `json:"sensor"`
	MagX      float64   `json:"mag_x"`
	MagY      float64   `json:"mag_y"`
	MagZ      float64   `json:"mag_z"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type FlightPhase int

const (
	Ground FlightPhase = iota
	Takeoff
	Climb
	CruiseStraight
	TurnLeft
	TurnRight
	Descent
	Landing
	OnGround
)

type Magnetometer struct {
	currentHeading      float64
	currentPitch        float64
	currentRoll         float64
	targetHeading       float64
	targetPitch         float64
	targetRoll          float64
	phase               FlightPhase
	phaseTime           time.Duration
	updateRate          time.Duration
	noiseLevel          float64
	hardIronBias        [3]float64 // Hard iron distortion (constant offset)
	magneticDeclination float64    // Local magnetic variation
	format              string
}

func NewMagnetometer(hz float64, format string, noiseLevel float64) *Magnetometer {
	return &Magnetometer{
		currentHeading: 0.0,
		currentPitch:   0.0,
		currentRoll:    0.0,
		targetHeading:  0.0,
		targetPitch:    0.0,
		targetRoll:     0.0,
		phase:          Ground,
		phaseTime:      0,
		updateRate:     time.Duration(float64(time.Second) / hz),
		noiseLevel:     noiseLevel,
		// Hard iron bias from aircraft's magnetic field
		hardIronBias: [3]float64{
			rand.Float64()*2.0 - 1.0,
			rand.Float64()*2.0 - 1.0,
			rand.Float64()*2.0 - 1.0,
		},
		magneticDeclination: 14.0, // Example: ~14° East for California
		format:              format,
	}
}

func (m *Magnetometer) updatePhase(elapsed time.Duration) {
	m.phaseTime += elapsed

	switch m.phase {
	case Ground:
		m.targetHeading = 0.0 // North
		m.targetPitch = 0.0
		m.targetRoll = 0.0
		if m.phaseTime > 20*time.Second {
			m.phase = Takeoff
			m.phaseTime = 0
		}

	case Takeoff:
		m.targetHeading = 0.0 // Maintain runway heading
		progress := math.Min(1.0, float64(m.phaseTime)/float64(5*time.Second))
		m.targetPitch = progress * 10.0
		m.targetRoll = 0.0

		if m.phaseTime > 15*time.Second {
			m.phase = Climb
			m.phaseTime = 0
		}

	case Climb:
		m.targetHeading = 0.0
		m.targetPitch = 8.0 + math.Sin(float64(m.phaseTime)/float64(8*time.Second))*1.0
		m.targetRoll = 0.0

		if m.phaseTime > 45*time.Second {
			m.phase = CruiseStraight
			m.phaseTime = 0
		}

	case CruiseStraight:
		m.targetHeading = 0.0
		m.targetPitch = 2.0 + math.Sin(float64(m.phaseTime)/float64(15*time.Second))*0.5
		m.targetRoll = 0.0

		if m.phaseTime > 20*time.Second {
			m.phase = TurnLeft
			m.phaseTime = 0
		}

	case TurnLeft:
		progress := float64(m.phaseTime) / float64(15*time.Second)

		// Turn 90 degrees left over 15 seconds
		m.targetHeading = -progress * 90.0

		// Bank angle during turn
		if progress < 0.2 {
			m.targetRoll = -progress * 5.0 * 18.0
		} else if progress < 0.8 {
			m.targetRoll = -18.0
		} else {
			rollProgress := (progress - 0.8) / 0.2
			m.targetRoll = -18.0 + rollProgress*18.0
		}

		m.targetPitch = 3.0

		if m.phaseTime > 15*time.Second {
			m.phase = CruiseStraight
			m.phaseTime = 0
		}

	case TurnRight:
		progress := float64(m.phaseTime) / float64(15*time.Second)

		// Turn 90 degrees right (from -90 to 0)
		m.targetHeading = -90.0 + progress*90.0

		if progress < 0.2 {
			m.targetRoll = progress * 5.0 * 18.0
		} else if progress < 0.8 {
			m.targetRoll = 18.0
		} else {
			rollProgress := (progress - 0.8) / 0.2
			m.targetRoll = 18.0 - rollProgress*18.0
		}

		m.targetPitch = 3.0

		if m.phaseTime > 15*time.Second {
			m.targetRoll = 0.0
			if m.phaseTime > 18*time.Second {
				m.phase = Descent
				m.phaseTime = 0
			}
		}

	case Descent:
		m.targetHeading = 0.0
		m.targetPitch = -2.0 + math.Sin(float64(m.phaseTime)/float64(10*time.Second))*0.5
		m.targetRoll = 0.0

		if m.phaseTime > 40*time.Second {
			m.phase = Landing
			m.phaseTime = 0
		}

	case Landing:
		m.targetHeading = 0.0
		progress := float64(m.phaseTime) / float64(20*time.Second)

		if progress < 0.8 {
			m.targetPitch = -1.0
		} else {
			flareProgress := (progress - 0.8) / 0.2
			m.targetPitch = -1.0 + flareProgress*6.0
		}

		m.targetRoll = math.Sin(float64(m.phaseTime)/float64(3*time.Second)) * 3.0

		if m.phaseTime > 20*time.Second {
			m.phase = OnGround
			m.phaseTime = 0
		}

	case OnGround:
		m.targetHeading = 0.0
		m.targetPitch = 0.0
		m.targetRoll = 0.0

		if m.phaseTime > 10*time.Second {
			m.phase = Ground
			m.phaseTime = 0
		}
	}

	// Check for phase transition
	if m.phase == CruiseStraight && m.phaseTime > 10*time.Second && math.Abs(m.currentRoll) < 2.0 {
		if rand.Float64() > 0.5 {
			m.phase = TurnRight
			m.phaseTime = 0
		}
	}
}

func (m *Magnetometer) updateState(dt float64) {
	// Smooth transitions
	headingDiff := m.targetHeading - m.currentHeading
	// Handle heading wraparound
	if headingDiff > 180 {
		headingDiff -= 360
	} else if headingDiff < -180 {
		headingDiff += 360
	}
	m.currentHeading += headingDiff * 0.05

	// Normalize heading to -180 to 180
	if m.currentHeading > 180 {
		m.currentHeading -= 360
	} else if m.currentHeading < -180 {
		m.currentHeading += 360
	}

	pitchDiff := m.targetPitch - m.currentPitch
	m.currentPitch += pitchDiff * 0.1

	rollDiff := m.targetRoll - m.currentRoll
	m.currentRoll += rollDiff * 0.15
}

// Calculate magnetometer readings based on heading and attitude
// Magnetometer measures Earth's magnetic field in body frame
func (m *Magnetometer) calculateMagneticField() (float64, float64, float64) {
	// Earth's magnetic field strength (typical: ~50 μT or 50,000 nT)
	// We'll use normalized units for simplicity
	const fieldStrength = 50.0 // μT

	// Convert angles to radians
	// Add magnetic declination to heading to get magnetic heading
	magneticHeading := (m.currentHeading + m.magneticDeclination) * math.Pi / 180.0
	pitchRad := m.currentPitch * math.Pi / 180.0
	rollRad := m.currentRoll * math.Pi / 180.0

	// Magnetic inclination (dip angle) - varies by location
	// At 37°N (California), inclination is about 60°
	const inclination = 60.0 * math.Pi / 180.0

	// Earth's field in NED (North-East-Down) frame
	northComponent := fieldStrength * math.Cos(inclination)
	downComponent := fieldStrength * math.Sin(inclination)
	eastComponent := 0.0

	// Rotate to body frame using heading, pitch, roll
	// This is a simplified rotation matrix

	// First rotate by heading (yaw)
	nx := northComponent*math.Cos(magneticHeading) - eastComponent*math.Sin(magneticHeading)
	ny := northComponent*math.Sin(magneticHeading) + eastComponent*math.Cos(magneticHeading)
	nz := downComponent

	// Then rotate by pitch
	nx2 := nx*math.Cos(pitchRad) + nz*math.Sin(pitchRad)
	ny2 := ny
	nz2 := -nx*math.Sin(pitchRad) + nz*math.Cos(pitchRad)

	// Finally rotate by roll
	magX := nx2
	magY := ny2*math.Cos(rollRad) - nz2*math.Sin(rollRad)
	magZ := ny2*math.Sin(rollRad) + nz2*math.Cos(rollRad)

	// Add hard iron bias (constant offset from aircraft's magnetic field)
	magX += m.hardIronBias[0]
	magY += m.hardIronBias[1]
	magZ += m.hardIronBias[2]

	// Add noise (magnetic interference, sensor noise)
	noiseX := (rand.Float64()*2.0 - 1.0) * m.noiseLevel
	noiseY := (rand.Float64()*2.0 - 1.0) * m.noiseLevel
	noiseZ := (rand.Float64()*2.0 - 1.0) * m.noiseLevel

	magX += noiseX
	magY += noiseY
	magZ += noiseZ

	return magX, magY, magZ
}

func (m *Magnetometer) getReading() MagnetometerReading {
	mx, my, mz := m.calculateMagneticField()

	return MagnetometerReading{
		Sensor:    "magnetometer",
		MagX:      math.Round(mx*100) / 100,
		MagY:      math.Round(my*100) / 100,
		MagZ:      math.Round(mz*100) / 100,
		Unit:      "μT",
		Timestamp: time.Now(),
	}
}

func (m *Magnetometer) outputReading(reading MagnetometerReading) {
	switch m.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.2f,%.2f,%.2f,%s,%s\n",
			reading.Sensor,
			reading.MagX,
			reading.MagY,
			reading.MagZ,
			reading.Unit,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("X:%.2f Y:%.2f Z:%.2f μT\n",
			reading.MagX,
			reading.MagY,
			reading.MagZ)
	}
}

func main() {
	hz := flag.Float64("hz", 20.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 0.5, "Noise level in μT")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	mag := NewMagnetometer(*hz, *format, *noise)

	if *format == "csv" {
		fmt.Println("sensor,mag_x,mag_y,mag_z,unit,timestamp")
	}

	ticker := time.NewTicker(mag.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := mag.updateRate.Seconds()

		mag.updatePhase(mag.updateRate)
		mag.updateState(dt)

		reading := mag.getReading()
		mag.outputReading(reading)
	}
}
