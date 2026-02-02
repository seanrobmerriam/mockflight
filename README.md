#  mockflight Mock Aviation Sensors

Realistic mock sensors for aircraft instrumentation development and testing.

**Features:**
- Realistic acceleration/deceleration physics
- Flight phase transitions
- Configurable sensor noise
- Multiple output formats
- Realistic climb/descent rates (typical small aircraft performance)
- Random field elevation
- Barometric noise simulation
- Multiple output formats

## Building

```bash
# Build [sensor]
go build -o [sensor] [sensor].go

# e.g. Build altimeter
go build -o altimeter altimeter.go
```

## Usage

### Basic Usage

```bash

git clone https://github.com/seanrobmerriam/mockflight

cd mockflight/sensors/

# For individual sensors:

# Run [sensor] with default settings
./[sensor]

# e.g. Run altimeter with default settings
./altimeter


### Command-line Options

# sensors support the following flags:

- `-hz <float>`: Update frequency in Hz (default: 10.0)
- `-format <string>`: Output format - json, csv, or simple (default: json)
- `-noise <float>`: Noise level (default varies by sensor)

For AHRS:

# Build everything
go build -o gyroscope gyroscope.go
go build -o accelerometer accelerometer.go  
go build -o magnetometer magnetometer.go
go build -o ahrs ahrs.go

# Run the AHRS (handles all sensors)
./ahrs -format simple

# Output:
# Roll:-18.2° Pitch:3.1° Heading:270.5°

```

### Examples

```bash
# High-frequency updates (50Hz) in simple format
./altimeter -hz 50 -format simple

# CSV output for logging
./altimeter -format csv > altitude_log.csv

# Reduce noise for cleaner data
./altimeter -noise 0.005

# Pipe to your application
./altimeter | your-app

# Run multiple sensors simultaneously
./altimeter > altitude.log &
```

## Output Formats

### JSON (default)
```json
{"sensor":"speedometer","value":145.3,"unit":"mph","timestamp":"2026-02-01T10:30:45Z"}
{"sensor":"altimeter","value":4523,"unit":"feet","timestamp":"2026-02-01T10:30:45Z"}
```

### CSV
```csv
sensor,value,unit,timestamp
speedometer,145.3,mph,2026-02-01T10:30:45Z
altimeter,4523,feet,2026-02-01T10:30:45Z
```

### Simple
```
145.3 mph
4523 ft
```

## Flight Profile

Both sensors follow a synchronized flight profile:

1. **Ground/Taxiing** (20s): Minimal speed, ground altitude
2. **Takeoff** (15s): Acceleration to rotation speed, initial climb
3. **Climb** (45s): Steady climb to cruise altitude and speed
4. **Cruise** (90s): Stable flight with minor variations
5. **Descent** (40s): Gradual descent and deceleration
6. **Landing** (20s): Final approach and touchdown
7. **Stopped/On Ground** (10s): Brief pause before restarting cycle

Total cycle time: ~240 seconds (4 minutes)

## Realistic Characteristics

### Altimeter
- Climb rate: ~1000 fpm (16.67 ft/s)
- Descent rate: ~800 fpm (13.33 ft/s)
- Cruise altitude: 5000 ft AGL
- Field elevation: Random 100-500 ft MSL
- Noise: ±1 foot (configurable)

# IMU and AHRS System

A realistic simulation of an Inertial Measurement Unit (IMU) and Attitude and Heading Reference System (AHRS) with separate raw sensors and sensor fusion.

## System Overview

In real aircraft, attitude determination uses multiple sensors working together:

```
┌─────────────┐
│ Gyroscope   │──┐
│ (3-axis)    │  │
└─────────────┘  │
                 │    ┌──────────┐      ┌─────────────┐
┌─────────────┐  ├───►│  AHRS    │─────►│   Attitude  │
│Accelerometer│──┤    │  Fusion  │      │ Roll, Pitch │
│ (3-axis)    │  │    └──────────┘      │   Heading   │
└─────────────┘  │                      └─────────────┘
                 │
┌─────────────┐  │
│Magnetometer │──┘
│ (3-axis)    │
└─────────────┘
```

