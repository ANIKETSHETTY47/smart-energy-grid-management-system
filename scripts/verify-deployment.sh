#!/bin/bash

# Deployment Verification Script
# Verifies that all AWS resources are properly configured

set -e

echo "🔍 AWS Resources Verification"
echo "=============================="
echo ""

# Check AWS credentials
echo "1. Checking AWS Credentials..."
if aws sts get-caller-identity --region eu-north-1 > /dev/null 2>&1; then
    echo "   ✅ AWS credentials are valid"
    aws sts get-caller-identity --region eu-north-1
else
    echo "   ❌ AWS credentials are invalid or expired"
    exit 1
fi
echo ""

# Check Elastic Beanstalk applications
echo "2. Checking Elastic Beanstalk Applications..."
EB_APPS=$(aws elasticbeanstalk describe-applications --region eu-north-1 --query 'Applications[*].ApplicationName' --output text)
if echo "$EB_APPS" | grep -q "smart-energy-grid"; then
    echo "   ✅ Backend application found: smart-energy-grid"
else
    echo "   ⚠️  Backend application not found: smart-energy-grid"
fi

if echo "$EB_APPS" | grep -q "energy-dashboard-frontend"; then
    echo "   ✅ Frontend application found: energy-dashboard-frontend"
else
    echo "   ⚠️  Frontend application not found: energy-dashboard-frontend"
fi
echo ""

# Check Elastic Beanstalk environments
echo "3. Checking Elastic Beanstalk Environments..."
EB_ENVS=$(aws elasticbeanstalk describe-environments --region eu-north-1 --query 'Environments[?Status!=`Terminated`].[EnvironmentName,Status,Health,CNAME]' --output table)
if [ -n "$EB_ENVS" ]; then
    echo "   Active Environments:"
    echo "$EB_ENVS"
else
    echo "   ⚠️  No active environments found"
fi
echo ""

# Check DynamoDB tables
echo "4. Checking DynamoDB Tables..."
EXPECTED_TABLES=("EnergyReadings" "Alerts" "Equipment" "AnalyticsSummaries")
for table in "${EXPECTED_TABLES[@]}"; do
    if aws dynamodb describe-table --table-name "$table" --region eu-north-1 > /dev/null 2>&1; then
        echo "   ✅ Table exists: $table"
    else
        echo "   ⚠️  Table not found: $table"
    fi
done
echo ""

# Check Lambda functions
echo "5. Checking Lambda Functions..."
EXPECTED_LAMBDAS=("analytics-processing" "anomaly-detection")
for func in "${EXPECTED_LAMBDAS[@]}"; do
    if aws lambda get-function --function-name "$func" --region eu-north-1 > /dev/null 2>&1; then
        echo "   ✅ Function exists: $func"
    else
        echo "   ⚠️  Function not found: $func"
    fi
done
echo ""

# Check S3 bucket
echo "6. Checking S3 Bucket..."
if aws s3 ls --region eu-north-1 | grep -q "energy-grid"; then
    echo "   ✅ S3 bucket for energy grid data found"
else
    echo "   ⚠️  S3 bucket not found"
fi
echo ""

# Check SNS topics
echo "7. Checking SNS Topics..."
SNS_TOPICS=$(aws sns list-topics --region eu-north-1 --query 'Topics[*].TopicArn' --output text)
if echo "$SNS_TOPICS" | grep -q "alert"; then
    echo "   ✅ Alert SNS topic found"
    echo "$SNS_TOPICS" | grep "alert"
else
    echo "   ⚠️  Alert SNS topic not found"
fi
echo ""

echo "=============================="
echo "Verification Complete!"
echo "=============================="
