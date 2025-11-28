#!/bin/bash

# Master Deployment Script - Runs Everything in Order
# Smart Energy Grid Management System

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Smart Energy Grid - Complete Deployment      ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}\n"

print_step() {
    echo -e "\n${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}  $1${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}\n"
}

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Make scripts executable
chmod +x setup-aws-infrastructure.sh
chmod +x deploy-elastic-beanstalk.sh

# Step 1: AWS Infrastructure
print_step "STEP 1: Setting up AWS Infrastructure"
print_info "Creating S3, DynamoDB, SNS, Lambda, IAM..."
./setup-aws-infrastructure.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}Infrastructure setup failed!${NC}"
    exit 1
fi

print_status "AWS Infrastructure setup complete"

# Wait for user confirmation
echo -e "\n${YELLOW}⚠ Important: Check that all AWS resources were created successfully.${NC}"
echo -e "${YELLOW}Press Enter to continue with Elastic Beanstalk deployment...${NC}"
read -r

# Step 2: Deploy Backend API
print_step "STEP 2: Deploying Backend API to Elastic Beanstalk"
print_info "Building and deploying Go application..."
./deploy-elastic-beanstalk.sh

if [ $? -ne 0 ]; then
    echo -e "${RED}Elastic Beanstalk deployment failed!${NC}"
    exit 1
fi

print_status "Backend API deployed"

# Step 3: Get deployment info
print_step "STEP 3: Deployment Summary"

REGION="eu-north-1"
APP_NAME="smart-energy-grid"
ENV_NAME="smart-energy-grid-env"

ENV_URL=$(aws elasticbeanstalk describe-environments \
    --application-name $APP_NAME \
    --environment-names $ENV_NAME \
    --region $REGION \
    --query "Environments[0].CNAME" \
    --output text 2>/dev/null || echo "Not found")

echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           DEPLOYMENT SUCCESSFUL!               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}\n"

echo -e "${GREEN}✓ AWS Services Created:${NC}"
echo -e "  • S3 Bucket: smart-energy-grid-reports-nci"
echo -e "  • DynamoDB Tables: 3 (readings, alerts, equipment)"
echo -e "  • SNS Topic: energy-grid-alerts"
echo -e "  • Lambda Functions: 2 (anomaly-detection, analytics-processing)"
echo -e "  • CloudWatch: Logging enabled"
echo -e "  • IAM Role: EnergyGridLambdaRole\n"

echo -e "${GREEN}✓ Application Deployed:${NC}"
echo -e "  • Elastic Beanstalk App: $APP_NAME"
echo -e "  • Environment: $ENV_NAME"
echo -e "  • Region: $REGION"
echo -e "  • URL: ${YELLOW}http://$ENV_URL${NC}\n"

echo -e "${BLUE}📋 Testing URLs:${NC}"
echo -e "  Health Check: ${YELLOW}curl http://$ENV_URL/health${NC}"
echo -e "  API Metrics: ${YELLOW}curl http://$ENV_URL/api/metrics${NC}"
echo -e "  Alerts: ${YELLOW}curl http://$ENV_URL/api/alerts${NC}\n"

echo -e "${BLUE}📊 AWS Console Links:${NC}"
echo -e "  Elastic Beanstalk: https://$REGION.console.aws.amazon.com/elasticbeanstalk"
echo -e "  DynamoDB: https://$REGION.console.aws.amazon.com/dynamodbv2"
echo -e "  Lambda: https://$REGION.console.aws.amazon.com/lambda"
echo -e "  S3: https://s3.console.aws.amazon.com/s3/buckets/smart-energy-grid-reports-nci"
echo -e "  CloudWatch Logs: https://$REGION.console.aws.amazon.com/cloudwatch/home?region=$REGION#logsV2:log-groups\n"

echo -e "${GREEN}✓ Deployment script completed!${NC}\n"

echo -e "${YELLOW}📝 IMPORTANT: Save these details for your project report!${NC}\n"
