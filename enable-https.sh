#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

DOMAIN="dashboard.aniketshetty.me"
REGION="eu-north-1"
ACM_REGION="us-east-1"
APP_NAME="energy-dashboard"
ENV_NAME="energy-dashboard-env"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}HTTPS Setup for $DOMAIN${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Step 1: Request Certificate
echo -e "${BLUE}Step 1: Requesting SSL Certificate...${NC}"
CERT_ARN=$(/usr/local/bin/aws acm request-certificate \
  --domain-name $DOMAIN \
  --validation-method DNS \
  --region $ACM_REGION \
  --query 'CertificateArn' \
  --output text)

echo -e "${GREEN}✓ Certificate requested${NC}"
echo -e "  ARN: ${YELLOW}$CERT_ARN${NC}\n"

# Step 2: Get validation details
echo -e "${BLUE}Step 2: Getting DNS validation records...${NC}"
sleep 5  # Wait for AWS to process

VALIDATION=$(/usr/local/bin/aws acm describe-certificate \
  --certificate-arn $CERT_ARN \
  --region $ACM_REGION \
  --query 'Certificate.DomainValidationOptions[0].ResourceRecord')

VALIDATION_NAME=$(echo $VALIDATION | /usr/bin/python3 -c "import sys, json; print(json.load(sys.stdin)['Name'])")
VALIDATION_VALUE=$(echo $VALIDATION | /usr/bin/python3 -c "import sys, json; print(json.load(sys.stdin)['Value'])")

echo -e "${GREEN}✓ Validation record:${NC}"
echo -e "  Name:  ${YELLOW}$VALIDATION_NAME${NC}"
echo -e "  Type:  ${YELLOW}CNAME${NC}"
echo -e "  Value: ${YELLOW}$VALIDATION_VALUE${NC}\n"

# Step 3: Instructions for Namecheap
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}ACTION REQUIRED: Add DNS Record${NC}"
echo -e "${YELLOW}========================================${NC}\n"

echo -e "${BLUE}Go to Namecheap:${NC}"
echo -e "  1. Login: https://www.namecheap.com/myaccount/login"
echo -e "  2. Domain List → Manage aniketshetty.me"
echo -e "  3. Advanced DNS tab"
echo -e "  4. Add New Record:\n"

# Extract just the subdomain part
VALIDATION_HOST=$(echo $VALIDATION_NAME | sed 's/.aniketshetty.me.*//')
echo -e "     ${GREEN}Type:${NC}  CNAME Record"
echo -e "     ${GREEN}Host:${NC}  ${YELLOW}$VALIDATION_HOST${NC}"
echo -e "     ${GREEN}Value:${NC} ${YELLOW}$VALIDATION_VALUE${NC}"
echo -e "     ${GREEN}TTL:${NC}   Automatic\n"

echo -e "${BLUE}Press Enter after adding the DNS record...${NC}"
read

# Step 4: Wait for validation
echo -e "\n${BLUE}Step 3: Waiting for certificate validation...${NC}"
echo -e "${YELLOW}This may take 5-30 minutes...${NC}\n"

WAIT_COUNT=0
MAX_WAIT=60
while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    STATUS=$(/usr/local/bin/aws acm describe-certificate \
      --certificate-arn $CERT_ARN \
      --region $ACM_REGION \
      --query 'Certificate.Status' \
      --output text)
    
    if [ "$STATUS" = "ISSUED" ]; then
        echo -e "\n${GREEN}✓ Certificate validated and issued!${NC}\n"
        break
    fi
    
    echo -ne "\r  Status: $STATUS | Waiting... ($WAIT_COUNT/$MAX_WAIT)"
    sleep 30
    WAIT_COUNT=$((WAIT_COUNT + 1))
done

if [ "$STATUS" != "ISSUED" ]; then
    echo -e "\n${RED}✗ Certificate validation timed out${NC}"
    echo -e "${YELLOW}Check DNS record and try again later${NC}"
    exit 1
fi

# Step 5: Get current EB URL
echo -e "${BLUE}Step 4: Getting Elastic Beanstalk URL...${NC}"
EB_URL=$(/usr/local/bin/aws elasticbeanstalk describe-environments \
  --application-name $APP_NAME \
  --environment-names $ENV_NAME \
  --region $REGION \
  --query 'Environments[0].CNAME' \
  --output text)

