# Deployment Architecture - Unified Backend + Dashboard

## Overview

The Smart Energy Grid Management System now runs as a **single unified Go application** that serves both:
- **API endpoints** (JSON) for data operations
- **Dashboard HTML pages** for web interface
- **Static assets** (CSS, JS, images)

All from one Elastic Beanstalk environment, accessible via:
- **Production URL**: `https://dashboard.aniketshetty.me` (via Cloudflare)
- **Direct EB URL**: `http://smart-energy-grid-env.eba-xxxx.eu-north-1.elasticbeanstalk.com`

## Architecture Changes

### Before (Separated Frontend/Backend)
```
┌─────────────────┐         ┌─────────────────┐
│ Frontend EB App │ ──────▶ │ Backend EB App  │
│ energy-dashboard│  CORS   │ smart-energy-   │
│ :3002           │         │ grid-env :8080  │
└─────────────────┘         └─────────────────┘
```

### After (Unified Application)
```
┌────────────────────────────────────┐
│   Single EB Environment            │
│   smart-energy-grid-env :8080      │
│                                    │
│   ┌─────────────┐  ┌────────────┐ │
│   │ Dashboard   │  │ API Routes │ │
│   │ HTML/Static │  │ JSON       │ │
│   │ /dashboard  │  │ /readings  │ │
│   │ /alerts     │  │ /alerts    │ │
│   │ /equipment  │  │ /analytics │ │
│   │ /analytics  │  │ /health    │ │
│   └─────────────┘  └────────────┘ │
└────────────────────────────────────┘
             ▲
             │
    ┌────────┴────────┐
    │   Cloudflare    │
    │   SSL/TLS Full  │
    │   dashboard.    │
    │   aniketshetty  │
    │   .me           │
    └─────────────────┘
```

## Project Structure

```
smart-energy-grid-management-system/
├── cmd/api/                          # Main application entry point
│   └── main.go                       # Integrates dashboard + API
├── internal/
│   ├── dashboard/                    # Dashboard package (NEW)
│   │   ├── api/
│   │   │   └── client.go            # API client (same-origin support)
│   │   ├── models/
│   │   │   └── types.go             # Dashboard data models
│   │   └── server/
│   │       └── server.go            # HTTP handler implementation
│   ├── http/                         # API handlers
│   ├── service/                      # Business logic
│   └── ...
├── web/
│   ├── dashboard/                    # Dashboard assets (NEW LOCATION)
│   │   ├── templates/               # HTML templates
│   │   │   ├── layout.html
│   │   │   ├── dashboard.html
│   │   │   ├── alerts.html
│   │   │   ├── equipment.html
│   │   │   └── analytics.html
│   │   └── static/                  # CSS, JS, images
│   │       └── css/
│   │           └── app.css
│   └── energy-dashboard/            # OLD - now deprecated
│       ├── go.mod.disabled          # Disabled separate module
│       └── ...
├── deploy-elastic-beanstalk.sh      # MAIN DEPLOYMENT SCRIPT
├── deploy-frontend*.sh              # DEPRECATED (marked)
├── Procfile                          # Runs: ./smart-energy-grid-api
└── .ebextensions/
    └── 01_environment.config        # Environment variables
```

## Key Code Integration

### 1. Dashboard as http.Handler

The dashboard package exposes a standard `http.Handler`:

```go
// internal/dashboard/server/server.go
func New(staticDir, templatesDir string, apiClient *api.Client) http.Handler {
    // Returns a Server that implements http.Handler
    // Can be mounted anywhere in your router
}
```

### 2. API Client with Same-Origin Support

```go
// internal/dashboard/api/client.go
func New() *Client {
    base := os.Getenv("API_URL")
    // If API_URL is empty, uses relative paths (same origin)
    return &Client{baseURL: base, ...}
}

func (c *Client) makeURL(path string) string {
    if c.baseURL == "" {
        return path  // Same-origin: /readings, /alerts, etc.
    }
    return strings.TrimRight(c.baseURL, "/") + path
}
```

### 3. Fiber Integration

```go
// cmd/api/main.go
import (
    dashboardAPI "github.com/.../internal/dashboard/api"
    dashboardServer "github.com/.../internal/dashboard/server"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
)

func main() {
    app := fiber.New()
    
    // API routes (existing)
    httpHandlers.Register(app, svcs)
    
    // Dashboard integration
    dashClient := dashboardAPI.New()
    dashHandler := dashboardServer.New(
        "./web/dashboard/static",
        "./web/dashboard/templates",
        dashClient,
    )
    
    // Mount static assets
    app.Static("/static", "./web/dashboard/static")
    
    // Mount dashboard HTML routes
    app.All("/", adaptor.HTTPHandler(dashHandler))
    app.All("/dashboard", adaptor.HTTPHandler(dashHandler))
    app.All("/alerts", adaptor.HTTPHandler(dashHandler))
    app.All("/equipment", adaptor.HTTPHandler(dashHandler))
    app.All("/analytics", adaptor.HTTPHandler(dashHandler))
    app.All("/ws", adaptor.HTTPHandler(dashHandler))
    
    app.Listen(":8080")
}
```

## Environment Variables

