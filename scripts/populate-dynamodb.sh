#!/bin/bash

# Script to populate DynamoDB tables with dummy data

echo "🚀 Starting DynamoDB population script..."
echo ""

# Check if AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ Error: AWS credentials not configured"
    echo "Please run: aws configure"
    exit 1
fi

echo "✓ AWS credentials verified"
echo ""

# Navigate to scripts directory
cd "$(dirname "$0")"

echo "📦 Installing dependencies..."
go mod init populate-dynamodb 2>/dev/null || true
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue

echo ""
echo "🔄 Running population script..."
echo "This will create:"
echo "  - ~500 energy readings (7 days of data)"
echo "  - 50 alerts"
echo "  - 5 equipment records"
echo "  - 60 analytics summaries (30 days)"
echo ""

go run populate-dynamodb.go

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ SUCCESS! All DynamoDB tables populated"
    echo ""
    echo "You can now view the data in AWS Console:"
    echo "https://eu-north-1.console.aws.amazon.com/dynamodbv2/home?region=eu-north-1#tables"
else
    echo ""
    echo "❌ Error occurred during population"
    exit 1
fi
