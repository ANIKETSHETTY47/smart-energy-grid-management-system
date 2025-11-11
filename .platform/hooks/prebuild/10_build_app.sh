#!/usr/bin/env bash
set -euo pipefail

echo "[prebuild] building Go binary"
cd /var/app/staging

# If your module root is the repo root and main is in ./cmd/api:
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o application ./cmd/api

chmod +x application
ls -la application
