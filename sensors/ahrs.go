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

type GyroReading struct {
	Sensor    string    `json:"sensor"`
	RollRate  float64   `json:"roll_rate"`
	PitchRate float64   `json:"pitch_rate"`
	YawRate   float64   `json:"yaw_rate"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type AccelReading struct {
	Sensor    string    `json:"sensor"`
	AccelX    float64   `json:"accel_x"`
	AccelY    float64   `json:"accel_y"`
	AccelZ    float64   `json:"accel_z"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type MagReading struct {
	Sensor    string    `json:"sensor"`
	MagX      float64   `json:"mag_x"`
	MagY      float64   `json:"mag_y"`
	MagZ      float64   `json:"mag_z"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

type AHRSReading struct {
	Sensor    string    `json:"sensor"`
	Roll      float64   `json:"roll"`
	Pitch     float64   `json:"pitch"`
	Yaw       float64   `json:"yaw"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

type AHRS struct {
	// State
	roll    float64
	pitch   float64
	yaw     float64
	heading float64

	// Latest sensor readings
	gyroRoll  float64
	gyroPitch float64
	gyroYaw   float64
	accelX    float64
	accelY    float64
	accelZ    float64
	magX      float64
	magY      float64
	magZ      float64

	// Filter parameters
	gyroWeight  float64 // Complementary filter weight for gyro (vs accel/mag)
	updateRate  time.Duration
	lastUpdate  time.Time
	format      string
	initialized bool
}

func NewAHRS(hz float64, format string, gyroWeight float64) *AHRS {
	return &AHRS{
		roll:        0.0,
		pitch:       0.0,
		yaw:         0.0,
		heading:     0.0,
		updateRate:  time.Duration(float64(time.Second) / hz),
		gyroWeight:  gyroWeight,
		format:      format,
		lastUpdate:  time.Now(),
		initialized: false,
	}
}

func (a *AHRS) processGyro(reading GyroReading) {
	a.gyroRoll = reading.RollRate
	a.gyroPitch = reading.PitchRate
	a.gyroYaw = reading.YawRate
}

func (a *AHRS) processAccel(reading AccelReading) {
	a.accelX = reading.AccelX
	a.accelY = reading.AccelY
	a.accelZ = reading.AccelZ
}

func (a *AHRS) processMag(reading MagReading) {
	a.magX = reading.MagX
	a.magY = reading.MagY
	a.magZ = reading.MagZ
}

// Calculate roll and pitch from accelerometer
// This gives us gravity vector direction
func (a *AHRS) calculateAccelAttitude() (float64, float64) {
	// Roll: atan2(ay, az)
	roll := math.Atan2(a.accelY, -a.accelZ) * 180.0 / math.Pi

	// Pitch: atan2(ax, sqrt(ay² + az²))
	pitch := math.Atan2(a.accelX, math.Sqrt(a.accelY*a.accelY+a.accelZ*a.accelZ)) * 180.0 / math.Pi

	return roll, pitch
}

// Calculate heading from magnetometer
// This needs to be tilt-compensated using roll and pitch
func (a *AHRS) calculateMagHeading() float64 {
	// Convert current roll and pitch to radians
	rollRad := a.roll * math.Pi / 180.0
	pitchRad := a.pitch * math.Pi / 180.0

	// Tilt compensation
	// Rotate magnetometer readings to horizontal plane
	magXh := a.magX*math.Cos(pitchRad) + a.magZ*math.Sin(pitchRad)
	magYh := a.magX*math.Sin(rollRad)*math.Sin(pitchRad) + a.magY*math.Cos(rollRad) - a.magZ*math.Sin(rollRad)*math.Cos(pitchRad)

	// Calculate heading
	heading := math.Atan2(-magYh, magXh) * 180.0 / math.Pi

	// Normalize to 0-360
	if heading < 0 {
		heading += 360
	}

	return heading
}

// Complementary filter: combines gyro (short-term accuracy) with accel/mag (long-term accuracy)
// Gyro integrates angular velocity but drifts over time
// Accel/Mag provide absolute reference but are noisy
func (a *AHRS) updateAttitude() {
	now := time.Now()
	dt := now.Sub(a.lastUpdate).Seconds()
	a.lastUpdate = now

	if dt > 1.0 || dt <= 0 {
		// Skip unrealistic dt values
		return
	}

	// Get accelerometer-based attitude (noisy but absolute)
	accelRoll, accelPitch := a.calculateAccelAttitude()

	// Get magnetometer-based heading (noisy but absolute)
	magHeading := a.calculateMagHeading()

	if !a.initialized {
		// First update - initialize from accel/mag
		a.roll = accelRoll
		a.pitch = accelPitch
		a.heading = magHeading
		a.yaw = magHeading
		a.initialized = true
		return
	}

	// Integrate gyro rates (accurate short-term)
	gyroRoll := a.roll + a.gyroRoll*dt
	gyroPitch := a.pitch + a.gyroPitch*dt
	gyroYaw := a.yaw + a.gyroYaw*dt

	// Complementary filter: blend gyro integration with accel/mag
	// High gyroWeight (e.g., 0.98) means trust gyro more (smooth but drifts)
	// Low gyroWeight (e.g., 0.5) means trust accel/mag more (noisy but no drift)
	a.roll = a.gyroWeight*gyroRoll + (1.0-a.gyroWeight)*accelRoll
	a.pitch = a.gyroWeight*gyroPitch + (1.0-a.gyroWeight)*accelPitch
	a.yaw = a.gyroWeight*gyroYaw + (1.0-a.gyroWeight)*magHeading

	// Update heading (from yaw)
	a.heading = a.yaw

	// Normalize angles
	if a.heading < 0 {
		a.heading += 360
	} else if a.heading >= 360 {
		a.heading -= 360
	}

	if a.yaw < 0 {
		a.yaw += 360
	} else if a.yaw >= 360 {
		a.yaw -= 360
	}

	// Clamp roll and pitch to reasonable ranges
	if a.roll > 180 {
		a.roll -= 360
	} else if a.roll < -180 {
		a.roll += 360
	}

	if a.pitch > 180 {
		a.pitch -= 360
	} else if a.pitch < -180 {
		a.pitch += 360
	}
}

func (a *AHRS) getReading() AHRSReading {
	return AHRSReading{
		Sensor:    "ahrs",
		Roll:      math.Round(a.roll*10) / 10,
		Pitch:     math.Round(a.pitch*10) / 10,
		Yaw:       math.Round(a.yaw*10) / 10,
		Heading:   math.Round(a.heading*10) / 10,
		Timestamp: time.Now(),
	}
}

func (a *AHRS) outputReading(reading AHRSReading) {
	switch a.format {
	case "json":
		data, _ := json.Marshal(reading)
		fmt.Println(string(data))
	case "csv":
		fmt.Printf("%s,%.1f,%.1f,%.1f,%.1f,%s\n",
			reading.Sensor,
			reading.Roll,
			reading.Pitch,
			reading.Yaw,
			reading.Heading,
			reading.Timestamp.Format(time.RFC3339))
	case "simple":
		fmt.Printf("Roll:%.1f° Pitch:%.1f° Heading:%.1f°\n",
			reading.Roll,
			reading.Pitch,
			reading.Heading)
	}
}

func readSensor(cmd *exec.Cmd, readingType string) chan interface{} {
	readings := make(chan interface{}, 100)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pipe for %s: %v\n", readingType, err)
		close(readings)
		return readings
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting %s sensor: %v\n", readingType, err)
		close(readings)
		return readings
	}

	go func() {
		defer close(readings)
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()

			switch readingType {
			case "gyro":
				var reading GyroReading
				if err := json.Unmarshal([]byte(line), &reading); err == nil {
					readings <- reading
				}
			case "accel":
				var reading AccelReading
				if err := json.Unmarshal([]byte(line), &reading); err == nil {
					readings <- reading
				}
			case "mag":
				var reading MagReading
				if err := json.Unmarshal([]byte(line), &reading); err == nil {
					readings <- reading
				}
			}
		}
	}()

	return readings
}

