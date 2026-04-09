//go:build legacycli
// +build legacycli

package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/mockflight/mockflight/internal/legacycli"
	"github.com/mockflight/mockflight/internal/sim"
)

type AirspeedReading struct {
	Sensor         string  `json:"sensor"`
	IndicatedSpeed float64 `json:"indicated_speed"`
	Unit           string  `json:"unit"`
	PitotPressure  float64 `json:"pitot_pressure"`
	StaticPressure float64 `json:"static_pressure"`
	Timestamp      string  `json:"timestamp"`
}

func main() {
	opts := legacycli.ParseOptions(func() {
		flag.String("pitot", "./pitot", "Deprecated: ignored, uses shared simulator")
		flag.String("static", "./static", "Deprecated: ignored, uses shared simulator")
	})
	if opts.Format == "csv" {
		fmt.Println("sensor,indicated_speed,unit,pitot_pressure,static_pressure,timestamp")
	}

	legacycli.Loop(opts, func(snapshot sim.Snapshot) {
		reading := AirspeedReading{
			Sensor:         "airspeed",
			IndicatedSpeed: snapshot.Airspeed.Indicated,
			Unit:           snapshot.Airspeed.Unit,
			PitotPressure:  snapshot.Airspeed.PitotPressure,
			StaticPressure: snapshot.Airspeed.StaticPressure,
			Timestamp:      snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		switch opts.Format {
		case "json":
			data, _ := json.Marshal(reading)
			fmt.Println(string(data))
		case "csv":
			fmt.Printf("%s,%.1f,%s,%.3f,%.3f,%s\n", reading.Sensor, reading.IndicatedSpeed, reading.Unit, reading.PitotPressure, reading.StaticPressure, reading.Timestamp)
		default:
			fmt.Printf("%.1f %s (pitot: %.3f, static: %.3f)\n", reading.IndicatedSpeed, reading.Unit, reading.PitotPressure, reading.StaticPressure)
		}
	})
}
