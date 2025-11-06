#!/bin/bash

# Smart Energy Grid - AWS Infrastructure Setup Script
# This script creates all necessary AWS resources for Elastic Beanstalk deployment

set -e  # Exit on error

# Configuration
AWS_REGION="us-east-1"
APP_NAME="smart-energy-grid"
ENV_NAME="smart-energy-grid-env"
S3_BUCKET="smart-energy-grid-deployments"
SOLUTION_STACK="64bit Amazon Linux 2023 v4.3.3 running Go 1"

echo "=========================================="
echo "Smart Energy Grid - AWS Setup"
echo "=========================================="
echo "Region: $AWS_REGION"
echo "Application: $APP_NAME"
echo "Environment: $ENV_NAME"
echo "=========================================="

# Step 1: Create S3 bucket for deployments
echo ""
echo "Step 1: Creating S3 bucket..."
if aws s3 ls "s3://$S3_BUCKET" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3 mb "s3://$S3_BUCKET" --region "$AWS_REGION"
    echo "✅ S3 bucket created: $S3_BUCKET"
else
    echo "✅ S3 bucket already exists: $S3_BUCKET"
fi

# Enable versioning
aws s3api put-bucket-versioning \
    --bucket "$S3_BUCKET" \
    --versioning-configuration Status=Enabled \
    --region "$AWS_REGION"
echo "✅ Versioning enabled"

# Step 2: Create SNS topic for alerts
echo ""
echo "Step 2: Creating SNS topic..."
TOPIC_ARN=$(aws sns create-topic \
    --name energy-grid-alerts \
    --region "$AWS_REGION" \
    --query 'TopicArn' \
    --output text)
echo "✅ SNS Topic created: $TOPIC_ARN"

# Subscribe your email (replace with your email)
read -p "Enter your email for alerts: " EMAIL
aws sns subscribe \
    --topic-arn "$TOPIC_ARN" \
    --protocol email \
    --notification-endpoint "$EMAIL" \
    --region "$AWS_REGION"
echo "✅ Email subscription pending (check your inbox to confirm)"

# Step 3: Create DynamoDB tables
echo ""
echo "Step 3: Creating DynamoDB tables..."

# EnergyReadings table
if ! aws dynamodb describe-table --table-name EnergyReadings --region "$AWS_REGION" 2>/dev/null; then
    aws dynamodb create-table \
        --table-name EnergyReadings \
        --attribute-definitions \
            AttributeName=facilityId,AttributeType=S \
            AttributeName=timestamp,AttributeType=N \
        --key-schema \
            AttributeName=facilityId,KeyType=HASH \
            AttributeName=timestamp,KeyType=RANGE \
        --billing-mode PAY_PER_REQUEST \
        --region "$AWS_REGION"
    echo "✅ EnergyReadings table created"
else
    echo "✅ EnergyReadings table already exists"
fi

# Alerts table
if ! aws dynamodb describe-table --table-name Alerts --region "$AWS_REGION" 2>/dev/null; then
    aws dynamodb create-table \
        --table-name Alerts \
        --attribute-definitions \
            AttributeName=alertId,AttributeType=S \
            AttributeName=facilityId,AttributeType=S \
            AttributeName=timestamp,AttributeType=N \
        --key-schema \
            AttributeName=alertId,KeyType=HASH \
        --global-secondary-indexes \
            "IndexName=facilityId-timestamp-index,\
            KeySchema=[{AttributeName=facilityId,KeyType=HASH},{AttributeName=timestamp,KeyType=RANGE}],\
            Projection={ProjectionType=ALL}" \
        --billing-mode PAY_PER_REQUEST \
        --region "$AWS_REGION"
    echo "✅ Alerts table created"
else
    echo "✅ Alerts table already exists"
fi

# AnalyticsSummaries table
if ! aws dynamodb describe-table --table-name AnalyticsSummaries --region "$AWS_REGION" 2>/dev/null; then
    aws dynamodb create-table \
        --table-name AnalyticsSummaries \
        --attribute-definitions \
            AttributeName=facilityId,AttributeType=S \
            AttributeName=date,AttributeType=S \
        --key-schema \
            AttributeName=facilityId,KeyType=HASH \
            AttributeName=date,KeyType=RANGE \
        --billing-mode PAY_PER_REQUEST \
        --region "$AWS_REGION"
    echo "✅ AnalyticsSummaries table created"
else
    echo "✅ AnalyticsSummaries table already exists"
fi

# Equipment table
if ! aws dynamodb describe-table --table-name Equipment --region "$AWS_REGION" 2>/dev/null; then
    aws dynamodb create-table \
        --table-name Equipment \
        --attribute-definitions \
            AttributeName=equipmentId,AttributeType=S \
            AttributeName=facilityId,AttributeType=S \
        --key-schema \
            AttributeName=equipmentId,KeyType=HASH \
        --global-secondary-indexes \
            "IndexName=facilityId-index,\
            KeySchema=[{AttributeName=facilityId,KeyType=HASH}],\
            Projection={ProjectionType=ALL}" \
        --billing-mode PAY_PER_REQUEST \
        --region "$AWS_REGION"
    echo "✅ Equipment table created"
else
    echo "✅ Equipment table already exists"
fi

echo "⏳ Waiting for tables to be active..."
aws dynamodb wait table-exists --table-name EnergyReadings --region "$AWS_REGION"
aws dynamodb wait table-exists --table-name Alerts --region "$AWS_REGION"
aws dynamodb wait table-exists --table-name AnalyticsSummaries --region "$AWS_REGION"
aws dynamodb wait table-exists --table-name Equipment --region "$AWS_REGION"
echo "✅ All DynamoDB tables are active"

