# 🚀 Complete Deployment Guide - Smart Energy Grid Management System

## 📋 Prerequisites Checklist

Before starting, ensure you have:

- [ ] AWS Account with credentials configured
- [ ] AWS CLI installed (`aws --version`)
- [ ] Go 1.21+ installed (`go version`)
- [ ] Git installed (`git --version`)
- [ ] Terminal access (bash/zsh)

---

## 🎯 What This Deployment Does

Your project uses **6 AWS Services** (meets 5+ requirement):

1. **DynamoDB** - NoSQL database for readings, alerts, equipment
2. **Lambda** - Serverless functions for anomaly detection & analytics
3. **S3** - Object storage for reports and analytics files
4. **SNS** - Simple Notification Service for alerts
5. **Elastic Beanstalk** - Platform-as-a-Service for hosting API
6. **CloudWatch** - Monitoring, logging, and metrics

---

## 📁 Current Project Structure

```
smart-energy-grid-management-system/
├── cmd/api/main.go              # Backend API entry point
├── lambda-functions/
│   ├── anomaly-detection/       # Lambda 1: Detects anomalies
│   └── analytics-processing/    # Lambda 2: Processes analytics
├── internal/
│   ├── cloud/                   # AWS SDK integrations
│   ├── service/                 # Business logic
│   └── http/                    # API handlers
├── .env                         # AWS credentials & config
├── Procfile                     # EB deployment config
├── .ebextensions/               # EB environment variables
└── setup-aws-infrastructure.sh  # Infrastructure creation script
```

---

## 🔥 DEPLOYMENT STEPS (Follow Exactly)

### **STEP 1: Verify AWS Credentials**

```bash
# Navigate to project directory
cd "/Users/shetty/Desktop/Sem 1 Projects/Cloud Progm/smart-energy-grid-management-system"

# Check AWS credentials are configured
aws sts get-caller-identity

# Should show:
# - Account: 366916330002
# - Region: eu-north-1
```

**If this fails:** Run `aws configure` and enter your credentials from `.env` file

---

### **STEP 2: Run Complete Deployment (Automated)**

This single command does everything:

```bash
# Make script executable and run
chmod +x deploy-all.sh
./deploy-all.sh
```

**What this script does:**
1. ✅ Creates S3 bucket
2. ✅ Creates 3 DynamoDB tables
3. ✅ Creates SNS topic
4. ✅ Creates IAM roles
5. ✅ Builds & deploys 2 Lambda functions
6. ✅ Builds Go application
7. ✅ Creates Elastic Beanstalk application
8. ✅ Deploys to EB environment
9. ✅ Provides public URL

**Expected Duration:** 10-15 minutes total

---

### **STEP 3: Verify Deployment**

After script completes, test your deployment:

```bash
# Get your application URL (shown at end of script)
# Example URL: smart-energy-grid-env.eba-xyz123.eu-north-1.elasticbeanstalk.com

# Test endpoints
curl http://YOUR-APP-URL/health
curl http://YOUR-APP-URL/api/metrics
curl http://YOUR-APP-URL/api/alerts
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T10:30:00Z"
}
```

---

## 🔧 Alternative: Manual Step-by-Step Deployment

If `deploy-all.sh` fails, run scripts individually:

### **Step 1: Create AWS Infrastructure**

```bash
chmod +x setup-aws-infrastructure.sh
./setup-aws-infrastructure.sh
```

**What this creates:**
- S3 bucket: `smart-energy-grid-reports-nci`
- DynamoDB tables: `energy-grid-readings`, `energy-grid-alerts`, `energy-grid-equipment`
- SNS topic: `energy-grid-alerts`
- Lambda functions: `anomaly-detection`, `analytics-processing`
- IAM role: `EnergyGridLambdaRole`

**Verification:**
```bash
# Check DynamoDB tables
aws dynamodb list-tables --region eu-north-1

# Check S3 bucket
aws s3 ls s3://smart-energy-grid-reports-nci

# Check Lambda functions
aws lambda list-functions --region eu-north-1
```

