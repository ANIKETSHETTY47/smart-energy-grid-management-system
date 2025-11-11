build:
  - set -euo pipefail
  - go mod download
  - GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o application ./cmd/api
  - chmod +x application
  - ls -la .