# Step 4: Create S3 bucket for reports
echo ""
echo "Step 4: Creating S3 bucket for reports..."
REPORTS_BUCKET="energy-grid-reports"
if aws s3 ls "s3://$REPORTS_BUCKET" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3 mb "s3://$REPORTS_BUCKET" --region "$AWS_REGION"
    echo "✅ Reports bucket created: $REPORTS_BUCKET"
else
    echo "✅ Reports bucket already exists: $REPORTS_BUCKET"
fi

# Step 5: Check for existing Elastic Beanstalk application
echo ""
echo "Step 5: Setting up Elastic Beanstalk..."

if ! aws elasticbeanstalk describe-applications \
    --application-names "$APP_NAME" \
    --region "$AWS_REGION" 2>/dev/null | grep -q "$APP_NAME"; then
    
    echo "Creating Elastic Beanstalk application..."
    aws elasticbeanstalk create-application \
        --application-name "$APP_NAME" \
        --description "Smart Energy Grid Management System" \
        --region "$AWS_REGION"
    echo "✅ Application created: $APP_NAME"
else
    echo "✅ Application already exists: $APP_NAME"
fi

# Step 6: Check available solution stacks
echo ""
echo "Step 6: Checking available Go solution stacks..."
aws elasticbeanstalk list-available-solution-stacks \
    --region "$AWS_REGION" \
    --query 'SolutionStacks[?contains(@, `Go`)]' \
    --output table

echo ""
echo "Note: Use one of the above solution stacks for your environment"
echo "Default: $SOLUTION_STACK"

# Step 7: Create environment (if it doesn't exist)
echo ""
echo "Step 7: Creating Elastic Beanstalk environment..."

if ! aws elasticbeanstalk describe-environments \
    --application-name "$APP_NAME" \
    --environment-names "$ENV_NAME" \
    --region "$AWS_REGION" 2>/dev/null | grep -q "$ENV_NAME"; then
    
    echo "Creating environment (this may take several minutes)..."
    aws elasticbeanstalk create-environment \
        --application-name "$APP_NAME" \
        --environment-name "$ENV_NAME" \
        --solution-stack-name "$SOLUTION_STACK" \
        --option-settings \
            Namespace=aws:autoscaling:launchconfiguration,OptionName=InstanceType,Value=t3.small \
            Namespace=aws:autoscaling:asg,OptionName=MinSize,Value=1 \
            Namespace=aws:autoscaling:asg,OptionName=MaxSize,Value=4 \
            Namespace=aws:elasticbeanstalk:environment,OptionName=EnvironmentType,Value=LoadBalanced \
            Namespace=aws:elasticbeanstalk:application:environment,OptionName=AWS_REGION,Value="$AWS_REGION" \
            Namespace=aws:elasticbeanstalk:application:environment,OptionName=USE_CLOUD_SERVICES,Value=true \
            Namespace=aws:elasticbeanstalk:application:environment,OptionName=AWS_S3_BUCKET,Value="$REPORTS_BUCKET" \
            Namespace=aws:elasticbeanstalk:application:environment,OptionName=AWS_SNS_TOPIC_ARN,Value="$TOPIC_ARN" \
        --region "$AWS_REGION"
    
    echo "⏳ Waiting for environment to be ready (this can take 5-10 minutes)..."
    aws elasticbeanstalk wait environment-exists \
        --application-name "$APP_NAME" \
        --environment-names "$ENV_NAME" \
        --region "$AWS_REGION"
    
    echo "✅ Environment created: $ENV_NAME"
else
    echo "✅ Environment already exists: $ENV_NAME"
fi

# Step 8: Get environment info
echo ""
echo "Step 8: Environment information..."
ENV_INFO=$(aws elasticbeanstalk describe-environments \
    --application-name "$APP_NAME" \
    --environment-names "$ENV_NAME" \
    --region "$AWS_REGION" \
    --query 'Environments[0].[CNAME,Health,Status]' \
    --output text)

echo "Environment URL: $(echo $ENV_INFO | awk '{print $1}')"
echo "Health: $(echo $ENV_INFO | awk '{print $2}')"
echo "Status: $(echo $ENV_INFO | awk '{print $3}')"

# Summary
echo ""
echo "=========================================="
echo "✅ AWS Infrastructure Setup Complete!"
echo "=========================================="
echo ""
echo "Next Steps:"
echo "1. Add these secrets to GitHub repository:"
echo "   - AWS_ACCESS_KEY_ID"
echo "   - AWS_SECRET_ACCESS_KEY"
echo "   - AWS_SESSION_TOKEN (if using temporary credentials)"
echo ""
echo "2. Update these values in .github/workflows/deploy-eb.yml:"
echo "   - EB_APPLICATION_NAME: $APP_NAME"
echo "   - EB_ENVIRONMENT_NAME: $ENV_NAME"
echo "   - S3_BUCKET: $S3_BUCKET"
echo ""
echo "3. Update .ebextensions/01_environment.config with:"
echo "   - AWS_SNS_TOPIC_ARN: $TOPIC_ARN"
echo "   - AWS_S3_BUCKET: $REPORTS_BUCKET"
echo ""
echo "4. Confirm your email subscription in AWS SNS"
echo ""
echo "5. Push to main branch to trigger deployment"
echo ""
echo "Resources created:"
echo "  - S3 Bucket (deployments): $S3_BUCKET"
echo "  - S3 Bucket (reports): $REPORTS_BUCKET"
echo "  - SNS Topic: $TOPIC_ARN"
echo "  - DynamoDB Tables: EnergyReadings, Alerts, AnalyticsSummaries, Equipment"
echo "  - EB Application: $APP_NAME"
echo "  - EB Environment: $ENV_NAME"
echo "=========================================="