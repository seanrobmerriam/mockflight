package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/mockflight/mockflight/internal/dashboard"
	"github.com/mockflight/mockflight/internal/sim"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	noise := flag.Float64("noise", 1.0, "Noise scaling for simulated sensors")
	pitotBlocked := flag.Bool("pitot-blocked", false, "Freeze pitot pressure at its current reading")
	pitotDrainBlocked := flag.Bool("pitot-drain-blocked", false, "Bleed off dynamic pressure from the pitot system")
	staticBlocked := flag.Bool("static-blocked", false, "Freeze static pressure at its current reading")
	staticLeak := flag.Bool("static-leak", false, "Bias static pressure toward a leaking reference pressure")
	magDisturbance := flag.Bool("mag-disturbance", false, "Inject magnetic disturbance into the magnetometer")
	gyroSaturation := flag.Bool("gyro-saturation", false, "Clamp gyro rates at the sensor saturation threshold")
	flag.Parse()

	simulator := sim.NewSimulator(sim.Config{
		Seed:       time.Now().UnixNano(),
		NoiseScale: *noise,
		Failures: sim.FailureConfig{
			PitotBlocked:          *pitotBlocked,
			PitotDrainBlocked:     *pitotDrainBlocked,
			StaticPortBlocked:     *staticBlocked,
			StaticLeak:            *staticLeak,
			MagnetometerDisturbed: *magDisturbance,
			GyroSaturation:        *gyroSaturation,
		},
	})

	go func() {
		ticker := time.NewTicker(simulator.TickInterval())
		defer ticker.Stop()

		for range ticker.C {
			simulator.Advance(simulator.TickInterval())
		}
	}()

	server := &http.Server{
		Addr:              *addr,
		Handler:           dashboard.NewHandler(simulator),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("mockflight dashboard listening on http://localhost%s", *addr)
	log.Fatal(server.ListenAndServe())
}
