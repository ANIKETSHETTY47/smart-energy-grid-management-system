#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Checking Elastic Beanstalk Frontend Status...${NC}\n"

# Check frontend environment
echo -e "${BLUE}Frontend Environment:${NC}"
/usr/local/bin/aws elasticbeanstalk describe-environments \
  --application-name energy-dashboard \
  --environment-names energy-dashboard-env \
  --region eu-north-1 \
  --query 'Environments[0].[EnvironmentName,Status,Health,CNAME]' \
  --output table

echo -e "\n${BLUE}Backend Environment:${NC}"
/usr/local/bin/aws elasticbeanstalk describe-environments \
  --application-name smart-energy-grid \
  --environment-names smart-energy-grid-env \
  --region eu-north-1 \
  --query 'Environments[0].[EnvironmentName,Status,Health,CNAME]' \
  --output table

echo -e "\n${BLUE}Testing Frontend URL Directly:${NC}"
FRONTEND_URL=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
  --application-name energy-dashboard \
  --environment-names energy-dashboard-env \
  --region eu-north-1 \
  --query 'Environments[0].CNAME' \
  --output text)

if [ "$FRONTEND_URL" != "None" ] && [ -n "$FRONTEND_URL" ]; then
    echo -e "Testing: ${YELLOW}http://$FRONTEND_URL${NC}"
    curl -I -m 10 "http://$FRONTEND_URL" 2>&1 | head -5
else
    echo -e "${RED}Frontend URL not found or environment terminated${NC}"
fi
