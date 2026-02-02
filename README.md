#  mockflight Mock Aviation Sensors

Realistic mock sensors for aircraft instrumentation development and testing.

## Sensors

### Speedometer
Simulates aircraft speed through a complete flight cycle: taxiing, takeoff, climb, cruise, descent, and landing.

**Features:**
- Realistic acceleration/deceleration physics
- Flight phase transitions
- Configurable sensor noise
- Multiple output formats

### Altimeter
Simulates aircraft altitude through a complete flight cycle with correlated flight phases.

**Features:**
- Realistic climb/descent rates (typical small aircraft performance)
- Random field elevation
- Barometric noise simulation
- Multiple output formats

## Building

```bash
# Build speedometer
go build -o speedometer speedometer.go

# Build altimeter
go build -o altimeter altimeter.go
```

## Usage

### Basic Usage

```bash
# Run speedometer with default settings (10Hz, JSON output)
./speedometer

# Run altimeter with default settings
./altimeter
```

### Command-line Options

Both sensors support the following flags:

- `-hz <float>`: Update frequency in Hz (default: 10.0)
- `-format <string>`: Output format - json, csv, or simple (default: json)
- `-noise <float>`: Noise level (default varies by sensor)

### Examples

```bash
# High-frequency updates (50Hz) in simple format
./speedometer -hz 50 -format simple

# CSV output for logging
./altimeter -format csv > altitude_log.csv

# Reduce noise for cleaner data
./speedometer -noise 0.005

# Pipe to your application
./speedometer | your-app

# Run multiple sensors simultaneously
./speedometer > speed.log &
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

### Speedometer
- Max acceleration: 20 mph/s
- Max deceleration: 30 mph/s
- Speed range: 0-150 mph (typical for small aircraft)
- Noise: 1% of current speed (configurable)

### Altimeter
- Climb rate: ~1000 fpm (16.67 ft/s)
- Descent rate: ~800 fpm (13.33 ft/s)
- Cruise altitude: 5000 ft AGL
- Field elevation: Random 100-500 ft MSL
- Noise: ±1 foot (configurable)

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

MIT
