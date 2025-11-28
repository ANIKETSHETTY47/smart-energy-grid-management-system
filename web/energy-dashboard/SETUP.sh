#!/bin/bash

echo "=========================================="
echo "Smart Energy Grid - Frontend Setup"
echo "=========================================="
echo ""

# Check if we're in the right directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: Please run this script from web/energy-dashboard directory"
    exit 1
fi

echo "Step 1: Creating directory structure..."
mkdir -p internal/server
mkdir -p internal/api
mkdir -p internal/models
mkdir -p templates
mkdir -p static/css

echo "✅ Directories created"
echo ""

echo "Step 2: Installing Go dependencies..."
go get github.com/gorilla/websocket@v1.5.3
go mod tidy

echo "✅ Dependencies installed"
echo ""

echo "Step 3: Verifying file structure..."
echo ""
echo "Please ensure you have these files:"
echo "  ✓ internal/server/server.go (with WebSocket support)"
echo "  ✓ internal/models/types.go (with Equipment model)"
echo "  ✓ internal/api/client.go (API client)"
echo "  ✓ templates/layout.html (base layout)"
echo "  ✓ templates/dashboard.html (dashboard page)"
echo "  ✓ templates/equipment.html (equipment page)"
echo "  ✓ templates/alerts.html (alerts page)"
echo "  ✓ templates/analytics.html (analytics page)"
echo "  ✓ static/css/app.css (enhanced styles)"
echo "  ✓ main.go (entry point)"
echo ""

echo "Step 4: Setting up environment variables..."
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "Creating .env file..."
    cat > .env << EOF
API_URL=http://localhost:8080
FACILITY_ID=facility-001
PORT=3000
EOF
    echo "✅ .env file created"
else
    echo "✅ .env file already exists"
fi

echo ""
echo "Step 5: Building the application..."
go build -o energy-dashboard-go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed. Please check the errors above."
    exit 1
fi

echo ""
echo "=========================================="
echo "✅ Setup Complete!"
echo "=========================================="
echo ""
echo "To start the application:"
echo ""
echo "  1. Ensure backend API is running at http://localhost:8080"
echo "  2. Run: export API_URL=http://localhost:8080"
echo "  3. Run: export FACILITY_ID=facility-001"
echo "  4. Run: ./energy-dashboard-go"
echo "  5. Open: http://localhost:3000"
echo ""
echo "Or simply run: go run main.go"
echo ""
echo "For production build:"
echo "  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o energy-dashboard-go"
echo ""
echo "=========================================="