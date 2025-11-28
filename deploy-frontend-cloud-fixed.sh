#!/bin/bash

# ⚠️ DEPRECATED - DO NOT USE ⚠️
# This script is NO LONGER USED as the frontend has been integrated into the backend.
# Use deploy-elastic-beanstalk.sh instead to deploy the unified application.
# 
# Smart Energy Grid - Frontend Cloud Deployment with Custom Domain
# Deploys frontend to AWS Elastic Beanstalk with domain configuration

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
DOMAIN="aniketshetty.me"
SUBDOMAIN="dashboard.aniketshetty.me"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Frontend Cloud Deployment${NC}"
echo -e "${BLUE}With Custom Domain: $SUBDOMAIN${NC}"
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
cd "/Users/shetty/Desktop/Sem 1 Projects/Cloud Progm/smart-energy-grid-management-system/web/energy-dashboard"

echo -e "${BLUE}Step 1: Building Frontend Application${NC}"
print_info "Building Go dashboard for Linux..."

# Build for Linux (EB environment)
GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o energy-dashboard-go main.go

if [ ! -f "energy-dashboard-go" ]; then
    print_error "Build failed!"
    exit 1
fi

print_status "Frontend application built successfully"

echo -e "\n${BLUE}Step 2: Creating Elastic Beanstalk Application${NC}"
if /usr/local/bin/aws elasticbeanstalk describe-applications \
    --application-names $APP_NAME \
    --region $REGION 2>&1 | grep -q "ApplicationName"; then
    print_warning "Application already exists: $APP_NAME"
else
    print_info "Creating application: $APP_NAME"
    /usr/local/bin/aws elasticbeanstalk create-application \
        --application-name $APP_NAME \
        --description "Smart Energy Grid Dashboard" \
        --region $REGION
    print_status "Application created"
fi

echo -e "\n${BLUE}Step 3: Creating Deployment Package${NC}"
VERSION_LABEL="v$(date +%Y%m%d-%H%M%S)"
print_info "Version: $VERSION_LABEL"

# Create deployment package
zip -q -r deployment.zip \
    energy-dashboard-go \
    Procfile \
    .ebextensions/ \
    templates/ \
    static/ \
    2>/dev/null || true

print_status "Deployment package created"

# Upload to S3
ACCOUNT_ID=$(/usr/local/bin/aws sts get-caller-identity --query Account --output text)
S3_BUCKET="elasticbeanstalk-$REGION-$ACCOUNT_ID"

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

print_status "Application version created"

echo -e "\n${BLUE}Step 4: Checking IAM Roles${NC}"

# Check if IAM instance profile exists
if /usr/local/bin/aws iam get-instance-profile --instance-profile-name aws-elasticbeanstalk-ec2-role 2>/dev/null; then
    print_status "IAM instance profile exists"
else
    print_warning "IAM instance profile not found, using existing role"
fi

echo -e "\n${BLUE}Step 5: Creating Environment${NC}"
print_info "This will take 5-8 minutes..."

# Check if environment exists and is not terminated
ENV_STATUS=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
    --application-name $APP_NAME \
    --environment-names $ENV_NAME \
    --region $REGION \
    --query 'Environments[0].Status' \
    --output text 2>/dev/null || echo "NotFound")

if [ "$ENV_STATUS" = "Ready" ] || [ "$ENV_STATUS" = "Updating" ]; then
    print_warning "Environment exists, updating..."
    /usr/local/bin/aws elasticbeanstalk update-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $VERSION_LABEL \
        --region $REGION \
        >/dev/null
else
    print_info "Creating new environment..."
    /usr/local/bin/aws elasticbeanstalk create-environment \
        --application-name $APP_NAME \
        --environment-name $ENV_NAME \
        --version-label $VERSION_LABEL \
        --solution-stack-name "64bit Amazon Linux 2023 v4.5.0 running Go 1" \
        --option-settings \
            Namespace=aws:autoscaling:launchconfiguration,OptionName=InstanceType,Value=t3.micro \
            Namespace=aws:autoscaling:launchconfiguration,OptionName=IamInstanceProfile,Value=aws-elasticbeanstalk-ec2-role \
            Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=SingleInstance \
            Namespace=aws:elasticbeanstalk:environment,OptionName=ServiceRole,Value=aws-elasticbeanstalk-service-role \
            Namespace=aws:elasticbeanstalk:environment:process:default,OptionName=HealthCheckPath,Value=/ \
            Namespace=aws:elasticbeanstalk:environment:process:default,OptionName=Port,Value=3000 \
        --region $REGION \
        >/dev/null