---

### **Step 2: Deploy to Elastic Beanstalk**

```bash
chmod +x deploy-elastic-beanstalk.sh
./deploy-elastic-beanstalk.sh
```

**What this does:**
- Builds Go application for Linux
- Creates EB application: `smart-energy-grid`
- Creates EB environment: `smart-energy-grid-env`
- Deploys application
- Returns public URL

**Expected Output:**
```
✓ Application built successfully
✓ Elastic Beanstalk application created
✓ Environment created successfully
✓ URL: http://smart-energy-grid-env.eu-north-1.elasticbeanstalk.com
```

---

### **Step 3: Update Lambda Functions (if needed)**

If you need to redeploy Lambda functions:

```bash
# Anomaly Detection Lambda
cd lambda-functions/anomaly-detection
make deploy

# Analytics Processing Lambda
cd ../analytics-processing
make deploy

cd ../..
```

---

## 🧪 Testing Your Deployment

### **1. Test Backend API**

```bash
# Set your URL (get from deployment output)
export API_URL="http://YOUR-APP-URL"

# Health check
curl $API_URL/health

# Get metrics
curl $API_URL/api/metrics

# Get alerts
curl $API_URL/api/alerts

# Get equipment status
curl $API_URL/api/equipment
```

---

### **2. Test Lambda Functions**

```bash
# Test anomaly detection Lambda
aws lambda invoke \
  --function-name anomaly-detection \
  --region eu-north-1 \
  --payload '{"node_id":"node-001","value":95.5}' \
  response.json

cat response.json

# Test analytics processing Lambda
aws lambda invoke \
  --function-name analytics-processing \
  --region eu-north-1 \
  --payload '{"timeRange":"24h"}' \
  analytics-response.json

cat analytics-response.json
```

---

### **3. Test DynamoDB Integration**

```bash
# Insert test reading
aws dynamodb put-item \
  --table-name energy-grid-readings \
  --region eu-north-1 \
  --item '{
    "node_id": {"S": "test-node-001"},
    "timestamp": {"N": "1705750000"},
    "voltage": {"N": "240.5"},
    "current": {"N": "12.3"},
    "power": {"N": "2958.15"}
  }'

# Query readings
aws dynamodb query \
  --table-name energy-grid-readings \
  --region eu-north-1 \
  --key-condition-expression "node_id = :nid" \
  --expression-attribute-values '{":nid":{"S":"test-node-001"}}'
```

---

## 📊 AWS Console Verification

Check these in AWS Console:

1. **Elastic Beanstalk**
   - URL: https://eu-north-1.console.aws.amazon.com/elasticbeanstalk
   - Check: Application status = "Green" (Healthy)

2. **DynamoDB**
   - URL: https://eu-north-1.console.aws.amazon.com/dynamodbv2
   - Check: 3 tables exist with correct names

3. **Lambda**
   - URL: https://eu-north-1.console.aws.amazon.com/lambda
   - Check: 2 functions exist and are active

4. **S3**
   - URL: https://s3.console.aws.amazon.com/s3
   - Check: Bucket `smart-energy-grid-reports-nci` exists

5. **CloudWatch Logs**
   - URL: https://eu-north-1.console.aws.amazon.com/cloudwatch
   - Check: Log groups for EB and Lambda functions exist

---

## 🐛 Troubleshooting

### **Problem: Script fails with "command not found"**

**Solution:**
```bash
# Install missing tools
pip3 install awsebcli --upgrade --user

# Add to PATH
export PATH=$PATH:$HOME/.local/bin
```

---

### **Problem: "Access Denied" errors**

**Solution:**
```bash
# Verify AWS credentials
aws sts get-caller-identity

# If incorrect, reconfigure
aws configure
# Enter:
# - Access Key: AKIAVK3PK4IJAZ2HKIPA
# - Secret Key: (from .env file)
# - Region: eu-north-1
# - Format: json
```

---

### **Problem: Elastic Beanstalk deployment fails**

