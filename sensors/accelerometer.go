package sensors
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type AccelerometerReading struct {
	Sensor    string    `json:"sensor"`
	AccelX    float64   `json:"accel_x"`
	AccelY    float64   `json:"accel_y"`
	AccelZ    float64   `json:"accel_z"`
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

type Accelerometer struct {
	currentPitch float64
	currentRoll  float64
	targetPitch  float64
	targetRoll   float64
	currentSpeed float64
	targetSpeed  float64
	phase        FlightPhase
	phaseTime    time.Duration
	updateRate   time.Duration
	noiseLevel   float64
	format       string
}

func NewAccelerometer(hz float64, format string, noiseLevel float64) *Accelerometer {
	return &Accelerometer{
		currentPitch: 0.0,
		currentRoll:  0.0,
		targetPitch:  0.0,
		targetRoll:   0.0,
		currentSpeed: 0.0,
		targetSpeed:  0.0,
		phase:        Ground,
		phaseTime:    0,
		updateRate:   time.Duration(float64(time.Second) / hz),
		noiseLevel:   noiseLevel,
		format:       format,
	}
}

func (a *Accelerometer) updatePhase(elapsed time.Duration) {
	a.phaseTime += elapsed

	switch a.phase {
	case Ground:
		a.targetPitch = 0.0
		a.targetRoll = 0.0
		a.targetSpeed = 0.0
		if a.phaseTime > 20*time.Second {
			a.phase = Takeoff
			a.phaseTime = 0
		}

	case Takeoff:
		progress := math.Min(1.0, float64(a.phaseTime)/float64(5*time.Second))
		a.targetPitch = progress * 10.0
		a.targetRoll = 0.0
		a.targetSpeed = 70.0 // kts

		if a.phaseTime > 15*time.Second {
			a.phase = Climb
			a.phaseTime = 0
		}

	case Climb:
		a.targetPitch = 8.0 + math.Sin(float64(a.phaseTime)/float64(8*time.Second))*1.0
		a.targetRoll = 0.0
		progress := math.Min(1.0, float64(a.phaseTime)/float64(45*time.Second))
		a.targetSpeed = 70.0 + progress*60.0 // kts

		if a.phaseTime > 45*time.Second {
			a.phase = CruiseStraight
			a.phaseTime = 0
		}

	case CruiseStraight:
		a.targetPitch = 2.0 + math.Sin(float64(a.phaseTime)/float64(15*time.Second))*0.5
		a.targetRoll = 0.0
		a.targetSpeed = 130.0 // kts

		if a.phaseTime > 20*time.Second {
			a.phase = TurnLeft
			a.phaseTime = 0
		}

	case TurnLeft:
		progress := math.Min(1.0, float64(a.phaseTime)/float64(3*time.Second))
		if progress < 0.5 {
			a.targetRoll = -progress * 2.0 * 18.0
		} else {
			a.targetRoll = -18.0
		}
		a.targetPitch = 3.0
		a.targetSpeed = 130.0 // kts

		if a.phaseTime > 15*time.Second {
			a.phase = CruiseStraight
			a.phaseTime = 0
		}

	case TurnRight:
		progress := math.Min(1.0, float64(a.phaseTime)/float64(3*time.Second))
		if progress < 0.5 {
			a.targetRoll = progress * 2.0 * 18.0
		} else {
			a.targetRoll = 18.0
		}
		a.targetPitch = 3.0
		a.targetSpeed = 130.0 // kts

		if a.phaseTime > 15*time.Second {
			a.targetRoll = 0.0
			if a.phaseTime > 18*time.Second {
				a.phase = Descent
				a.phaseTime = 0
			}
		}

	case Descent:
		a.targetPitch = -2.0 + math.Sin(float64(a.phaseTime)/float64(10*time.Second))*0.5
		a.targetRoll = 0.0
		progress := math.Min(1.0, float64(a.phaseTime)/float64(40*time.Second))
		a.targetSpeed = 130.0 - progress*60.0 // kts

		if a.phaseTime > 40*time.Second {
			a.phase = Landing
			a.phaseTime = 0
		}

	case Landing:
		progress := float64(a.phaseTime) / float64(20*time.Second)

		if progress < 0.8 {
			a.targetPitch = -1.0
		} else {
			flareProgress := (progress - 0.8) / 0.2
			a.targetPitch = -1.0 + flareProgress*6.0
		}

		a.targetRoll = math.Sin(float64(a.phaseTime)/float64(3*time.Second)) * 3.0
		a.targetSpeed = 70.0 - progress*57.0 // kts (down to ~13 kts)

		if a.phaseTime > 20*time.Second {
			a.phase = OnGround
			a.phaseTime = 0
		}

	case OnGround:
		a.targetPitch = 0.0
		a.targetRoll = 0.0
		a.targetSpeed = 0.0

		if a.phaseTime > 10*time.Second {
			a.phase = Ground
			a.phaseTime = 0
		}
	}

	// Check for phase transition
	if a.phase == CruiseStraight && a.phaseTime > 10*time.Second && math.Abs(a.currentRoll) < 2.0 {
		if rand.Float64() > 0.5 {
			a.phase = TurnRight
			a.phaseTime = 0
		}
	}
}

func (a *Accelerometer) updateState(dt float64) {
	// Smooth transitions
	pitchDiff := a.targetPitch - a.currentPitch
	a.currentPitch += pitchDiff * 0.1

	rollDiff := a.targetRoll - a.currentRoll
	a.currentRoll += rollDiff * 0.15

	speedDiff := a.targetSpeed - a.currentSpeed
	maxSpeedChange := 17.4 * dt // kts/s
	if math.Abs(speedDiff) <= maxSpeedChange {
		a.currentSpeed = a.targetSpeed
	} else {
		a.currentSpeed += math.Copysign(maxSpeedChange, speedDiff)
	}
}

// Calculate accelerometer readings based on attitude and motion
// Accelerometer measures specific force (acceleration - gravity) in body frame
func (a *Accelerometer) calculateAcceleration() (float64, float64, float64) {
	// Convert angles to radians
	pitchRad := a.currentPitch * math.Pi / 180.0
	rollRad := a.currentRoll * math.Pi / 180.0

	// Gravity in body frame (standard 1g = 9.81 m/s²)
	// When level: ax=0, ay=0, az=-1g (gravity pulls down)
	const g = 9.81

	// Gravity components in body frame
	ax := -g * math.Sin(pitchRad)
	ay := g * math.Sin(rollRad) * math.Cos(pitchRad)
	az := -g * math.Cos(rollRad) * math.Cos(pitchRad)

	// Add centripetal acceleration during turns
	// In a coordinated turn: lateral acceleration = g * tan(bank_angle)
	if math.Abs(a.currentRoll) > 5.0 {
		// Turning - add centripetal acceleration
		turnAccel := g * math.Tan(rollRad)
		ay += turnAccel * 0.1 // Simplified model
	}

	// Add longitudinal acceleration from speed changes
	// a = dv/dt (very simplified - would need proper dynamics)
	speedChange := a.targetSpeed - a.currentSpeed
	longitudinalAccel := speedChange * 0.05
	ax += longitudinalAccel

	// Add noise (vibration, turbulence)
	noiseX := (rand.Float64()*2.0 - 1.0) * a.noiseLevel
	noiseY := (rand.Float64()*2.0 - 1.0) * a.noiseLevel
	noiseZ := (rand.Float64()*2.0 - 1.0) * a.noiseLevel

	ax += noiseX
	ay += noiseY
	az += noiseZ

	return ax, ay, az
}

func (a *Accelerometer) getReading() AccelerometerReading {
	ax, ay, az := a.calculateAcceleration()

	return AccelerometerReading{
		Sensor:    "accelerometer",
		AccelX:    math.Round(ax*100) / 100,
		AccelY:    math.Round(ay*100) / 100,
		AccelZ:    math.Round(az*100) / 100,
		Unit:      "m/s²",
		Timestamp: time.Now(),
	}
}

func (a *Accelerometer) outputReading(reading AccelerometerReading) {
	switch a.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.2f,%.2f,%.2f,%s,%s\n",
			reading.Sensor,
			reading.AccelX,
			reading.AccelY,
			reading.AccelZ,
			reading.Unit,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("X:%.2f Y:%.2f Z:%.2f m/s²\n",
			reading.AccelX,
			reading.AccelY,
			reading.AccelZ)
	}
}

func main() {
	hz := flag.Float64("hz", 50.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 0.1, "Noise level in m/s²")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	accel := NewAccelerometer(*hz, *format, *noise)

	if *format == "csv" {
		fmt.Println("sensor,accel_x,accel_y,accel_z,unit,timestamp")
	}

	ticker := time.NewTicker(accel.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := accel.updateRate.Seconds()

		accel.updatePhase(accel.updateRate)
		accel.updateState(dt)

		reading := accel.getReading()
		accel.outputReading(reading)
	}
}