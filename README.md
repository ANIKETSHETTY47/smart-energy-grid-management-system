# Smart Energy Grid — Golang Starter

A production-ready starter implementing a **smart energy grid management** backend in Go.

## What you get

- Go HTTP API (Fiber) with clean layered structure (`cmd/`, `internal/`)
- MQTT ingestion worker (Eclipse Mosquitto broker)
- PostgreSQL (Timescale-ready) + Redis cache
- JWT auth, role-based middleware
- Docker Compose for local dev
- Basic Grafana dashboard & Prometheus metrics
- Seed data and example device simulator

## Quick start

```bash
# 1) dev prerequisites
docker compose up -d

# 2) run API
make dev

# 3) run ingestor
make ingestor

# 4) simulate device data
make simulate
```




API will listen on `http://localhost:8080`.

## Endpoints

- `GET /health` — liveness
- `POST /auth/login` — get JWT (demo user: admin@example.com / admin123)
- `GET /facilities` — list (requires JWT)
- `GET /meters` — list (requires JWT)
- `POST /readings` — ingest single reading (requires JWT)
- `GET /analytics/summary` — daily aggregates (requires JWT)

## Project layout

```
README.md  .env.example  .github/workflows/deploy.yml  energy-grid-analytics/package.json  energy-grid-analytics/README.md  energy-grid-analytics/.gitignore  energy-grid-analytics/src/EnergyConverter.js  energy-grid-analytics/src/AnomalyDetector.js  energy-grid-analytics/src/MaintenancePredictor.js  energy-grid-analytics/src/LoadBalancer.js  energy-grid-analytics/src/DataAggregator.js  energy-grid-analytics/src/index.js  energy-grid-analytics/test/AnomalyDetector.test.js  energy-grid-analytics/test/EnergyConverter.test.js  energy-grid-dashboard/package.json  energy-grid-dashboard/public/index.html  energy-grid-dashboard/src/index.js  energy-grid-dashboard/src/App.jsx  energy-grid-dashboard/src/App.css  energy-grid-dashboard/src/components/Dashboard/Dashboard.jsx  energy-grid-dashboard/src/components/Alerts/AlertsList.jsx  energy-grid-dashboard/src/components/Equipment/EquipmentList.jsx  energy-grid-dashboard/src/services/api.js  lambda-functions/data-ingestion/package.json  lambda-functions/data-ingestion/index.js  lambda-functions/anomaly-detection/package.json  lambda-functions/anomaly-detection/index.js  lambda-functions/analytics-processing/package.json  lambda-functions/analytics-processing/index.js  lambda-functions/report-generation/package.json  lambda-functions/report-generation/index.js  lambda-functions/alert-management/package.json  lambda-functions/alert-management/index.js  lambda-functions/predictive-maintenance/package.json  lambda-functions/predictive-maintenance/index.js  lambda-functions/user-management/package.json  lambda-functions/user-management/index.js  scripts/generate-test-data.js

cmd/
  api/
  ingestor/
internal/
  config/
  database/
  domain/
  http/
  mqtt/
  repository/
  service/
pkg/
scripts/
deploy/
```

## Credentials (dev)

- Postgres: `postgres:postgres@energy_db:5432/energy`
- Redis: `energy_redis:6379`
- Mosquitto: `energy_mqtt:1883` (topic: `energy/readings`)
- Grafana: `http://localhost:3000` (`admin`/`admin`)

## License

MIT
