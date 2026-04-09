//go:build legacycli
// +build legacycli

package main

import (
	"encoding/json"
	"fmt"

	"github.com/mockflight/mockflight/internal/legacycli"
	"github.com/mockflight/mockflight/internal/sim"
)

type StaticReading struct {
	Sensor    string  `json:"sensor"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

func main() {
	opts := legacycli.ParseOptions(nil)
	if opts.Format == "csv" {
		fmt.Println("sensor,value,unit,timestamp")
	}

	legacycli.Loop(opts, func(snapshot sim.Snapshot) {
		reading := StaticReading{
			Sensor:    "static",
			Value:     snapshot.StaticAir.Value,
			Unit:      snapshot.StaticAir.Unit,
			Timestamp: snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		switch opts.Format {
		case "json":
			data, _ := json.Marshal(reading)
			fmt.Println(string(data))
		case "csv":
			fmt.Printf("%s,%.3f,%s,%s\n", reading.Sensor, reading.Value, reading.Unit, reading.Timestamp)
		default:
			fmt.Printf("%.3f %s\n", reading.Value, reading.Unit)
		}
	})
}
