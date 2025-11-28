#!/bin/bash

# GitHub Secrets Verification Script
# This script helps verify that GitHub secrets are properly configured

echo "🔐 GitHub Secrets Verification Guide"
echo "======================================"
echo ""
echo "Please ensure the following secrets are set in your GitHub repository:"
echo "Repository → Settings → Secrets and variables → Actions → Repository secrets"
echo ""
echo "Required Secrets:"
echo "----------------"
echo "1. AWS_ACCESS_KEY_ID"
echo "   Value: AKIAVK3PK4IJAZ2HKIPA"
echo ""
echo "2. AWS_SECRET_ACCESS_KEY"
echo "   Value: 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"
echo ""
echo "To set secrets via GitHub CLI (if installed):"
echo "-----------------------------------------------"
echo "gh secret set AWS_ACCESS_KEY_ID -b 'AKIAVK3PK4IJAZ2HKIPA'"
echo "gh secret set AWS_SECRET_ACCESS_KEY -b '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'"
echo ""
echo "To set secrets via GitHub Web UI:"
echo "----------------------------------"
echo "1. Go to: https://github.com/ANIKETSHETTY47/smart-energy-grid-management-system/settings/secrets/actions"
echo "2. Click 'New repository secret'"
echo "3. Add each secret with the name and value shown above"
echo ""
echo "✅ After setting secrets, push your code to trigger the CI/CD pipeline"
echo ""

# Check if gh CLI is installed
if command -v gh &> /dev/null; then
    echo "GitHub CLI detected. Would you like to set secrets now? (y/n)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        gh secret set AWS_ACCESS_KEY_ID -b 'AKIAVK3PK4IJAZ2HKIPA'
        gh secret set AWS_SECRET_ACCESS_KEY -b '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'
        echo "✅ Secrets set successfully!"
    fi
else
    echo "ℹ️  GitHub CLI not found. Please set secrets manually via GitHub web UI."
fi