Set these in `.ebextensions/01_environment.config` or locally:

```yaml
# Production (same-origin)
FACILITY_ID: facility-001
# API_URL: <leave empty for same-origin>

# Local development (if needed)
API_URL: http://localhost:8080
FACILITY_ID: facility-001
PORT: 8080
```

## Running Locally

### 1. From Project Root

```bash
# Build and run
go build -o bin/api ./cmd/api
./bin/api

# Or run directly
go run ./cmd/api
```

### 2. Test Endpoints

```bash
# Dashboard pages
curl -I http://localhost:8080/
curl -I http://localhost:8080/dashboard
curl -I http://localhost:8080/alerts
curl -I http://localhost:8080/equipment
curl -I http://localhost:8080/analytics

# API endpoints
curl http://localhost:8080/health
curl "http://localhost:8080/readings/recent?facility_id=facility-001&hours=24"
curl http://localhost:8080/alerts
```

### 3. Check Static Assets

```bash
# Verify CSS loads
curl -I http://localhost:8080/static/css/app.css
```

## Deployment to Elastic Beanstalk

### Build and Deploy

```bash
# Make script executable
chmod +x deploy-elastic-beanstalk.sh

# Deploy unified app
./deploy-elastic-beanstalk.sh
```

### What the Script Does

1. **Builds** the Go binary for Linux:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o smart-energy-grid-api ./cmd/api
   ```

2. **Creates deployment package** with:
   - Binary: `smart-energy-grid-api`
   - Templates: `web/dashboard/templates/**`
   - Static: `web/dashboard/static/**`
   - Config: `.ebextensions/`, `Procfile`

3. **Uploads to EB**:
   - Application: `smart-energy-grid`
   - Environment: `smart-energy-grid-env`
   - Region: `eu-north-1`

4. **Returns URL**: `http://smart-energy-grid-env.eba-xxxx.eu-north-1.elasticbeanstalk.com`

## Cloudflare Configuration

### DNS Setup

```
Type: CNAME
Name: dashboard
Target: smart-energy-grid-env.eba-xxxx.eu-north-1.elasticbeanstalk.com
Proxy: On (orange cloud)
```

### SSL/TLS

- **Mode**: Full (not Flexible)
- **Always Use HTTPS**: On
- **Auto Minify**: CSS, JS, HTML

### Result

- **Production URL**: `https://dashboard.aniketshetty.me`
- All HTTP traffic auto-redirects to HTTPS
- Cloudflare provides SSL certificate
- EB handles backend over HTTP (no SSL cert needed on EB)

## Monitoring & Logs

### View Logs

```bash
# Real-time logs
aws logs tail /aws/elasticbeanstalk/smart-energy-grid-env/var/log/web.stdout.log --follow

# Recent logs
aws logs tail /aws/elasticbeanstalk/smart-energy-grid-env/var/log/web.stdout.log --since 1h
```

### Health Check

```bash
# Via Cloudflare
curl https://dashboard.aniketshetty.me/health

# Direct to EB
curl http://smart-energy-grid-env.eba-xxxx.eu-north-1.elasticbeanstalk.com/health
```

### EB Console

Visit: [https://eu-north-1.console.aws.amazon.com/elasticbeanstalk](https://eu-north-1.console.aws.amazon.com/elasticbeanstalk)

## Troubleshooting

### Dashboard Not Loading

1. Check if binary includes templates:
   ```bash
   unzip -l deployment.zip | grep templates
   ```

2. Verify paths in code match deployment:
   ```go
   // Should be relative to where binary runs (EB app root)
   "./web/dashboard/templates"
   "./web/dashboard/static"
   ```

3. Check EB logs for template errors

### API Calls Failing

1. Ensure `API_URL` is **empty** in production (same-origin)
2. Check CORS isn't blocking requests (shouldn't happen with same-origin)
3. Verify API routes are registered before dashboard routes

### Static Assets Not Loading

1. Verify `app.Static("/static", "./web/dashboard/static")` comes before HTML routes
2. Check file paths in templates: `/static/css/app.css`
3. Ensure files are in deployment.zip:
   ```bash
   unzip -l deployment.zip | grep static
   ```

## Migration Checklist

- [x] Dashboard moved to `internal/dashboard/` package
- [x] Import paths updated to root module
- [x] `http.Handler` interface implemented
- [x] API client supports same-origin calls
- [x] Integrated into Fiber backend
- [x] Static assets mounted
- [x] Build script updated
- [x] Deployment script updated
- [x] Procfile updated
- [x] Old frontend scripts marked deprecated
- [x] Environment variables configured
- [x] Local testing successful
- [x] Documentation created

## Benefits of Unified Architecture

1. **Single Deployment**: One EB environment instead of two
2. **No CORS Issues**: Dashboard and API on same origin
3. **Simpler Infrastructure**: Less to manage and monitor
4. **Cost Savings**: One instance instead of two
5. **Faster**: No cross-origin latency
6. **Easier SSL**: Single domain to secure

## Support

For issues or questions:
- Check logs: `aws logs tail ...`
- EB console: AWS Console > Elastic Beanstalk
- GitHub issues: [repository URL]
