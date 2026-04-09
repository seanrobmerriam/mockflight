//go:build legacycli
// +build legacycli

package main

import (
	"encoding/json"
	"fmt"

	"github.com/mockflight/mockflight/internal/legacycli"
	"github.com/mockflight/mockflight/internal/sim"
)

type AccelerometerReading struct {
	Sensor    string  `json:"sensor"`
	AccelX    float64 `json:"accel_x"`
	AccelY    float64 `json:"accel_y"`
	AccelZ    float64 `json:"accel_z"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

func main() {
	opts := legacycli.ParseOptions(nil)
	if opts.Format == "csv" {
		fmt.Println("sensor,accel_x,accel_y,accel_z,unit,timestamp")
	}

	legacycli.Loop(opts, func(snapshot sim.Snapshot) {
		reading := AccelerometerReading{
			Sensor:    "accelerometer",
			AccelX:    snapshot.Accelerometer.X,
			AccelY:    snapshot.Accelerometer.Y,
			AccelZ:    snapshot.Accelerometer.Z,
			Unit:      snapshot.Accelerometer.Unit,
			Timestamp: snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		switch opts.Format {
		case "json":
			data, _ := json.Marshal(reading)
			fmt.Println(string(data))
		case "csv":
			fmt.Printf("%s,%.2f,%.2f,%.2f,%s,%s\n", reading.Sensor, reading.AccelX, reading.AccelY, reading.AccelZ, reading.Unit, reading.Timestamp)
		default:
			fmt.Printf("X:%.2f Y:%.2f Z:%.2f %s\n", reading.AccelX, reading.AccelY, reading.AccelZ, reading.Unit)
		}
	})
}