echo -e "${GREEN}✓ EB URL: $EB_URL${NC}\n"

# Step 6: Create CloudFront distribution
echo -e "${BLUE}Step 5: Creating CloudFront distribution...${NC}"
echo -e "${YELLOW}This takes 10-15 minutes...${NC}\n"

cat > /tmp/cf-config.json << EOF
{
  "CallerReference": "$(date +%s)",
  "Comment": "Energy Dashboard HTTPS",
  "Enabled": true,
  "Origins": {
    "Quantity": 1,
    "Items": [{
      "Id": "eb-origin",
      "DomainName": "$EB_URL",
      "CustomOriginConfig": {
        "HTTPPort": 80,
        "HTTPSPort": 443,
        "OriginProtocolPolicy": "http-only"
      }
    }]
  },
  "DefaultCacheBehavior": {
    "TargetOriginId": "eb-origin",
    "ViewerProtocolPolicy": "redirect-to-https",
    "AllowedMethods": {
      "Quantity": 7,
      "Items": ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"],
      "CachedMethods": {
        "Quantity": 2,
        "Items": ["GET", "HEAD"]
      }
    },
    "ForwardedValues": {
      "QueryString": true,
      "Cookies": {"Forward": "all"},
      "Headers": {
        "Quantity": 1,
        "Items": ["Host"]
      }
    },
    "MinTTL": 0,
    "DefaultTTL": 0,
    "MaxTTL": 31536000,
    "Compress": true
  },
  "Aliases": {
    "Quantity": 1,
    "Items": ["$DOMAIN"]
  },
  "ViewerCertificate": {
    "ACMCertificateArn": "$CERT_ARN",
    "SSLSupportMethod": "sni-only",
    "MinimumProtocolVersion": "TLSv1.2_2021"
  },
  "PriceClass": "PriceClass_100"
}
EOF

DIST_ID=$(/usr/local/bin/aws cloudfront create-distribution \
  --distribution-config file:///tmp/cf-config.json \
  --query 'Distribution.Id' \
  --output text)

echo -e "${GREEN}✓ CloudFront distribution created${NC}"
echo -e "  Distribution ID: ${YELLOW}$DIST_ID${NC}\n"

# Get CloudFront domain
CF_DOMAIN=$(/usr/local/bin/aws cloudfront get-distribution \
  --id $DIST_ID \
  --query 'Distribution.DomainName' \
  --output text)

echo -e "${GREEN}✓ CloudFront domain: $CF_DOMAIN${NC}\n"

# Step 7: Update DNS
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}FINAL STEP: Update DNS Record${NC}"
echo -e "${YELLOW}========================================${NC}\n"

echo -e "${BLUE}Go to Namecheap again:${NC}"
echo -e "  1. Find your existing CNAME record for 'dashboard'"
echo -e "  2. Edit it (pencil icon)"
echo -e "  3. Change Value to: ${YELLOW}$CF_DOMAIN${NC}"
echo -e "  4. Save\n"

echo -e "${GREEN}After DNS propagates (5-30 min), access:${NC}"
echo -e "  ${YELLOW}https://dashboard.aniketshetty.me${NC}\n"

echo -e "${BLUE}To check CloudFront deployment status:${NC}"
echo -e "  /usr/local/bin/aws cloudfront get-distribution --id $DIST_ID --query 'Distribution.Status'\n"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}HTTPS Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Summary:${NC}"
echo -e "  Certificate ARN: ${YELLOW}$CERT_ARN${NC}"
echo -e "  CloudFront ID: ${YELLOW}$DIST_ID${NC}"
echo -e "  CloudFront Domain: ${YELLOW}$CF_DOMAIN${NC}"
echo -e "  Your Domain: ${YELLOW}https://$DOMAIN${NC}\n"

echo -e "${YELLOW}Next Steps:${NC}"
echo -e "  1. Update Namecheap CNAME to point to CloudFront"
echo -e "  2. Wait for DNS propagation (5-30 minutes)"
echo -e "  3. Test: https://dashboard.aniketshetty.me"
echo -e "  4. Take screenshots for your report!\n"