func main() {
	hz := flag.Float64("hz", 50.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	gyroPath := flag.String("gyro", "./gyroscope", "Path to gyroscope sensor")
	accelPath := flag.String("accel", "./accelerometer", "Path to accelerometer sensor")
	magPath := flag.String("mag", "./magnetometer", "Path to magnetometer sensor")
	gyroWeight := flag.Float64("gyro-weight", 0.98, "Complementary filter gyro weight (0.9-0.99)")
	flag.Parse()

	// Start all three sensors
	gyroCmd := exec.Command(*gyroPath, "-format", "json", "-hz", fmt.Sprintf("%.1f", *hz))
	accelCmd := exec.Command(*accelPath, "-format", "json", "-hz", fmt.Sprintf("%.1f", *hz))
	magCmd := exec.Command(*magPath, "-format", "json", "-hz", fmt.Sprintf("%.1f", *hz))

	gyroReadings := readSensor(gyroCmd, "gyro")
	accelReadings := readSensor(accelCmd, "accel")
	magReadings := readSensor(magCmd, "mag")

	ahrs := NewAHRS(*hz, *format, *gyroWeight)

	if *format == "csv" {
		fmt.Println("sensor,roll,pitch,yaw,heading,timestamp")
	}

	ticker := time.NewTicker(ahrs.updateRate)
	defer ticker.Stop()

	for {
		select {
		case reading, ok := <-gyroReadings:
			if !ok {
				fmt.Fprintln(os.Stderr, "Gyroscope sensor stopped")
				return
			}
			ahrs.processGyro(reading.(GyroReading))

		case reading, ok := <-accelReadings:
			if !ok {
				fmt.Fprintln(os.Stderr, "Accelerometer sensor stopped")
				return
			}
			ahrs.processAccel(reading.(AccelReading))

		case reading, ok := <-magReadings:
			if !ok {
				fmt.Fprintln(os.Stderr, "Magnetometer sensor stopped")
				return
			}
			ahrs.processMag(reading.(MagReading))

		case <-ticker.C:
			ahrs.updateAttitude()
			reading := ahrs.getReading()
			ahrs.outputReading(reading)
		}
	}
}
