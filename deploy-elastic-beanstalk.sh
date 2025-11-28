#!/bin/bash

# Elastic Beanstalk Deployment Script for Smart Energy Grid Management System
# This script deploys the UNIFIED backend + dashboard application

set -e  # Exit on error

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
REGION="eu-north-1"
APP_NAME="smart-energy-grid"
ENV_NAME="smart-energy-grid-env"
PLATFORM="Go 1 running on 64bit Amazon Linux 2023"
INSTANCE_TYPE="t3.micro"  # Free tier eligible

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Elastic Beanstalk Deployment${NC}"
echo -e "${BLUE}Unified Backend + Dashboard${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Function to print status
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

# Check AWS CLI and EB CLI installation
if ! command -v aws &> /dev/null; then
    print_error "AWS CLI not found. Please install it first."
    exit 1
fi

if ! command -v eb &> /dev/null; then
    print_error "EB CLI not found. Installing via pip..."
    pip3 install awsebcli --upgrade --user
    if ! command -v eb &> /dev/null; then
        print_error "EB CLI installation failed. Please install manually: pip3 install awsebcli"
        exit 1
    fi
fi

print_status "AWS CLI and EB CLI found"

# Configure AWS region
export AWS_DEFAULT_REGION=$REGION
print_info "Using AWS Region: $REGION"

echo -e "\n${BLUE}Step 1: Building Application${NC}"
print_info "Building unified Go application (API + Dashboard)..."
GOOS=linux GOARCH=amd64 go build -o smart-energy-grid-api ./cmd/api
if [ ! -f "smart-energy-grid-api" ]; then
    print_error "Build failed!"
    exit 1
fi
print_status "Application built successfully"

echo -e "\n${BLUE}Step 2: Creating/Updating Elastic Beanstalk Application${NC}"
if aws elasticbeanstalk describe-applications --application-names $APP_NAME --region $REGION 2>&1 | grep -q "ApplicationName"; then
    print_warning "Application already exists: $APP_NAME"
else
    print_info "Creating Elastic Beanstalk application: $APP_NAME"
    aws elasticbeanstalk create-application \
        --application-name $APP_NAME \
        --description "Smart Energy Grid Management System - Unified Backend + Dashboard" \
        --region $REGION
    print_status "Application created: $APP_NAME"
fi

echo -e "\n${BLUE}Step 3: Creating Application Version${NC}"
VERSION_LABEL="v$(date +%Y%m%d-%H%M%S)"
print_info "Creating version: $VERSION_LABEL"

# Create deployment package
print_info "Creating deployment package..."
print_info "Including: binary, dashboard templates, static assets, configs..."
zip -q -r deployment.zip \
    smart-energy-grid-api \
    web/dashboard/templates \
    web/dashboard/static \
    .ebextensions \
    Procfile
print_status "Deployment package created"

# Upload to S3 (EB creates a bucket automatically)
S3_BUCKET="elasticbeanstalk-$REGION-$(aws sts get-caller-identity --query Account --output text)"
print_info "Uploading to S3: $S3_BUCKET"

aws s3 mb "s3://$S3_BUCKET" --region $REGION 2>/dev/null || true
aws s3 cp deployment.zip "s3://$S3_BUCKET/$APP_NAME/$VERSION_LABEL.zip" --region $REGION

# Create application version
aws elasticbeanstalk create-application-version \
    --application-name $APP_NAME \
    --version-label $VERSION_LABEL \
    --source-bundle S3Bucket="$S3_BUCKET",S3Key="$APP_NAME/$VERSION_LABEL.zip" \
    --region $REGION

print_status "Application version created: $VERSION_LABEL"

echo -e "\n${BLUE}Step 4: Creating/Updating Environment${NC}"
if aws elasticbeanstalk describe-environments --application-name $APP_NAME --environment-names $ENV_NAME --region $REGION 2>&1 | grep -q "EnvironmentName"; then
    print_warning "Environment already exists: $ENV_NAME"
    print_info "Updating environment with new version..."
    
    aws elasticbeanstalk update-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $VERSION_LABEL \
        --region $REGION
    
    print_info "Waiting for environment to update (this may take 5-10 minutes)..."
    aws elasticbeanstalk wait environment-updated \
        --application-name $APP_NAME \
        --environment-names $ENV_NAME \
        --region $REGION
    
    print_status "Environment updated successfully"
else
    print_info "Creating Elastic Beanstalk environment: $ENV_NAME"
    print_info "This will take 5-10 minutes..."
    
    # Create environment
    aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $VERSION_LABEL \
        --solution-stack-name "64bit Amazon Linux 2023 v4.5.0 running Go 1" \
        --option-settings \
            Namespace=aws:autoscaling:launchconfiguration,OptionName=InstanceType,Value=$INSTANCE_TYPE \
            Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=SingleInstance \
            Namespace=aws:elasticbeanstalk:environment,OptionName=ServiceRole,Value=aws-elasticbeanstalk-service-role \
            Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=aws-elasticbeanstalk-ec2-role \
        --region $REGION
    
    print_info "Waiting for environment to be ready..."
    aws elasticbeanstalk wait environment-exists \
        --application-name $APP_NAME \
        --environment-names $ENV_NAME \
        --region $REGION
    
    print_status "Environment created successfully"
fi

echo -e "\n${BLUE}Step 5: Getting Environment URL${NC}"
ENV_URL=$(aws elasticbeanstalk describe-environments \
    --application-name $APP_NAME \
    --environment-names $ENV_NAME \
    --region $REGION \
    --query "Environments[0].CNAME" \
    --output text)

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Deployment Summary:${NC}"
echo -e "  Application: ${GREEN}$APP_NAME${NC}"
echo -e "  Environment: ${GREEN}$ENV_NAME${NC}"
echo -e "  Version: ${GREEN}$VERSION_LABEL${NC}"
echo -e "  Region: ${GREEN}$REGION${NC}"
echo -e "  URL: ${GREEN}http://$ENV_URL${NC}\n"

echo -e "${BLUE}Dashboard & API Endpoints:${NC}"
echo -e "  Dashboard: ${YELLOW}http://$ENV_URL/${NC}"
echo -e "  Dashboard Home: ${YELLOW}http://$ENV_URL/dashboard${NC}"
echo -e "  Alerts Page: ${YELLOW}http://$ENV_URL/alerts${NC}"
echo -e "  Equipment Page: ${YELLOW}http://$ENV_URL/equipment${NC}"
echo -e "  Analytics Page: ${YELLOW}http://$ENV_URL/analytics${NC}"
echo -e "  Health Check: ${YELLOW}http://$ENV_URL/health${NC}"
echo -e "  API Readings: ${YELLOW}http://$ENV_URL/readings${NC}\n"

echo -e "${BLUE}With Cloudflare:${NC}"
echo -e "  Dashboard URL: ${GREEN}https://dashboard.aniketshetty.me${NC}\n"

echo -e "${BLUE}Next Steps:${NC}"
echo -e "  1. Test the endpoints above"
echo -e "  2. Configure Cloudflare DNS: CNAME 'dashboard' → $ENV_URL"
echo -e "  3. Set Cloudflare SSL/TLS to 'Full'"
echo -e "  4. Check CloudWatch logs: ${YELLOW}aws logs tail /aws/elasticbeanstalk/$ENV_NAME/var/log/web.stdout.log --follow${NC}"
echo -e "  5. Monitor in AWS Console: https://$REGION.console.aws.amazon.com/elasticbeanstalk"

print_status "Deployment script completed successfully!"

# Cleanup
rm -f deployment.zip smart-energy-grid-api
