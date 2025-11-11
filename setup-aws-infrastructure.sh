#!/bin/bash
set -e

AWS_REGION="us-east-1"
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

echo "=========================================="
echo "Smart Energy Grid - AWS Infrastructure"
echo "Account: $ACCOUNT_ID"
echo "Region: $AWS_REGION"
echo "=========================================="

# 1. Create S3 Buckets
echo "Creating S3 buckets..."
aws s3 mb s3://energy-grid-reports --region $AWS_REGION || echo "Bucket exists"
aws s3 mb s3://smart-energy-grid-deployments --region $AWS_REGION || echo "Bucket exists"

# 2. Create SNS Topic
echo "Creating SNS topic..."
TOPIC_ARN=$(aws sns create-topic --name energy-grid-alerts --region $AWS_REGION --query TopicArn --output text)
echo "SNS Topic: $TOPIC_ARN"

# 3. Create DynamoDB Tables
echo "Creating DynamoDB tables..."

# EnergyReadings
aws dynamodb create-table \
  --table-name EnergyReadings \
  --attribute-definitions \
    AttributeName=facilityId,AttributeType=S \
    AttributeName=timestamp,AttributeType=N \
  --key-schema \
    AttributeName=facilityId,KeyType=HASH \
    AttributeName=timestamp,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --region $AWS_REGION 2>/dev/null || echo "Table exists"

# Alerts
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
  --region $AWS_REGION 2>/dev/null || echo "Table exists"

# AnalyticsSummaries
aws dynamodb create-table \
  --table-name AnalyticsSummaries \
  --attribute-definitions \
    AttributeName=facilityId,AttributeType=S \
    AttributeName=date,AttributeType=S \
  --key-schema \
    AttributeName=facilityId,KeyType=HASH \
    AttributeName=date,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --region $AWS_REGION 2>/dev/null || echo "Table exists"

# Equipment
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
  --region $AWS_REGION 2>/dev/null || echo "Table exists"

echo "Waiting for tables..."
aws dynamodb wait table-exists --table-name EnergyReadings --region $AWS_REGION
aws dynamodb wait table-exists --table-name Alerts --region $AWS_REGION

# 4. Deploy Lambda Functions
echo "Deploying Lambda functions..."

# Anomaly Detection Lambda
cd lambda-functions/anomaly-detection
make build
aws lambda create-function \
  --function-name anomaly-detection \
  --runtime provided.al2 \
  --role arn:aws:iam::${ACCOUNT_ID}:role/aws-elasticbeanstalk-ec2-role \
  --handler bootstrap \
  --zip-file fileb://function.zip \
  --environment "Variables={AWS_REGION=${AWS_REGION},SNS_TOPIC_ARN=${TOPIC_ARN}}" \
  --region $AWS_REGION 2>/dev/null || \
aws lambda update-function-code \
  --function-name anomaly-detection \
  --zip-file fileb://function.zip \
  --region $AWS_REGION
cd ../..

# Analytics Processing Lambda
cd lambda-functions/analytics-processing
make build
aws lambda create-function \
  --function-name analytics-processing \
  --runtime provided.al2 \
  --role arn:aws:iam::${ACCOUNT_ID}:role/aws-elasticbeanstalk-ec2-role \
  --handler bootstrap \
  --zip-file fileb://function.zip \
  --timeout 60 \
  --memory-size 512 \
  --environment "Variables={AWS_REGION=${AWS_REGION},S3_BUCKET=energy-grid-reports}" \
  --region $AWS_REGION 2>/dev/null || \
aws lambda update-function-code \
  --function-name analytics-processing \
  --zip-file fileb://function.zip \
  --region $AWS_REGION
cd ../..

echo "=========================================="
echo "✅ Infrastructure Setup Complete!"
echo "=========================================="
echo ""
echo "Add these to GitHub Secrets:"
echo "AWS_ACCESS_KEY_ID: (from IAM user)"
echo "AWS_SECRET_ACCESS_KEY: (from IAM user)"
echo ""
echo "SNS Topic ARN: $TOPIC_ARN"
echo "=========================================="
