#!/bin/bash

echo "🔍 DynamoDB Region Diagnostic Tool"
echo "=================================="
echo ""

# Check credentials
if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "❌ Please set AWS credentials first:"
    echo ""
    echo "export AWS_ACCESS_KEY_ID=\"AKIAVK3PK4IJAZ2HKIPA\""
    echo "export AWS_SECRET_ACCESS_KEY=\"your-secret\""
    echo ""
    exit 1
fi

echo "📍 Checking DynamoDB tables in both regions..."
echo ""

# Check us-east-1
echo "Region: us-east-1 (N. Virginia)"
echo "--------------------------------"
us_east_tables=$(aws dynamodb list-tables --region us-east-1 --query 'TableNames' --output json 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "$us_east_tables" | jq -r '.[]' | while read table; do
        # Get item count
        count=$(aws dynamodb scan --table-name "$table" --region us-east-1 --select "COUNT" --query 'Count' --output text 2>/dev/null)
        echo "  ✓ $table ($count items)"
    done
else
    echo "  ❌ No access or no tables"
fi
echo ""

# Check eu-north-1  
echo "Region: eu-north-1 (Stockholm)"
echo "--------------------------------"
eu_north_tables=$(aws dynamodb list-tables --region eu-north-1 --query 'TableNames' --output json 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "$eu_north_tables" | jq -r '.[]' | while read table; do
        # Get item count
        count=$(aws dynamodb scan --table-name "$table" --region eu-north-1 --select "COUNT" --query 'Count' --output text 2>/dev/null)
        echo "  ✓ $table ($count items)"
    done
else
    echo "  ❌ No access or no tables"
fi
echo ""

echo "📋 Application Configuration"
echo "--------------------------------"
echo "Current AWS_REGION in .env: eu-north-1"
echo ""
echo "Expected table names:"
echo "  - EnergyReadings"
echo "  - Alerts"
echo "  - Equipment"  
echo "  - AnalyticsSummaries"
echo ""

echo "💡 Recommendations:"
echo "--------------------------------"
echo ""
echo "Option 1: Use us-east-1 (where data already exists)"
echo "  → Change .env: AWS_REGION=us-east-1"
echo "  → Restart your application"
echo ""
echo "Option 2: Populate eu-north-1 (match current config)"
echo "  → Run: export AWS_REGION=eu-north-1"
echo "  → Run: go run scripts/populate-dynamodb-simple.go"
echo ""