fi

print_info "Waiting for environment to be ready..."

# Wait for environment (with timeout)
WAIT_COUNT=0
MAX_WAIT=40
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    STATUS=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
        --application-name $APP_NAME \
        --environment-names $ENV_NAME \
        --region $REGION \
        --query 'Environments[0].Status' \
        --output text 2>/dev/null || echo "Unknown")
    
    HEALTH=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
        --application-name $APP_NAME \
        --environment-names $ENV_NAME \
        --region $REGION \
        --query 'Environments[0].Health' \
        --output text 2>/dev/null || echo "Unknown")
    
    if [ "$STATUS" = "Ready" ] && [ "$HEALTH" = "Green" ]; then
        print_status "Environment is ready!"
        break
    fi
    
    if [ "$STATUS" = "Terminated" ]; then
        print_error "Environment terminated! Check AWS Console for errors."
        break
    fi
    
    echo -ne "\r  Status: $STATUS | Health: $HEALTH | Waiting... ($WAIT_COUNT/$MAX_WAIT)"
    sleep 15
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

echo ""

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
    print_warning "Environment is taking longer than expected"
    print_info "Check status manually in AWS Console"
fi

echo -e "\n${BLUE}Step 6: Getting Environment URL${NC}"
DASHBOARD_URL=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
    --application-name $APP_NAME \
    --environment-names $ENV_NAME \
    --region $REGION \
    --query "Environments[0].CNAME" \
    --output text)

if [ "$DASHBOARD_URL" != "None" ] && [ -n "$DASHBOARD_URL" ]; then
    print_status "Default URL: http://$DASHBOARD_URL"
else
    print_error "Could not retrieve environment URL"
    print_info "Check AWS Console for environment status"
    DASHBOARD_URL="Check-AWS-Console"
fi

echo -e "\n${BLUE}Step 7: Domain Configuration${NC}"
print_warning "To configure your custom domain ($SUBDOMAIN):"
echo ""
echo "Using Namecheap (Your Registrar):"
echo "  1. Log in to Namecheap: https://www.namecheap.com/"
echo "  2. Go to Domain List → Manage $DOMAIN"
echo "  3. Click 'Advanced DNS' tab"
echo "  4. Add New Record:"
echo "     Type: CNAME Record"
echo "     Host: dashboard"
echo "     Value: $DASHBOARD_URL"
echo "     TTL: Automatic (or 300)"
echo ""
print_info "DNS propagation takes 5-60 minutes"

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Frontend Deployment Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Access URLs:${NC}"
echo -e "  Default: ${YELLOW}http://$DASHBOARD_URL${NC}"
echo -e "  Custom (after DNS): ${YELLOW}http://$SUBDOMAIN${NC}\n"

if [ "$DASHBOARD_URL" != "Check-AWS-Console" ] && [ "$DASHBOARD_URL" != "None" ]; then
    echo -e "${BLUE}Dashboard Pages:${NC}"
    echo -e "  Home:       http://$DASHBOARD_URL/"
    echo -e "  Dashboard:  http://$DASHBOARD_URL/dashboard"
    echo -e "  Alerts:     http://$DASHBOARD_URL/alerts"
    echo -e "  Equipment:  http://$DASHBOARD_URL/equipment"
    echo -e "  Analytics:  http://$DASHBOARD_URL/analytics\n"
fi

echo -e "${BLUE}Namecheap DNS Setup:${NC}"
echo -e "  1. Login: ${YELLOW}https://www.namecheap.com/myaccount/login${NC}"
echo -e "  2. Domain List → Manage → Advanced DNS"
echo -e "  3. Add CNAME: dashboard → $DASHBOARD_URL"
echo -e "  4. Wait 5-60 minutes for DNS propagation\n"

echo -e "${YELLOW}Cost Estimate:${NC}"
echo -e "  t3.micro EC2: ~\$0.0104/hour (~\$7.50/month)"
echo -e "  For 9 days: ~\$2.25"
echo -e "  ${GREEN}Remember to delete after demo!${NC}\n"

print_status "Deployment completed successfully!"

# Cleanup
rm -f deployment.zip

# Return to root
cd "../.."
