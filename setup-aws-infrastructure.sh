#!/bin/bash

# AWS Infrastructure Setup Script for Smart Energy Grid Management System
# This script creates all necessary AWS resources in eu-north-1 region

set -e  # Exit on error

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
REGION="eu-north-1"
AWS_ACCOUNT_ID="366916330002"
S3_BUCKET="smart-energy-grid-reports-nci"
DYNAMODB_TABLE_READINGS="energy-grid-readings"
DYNAMODB_TABLE_ALERTS="energy-grid-alerts"
DYNAMODB_TABLE_EQUIPMENT="energy-grid-equipment"
SNS_TOPIC_NAME="energy-grid-alerts"
LAMBDA_ROLE_NAME="EnergyGridLambdaRole"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Smart Energy Grid AWS Infrastructure Setup${NC}"
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

# Check AWS CLI installation
AWS_CLI="/usr/local/bin/aws"
if ! command -v $AWS_CLI &> /dev/null; then
    if [ -f "/usr/local/bin/aws" ]; then
        AWS_CLI="/usr/local/bin/aws"
    else
        print_error "AWS CLI not found. Please install it first."
        exit 1
    fi
fi

print_status "AWS CLI found"

# Configure AWS region
export AWS_DEFAULT_REGION=$REGION
print_info "Using AWS Region: $REGION"

echo -e "\n${BLUE}Step 1: Creating S3 Bucket${NC}"
if aws s3 ls "s3://$S3_BUCKET" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3 mb "s3://$S3_BUCKET" --region $REGION
    
    # Enable versioning
    aws s3api put-bucket-versioning \
        --bucket $S3_BUCKET \
        --versioning-configuration Status=Enabled
    
    # Add bucket policy for public read of reports (optional)
    # Uncomment if you want reports to be publicly accessible
    # cat > /tmp/bucket-policy.json <<EOF
    # {
    #     "Version": "2012-10-17",
    #     "Statement": [{
    #         "Sid": "PublicReadGetObject",
    #         "Effect": "Allow",
    #         "Principal": "*",
    #         "Action": "s3:GetObject",
    #         "Resource": "arn:aws:s3:::$S3_BUCKET/reports/*"
    #     }]
    # }
    # EOF
    # aws s3api put-bucket-policy --bucket $S3_BUCKET --policy file:///tmp/bucket-policy.json
    
    print_status "S3 Bucket created: $S3_BUCKET"
else
    print_warning "S3 Bucket already exists: $S3_BUCKET"
fi

echo -e "\n${BLUE}Step 2: Creating DynamoDB Tables${NC}"

# Create energy-grid-readings table
print_info "Creating DynamoDB table: $DYNAMODB_TABLE_READINGS"
if ! aws dynamodb describe-table --table-name $DYNAMODB_TABLE_READINGS --region $REGION 2>&1 | grep -q "TableName"; then
    aws dynamodb create-table \
        --table-name $DYNAMODB_TABLE_READINGS \
        --attribute-definitions \
            AttributeName=node_id,AttributeType=S \
            AttributeName=timestamp,AttributeType=N \
        --key-schema \
            AttributeName=node_id,KeyType=HASH \
            AttributeName=timestamp,KeyType=RANGE \
        --provisioned-throughput \
            ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region $REGION
    
    aws dynamodb wait table-exists --table-name $DYNAMODB_TABLE_READINGS --region $REGION
    print_status "Table created: $DYNAMODB_TABLE_READINGS"
else
    print_warning "Table already exists: $DYNAMODB_TABLE_READINGS"
fi

# Create energy-grid-alerts table
print_info "Creating DynamoDB table: $DYNAMODB_TABLE_ALERTS"
if ! aws dynamodb describe-table --table-name $DYNAMODB_TABLE_ALERTS --region $REGION 2>&1 | grep -q "TableName"; then
    aws dynamodb create-table \
        --table-name $DYNAMODB_TABLE_ALERTS \
        --attribute-definitions \
            AttributeName=alert_id,AttributeType=S \
            AttributeName=timestamp,AttributeType=N \
        --key-schema \
            AttributeName=alert_id,KeyType=HASH \
            AttributeName=timestamp,KeyType=RANGE \
        --provisioned-throughput \
            ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region $REGION
    
    aws dynamodb wait table-exists --table-name $DYNAMODB_TABLE_ALERTS --region $REGION
    print_status "Table created: $DYNAMODB_TABLE_ALERTS"
else
    print_warning "Table already exists: $DYNAMODB_TABLE_ALERTS"
fi

# Create energy-grid-equipment table
print_info "Creating DynamoDB table: $DYNAMODB_TABLE_EQUIPMENT"
if ! aws dynamodb describe-table --table-name $DYNAMODB_TABLE_EQUIPMENT --region $REGION 2>&1 | grep -q "TableName"; then
    aws dynamodb create-table \
        --table-name $DYNAMODB_TABLE_EQUIPMENT \
        --attribute-definitions \
            AttributeName=equipment_id,AttributeType=S \
        --key-schema \
            AttributeName=equipment_id,KeyType=HASH \
        --provisioned-throughput \
            ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region $REGION
    
    aws dynamodb wait table-exists --table-name $DYNAMODB_TABLE_EQUIPMENT --region $REGION
    print_status "Table created: $DYNAMODB_TABLE_EQUIPMENT"
else
    print_warning "Table already exists: $DYNAMODB_TABLE_EQUIPMENT"
fi

