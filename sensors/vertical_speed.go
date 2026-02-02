package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type VSIReading struct {
	Sensor    string    `json:"sensor"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type FlightPhase int

const (
	Ground FlightPhase = iota
	Takeoff
	Climb
	Cruise
	Descent
	Landing
	OnGround
)

type VerticalSpeedIndicator struct {
	currentVS  float64
	targetVS   float64
	phase      FlightPhase
	phaseTime  time.Duration
	updateRate time.Duration
	noiseLevel float64
	format     string
	lagFactor  float64 // VSI has mechanical lag in real instruments
}

func NewVSI(hz float64, format string, noiseLevel float64) *VerticalSpeedIndicator {
	return &VerticalSpeedIndicator{
		currentVS:  0.0,
		targetVS:   0.0,
		phase:      Ground,
		phaseTime:  0,
		updateRate: time.Duration(float64(time.Second) / hz),
		noiseLevel: noiseLevel,
		format:     format,
		lagFactor:  0.15, // VSI lags behind actual vertical speed
	}
}

func (v *VerticalSpeedIndicator) updatePhase(elapsed time.Duration) {
	v.phaseTime += elapsed

	switch v.phase {
	case Ground:
		v.targetVS = 0.0
		if v.phaseTime > 20*time.Second {
			v.phase = Takeoff
			v.phaseTime = 0
		}

	case Takeoff:
		// Initial rotation and climb
		progress := math.Min(1.0, float64(v.phaseTime)/float64(15*time.Second))
		v.targetVS = progress * 6.1 // Build up to 6.1 m/s climb (~1200 fpm)
		if v.phaseTime > 15*time.Second {
			v.phase = Climb
			v.phaseTime = 0
		}

	case Climb:
		// Steady climb with slight variations
		baseVS := 5.1 // ~1000 fpm
		variation := math.Sin(float64(v.phaseTime)/float64(10*time.Second)) * 0.5
		v.targetVS = baseVS + variation
		if v.phaseTime > 45*time.Second {
			v.phase = Cruise
			v.phaseTime = 0
		}

	case Cruise:
		// Small variations due to turbulence and altitude corrections
		variation := math.Sin(float64(v.phaseTime)/float64(20*time.Second)) * 0.25
		v.targetVS = variation
		if v.phaseTime > 90*time.Second {
			v.phase = Descent
			v.phaseTime = 0
		}

	case Descent:
		// Controlled descent
		baseVS := -4.1 // ~-800 fpm
		variation := math.Sin(float64(v.phaseTime)/float64(12*time.Second)) * 0.5
		v.targetVS = baseVS + variation
		if v.phaseTime > 40*time.Second {
			v.phase = Landing
			v.phaseTime = 0
		}

	case Landing:
		// Final approach - more controlled descent
		progress := float64(v.phaseTime) / float64(20*time.Second)
		if progress < 0.7 {
			// Steady descent on approach
			v.targetVS = -2.5 // ~-500 fpm
		} else {
			// Flare - reduce descent rate
			flareProgress := (progress - 0.7) / 0.3
			v.targetVS = -2.5 + flareProgress*2.5
		}

		if v.phaseTime > 20*time.Second {
			v.phase = OnGround
			v.phaseTime = 0
		}

	case OnGround:
		v.targetVS = 0.0
		if v.phaseTime > 10*time.Second {
			v.phase = Ground
			v.phaseTime = 0
		}
	}
}

func (v *VerticalSpeedIndicator) updateVS(dt float64) {
	// VSI has inherent lag due to mechanical design (pressure differential)
	// It doesn't instantly show the actual vertical speed
	diff := v.targetVS - v.currentVS
	v.currentVS += diff * v.lagFactor

	// Add realistic noise and oscillation
	noise := (rand.Float64()*2.0 - 1.0) * v.noiseLevel
	v.currentVS += noise

	// Add slight oscillation common in VSI instruments
	oscillation := math.Sin(float64(time.Now().UnixNano())/1e9*10) * 0.05
	v.currentVS += oscillation
}

func (v *VerticalSpeedIndicator) getReading() VSIReading {
	// Round to nearest 0.1 m/s
	rounded := math.Round(v.currentVS*10.0) / 10.0

	return VSIReading{
		Sensor:    "vertical_speed",
		Value:     rounded,
		Unit:      "m/s",
		Timestamp: time.Now(),
	}
}

func (v *VerticalSpeedIndicator) outputReading(reading VSIReading) {
	switch v.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.1f,%s,%s\n",
			reading.Sensor,
			reading.Value,
			reading.Unit,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		sign := ""
		if reading.Value > 0 {
			sign = "+"
		}
		fmt.Printf("%s%.1f m/s\n", sign, reading.Value)
	}
}

func main() {
	hz := flag.Float64("hz", 10.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 0.025, "Noise level in m/s")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	vsi := NewVSI(*hz, *format, *noise)

	// Print CSV header if using CSV format
	if *format == "csv" {
		fmt.Println("sensor,value,unit,timestamp")
	}

	ticker := time.NewTicker(vsi.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := vsi.updateRate.Seconds()

		vsi.updatePhase(vsi.updateRate)
		vsi.updateVS(dt)

		reading := vsi.getReading()
		vsi.outputReading(reading)
	}
}
