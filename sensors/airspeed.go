package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"time"
)

type PressureReading struct {
	Sensor    string    `json:"sensor"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type AirspeedReading struct {
	Sensor         string    `json:"sensor"`
	IndicatedSpeed float64   `json:"indicated_speed"`
	Unit           string    `json:"unit"`
	PitotPressure  float64   `json:"pitot_pressure"`
	StaticPressure float64   `json:"static_pressure"`
	Timestamp      time.Time `json:"timestamp"`
}

type AirspeedIndicator struct {
	pitotPressure  float64
	staticPressure float64
	format         string
	updateRate     time.Duration
}

func NewASI(hz float64, format string) *AirspeedIndicator {
	return &AirspeedIndicator{
		pitotPressure:  29.92, // Initialize to sea level
		staticPressure: 29.92,
		format:         format,
		updateRate:     time.Duration(float64(time.Second) / hz),
	}
}

// Calculate indicated airspeed from pitot and static pressure
// IAS is based on: V = sqrt(2 * ΔP / ρ₀)
// Where ΔP = P_pitot - P_static (dynamic pressure)
// ρ₀ is standard sea level air density
func (asi *AirspeedIndicator) calculateIndicatedAirspeed() float64 {
	// Dynamic pressure in inHg
	dynamicPressure := asi.pitotPressure - asi.staticPressure

	// Ensure we don't have negative dynamic pressure
	if dynamicPressure < 0 {
		dynamicPressure = 0
	}

	// Convert dynamic pressure to airspeed
	// This formula converts inHg dynamic pressure to knots (IAS)
	// V (knots) ≈ sqrt(ΔP * 77620)
	// The constant comes from: sqrt(2 * ΔP * 703.07 / ρ₀)
	// where ρ₀ = 0.002377 slugs/ft³ at sea level
	speedKnots := math.Sqrt(dynamicPressure * 77620)

	// Convert knots to mph for consistency with other sensors
	speedKnots := math.Sqrt(dynamicPressure * 77620)

	return speedKnots
}

func (asi *AirspeedIndicator) processReading(reading PressureReading) {
	switch reading.Sensor {
	case "pitot":
		asi.pitotPressure = reading.Value
	case "static":
		asi.staticPressure = reading.Value
	}
}

func (asi *AirspeedIndicator) getReading() AirspeedReading {
	speed := asi.calculateIndicatedAirspeed()

	return AirspeedReading{
		Sensor:         "airspeed",
		IndicatedSpeed: math.Round(speed*10) / 10,
		Unit:           "kts",
		PitotPressure:  asi.pitotPressure,
		StaticPressure: asi.staticPressure,
		Timestamp:      time.Now(),
	}
}

func (asi *AirspeedIndicator) outputReading(reading AirspeedReading) {
	switch asi.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.1f,%s,%.3f,%.3f,%s\n",
			reading.Sensor,
			reading.IndicatedSpeed,
			reading.Unit,
			reading.PitotPressure,
			reading.StaticPressure,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("%.1f kts (pitot: %.3f, static: %.3f)\n",
			reading.IndicatedSpeed,
			reading.PitotPressure,
			reading.StaticPressure)
	}
}

func readSensor(cmd *exec.Cmd) chan PressureReading {
	readings := make(chan PressureReading, 100)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pipe: %v\n", err)
		close(readings)
		return readings
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting sensor: %v\n", err)
		close(readings)
		return readings
	}

	go func() {
		defer close(readings)
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			var reading PressureReading
			if err := json.Unmarshal(scanner.Bytes(), &reading); err != nil {
				continue // Skip malformed lines (like CSV headers)
			}
			readings <- reading
		}
	}()

	return readings
}

func main() {
	hz := flag.Float64("hz", 10.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	pitotPath := flag.String("pitot", "./pitot", "Path to pitot sensor executable")
	staticPath := flag.String("static", "./static", "Path to static sensor executable")
	flag.Parse()

	// Start pitot and static sensors
	pitotCmd := exec.Command(*pitotPath, "-format", "json", "-hz", fmt.Sprintf("%.1f", *hz))
	staticCmd := exec.Command(*staticPath, "-format", "json", "-hz", fmt.Sprintf("%.1f", *hz))

	pitotReadings := readSensor(pitotCmd)
	staticReadings := readSensor(staticCmd)

	asi := NewASI(*hz, *format)

	if *format == "csv" {
		fmt.Println("sensor,indicated_speed,unit,pitot_pressure,static_pressure,timestamp")
	}

	ticker := time.NewTicker(asi.updateRate)
	defer ticker.Stop()

	for {
		select {
		case reading, ok := <-pitotReadings:
			if !ok {
				fmt.Fprintln(os.Stderr, "Pitot sensor stopped")
				return
			}
			asi.processReading(reading)

		case reading, ok := <-staticReadings:
			if !ok {
				fmt.Fprintln(os.Stderr, "Static sensor stopped")
				return
			}
			asi.processReading(reading)

		case <-ticker.C:
			reading := asi.getReading()
			asi.outputReading(reading)
		}
	}
}