## Components

### 1. Gyroscope (gyroscope.go)
Measures angular velocity (rotation rates) around three axes.

**What it measures:**
- Roll rate (rotation around longitudinal axis)
- Pitch rate (rotation around lateral axis)  
- Yaw rate (rotation around vertical axis)

**Output:** Degrees per second (°/s)

**Characteristics:**
- Very accurate short-term
- Drifts over time (bias error)
- No absolute reference

**Example reading:**
```json
{
  "sensor": "gyroscope",
  "roll_rate": -6.24,
  "pitch_rate": 0.31,
  "yaw_rate": -3.02,
  "unit": "deg/s"
}
```

When turning left at standard rate:
- Roll rate: -6°/s (rolling left)
- Yaw rate: -3°/s (turning left)
- Pitch rate: ~0°/s (holding altitude)

### 2. Accelerometer (accelerometer.go)
Measures specific force (acceleration - gravity) in body frame.

**What it measures:**
- X-axis: Forward/backward acceleration + gravity component
- Y-axis: Left/right acceleration + gravity component
- Z-axis: Up/down acceleration + gravity component

**Output:** m/s²

**Characteristics:**
- Provides absolute attitude reference via gravity
- Noisy (vibration, turbulence)
- Can't distinguish between gravity and acceleration
- Cannot measure heading

**Example reading (level flight):**
```json
{
  "sensor": "accelerometer",
  "accel_x": -0.12,
  "accel_y": 0.05,
  "accel_z": -9.81,
  "unit": "m/s²"
}
```

When level: az ≈ -9.81 m/s² (gravity pulls down)
When pitched up 10°: ax ≈ -1.70, az ≈ -9.66
When rolled 20°: ay ≈ 3.36, az ≈ -9.22

### 3. Magnetometer (magnetometer.go)
Measures Earth's magnetic field in body frame.

**What it measures:**
- X, Y, Z components of magnetic field
- Used to determine heading (compass)

**Output:** μT (microtesla)

**Characteristics:**
- Provides absolute heading reference
- Affected by magnetic interference (hard/soft iron)
- Needs tilt compensation (uses roll/pitch)
- Subject to magnetic declination

**Example reading (heading north, level):**
```json
{
  "sensor": "magnetometer",
  "mag_x": 25.34,
  "mag_y": -0.52,
  "mag_z": 43.18,
  "unit": "μT"
}
```

Field strength: ~50 μT (typical at mid-latitudes)
Inclination: ~60° (points downward in Northern Hemisphere)

### 4. AHRS (ahrs.go)
Fuses all three sensors using a complementary filter.

**What it does:**
- Integrates gyro for smooth, responsive attitude
- Corrects gyro drift using accel/mag
- Outputs absolute roll, pitch, and heading

**Algorithm:** Complementary Filter
```
attitude = gyro_weight × gyro_integration + (1 - gyro_weight) × accel_mag_reference
```

**Output:**
```json
{
  "sensor": "ahrs",
  "roll": -18.2,
  "pitch": 3.1,
  "yaw": 270.5,
  "heading": 270.5
}
```

## Building

```bash
go build -o gyroscope gyroscope.go
go build -o accelerometer accelerometer.go
go build -o magnetometer magnetometer.go
go build -o ahrs ahrs.go
```

## Usage

### Option 1: Use AHRS (Recommended)
The AHRS automatically starts all three sensors and fuses them:

```bash
./ahrs
./ahrs -format simple
./ahrs -format csv > attitude_log.csv
```

### Option 2: Run Sensors Independently
```bash
# Run each sensor separately
./gyroscope -format simple
./accelerometer -format simple
./magnetometer -format simple
```

### Advanced Options

**Adjust filter characteristics:**
```bash
# More responsive (trusts gyro more, but drifts faster)
./ahrs -gyro-weight 0.99

# More stable (trusts accel/mag more, but noisier)
./ahrs -gyro-weight 0.90
```