echo -e "\n${BLUE}Step 3: Creating SNS Topic${NC}"
SNS_TOPIC_ARN=$(aws sns create-topic --name $SNS_TOPIC_NAME --region $REGION --query 'TopicArn' --output text 2>/dev/null || \
                aws sns list-topics --region $REGION --query "Topics[?contains(TopicArn, '$SNS_TOPIC_NAME')].TopicArn" --output text)
print_status "SNS Topic ARN: $SNS_TOPIC_ARN"

echo -e "\n${BLUE}Step 4: Creating IAM Role for Lambda Functions${NC}"
if ! aws iam get-role --role-name $LAMBDA_ROLE_NAME 2>&1 | grep -q "RoleName"; then
    # Create trust policy
    cat > /tmp/lambda-trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF

    # Create role
    aws iam create-role \
        --role-name $LAMBDA_ROLE_NAME \
        --assume-role-policy-document file:///tmp/lambda-trust-policy.json
    
    # Attach policies
    aws iam attach-role-policy \
        --role-name $LAMBDA_ROLE_NAME \
        --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
    
    aws iam attach-role-policy \
        --role-name $LAMBDA_ROLE_NAME \
        --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess
    
    aws iam attach-role-policy \
        --role-name $LAMBDA_ROLE_NAME \
        --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
    
    aws iam attach-role-policy \
        --role-name $LAMBDA_ROLE_NAME \
        --policy-arn arn:aws:iam::aws:policy/AmazonSNSFullAccess
    
    print_status "IAM Role created: $LAMBDA_ROLE_NAME"
    print_info "Waiting 10 seconds for IAM role to propagate..."
    sleep 10
else
    print_warning "IAM Role already exists: $LAMBDA_ROLE_NAME"
fi

LAMBDA_ROLE_ARN="arn:aws:iam::$AWS_ACCOUNT_ID:role/$LAMBDA_ROLE_NAME"

echo -e "\n${BLUE}Step 5: Deploying Lambda Functions${NC}"

# Build and deploy anomaly-detection Lambda
print_info "Building anomaly-detection Lambda function..."
cd lambda-functions/anomaly-detection
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go
zip -q function.zip bootstrap
print_status "Built anomaly-detection Lambda"

if aws lambda get-function --function-name anomaly-detection --region $REGION 2>&1 | grep -q "FunctionName"; then
    print_info "Updating existing Lambda function: anomaly-detection"
    aws lambda update-function-code \
        --function-name anomaly-detection \
        --zip-file fileb://function.zip \
        --region $REGION
else
    print_info "Creating Lambda function: anomaly-detection"
    aws lambda create-function \
        --function-name anomaly-detection \
        --runtime provided.al2023 \
        --role $LAMBDA_ROLE_ARN \
        --handler bootstrap \
        --zip-file fileb://function.zip \
        --timeout 30 \
        --memory-size 256 \
        --environment "Variables={REGION=$REGION,DYNAMODB_TABLE_READINGS=$DYNAMODB_TABLE_READINGS,DYNAMODB_TABLE_ALERTS=$DYNAMODB_TABLE_ALERTS,SNS_TOPIC_ARN=$SNS_TOPIC_ARN}" \
        --region $REGION
fi
print_status "Deployed anomaly-detection Lambda"
cd ../..

# Build and deploy analytics-processing Lambda
print_info "Building analytics-processing Lambda function..."
cd lambda-functions/analytics-processing
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap main.go
zip -q function.zip bootstrap
print_status "Built analytics-processing Lambda"

if aws lambda get-function --function-name analytics-processing --region $REGION 2>&1 | grep -q "FunctionName"; then
    print_info "Updating existing Lambda function: analytics-processing"
    aws lambda update-function-code \
        --function-name analytics-processing \
        --zip-file fileb://function.zip \
        --region $REGION
else
    print_info "Creating Lambda function: analytics-processing"
    aws lambda create-function \
        --function-name analytics-processing \
        --runtime provided.al2023 \
        --role $LAMBDA_ROLE_ARN \
        --handler bootstrap \
        --zip-file fileb://function.zip \
        --timeout 60 \
        --memory-size 512 \
        --environment "Variables={REGION=$REGION,DYNAMODB_TABLE_READINGS=$DYNAMODB_TABLE_READINGS,S3_BUCKET=$S3_BUCKET}" \
        --region $REGION
fi
print_status "Deployed analytics-processing Lambda"
cd ../..

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}AWS Infrastructure Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Resource Summary:${NC}"
echo -e "  S3 Bucket: ${GREEN}$S3_BUCKET${NC}"
echo -e "  DynamoDB Tables:"
echo -e "    - ${GREEN}$DYNAMODB_TABLE_READINGS${NC}"
echo -e "    - ${GREEN}$DYNAMODB_TABLE_ALERTS${NC}"
echo -e "    - ${GREEN}$DYNAMODB_TABLE_EQUIPMENT${NC}"
echo -e "  SNS Topic: ${GREEN}$SNS_TOPIC_ARN${NC}"
echo -e "  Lambda Functions:"
echo -e "    - ${GREEN}anomaly-detection${NC}"
echo -e "    - ${GREEN}analytics-processing${NC}"
echo -e "  IAM Role: ${GREEN}$LAMBDA_ROLE_NAME${NC}\n"

echo -e "${BLUE}Next Steps:${NC}"
echo -e "  1. Update your .env file with:"
echo -e "     ${YELLOW}AWS_SNS_TOPIC_ARN=$SNS_TOPIC_ARN${NC}"
echo -e "  2. Test the Lambda functions"
echo -e "  3. Deploy your main application to Elastic Beanstalk"

print_status "Setup script completed successfully!"
