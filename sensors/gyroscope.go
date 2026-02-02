package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type GyroscopeReading struct {
	Sensor    string    `json:"sensor"`
	RollRate  float64   `json:"roll_rate"`
	PitchRate float64   `json:"pitch_rate"`
	YawRate   float64   `json:"yaw_rate"`
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

type Gyroscope struct {
	currentRollRate  float64
	currentPitchRate float64
	currentYawRate   float64
	targetRollRate   float64
	targetPitchRate  float64
	targetYawRate    float64
	phase            FlightPhase
	phaseTime        time.Duration
	updateRate       time.Duration
	noiseLevel       float64
	driftRate        float64
	bias             [3]float64 // Gyro bias (drift) for each axis
	format           string
}

func NewGyroscope(hz float64, format string, noiseLevel float64) *Gyroscope {
	return &Gyroscope{
		currentRollRate:  0.0,
		currentPitchRate: 0.0,
		currentYawRate:   0.0,
		targetRollRate:   0.0,
		targetPitchRate:  0.0,
		targetYawRate:    0.0,
		phase:            Ground,
		phaseTime:        0,
		updateRate:       time.Duration(float64(time.Second) / hz),
		noiseLevel:       noiseLevel,
		driftRate:        0.01, // degrees/sec drift
		bias:             [3]float64{rand.Float64()*0.1 - 0.05, rand.Float64()*0.1 - 0.05, rand.Float64()*0.1 - 0.05},
		format:           format,
	}
}

func (g *Gyroscope) updatePhase(elapsed time.Duration) {
	g.phaseTime += elapsed

	switch g.phase {
	case Ground:
		g.targetRollRate = 0.0
		g.targetPitchRate = 0.0
		g.targetYawRate = 0.0
		if g.phaseTime > 20*time.Second {
			g.phase = Takeoff
			g.phaseTime = 0
		}

	case Takeoff:
		// Rotation - pitch up
		progress := math.Min(1.0, float64(g.phaseTime)/float64(5*time.Second))
		if progress < 0.8 {
			g.targetPitchRate = 2.0 // Pitch up at 2 deg/sec
		} else {
			g.targetPitchRate = 0.0 // Level off
		}
		g.targetRollRate = 0.0
		g.targetYawRate = 0.0

		if g.phaseTime > 15*time.Second {
			g.phase = Climb
			g.phaseTime = 0
		}

	case Climb:
		// Small pitch adjustments during climb
		g.targetPitchRate = math.Sin(float64(g.phaseTime)/float64(8*time.Second)) * 0.5
		g.targetRollRate = 0.0
		g.targetYawRate = 0.0

		if g.phaseTime > 45*time.Second {
			g.phase = CruiseStraight
			g.phaseTime = 0
		}

	case CruiseStraight:
		// Very small rates during level flight
		g.targetPitchRate = math.Sin(float64(g.phaseTime)/float64(15*time.Second)) * 0.2
		g.targetRollRate = 0.0
		g.targetYawRate = 0.0

		if g.phaseTime > 20*time.Second {
			g.phase = TurnLeft
			g.phaseTime = 0
		}

	case TurnLeft:
		progress := float64(g.phaseTime) / float64(15*time.Second)

		if progress < 0.2 {
			// Rolling into turn - high roll rate
			g.targetRollRate = -6.0 // Roll left at 6 deg/sec
			g.targetYawRate = 0.0
		} else if progress < 0.8 {
			// In the turn - coordinated
			g.targetRollRate = 0.0
			// Standard rate turn: 3 deg/sec turn rate
			g.targetYawRate = -3.0
		} else {
			// Rolling out
			g.targetRollRate = 6.0 // Roll right to level
			g.targetYawRate = 0.0
		}

		// Slight pitch up during turn
		g.targetPitchRate = 0.3

		if g.phaseTime > 15*time.Second {
			g.phase = CruiseStraight
			g.phaseTime = 0
		}

	case TurnRight:
		progress := float64(g.phaseTime) / float64(15*time.Second)

		if progress < 0.2 {
			// Rolling into turn
			g.targetRollRate = 6.0
			g.targetYawRate = 0.0
		} else if progress < 0.8 {
			// In the turn
			g.targetRollRate = 0.0
			g.targetYawRate = 3.0
		} else {
			// Rolling out
			g.targetRollRate = -6.0
			g.targetYawRate = 0.0
		}

		g.targetPitchRate = 0.3

		if g.phaseTime > 15*time.Second {
			g.targetRollRate = 0.0
			g.targetYawRate = 0.0
			if g.phaseTime > 18*time.Second {
				g.phase = Descent
				g.phaseTime = 0
			}
		}

	case Descent:
		// Small nose-down pitch rate initially
		if g.phaseTime < 5*time.Second {
			g.targetPitchRate = -1.0
		} else {
			g.targetPitchRate = math.Sin(float64(g.phaseTime)/float64(10*time.Second)) * 0.3
		}
		g.targetRollRate = 0.0
		g.targetYawRate = 0.0

		if g.phaseTime > 40*time.Second {
			g.phase = Landing
			g.phaseTime = 0
		}

	case Landing:
		progress := float64(g.phaseTime) / float64(20*time.Second)

		if progress < 0.8 {
			// Approach - stable
			g.targetPitchRate = 0.1
		} else {
			// Flare - pitch up quickly
			g.targetPitchRate = 2.5
		}

		// Small roll corrections
		g.targetRollRate = math.Sin(float64(g.phaseTime)/float64(3*time.Second)) * 1.0
		g.targetYawRate = 0.0

		if g.phaseTime > 20*time.Second {
			g.phase = OnGround
			g.phaseTime = 0
		}

	case OnGround:
		g.targetRollRate = 0.0
		g.targetPitchRate = 0.0
		g.targetYawRate = 0.0

		if g.phaseTime > 10*time.Second {
			g.phase = Ground
			g.phaseTime = 0
		}
	}

	// Check for phase transition to TurnRight
	if g.phase == CruiseStraight && g.phaseTime > 10*time.Second && g.currentRollRate < 1.0 && g.currentRollRate > -1.0 {
		// Only transition if we're close to wings level
		if rand.Float64() > 0.5 {
			g.phase = TurnRight
			g.phaseTime = 0
		}
	}
}

func (g *Gyroscope) updateRates(dt float64) {
	// Smooth transition to target rates
	smoothing := 0.1

	rollDiff := g.targetRollRate - g.currentRollRate
	g.currentRollRate += rollDiff * smoothing

	pitchDiff := g.targetPitchRate - g.currentPitchRate
	g.currentPitchRate += pitchDiff * smoothing

	yawDiff := g.targetYawRate - g.currentYawRate
	g.currentYawRate += yawDiff * smoothing

	// Add gyro bias (drift over time)
	g.currentRollRate += g.bias[0]
	g.currentPitchRate += g.bias[1]
	g.currentYawRate += g.bias[2]

	// Add noise (random walk)
	noiseRoll := (rand.Float64()*2.0 - 1.0) * g.noiseLevel
	noisePitch := (rand.Float64()*2.0 - 1.0) * g.noiseLevel
	noiseYaw := (rand.Float64()*2.0 - 1.0) * g.noiseLevel

	g.currentRollRate += noiseRoll
	g.currentPitchRate += noisePitch
	g.currentYawRate += noiseYaw
}

func (g *Gyroscope) getReading() GyroscopeReading {
	return GyroscopeReading{
		Sensor:    "gyroscope",
		RollRate:  math.Round(g.currentRollRate*100) / 100,
		PitchRate: math.Round(g.currentPitchRate*100) / 100,
		YawRate:   math.Round(g.currentYawRate*100) / 100,
		Unit:      "deg/s",
		Timestamp: time.Now(),
	}
}

func (g *Gyroscope) outputReading(reading GyroscopeReading) {
	switch g.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.2f,%.2f,%.2f,%s,%s\n",
			reading.Sensor,
			reading.RollRate,
			reading.PitchRate,
			reading.YawRate,
			reading.Unit,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("Roll:%.2f Pitch:%.2f Yaw:%.2f deg/s\n",
			reading.RollRate,
			reading.PitchRate,
			reading.YawRate)
	}
}

func main() {
	hz := flag.Float64("hz", 50.0, "Update frequency in Hz (gyros need high rate)")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 0.05, "Noise level in deg/s")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	gyro := NewGyroscope(*hz, *format, *noise)

	if *format == "csv" {
		fmt.Println("sensor,roll_rate,pitch_rate,yaw_rate,unit,timestamp")
	}

	ticker := time.NewTicker(gyro.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := gyro.updateRate.Seconds()

		gyro.updatePhase(gyro.updateRate)
		gyro.updateRates(dt)

		reading := gyro.getReading()
		gyro.outputReading(reading)
	}
}
