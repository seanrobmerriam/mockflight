package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/mockflight/mockflight/internal/sim"
)

type SnapshotProvider interface {
	Snapshot() sim.Snapshot
}

type FailureProvider interface {
	SetFailures(sim.FailureConfig)
	Failures() sim.FailureConfig
}

func NewHandler(provider SnapshotProvider) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pageTemplate.Execute(w, map[string]string{"Title": "MockFlight Sensor Deck"})
	})
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(provider.Snapshot())
	})
	mux.HandleFunc("/api/failures", func(w http.ResponseWriter, r *http.Request) {
		controller, ok := provider.(FailureProvider)
		if !ok {
			http.Error(w, "failure controls unavailable", http.StatusNotImplemented)
			return
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(controller.Failures())
		case http.MethodPost:
			var failures sim.FailureConfig
			if err := json.NewDecoder(r.Body).Decode(&failures); err != nil {
				http.Error(w, "invalid failure payload", http.StatusBadRequest)
				return
			}
			controller.SetFailures(failures)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(controller.Failures())
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}

var pageTemplate = template.Must(template.New("dashboard").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root {
      --paper: #f7f2e7;
      --paper-shadow: #eadfc6;
      --ink: #16212f;
      --muted: #5e6b78;
      --accent: #0c7c86;
      --accent-strong: #054d54;
      --warning: #b6662a;
      --panel: rgba(255, 252, 245, 0.88);
      --line: rgba(22, 33, 47, 0.12);
      --glass: rgba(12, 124, 134, 0.12);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: "Avenir Next", "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(circle at top left, rgba(12,124,134,0.2), transparent 26%),
        radial-gradient(circle at top right, rgba(182,102,42,0.15), transparent 24%),
        linear-gradient(180deg, #fbf7ef 0%, #efe5d1 100%);
    }
    .shell {
      max-width: 1280px;
      margin: 0 auto;
      padding: 32px 20px 56px;
    }
    .hero {
      display: grid;
      gap: 18px;
      grid-template-columns: 1.4fr 1fr;
      align-items: end;
      margin-bottom: 22px;
    }
    .title-block {
      background: linear-gradient(145deg, rgba(255,255,255,0.85), rgba(255,247,232,0.9));
      border: 1px solid rgba(255,255,255,0.65);
      border-radius: 26px;
      padding: 24px;
      box-shadow: 0 18px 48px rgba(75, 62, 45, 0.12);
      backdrop-filter: blur(6px);
    }
    .eyebrow {
      margin: 0 0 8px;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.18em;
      text-transform: uppercase;
      color: var(--warning);
    }
    h1 {
      margin: 0;
      font-family: "Iowan Old Style", "Palatino Linotype", serif;
      font-size: clamp(2.2rem, 5vw, 4.2rem);
      line-height: 0.95;
      letter-spacing: -0.04em;
    }
    .subtitle {
      margin: 12px 0 0;
      color: var(--muted);
      max-width: 56ch;
      line-height: 1.45;
    }
    .status {
      display: grid;
      gap: 14px;
      background: rgba(22, 33, 47, 0.92);
      color: white;
      border-radius: 26px;
      padding: 24px;
      box-shadow: 0 20px 44px rgba(22, 33, 47, 0.18);
    }
    .status-grid,
    .sensor-grid {
      display: grid;
      gap: 16px;
    }
    .status-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    .sensor-grid {
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    }
    .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 22px;
      padding: 18px;
      box-shadow: inset 0 1px 0 rgba(255,255,255,0.45), 0 16px 38px rgba(78, 66, 49, 0.08);
    }
    .status .panel {
      background: rgba(255,255,255,0.08);
      border-color: rgba(255,255,255,0.08);
      box-shadow: none;
    }
    .label {
      margin: 0;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.16em;
      text-transform: uppercase;
      color: inherit;
      opacity: 0.72;
    }
    .value {
      margin: 8px 0 0;
      font-size: clamp(1.4rem, 4vw, 2.6rem);
      font-weight: 700;
    }
    .subvalue {
      margin: 8px 0 0;
      color: var(--muted);
      line-height: 1.45;
      font-size: 0.95rem;
    }
    .status .subvalue { color: rgba(255,255,255,0.72); }
    .failure-list {
      margin: 0;
      padding: 0;
      list-style: none;
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }
    .failure-list li {
      padding: 8px 10px;
      border-radius: 999px;
      background: rgba(182, 102, 42, 0.18);
      color: #fff3e8;
      border: 1px solid rgba(255,255,255,0.12);
      font-size: 0.86rem;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }
    .failure-controls {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 10px;
      margin-top: 14px;
    }
    .failure-controls label {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 12px;
      border-radius: 14px;
      background: rgba(255,255,255,0.08);
      color: rgba(255,255,255,0.92);
      font-size: 0.9rem;
    }
    .failure-controls input {
      accent-color: #f1bb7b;
    }
    .meter {
      position: relative;
      overflow: hidden;
      height: 10px;
      margin-top: 14px;
      border-radius: 999px;
      background: rgba(22, 33, 47, 0.08);
    }
    .status .meter { background: rgba(255,255,255,0.12); }
    .meter > span {
      display: block;
      height: 100%;
      width: 0;
      border-radius: inherit;
      background: linear-gradient(90deg, var(--accent), var(--warning));
      transition: width 260ms ease;
    }
    .vector {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 10px;
      margin-top: 12px;
    }
    .vector div {
      padding: 10px;
      border-radius: 16px;
      background: var(--glass);
    }
    .vector strong {
      display: block;
      font-size: 0.78rem;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--muted);
    }
    .vector span {
      display: block;
      margin-top: 6px;
      font-size: 1.1rem;
      font-weight: 700;
    }
    .footer-note {
      margin-top: 20px;
      color: var(--muted);
      font-size: 0.92rem;
      text-align: right;
    }
    @media (max-width: 860px) {
      .hero { grid-template-columns: 1fr; }
      .status-grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="hero">
      <div class="title-block">
        <p class="eyebrow">Local Flight Simulation Deck</p>
        <h1>MockFlight Sensor Deck</h1>
        <p class="subtitle">A single local dashboard for the full mock instrument stack: pitot-static, inertial sensors, and fused AHRS. Values update from one shared flight model so the instruments stay in sync.</p>
      </div>
      <section class="status">
        <div class="status-grid">
          <article class="panel">
            <p class="label">Flight Phase</p>
            <p class="value" id="phase">Ground</p>
            <p class="subvalue">Shared state for every sensor stream.</p>
          </article>
          <article class="panel">
            <p class="label">Timestamp</p>
            <p class="value" id="timestamp">--</p>
            <p class="subvalue">UTC snapshot time</p>
          </article>
          <article class="panel">
            <p class="label">AHRS Heading</p>
            <p class="value" id="heading">--</p>
            <p class="subvalue">Roll <span id="roll">--</span> | Pitch <span id="pitch">--</span></p>
          </article>
        </div>
        <article class="panel">
          <p class="label">Active Failures</p>
          <ul class="failure-list" id="failures">
            <li>Nominal</li>
          </ul>
          <p class="subvalue">Blocked pitot/static ports and magnetic disturbance flags are injected by the shared simulator.</p>
          <div class="failure-controls">
            <label><input type="checkbox" id="pitotBlocked"> Pitot Blocked</label>
            <label><input type="checkbox" id="pitotDrainBlocked"> Pitot Drain Blocked</label>
            <label><input type="checkbox" id="staticBlocked"> Static Port Blocked</label>
            <label><input type="checkbox" id="staticLeak"> Static Leak</label>
            <label><input type="checkbox" id="magDisturbed"> Magnetometer Disturbed</label>
            <label><input type="checkbox" id="gyroSaturation"> Gyro Saturation</label>
          </div>
        </article>
      </section>
    </section>

    <section class="sensor-grid">
      <article class="panel">
        <p class="label">Altitude</p>
        <p class="value" id="altitude">--</p>
        <p class="subvalue">Vertical speed <span id="verticalSpeed">--</span></p>
        <div class="meter"><span id="altitudeMeter"></span></div>
      </article>
      <article class="panel">
        <p class="label">Airspeed</p>
        <p class="value" id="airspeed">--</p>
        <p class="subvalue">Pitot <span id="pitot">--</span> | Static <span id="static">--</span></p>
        <div class="meter"><span id="airspeedMeter"></span></div>
      </article>
      <article class="panel">
        <p class="label">Accelerometer</p>
        <div class="vector">
          <div><strong>X</strong><span id="accelX">--</span></div>
          <div><strong>Y</strong><span id="accelY">--</span></div>
          <div><strong>Z</strong><span id="accelZ">--</span></div>
        </div>
      </article>
      <article class="panel">
        <p class="label">Gyroscope</p>
        <div class="vector">
          <div><strong>Roll</strong><span id="gyroX">--</span></div>
          <div><strong>Pitch</strong><span id="gyroY">--</span></div>
          <div><strong>Yaw</strong><span id="gyroZ">--</span></div>
        </div>
      </article>
      <article class="panel">
        <p class="label">Magnetometer</p>
        <div class="vector">
          <div><strong>X</strong><span id="magX">--</span></div>
          <div><strong>Y</strong><span id="magY">--</span></div>
          <div><strong>Z</strong><span id="magZ">--</span></div>
        </div>
      </article>
      <article class="panel">
        <p class="label">AHRS</p>
        <div class="vector">
          <div><strong>Roll</strong><span id="ahrsRoll">--</span></div>
          <div><strong>Pitch</strong><span id="ahrsPitch">--</span></div>
          <div><strong>Yaw</strong><span id="ahrsYaw">--</span></div>
        </div>
      </article>
    </section>

    <p class="footer-note">Polling /api/snapshot every 750ms from a single in-process simulator.</p>
  </main>

  <script>
    const text = (id, value) => { document.getElementById(id).textContent = value; };
    const meter = (id, percent) => { document.getElementById(id).style.width = Math.max(0, Math.min(100, percent)) + '%'; };
    const fmt = (value, unit, digits = 1) => Number(value).toFixed(digits) + ' ' + unit;
    const renderFailures = (items) => {
      const target = document.getElementById('failures');
      target.innerHTML = '';
      const failures = items && items.length ? items : ['nominal'];
      failures.forEach((failure) => {
        const node = document.createElement('li');
        node.textContent = failure.replaceAll('_', ' ');
        target.appendChild(node);
      });
    };
    const syncFailureControls = (items) => {
      const active = new Set(items || []);
      document.getElementById('pitotBlocked').checked = active.has('pitot_blocked');
      document.getElementById('pitotDrainBlocked').checked = active.has('pitot_drain_blocked');
      document.getElementById('staticBlocked').checked = active.has('static_port_blocked');
      document.getElementById('staticLeak').checked = active.has('static_leak');
      document.getElementById('magDisturbed').checked = active.has('magnetometer_disturbed');
      document.getElementById('gyroSaturation').checked = active.has('gyro_saturation');
    };
    const failurePayload = () => ({
      pitot_blocked: document.getElementById('pitotBlocked').checked,
      pitot_drain_blocked: document.getElementById('pitotDrainBlocked').checked,
      static_port_blocked: document.getElementById('staticBlocked').checked,
      static_leak: document.getElementById('staticLeak').checked,
      magnetometer_disturbed: document.getElementById('magDisturbed').checked,
      gyro_saturation: document.getElementById('gyroSaturation').checked,
    });
    let savingFailures = false;
    async function saveFailures() {
      if (savingFailures) {
        return;
      }
      savingFailures = true;
      try {
        await fetch('/api/failures', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(failurePayload()),
        });
      } finally {
        savingFailures = false;
      }
    }

    async function refresh() {
      const response = await fetch('/api/snapshot', { cache: 'no-store' });
      const data = await response.json();

      text('phase', data.phase.replace('_', ' '));
      text('timestamp', new Date(data.timestamp).toISOString().slice(11, 19));
      text('heading', fmt(data.ahrs.heading, 'deg'));
      text('roll', fmt(data.ahrs.roll, 'deg'));
      text('pitch', fmt(data.ahrs.pitch, 'deg'));
      renderFailures(data.active_failures);
      if (!savingFailures) {
        syncFailureControls(data.active_failures);
      }

      text('altitude', fmt(data.altimeter.value, data.altimeter.unit, 0));
      text('verticalSpeed', fmt(data.vertical_speed.value, data.vertical_speed.unit, 2));
      meter('altitudeMeter', (data.altimeter.value / 5500) * 100);

      text('airspeed', fmt(data.airspeed.indicated, data.airspeed.unit));
      text('pitot', fmt(data.pitot.value, data.pitot.unit, 3));
      text('static', fmt(data.static_air.value, data.static_air.unit, 3));
      meter('airspeedMeter', (data.airspeed.indicated / 160) * 100);

      text('accelX', fmt(data.accelerometer.x, data.accelerometer.unit, 2));
      text('accelY', fmt(data.accelerometer.y, data.accelerometer.unit, 2));
      text('accelZ', fmt(data.accelerometer.z, data.accelerometer.unit, 2));

      text('gyroX', fmt(data.gyroscope.x, data.gyroscope.unit, 2));
      text('gyroY', fmt(data.gyroscope.y, data.gyroscope.unit, 2));
      text('gyroZ', fmt(data.gyroscope.z, data.gyroscope.unit, 2));

      text('magX', fmt(data.magnetometer.x, data.magnetometer.unit, 2));
      text('magY', fmt(data.magnetometer.y, data.magnetometer.unit, 2));
      text('magZ', fmt(data.magnetometer.z, data.magnetometer.unit, 2));

      text('ahrsRoll', fmt(data.ahrs.roll, data.ahrs.unit, 1));
      text('ahrsPitch', fmt(data.ahrs.pitch, data.ahrs.unit, 1));
      text('ahrsYaw', fmt(data.ahrs.yaw, data.ahrs.unit, 1));
    }

    refresh();
    setInterval(refresh, 750);
    document.querySelectorAll('.failure-controls input').forEach((node) => {
      node.addEventListener('change', saveFailures);
    });
  </script>
</body>
</html>`))