**Solution:**
```bash
# Check EB CLI version
eb --version

# Check application logs
eb logs

# Or use AWS CLI
aws logs tail /aws/elasticbeanstalk/smart-energy-grid-env/var/log/web.stdout.log \
  --follow --region eu-north-1
```

---

### **Problem: Lambda deployment fails**

**Solution:**
```bash
# Rebuild Lambda manually
cd lambda-functions/anomaly-detection
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
zip function.zip bootstrap

# Check if function exists
aws lambda get-function --function-name anomaly-detection --region eu-north-1

# If exists, update
aws lambda update-function-code \
  --function-name anomaly-detection \
  --zip-file fileb://function.zip \
  --region eu-north-1
```

---

### **Problem: Application returns 502/503 errors**

**Solution:**
```bash
# Check if application is listening on correct port
# Application should use PORT environment variable

# Check logs
eb logs --all

# Verify Procfile
cat Procfile
# Should contain: web: ./application
```

---

## 📝 What to Include in Your Project Report

After successful deployment, document these:

### **1. AWS Services Used (6 services):**
- DynamoDB (3 tables)
- Lambda (2 functions)
- S3 (1 bucket)
- SNS (1 topic)
- Elastic Beanstalk (1 application + environment)
- CloudWatch (logging & monitoring)

### **2. Deployment URLs:**
- Main Application: `http://[your-eb-url].elasticbeanstalk.com`
- Health Check: `http://[your-eb-url]/health`
- API Endpoints: `/api/metrics`, `/api/alerts`, `/api/equipment`

### **3. Screenshots to Take:**
- ✅ Elastic Beanstalk dashboard (green status)
- ✅ DynamoDB tables list
- ✅ Lambda functions list
- ✅ S3 bucket contents
- ✅ CloudWatch logs showing activity
- ✅ Browser showing your application running
- ✅ Terminal showing successful curl requests

### **4. Code Repository:**
- GitHub URL: https://github.com/ANIKETSHETTY47/smart-energy-grid-management-system
- Custom Library URL: https://github.com/ANIKETSHETTY47/energy-grid-analytics

---

## 🎯 Success Criteria Checklist

After deployment, verify:

- [ ] Application is accessible via public URL
- [ ] Health check endpoint returns 200 OK
- [ ] At least 5 AWS services are active
- [ ] Lambda functions can be invoked successfully
- [ ] DynamoDB tables can be queried
- [ ] CloudWatch logs show application activity
- [ ] Custom Go library is imported and used
- [ ] Code is in GitHub repositories

---

## 🚨 Important Notes

1. **Free Tier Usage:**
   - t3.micro instance is free tier eligible
   - DynamoDB has 25GB free storage
   - Lambda has 1M free requests/month
   - Monitor usage to avoid charges

2. **Cleanup When Done:**
   ```bash
   # Delete EB environment (costs $)
   eb terminate smart-energy-grid-env
   
   # Delete other resources
   aws dynamodb delete-table --table-name energy-grid-readings --region eu-north-1
   aws s3 rb s3://smart-energy-grid-reports-nci --force
   aws lambda delete-function --function-name anomaly-detection --region eu-north-1
   ```

3. **Custom Library:**
   - Already published: `github.com/ANIKETSHETTY47/energy-grid-analytics`
   - Used in main project via `go.mod`
   - Contains energy calculation utilities

---

## 📞 Need Help?

If deployment fails:

1. Check this guide's Troubleshooting section
2. Check AWS CloudWatch logs for errors
3. Verify all prerequisites are installed
4. Ensure AWS credentials are correct
5. Check AWS service quotas/limits

---

## ✅ Final Checklist Before Submission

- [ ] All AWS services created successfully
- [ ] Application deployed with public URL
- [ ] All endpoints tested and working
- [ ] Screenshots taken
- [ ] GitHub repositories updated
- [ ] Project report written
- [ ] Demo video recorded (if required)

---

**Good luck with your project! 🚀**
