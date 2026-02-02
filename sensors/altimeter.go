package sensors

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

type AltimeterReading struct {
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

type Altimeter struct {
	currentAltitude float64
	targetAltitude  float64
	phase           FlightPhase
	phaseTime       time.Duration
	updateRate      time.Duration
	noiseLevel      float64
	format          string
	maxClimbRate    float64 // feet per second
	maxDescentRate  float64 // feet per second
	groundAltitude  float64
}

func NewAltimeter(hz float64, format string, noiseLevel float64) *Altimeter {
	groundAlt := 100.0 + rand.Float64()*400.0 // Random field elevation 100-500 ft MSL
	return &Altimeter{
		currentAltitude: groundAlt,
		targetAltitude:  groundAlt,
		phase:           Ground,
		phaseTime:       0,
		updateRate:      time.Duration(float64(time.Second) / hz),
		noiseLevel:      noiseLevel,
		format:          format,
		maxClimbRate:    16.67,  // ~1000 fpm typical for small aircraft
		maxDescentRate:  13.33,  // ~800 fpm
		groundAltitude:  groundAlt,
	}
}

func (a *Altimeter) updatePhase(elapsed time.Duration) {
	a.phaseTime += elapsed

	switch a.phase {
	case Ground:
		a.targetAltitude = a.groundAltitude
		if a.phaseTime > 20*time.Second {
			a.phase = Takeoff
			a.phaseTime = 0
		}

	case Takeoff:
		// Rapid initial climb
		a.targetAltitude = a.groundAltitude + 500.0
		if a.phaseTime > 15*time.Second {
			a.phase = Climb
			a.phaseTime = 0
		}

	case Climb:
		// Steady climb to cruise altitude
		progress := math.Min(1.0, float64(a.phaseTime)/float64(45*time.Second))
		a.targetAltitude = a.groundAltitude + 500.0 + progress*4500.0 // Climb to 5000 ft AGL
		if a.phaseTime > 45*time.Second {
			a.phase = Cruise
			a.phaseTime = 0
		}

	case Cruise:
		// Cruise altitude with small variations (turbulence, corrections)
		baseAltitude := a.groundAltitude + 5000.0
		variation := math.Sin(float64(a.phaseTime)/float64(20*time.Second)) * 50.0
		a.targetAltitude = baseAltitude + variation
		if a.phaseTime > 90*time.Second {
			a.phase = Descent
			a.phaseTime = 0
		}

	case Descent:
		// Gradual descent
		progress := math.Min(1.0, float64(a.phaseTime)/float64(40*time.Second))
		startAlt := a.groundAltitude + 5000.0
		a.targetAltitude = startAlt - progress*4500.0 // Descend to pattern altitude
		if a.phaseTime > 40*time.Second {
			a.phase = Landing
			a.phaseTime = 0
		}

	case Landing:
		// Final approach
		progress := math.Min(1.0, float64(a.phaseTime)/float64(20*time.Second))
		a.targetAltitude = a.groundAltitude + 500.0 - progress*500.0
		if a.phaseTime > 20*time.Second {
			a.phase = OnGround
			a.phaseTime = 0
		}

	case OnGround:
		a.targetAltitude = a.groundAltitude
		if a.phaseTime > 10*time.Second {
			// Restart the cycle
			a.phase = Ground
			a.phaseTime = 0
		}
	}
}

func (a *Altimeter) updateAltitude(dt float64) {
	// Move toward target altitude with realistic climb/descent rates
	diff := a.targetAltitude - a.currentAltitude
	
	var maxChange float64
	if diff > 0 {
		maxChange = a.maxClimbRate * dt
	} else {
		maxChange = a.maxDescentRate * dt
	}

	if math.Abs(diff) <= maxChange {
		a.currentAltitude = a.targetAltitude
	} else {
		a.currentAltitude += math.Copysign(maxChange, diff)
	}

	// Add realistic altimeter noise (barometric variations, etc.)
	noise := (rand.Float64()*2.0 - 1.0) * a.noiseLevel * 10.0 // ±noise feet
	a.currentAltitude += noise

	// Ensure we don't go below ground
	if a.currentAltitude < a.groundAltitude {
		a.currentAltitude = a.groundAltitude
	}
}

func (a *Altimeter) getReading() AltimeterReading {
	return AltimeterReading{
		Sensor:    "altimeter",
		Value:     math.Round(a.currentAltitude),
		Unit:      "feet",
		Timestamp: time.Now(),
	}
}

func (a *Altimeter) outputReading(reading AltimeterReading) {
	switch a.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.0f,%s,%s\n",
			reading.Sensor,
			reading.Value,
			reading.Unit,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("%.0f ft\n", reading.Value)
	}
}

func main() {
	hz := flag.Float64("hz", 10.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 1.0, "Noise level in feet")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	altimeter := NewAltimeter(*hz, *format, *noise)

	// Print CSV header if using CSV format
	if *format == "csv" {
		fmt.Println("sensor,value,unit,timestamp")
	}

	ticker := time.NewTicker(altimeter.updateRate)
	defer ticker.Stop()

	for range ticker.C {
		dt := altimeter.updateRate.Seconds()
		
		altimeter.updatePhase(altimeter.updateRate)
		altimeter.updateAltitude(dt)
		
		reading := altimeter.getReading()
		altimeter.outputReading(reading)
	}
}
