//go:build legacycli
// +build legacycli

package main

import (
	"encoding/json"
	"fmt"

	"github.com/mockflight/mockflight/internal/legacycli"
	"github.com/mockflight/mockflight/internal/sim"
)

type MagnetometerReading struct {
	Sensor    string  `json:"sensor"`
	MagX      float64 `json:"mag_x"`
	MagY      float64 `json:"mag_y"`
	MagZ      float64 `json:"mag_z"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

func main() {
	opts := legacycli.ParseOptions(nil)
	if opts.Format == "csv" {
		fmt.Println("sensor,mag_x,mag_y,mag_z,unit,timestamp")
	}

	legacycli.Loop(opts, func(snapshot sim.Snapshot) {
		reading := MagnetometerReading{
			Sensor:    "magnetometer",
			MagX:      snapshot.Magnetometer.X,
			MagY:      snapshot.Magnetometer.Y,
			MagZ:      snapshot.Magnetometer.Z,
			Unit:      snapshot.Magnetometer.Unit,
			Timestamp: snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		switch opts.Format {
		case "json":
			data, _ := json.Marshal(reading)
			fmt.Println(string(data))
		case "csv":
			fmt.Printf("%s,%.2f,%.2f,%.2f,%s,%s\n", reading.Sensor, reading.MagX, reading.MagY, reading.MagZ, reading.Unit, reading.Timestamp)
		default:
			fmt.Printf("X:%.2f Y:%.2f Z:%.2f %s\n", reading.MagX, reading.MagY, reading.MagZ, reading.Unit)
		}
	})
}
