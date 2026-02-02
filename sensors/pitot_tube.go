package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type PitotReading struct {
	Sensor    string    `json:"sensor"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type FlightPhase int

const (
	Taxiing FlightPhase = iota
	Takeoff
	Climb
	Cruise
	Descent
	Landing
	Stopped
)

type PitotTube struct {
	currentSpeed    float64
	targetSpeed     float64
	currentAltitude float64
	targetAltitude  float64
	phase           FlightPhase
	phaseTime       time.Duration
	updateRate      time.Duration
	noiseLevel      float64
	format          string
	maxAccel        float64
	maxDecel        float64
	groundAltitude  float64
}

func NewPitotTube(hz float64, format string, noiseLevel float64) *PitotTube {
	groundAlt := 100.0 + rand.Float64()*400.0
	return &PitotTube{
		currentSpeed:    0.0,
		targetSpeed:     0.0,
		currentAltitude: groundAlt,
		targetAltitude:  groundAlt,
		phase:           Taxiing,
		phaseTime:       0,
		updateRate:      time.Duration(float64(time.Second) / hz),
		noiseLevel:      noiseLevel,
		format:          format,
		maxAccel:        20.0,
		maxDecel:        30.0,
		groundAltitude:  groundAlt,
	}
}

func (p *PitotTube) updatePhase(elapsed time.Duration) {
	p.phaseTime += elapsed

	switch p.phase {
	case Taxiing:
		p.targetSpeed = 15.0 + rand.Float64()*10.0
		p.targetAltitude = p.groundAltitude
		if p.phaseTime > 20*time.Second {
			p.phase = Takeoff
			p.phaseTime = 0
		}

	case Takeoff:
		p.targetSpeed = 80.0
		p.targetAltitude = p.groundAltitude + 500.0
		if p.phaseTime > 15*time.Second {
			p.phase = Climb
			p.phaseTime = 0
		}

	case Climb:
		progress := math.Min(1.0, float64(p.phaseTime)/float64(45*time.Second))
		p.targetSpeed = 80.0 + progress*70.0
		p.targetAltitude = p.groundAltitude + 500.0 + progress*4500.0
		if p.phaseTime > 45*time.Second {
			p.phase = Cruise
			p.phaseTime = 0
		}

	case Cruise:
		baseSpeed := 150.0
		variation := math.Sin(float64(p.phaseTime)/float64(30*time.Second)) * 5.0
		p.targetSpeed = baseSpeed + variation
		baseAltitude := p.groundAltitude + 5000.0
		altVariation := math.Sin(float64(p.phaseTime)/float64(20*time.Second)) * 50.0
		p.targetAltitude = baseAltitude + altVariation
		if p.phaseTime > 90*time.Second {
			p.phase = Descent
			p.phaseTime = 0
		}

	case Descent:
		progress := math.Min(1.0, float64(p.phaseTime)/float64(40*time.Second))
		p.targetSpeed = 150.0 - progress*70.0
		startAlt := p.groundAltitude + 5000.0
		p.targetAltitude = startAlt - progress*4500.0
		if p.phaseTime > 40*time.Second {
			p.phase = Landing
			p.phaseTime = 0
		}

	case Landing:
		p.targetSpeed = 15.0
		progress := math.Min(1.0, float64(p.phaseTime)/float64(20*time.Second))
		p.targetAltitude = p.groundAltitude + 500.0 - progress*500.0
		if p.phaseTime > 20*time.Second {
			p.phase = Stopped
			p.phaseTime = 0
		}

	case Stopped:
		p.targetSpeed = 0.0
		p.targetAltitude = p.groundAltitude
		if p.phaseTime > 10*time.Second {
			p.phase = Taxiing
			p.phaseTime = 0
		}
	}
}

func (p *PitotTube) updateState(dt float64) {
	// Update speed
	diff := p.targetSpeed - p.currentSpeed
	var maxChange float64
	if diff > 0 {
		maxChange = p.maxAccel * dt
	} else {
		maxChange = p.maxDecel * dt
	}
	if math.Abs(diff) <= maxChange {
		p.currentSpeed = p.targetSpeed
	} else {
		p.currentSpeed += math.Copysign(maxChange, diff)
	}
	if p.currentSpeed < 0 {
		p.currentSpeed = 0
	}

	// Update altitude
	altDiff := p.targetAltitude - p.currentAltitude
	maxAltChange := 16.67 * dt // ~1000 fpm climb rate
	if math.Abs(altDiff) <= maxAltChange {
		p.currentAltitude = p.targetAltitude
	} else {
		p.currentAltitude += math.Copysign(maxAltChange, altDiff)
	}
}

// Calculate total pressure from airspeed and altitude
// Pitot tube measures: P_total = P_static + P_dynamic
// P_dynamic = 0.5 * ρ * v^2
func (p *PitotTube) calculatePitotPressure() float64 {
	// Standard atmospheric pressure at sea level (inHg)
	const P0 = 29.92
	
	// Calculate static pressure based on altitude (simplified barometric formula)
	// P = P0 * (1 - 0.0000068756 * altitude)^5.2559
	altitudeFeet := p.currentAltitude
	staticPressure := P0 * math.Pow(1.0-0.0000068756*altitudeFeet, 5.2559)
	
	// Calculate dynamic pressure from airspeed
	// Dynamic pressure in inHg: q = 0.00002378 * V^2 / 2
	// This is a simplified formula for low speeds
	speedKnots := p.currentSpeed * 0.868976 // Convert mph to knots
	dynamicPressure := 0.00002378 * speedKnots * speedKnots / 2.0
	
	// Total pressure = static + dynamic
	totalPressure := staticPressure + dynamicPressure
	
	// Add sensor noise
	noise := (rand.Float64()*2.0 - 1.0) * p.noiseLevel
	totalPressure += noise
	
	return totalPressure
}

func (p *PitotTube) getReading() PitotReading {
	pressure := p.calculatePitotPressure()
	
	return PitotReading{
		Sensor:    "pitot",
		Value:     math.Round(pressure*1000) / 1000, // Round to 3 decimals
		Unit:      "inHg",
		Timestamp: time.Now(),
	}
}

func (p *PitotTube) outputReading(reading PitotReading) {
	switch p.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.3f,%s,%s\n",
			reading.Sensor,
			reading.Value,
			reading.Unit,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("%.3f inHg\n", reading.Value)
	}
}

func main() {
	hz := flag.Float64("hz", 10.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 0.001, "Noise level in inHg")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	pitot := NewPitotTube(*hz, *format, *noise)

	if *format == "csv" {
		fmt.Println("sensor,value,unit,timestamp")
	}

	ticker := time.NewTicker(pitot.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := pitot.updateRate.Seconds()
		
		pitot.updatePhase(pitot.updateRate)
		pitot.updateState(dt)
		
		reading := pitot.getReading()
		pitot.outputReading(reading)
	}
}
