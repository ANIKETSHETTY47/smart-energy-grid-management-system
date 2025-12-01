#!/bin/bash

echo "🔄 Populating DynamoDB tables in eu-north-1 region..."
echo ""

# Check if credentials are set
if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "❌ AWS credentials not set!"
    echo ""
    echo "Please run:"
    echo "export AWS_ACCESS_KEY_ID=\"your-key\""
    echo "export AWS_SECRET_ACCESS_KEY=\"your-secret\""
    exit 1
fi

# Set region to eu-north-1
export AWS_REGION="eu-north-1"

echo "📍 Region: eu-north-1 (Stockholm)"
echo ""

# Navigate to scripts directory
cd "$(dirname "$0")"

# Check if tables exist in eu-north-1
echo "🔍 Checking for tables in eu-north-1..."
tables=$(aws dynamodb list-tables --region eu-north-1 --query 'TableNames[*]' --output text 2>/dev/null)

if [ -z "$tables" ]; then
    echo "⚠️  No tables found in eu-north-1"
    echo ""
    echo "You need to create these tables first:"
    echo "  - EnergyReadings"
    echo "  - Alerts"  
    echo "  - Equipment"
    echo "  - AnalyticsSummaries"
    echo ""
    echo "Would you like me to create them? (This will use the Terraform/CloudFormation scripts)"
    exit 1
fi

echo "✓ Found tables in eu-north-1:"
echo "$tables"
echo ""

# Run the population script
echo "🚀 Running population script..."
cd /Users/shetty/Desktop/Sem\ 1\ Projects/Cloud\ Progm/smart-energy-grid-management-system/scripts

go run populate-dynamodb-simple.go

echo ""
echo "✅ Done! Check your dashboard now."
