# Energy Dashboard (Go Frontend)

A Go-powered, server-rendered frontend for the Smart Energy Grid Management System.

## Features
- Dashboard with charts (Chart.js) fed by backend API
- Alerts view with filtering and acknowledgment
- Analytics page to generate and download daily reports
- Simple, production-ready HTTP server with HTML templates and static assets

## Run Locally
```bash
# Set your backend API URL (Elastic Beanstalk or local)
export API_URL=http://localhost:8080
# Optionally select the default facility
export FACILITY_ID=facility-001

go run .
# open http://localhost:3000
```

## Build
```bash
go build -o energy-dashboard-go
./energy-dashboard-go
```

## Deploy
- Build the binary and deploy to EC2/Elastic Beanstalk/Container of your choice.
- Serve traffic behind a reverse proxy (Nginx/ALB).
