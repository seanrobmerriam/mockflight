//go:build legacycli
// +build legacycli

package main

import (
	"encoding/json"
	"fmt"

	"github.com/mockflight/mockflight/internal/legacycli"
	"github.com/mockflight/mockflight/internal/sim"
)

type GyroscopeReading struct {
	Sensor    string  `json:"sensor"`
	RollRate  float64 `json:"roll_rate"`
	PitchRate float64 `json:"pitch_rate"`
	YawRate   float64 `json:"yaw_rate"`
	Unit      string  `json:"unit"`
	Timestamp string  `json:"timestamp"`
}

func main() {
	opts := legacycli.ParseOptions(nil)
	if opts.Format == "csv" {
		fmt.Println("sensor,roll_rate,pitch_rate,yaw_rate,unit,timestamp")
	}

	legacycli.Loop(opts, func(snapshot sim.Snapshot) {
		reading := GyroscopeReading{
			Sensor:    "gyroscope",
			RollRate:  snapshot.Gyroscope.X,
			PitchRate: snapshot.Gyroscope.Y,
			YawRate:   snapshot.Gyroscope.Z,
			Unit:      snapshot.Gyroscope.Unit,
			Timestamp: snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		switch opts.Format {
		case "json":
			data, _ := json.Marshal(reading)
			fmt.Println(string(data))
		case "csv":
			fmt.Printf("%s,%.2f,%.2f,%.2f,%s,%s\n", reading.Sensor, reading.RollRate, reading.PitchRate, reading.YawRate, reading.Unit, reading.Timestamp)
		default:
			fmt.Printf("Roll:%.2f Pitch:%.2f Yaw:%.2f %s\n", reading.RollRate, reading.PitchRate, reading.YawRate, reading.Unit)
		}
	})
}
