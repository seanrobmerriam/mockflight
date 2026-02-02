package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type StaticReading struct {
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

type StaticPort struct {
	currentAltitude float64
	targetAltitude  float64
	phase           FlightPhase
	phaseTime       time.Duration
	updateRate      time.Duration
	noiseLevel      float64
	format          string
	groundAltitude  float64
	maxClimbRate    float64
	maxDescentRate  float64
}

func NewStaticPort(hz float64, format string, noiseLevel float64) *StaticPort {
	groundAlt := 100.0 + rand.Float64()*400.0
	return &StaticPort{
		currentAltitude: groundAlt,
		targetAltitude:  groundAlt,
		phase:           Ground,
		phaseTime:       0,
		updateRate:      time.Duration(float64(time.Second) / hz),
		noiseLevel:      noiseLevel,
		format:          format,
		groundAltitude:  groundAlt,
		maxClimbRate:    16.67,  // ~1000 fpm
		maxDescentRate:  13.33,  // ~800 fpm
	}
}

func (s *StaticPort) updatePhase(elapsed time.Duration) {
	s.phaseTime += elapsed

	switch s.phase {
	case Ground:
		s.targetAltitude = s.groundAltitude
		if s.phaseTime > 20*time.Second {
			s.phase = Takeoff
			s.phaseTime = 0
		}

	case Takeoff:
		s.targetAltitude = s.groundAltitude + 500.0
		if s.phaseTime > 15*time.Second {
			s.phase = Climb
			s.phaseTime = 0
		}

	case Climb:
		progress := math.Min(1.0, float64(s.phaseTime)/float64(45*time.Second))
		s.targetAltitude = s.groundAltitude + 500.0 + progress*4500.0
		if s.phaseTime > 45*time.Second {
			s.phase = Cruise
			s.phaseTime = 0
		}

	case Cruise:
		baseAltitude := s.groundAltitude + 5000.0
		variation := math.Sin(float64(s.phaseTime)/float64(20*time.Second)) * 50.0
		s.targetAltitude = baseAltitude + variation
		if s.phaseTime > 90*time.Second {
			s.phase = Descent
			s.phaseTime = 0
		}

	case Descent:
		progress := math.Min(1.0, float64(s.phaseTime)/float64(40*time.Second))
		startAlt := s.groundAltitude + 5000.0
		s.targetAltitude = startAlt - progress*4500.0
		if s.phaseTime > 40*time.Second {
			s.phase = Landing
			s.phaseTime = 0
		}

	case Landing:
		progress := math.Min(1.0, float64(s.phaseTime)/float64(20*time.Second))
		s.targetAltitude = s.groundAltitude + 500.0 - progress*500.0
		if s.phaseTime > 20*time.Second {
			s.phase = OnGround
			s.phaseTime = 0
		}

	case OnGround:
		s.targetAltitude = s.groundAltitude
		if s.phaseTime > 10*time.Second {
			s.phase = Ground
			s.phaseTime = 0
		}
	}
}

func (s *StaticPort) updateAltitude(dt float64) {
	diff := s.targetAltitude - s.currentAltitude
	
	var maxChange float64
	if diff > 0 {
		maxChange = s.maxClimbRate * dt
	} else {
		maxChange = s.maxDescentRate * dt
	}

	if math.Abs(diff) <= maxChange {
		s.currentAltitude = s.targetAltitude
	} else {
		s.currentAltitude += math.Copysign(maxChange, diff)
	}

	if s.currentAltitude < s.groundAltitude {
		s.currentAltitude = s.groundAltitude
	}
}

// Calculate static pressure based on altitude using standard atmosphere model
// This is what the static port measures - ambient atmospheric pressure
func (s *StaticPort) calculateStaticPressure() float64 {
	// Standard atmospheric pressure at sea level (inHg)
	const P0 = 29.92
	
	// Standard atmosphere barometric formula
	// P = P0 * (1 - L * h / T0)^(g * M / (R * L))
	// Simplified for troposphere: P = P0 * (1 - 0.0000068756 * h)^5.2559
	altitudeFeet := s.currentAltitude
	staticPressure := P0 * math.Pow(1.0-0.0000068756*altitudeFeet, 5.2559)
	
	// Add sensor noise (turbulence, position error, etc.)
	noise := (rand.Float64()*2.0 - 1.0) * s.noiseLevel
	staticPressure += noise
	
	// Ensure pressure doesn't go negative
	if staticPressure < 0 {
		staticPressure = 0.001
	}
	
	return staticPressure
}

func (s *StaticPort) getReading() StaticReading {
	pressure := s.calculateStaticPressure()
	
	return StaticReading{
		Sensor:    "static",
		Value:     math.Round(pressure*1000) / 1000, // Round to 3 decimals
		Unit:      "inHg",
		Timestamp: time.Now(),
	}
}

func (s *StaticPort) outputReading(reading StaticReading) {
	switch s.format {
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

	staticPort := NewStaticPort(*hz, *format, *noise)

	if *format == "csv" {
		fmt.Println("sensor,value,unit,timestamp")
	}

	ticker := time.NewTicker(staticPort.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := staticPort.updateRate.Seconds()
		
		staticPort.updatePhase(staticPort.updateRate)
		staticPort.updateAltitude(dt)
		
		reading := staticPort.getReading()
		staticPort.outputReading(reading)
	}
}
