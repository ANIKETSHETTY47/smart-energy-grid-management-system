#!/bin/bash

# Frontend Dashboard - Local Run Script
# Starts the Energy Grid Dashboard locally

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Starting Energy Grid Dashboard${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Set environment variables
export PORT=3000
export API_URL="http://smart-energy-grid-env.eba-8hsepki2.eu-north-1.elasticbeanstalk.com"
export FACILITY_ID="facility-001"

echo -e "${GREEN}Configuration:${NC}"
echo -e "  Port: ${YELLOW}$PORT${NC}"
echo -e "  API URL: ${YELLOW}$API_URL${NC}"
echo -e "  Facility ID: ${YELLOW}$FACILITY_ID${NC}\n"

echo -e "${GREEN}Dashboard Pages:${NC}"
echo -e "  Home:       ${YELLOW}http://localhost:$PORT/${NC}"
echo -e "  Dashboard:  ${YELLOW}http://localhost:$PORT/dashboard${NC}"
echo -e "  Alerts:     ${YELLOW}http://localhost:$PORT/alerts${NC}"
echo -e "  Equipment:  ${YELLOW}http://localhost:$PORT/equipment${NC}"
echo -e "  Analytics:  ${YELLOW}http://localhost:$PORT/analytics${NC}\n"

echo -e "${BLUE}Starting server...${NC}\n"

# Navigate to dashboard directory
cd web/energy-dashboard

# Run the dashboard
go run main.go