**Custom sensor paths:**
```bash
./ahrs -gyro /path/to/gyroscope -accel /path/to/accelerometer -mag /path/to/magnetometer
```

**Higher update rate:**
```bash
./ahrs -hz 100
```

## How Sensor Fusion Works

### The Problem
Each sensor has limitations:
- **Gyro:** Accurate but drifts over time
- **Accelerometer:** Absolute reference but noisy and affected by motion
- **Magnetometer:** Absolute heading but affected by interference

### The Solution: Complementary Filter
Combines the best of each sensor:

1. **Short-term (0-1 second):** Trust gyro
   - Smooth, responsive attitude tracking
   - Not affected by turbulence/acceleration

2. **Long-term (>10 seconds):** Trust accel/mag
   - Correct gyro drift
   - Maintain absolute reference

### Filter Weight Parameter
The `gyro-weight` parameter (default: 0.98) controls the balance:

```
gyro-weight = 0.98 (typical):
- 98% gyro, 2% accel/mag per update
- Smooth, responsive
- Very slow drift correction

gyro-weight = 0.90:
- 90% gyro, 10% accel/mag per update
- Still smooth but noisier
- Faster drift correction

gyro-weight = 0.50:
- Equal trust in gyro and accel/mag
- Noisy but no drift
- Not recommended for aircraft
```

## Understanding the Outputs

### Gyroscope Interpretation
```
Roll Rate > 0: Rolling right
Roll Rate < 0: Rolling left
Pitch Rate > 0: Pitching up
Pitch Rate < 0: Pitching down
Yaw Rate > 0: Turning right
Yaw Rate < 0: Turning left
```

### Accelerometer Interpretation
When level and stationary:
- ax ≈ 0
- ay ≈ 0
- az ≈ -9.81 m/s² (gravity)

When pitched up 10°:
- ax ≈ -1.70 m/s² (gravity component forward)
- az ≈ -9.66 m/s² (gravity component down)

When rolled right 20°:
- ay ≈ 3.36 m/s² (gravity component right)
- az ≈ -9.22 m/s² (gravity component down)

### Magnetometer Interpretation
The magnetometer measures Earth's field (~50 μT) projected into body frame.

Heading North, level:
- Strong X component (north)
- Strong Z component (dip angle)
- Weak Y component

Heading East, level:
- Weak X component
- Strong Y component (east)
- Strong Z component (dip angle)

### AHRS Interpretation
- **Roll:** -180° to +180° (negative = left wing down)
- **Pitch:** -90° to +90° (positive = nose up)
- **Heading:** 0° to 360° (0° = North, 90° = East, 180° = South, 270° = West)
- **Yaw:** Same as heading but can wrap differently

## Realistic Flight Scenarios

### Takeoff Roll
```
Gyro: pitch_rate increases to ~2°/s
Accel: ax increases (acceleration), az stays ~-9.81
Mag: heading constant (runway heading)
AHRS: pitch increases to ~10°
```

### Level Turn (18° bank)
```
Gyro: roll_rate ~6°/s entering turn, yaw_rate ~3°/s in turn
Accel: ay increases (centripetal), |a| > 9.81 (load factor)
Mag: rotates through turn
AHRS: roll = 18°, heading changes 3°/s
```

### Landing Flare
```
Gyro: pitch_rate increases to ~2.5°/s
Accel: az becomes less negative (reducing descent)
Mag: heading constant (runway)
AHRS: pitch increases to ~5°
```

## Common Issues and Troubleshooting

### Issue: AHRS shows 0,0,0
**Cause:** Sensors not started or AHRS can't find them
**Solution:** Ensure gyroscope, accelerometer, magnetometer are built and in the same directory, or provide paths with flags

### Issue: Attitude drifts over time
**Cause:** Gyro-weight too high, not enough accel/mag correction
**Solution:** Decrease gyro-weight (try 0.95 or 0.90)

### Issue: Attitude is very noisy
**Cause:** Gyro-weight too low, too much accel/mag influence
**Solution:** Increase gyro-weight (try 0.98 or 0.99)

