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

type AHRSReading struct {
	Sensor    string  `json:"sensor"`
	Roll      float64 `json:"roll"`
	Pitch     float64 `json:"pitch"`
	Yaw       float64 `json:"yaw"`
	Heading   float64 `json:"heading"`
	Timestamp string  `json:"timestamp"`
}

func main() {
	opts := legacycli.ParseOptions(func() {
		flag.String("gyro", "./gyroscope", "Deprecated: ignored, uses shared simulator")
		flag.String("accel", "./accelerometer", "Deprecated: ignored, uses shared simulator")
		flag.String("mag", "./magnetometer", "Deprecated: ignored, uses shared simulator")
		flag.Float64("gyro-weight", 0.98, "Deprecated: ignored, uses shared simulator AHRS")
	})
	if opts.Format == "csv" {
		fmt.Println("sensor,roll,pitch,yaw,heading,timestamp")
	}

	legacycli.Loop(opts, func(snapshot sim.Snapshot) {
		reading := AHRSReading{
			Sensor:    "ahrs",
			Roll:      snapshot.AHRS.Roll,
			Pitch:     snapshot.AHRS.Pitch,
			Yaw:       snapshot.AHRS.Yaw,
			Heading:   snapshot.AHRS.Heading,
			Timestamp: snapshot.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		switch opts.Format {
		case "json":
			data, _ := json.Marshal(reading)
			fmt.Println(string(data))
		case "csv":
			fmt.Printf("%s,%.1f,%.1f,%.1f,%.1f,%s\n", reading.Sensor, reading.Roll, reading.Pitch, reading.Yaw, reading.Heading, reading.Timestamp)
		default:
			fmt.Printf("Roll:%.1f Pitch:%.1f Heading:%.1f\n", reading.Roll, reading.Pitch, reading.Heading)
		}
	})
}
