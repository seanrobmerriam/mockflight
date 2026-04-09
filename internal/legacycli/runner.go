package legacycli

import (
	"flag"
	"time"

	"github.com/mockflight/mockflight/internal/sim"
)

type Options struct {
	Format    string
	Interval  time.Duration
	Simulator *sim.Simulator
}

func ParseOptions(registerAdditional func()) Options {
	hz := flag.Float64("hz", 10.0, "Update frequency in Hz")
	format := flag.String("format", "json", "Output format: json, csv, or simple")
	noise := flag.Float64("noise", 1.0, "Noise scaling for simulated sensors")
	seed := flag.Int64("seed", 0, "Random seed for deterministic output (0 uses current time)")
	pitotBlocked := flag.Bool("pitot-blocked", false, "Freeze pitot pressure at its current reading")
	pitotDrainBlocked := flag.Bool("pitot-drain-blocked", false, "Bleed off dynamic pressure from the pitot system")
	staticBlocked := flag.Bool("static-blocked", false, "Freeze static pressure at its current reading")
	staticLeak := flag.Bool("static-leak", false, "Bias static pressure toward a leaking reference pressure")
	magDisturbed := flag.Bool("mag-disturbance", false, "Inject magnetic disturbance into the magnetometer")
	gyroSaturation := flag.Bool("gyro-saturation", false, "Clamp gyro rates at the sensor saturation threshold")

	if registerAdditional != nil {
		registerAdditional()
	}
	flag.Parse()

	return Options{
		Format:   *format,
		Interval: intervalFromHz(*hz),
		Simulator: sim.NewSimulator(sim.Config{
			Seed:       *seed,
			NoiseScale: *noise,
			Failures: sim.FailureConfig{
				PitotBlocked:          *pitotBlocked,
				PitotDrainBlocked:     *pitotDrainBlocked,
				StaticPortBlocked:     *staticBlocked,
				StaticLeak:            *staticLeak,
				MagnetometerDisturbed: *magDisturbed,
				GyroSaturation:        *gyroSaturation,
			},
		}),
	}
}

func Loop(opts Options, emit func(snapshot sim.Snapshot)) {
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for range ticker.C {
		opts.Simulator.Advance(opts.Interval)
		emit(opts.Simulator.Snapshot())
	}
}

func intervalFromHz(hz float64) time.Duration {
	if hz <= 0 {
		return time.Second
	}
	return time.Duration(float64(time.Second) / hz)
}