### Issue: Heading jumps around
**Cause:** Magnetometer interference or poor tilt compensation
**Solution:** This is realistic! Real magnetometers are noisy. Increase gyro-weight or add additional filtering.

### Issue: Roll/Pitch correct but heading wrong
**Cause:** Magnetic declination or local magnetic anomalies
**Solution:** This is realistic. Magnetometers need calibration in real systems.

## Integration Example

```go
package main

import (
    "bufio"
    "encoding/json"
    "os/exec"
)

type AHRSReading struct {
    Roll    float64 `json:"roll"`
    Pitch   float64 `json:"pitch"`
    Heading float64 `json:"heading"`
}

func main() {
    cmd := exec.Command("./ahrs", "-format", "json")
    stdout, _ := cmd.StdoutPipe()
    cmd.Start()
    
    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        var reading AHRSReading
        json.Unmarshal(scanner.Bytes(), &reading)
        
        fmt.Printf("Attitude: Roll=%.1f° Pitch=%.1f° Hdg=%.1f°\n",
            reading.Roll, reading.Pitch, reading.Heading)
    }
}
```

## Technical Details

### Coordinate Systems

**Body Frame (Aircraft):**
- X: Forward (nose)
- Y: Right wing
- Z: Down

**NED Frame (Navigation):**
- N: North
- E: East
- D: Down

### Rotation Order
Euler angles are applied in this order:
1. Yaw (heading)
2. Pitch
3. Roll

### Magnetic Declination
The difference between true north and magnetic north:
- Varies by location (e.g., 14°E in California)
- Changes over time
- Must be accounted for in navigation

### Magnetic Inclination (Dip Angle)
Earth's field points downward in Northern Hemisphere:
- Equator: ~0° (horizontal)
- Mid-latitudes: ~60° (downward)
- North Pole: ~90° (straight down)

## Advanced Topics

### Sensor Calibration
Real systems require calibration:
- **Gyro:** Measure bias when stationary
- **Accel:** 6-position calibration (each axis up/down)
- **Mag:** Hard/soft iron calibration (rotate through all orientations)

### Alternative Fusion Algorithms
This implementation uses a simple complementary filter. More advanced options:
- **Kalman Filter:** Optimal but more complex
- **Madgwick Filter:** Efficient quaternion-based
- **Mahony Filter:** Explicit complementary filter with PI controller

### Gimbal Lock
Euler angles suffer from gimbal lock at ±90° pitch. Professional systems use quaternions to avoid this.

## Performance Notes

- Default update rate: 50 Hz (adequate for most applications)
- Higher rates (100+ Hz) better for fast maneuvers
- Lower rates (10-20 Hz) acceptable for slow-moving platforms
- Gyro needs highest sample rate (>50 Hz)
- Magnetometer can be slower (10-20 Hz)

## Integration Examples

### Reading from Go
```go
cmd := exec.Command("./speedometer", "-format", "json")
stdout, _ := cmd.StdoutPipe()
cmd.Start()

scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
    var reading SpeedometerReading
    json.Unmarshal(scanner.Bytes(), &reading)
    fmt.Printf("Speed: %.1f %s\n", reading.Value, reading.Unit)
}
```

### Reading from Python
```python
import subprocess
import json

proc = subprocess.Popen(
    ['./speedometer', '-format', 'json'],
    stdout=subprocess.PIPE,
    text=True
)

for line in proc.stdout:
    reading = json.loads(line)
    print(f"Speed: {reading['value']} {reading['unit']}")
```

### Using with Unix pipes
```bash
# Combine multiple sensors
./speedometer -format simple &
./altimeter -format simple &
wait

# Log to file with timestamps
./speedometer | ts '%Y-%m-%d %H:%M:%S' >> flight_data.log
```

## Adding More Sensors

To create additional sensors (angle of attack, vertical speed, heading, etc.), follow this pattern:

1. Define realistic min/max values and rates of change
2. Create flight phases appropriate to the sensor
3. Add realistic noise characteristics
4. Implement smooth transitions between states
5. Use the same command-line interface for consistency

## License

0BSD
