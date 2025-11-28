#!/bin/bash

# ⚠️ DEPRECATED - DO NOT USE ⚠️
# This script is NO LONGER USED as the frontend has been integrated into the backend.
# Use deploy-elastic-beanstalk.sh instead to deploy the unified application.
# 
# Frontend Dashboard Deployment Script
# Deploys the Energy Grid Dashboard to Elastic Beanstalk

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
REGION="eu-north-1"
APP_NAME="energy-dashboard"
ENV_NAME="energy-dashboard-env"
BACKEND_URL="http://smart-energy-grid-env.eba-8hsepki2.eu-north-1.elasticbeanstalk.com"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Frontend Dashboard Deployment${NC}"
echo -e "${BLUE}========================================${NC}\n"

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Navigate to dashboard directory
cd web/energy-dashboard

echo -e "\n${BLUE}Step 1: Building Frontend Application${NC}"
print_info "Building Go dashboard application..."

# Build for Linux
GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o energy-dashboard-go main.go

if [ ! -f "energy-dashboard-go" ]; then
    print_error "Build failed!"
    exit 1
fi

print_status "Frontend application built successfully"

echo -e "\n${BLUE}Step 2: Creating/Updating Elastic Beanstalk Application${NC}"
if /usr/local/bin/aws elasticbeanstalk describe-applications \
    --application-names $APP_NAME \
    --region $REGION 2>&1 | grep -q "ApplicationName"; then
    print_warning "Application already exists: $APP_NAME"
else
    print_info "Creating Elastic Beanstalk application: $APP_NAME"
    /usr/local/bin/aws elasticbeanstalk create-application \
        --application-name $APP_NAME \
        --description "Smart Energy Grid Dashboard (Frontend)" \
        --region $REGION
    print_status "Application created: $APP_NAME"
fi

echo -e "\n${BLUE}Step 3: Creating Deployment Package${NC}"
VERSION_LABEL="v$(date +%Y%m%d-%H%M%S)"
print_info "Creating version: $VERSION_LABEL"

# Create deployment package
print_info "Packaging frontend files..."
zip -q -r deployment.zip \
    energy-dashboard-go \
    Procfile \
    .ebextensions/ \
    templates/ \
    static/ \
    2>/dev/null || true

print_status "Deployment package created"

# Upload to S3
S3_BUCKET="elasticbeanstalk-$REGION-$(/usr/local/bin/aws sts get-caller-identity --query Account --output text)"
print_info "Uploading to S3: $S3_BUCKET"

/usr/local/bin/aws s3 cp deployment.zip \
    "s3://$S3_BUCKET/$APP_NAME/$VERSION_LABEL.zip" \
    --region $REGION

# Create application version
/usr/local/bin/aws elasticbeanstalk create-application-version \
    --application-name $APP_NAME \
    --version-label $VERSION_LABEL \
    --source-bundle S3Bucket="$S3_BUCKET",S3Key="$APP_NAME/$VERSION_LABEL.zip" \
    --region $REGION \
    >/dev/null

print_status "Application version created: $VERSION_LABEL"

echo -e "\n${BLUE}Step 4: Creating/Updating Environment${NC}"
if /usr/local/bin/aws elasticbeanstalk describe-environments \
    --application-name $APP_NAME \
    --environment-names $ENV_NAME \
    --region $REGION 2>&1 | grep -q "EnvironmentName"; then
    
    print_warning "Environment already exists: $ENV_NAME"
    print_info "Updating environment with new version..."
    
    /usr/local/bin/aws elasticbeanstalk update-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $VERSION_LABEL \
        --region $REGION \
        >/dev/null
    
    print_info "Waiting for environment to update (3-5 minutes)..."
    /usr/local/bin/aws elasticbeanstalk wait environment-updated \
        --application-name $APP_NAME \
        --environment-names $ENV_NAME \
        --region $REGION
    
    print_status "Environment updated successfully"
else
    print_info "Creating Elastic Beanstalk environment: $ENV_NAME"
    print_info "This will take 5-10 minutes..."
    
    /usr/local/bin/aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $VERSION_LABEL \
        --solution-stack-name "64bit Amazon Linux 2023 v4.5.0 running Go 1" \
        --option-settings \
            Namespace=aws:autoscaling:launchconfiguration,OptionName=InstanceType,Value=t3.micro \
            Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=SingleInstance \
        --region $REGION \
        >/dev/null
    
    print_info "Waiting for environment to be ready..."
    /usr/local/bin/aws elasticbeanstalk wait environment-exists \
        --application-name $APP_NAME \
        --environment-names $ENV_NAME \
        --region $REGION
    
    print_status "Environment created successfully"
fi

echo -e "\n${BLUE}Step 5: Getting Environment URL${NC}"
DASHBOARD_URL=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
    --application-name $APP_NAME \
    --environment-names $ENV_NAME \
    --region $REGION \
    --query "Environments[0].CNAME" \
    --output text)

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Frontend Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Deployment Summary:${NC}"
echo -e "  Application: ${GREEN}$APP_NAME${NC}"
echo -e "  Environment: ${GREEN}$ENV_NAME${NC}"
echo -e "  Version: ${GREEN}$VERSION_LABEL${NC}"
echo -e "  Region: ${GREEN}$REGION${NC}"
echo -e "  Dashboard URL: ${GREEN}http://$DASHBOARD_URL${NC}\n"

echo -e "${BLUE}Dashboard Pages:${NC}"
echo -e "  Home: ${YELLOW}http://$DASHBOARD_URL/${NC}"
echo -e "  Dashboard: ${YELLOW}http://$DASHBOARD_URL/dashboard${NC}"
echo -e "  Alerts: ${YELLOW}http://$DASHBOARD_URL/alerts${NC}"
echo -e "  Equipment: ${YELLOW}http://$DASHBOARD_URL/equipment${NC}"
echo -e "  Analytics: ${YELLOW}http://$DASHBOARD_URL/analytics${NC}\n"

echo -e "${BLUE}Connected to Backend:${NC}"
echo -e "  API URL: ${YELLOW}$BACKEND_URL${NC}\n"

echo -e "${BLUE}Next Steps:${NC}"
echo -e "  1. Open dashboard in browser: ${YELLOW}http://$DASHBOARD_URL${NC}"
echo -e "  2. Take screenshots of the dashboard"
echo -e "  3. Include frontend URL in your report\n"

print_status "Frontend deployment completed successfully!"

# Cleanup
rm -f deployment.zip

# Return to root directory
cd ../..
