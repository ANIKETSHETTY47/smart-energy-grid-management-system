# 🚀 QUICK START - Smart Energy Grid Deployment

## ⚡ TL;DR - Just Deploy Everything

```bash
cd "/Users/shetty/Desktop/Sem 1 Projects/Cloud Progm/smart-energy-grid-management-system"
./deploy-all.sh
```

**That's it!** Script runs for ~10-15 minutes and deploys everything.

---

## 📝 What Gets Deployed

✅ **6 AWS Services:**
1. **DynamoDB** - 3 tables (readings, alerts, equipment)
2. **Lambda** - 2 functions (anomaly-detection, analytics-processing)
3. **S3** - 1 bucket (smart-energy-grid-reports-nci)
4. **SNS** - 1 topic (energy-grid-alerts)
5. **Elastic Beanstalk** - Backend API server
6. **CloudWatch** - Logs & monitoring

---

## 🧪 Quick Test Commands

```bash
# Get your URL from script output, then:
export API_URL="http://YOUR-APP-URL"

# Test API
curl $API_URL/health
curl $API_URL/api/metrics
curl $API_URL/api/alerts
```

---

## 🔧 Manual Deployment (If Automated Fails)

### Step 1: Infrastructure
```bash
./setup-aws-infrastructure.sh
# Creates: S3, DynamoDB, Lambda, SNS, IAM
```

### Step 2: Application
```bash
./deploy-elastic-beanstalk.sh
# Deploys: Backend API to Elastic Beanstalk
```

---

## 📊 Check AWS Console

After deployment, verify in AWS Console (eu-north-1 region):

1. **Elastic Beanstalk**: Application should be "Green" (Healthy)
   - https://eu-north-1.console.aws.amazon.com/elasticbeanstalk

2. **DynamoDB**: 3 tables should exist
   - https://eu-north-1.console.aws.amazon.com/dynamodbv2

3. **Lambda**: 2 functions should be active
   - https://eu-north-1.console.aws.amazon.com/lambda

4. **S3**: 1 bucket should exist
   - https://s3.console.aws.amazon.com/s3

5. **CloudWatch**: Logs should show activity
   - https://eu-north-1.console.aws.amazon.com/cloudwatch

---

## 🐛 Common Issues

### "AWS CLI not found"
```bash
# Install AWS CLI
brew install awscli  # macOS

# Configure credentials
aws configure
# Enter: Access Key, Secret Key, Region (eu-north-1), Format (json)
```

### "EB CLI not found"
```bash
pip3 install awsebcli --upgrade --user
export PATH=$PATH:$HOME/.local/bin
```

### "Access Denied"
```bash
# Verify credentials
aws sts get-caller-identity

# Should show Account: 366916330002
```

### Check Logs
```bash
# Elastic Beanstalk logs
aws logs tail /aws/elasticbeanstalk/smart-energy-grid-env/var/log/web.stdout.log \
  --follow --region eu-north-1

# Lambda logs
aws logs tail /aws/lambda/anomaly-detection --follow --region eu-north-1
```

---

## 📸 Screenshots Needed for Report

1. ✅ Elastic Beanstalk dashboard (green status)
2. ✅ DynamoDB tables list (3 tables)
3. ✅ Lambda functions list (2 functions)
4. ✅ S3 bucket with reports
5. ✅ CloudWatch logs showing activity
6. ✅ Browser with application running
7. ✅ Terminal with successful curl tests

---

## 🎯 Success Indicators

After deployment completes, you should have:

✅ Public URL working: `http://[app-url].elasticbeanstalk.com`
✅ Health check returns: `{"status":"healthy"}`
✅ API endpoints respond with data
✅ 6 AWS services visible in console
✅ CloudWatch logs show requests
✅ Lambda functions can be invoked

---

## 🧹 Cleanup (After Grading)

**⚠️ Only run this AFTER project is graded!**

```bash
# Delete Elastic Beanstalk (most expensive)
eb terminate smart-energy-grid-env

# Delete other resources
aws dynamodb delete-table --table-name energy-grid-readings --region eu-north-1
aws dynamodb delete-table --table-name energy-grid-alerts --region eu-north-1
aws dynamodb delete-table --table-name energy-grid-equipment --region eu-north-1
aws s3 rb s3://smart-energy-grid-reports-nci --force --region eu-north-1
aws lambda delete-function --function-name anomaly-detection --region eu-north-1
aws lambda delete-function --function-name analytics-processing --region eu-north-1
aws sns delete-topic --topic-arn [YOUR-SNS-ARN] --region eu-north-1
```

---

## 📚 Full Documentation

See `DEPLOYMENT-GUIDE.md` for:
- Detailed step-by-step instructions
- Troubleshooting guide
- Testing procedures
- Report requirements

---

## ✅ Pre-Deployment Checklist

- [ ] In correct directory
- [ ] AWS credentials configured
- [ ] AWS CLI installed
- [ ] Go 1.21+ installed
- [ ] Internet connection active

---

## 🚀 Ready? Deploy Now!

```bash
./deploy-all.sh
```

**Duration:** 10-15 minutes
**Output:** Public URL + AWS resources created

---

**Questions?** Check `DEPLOYMENT-GUIDE.md` or AWS CloudWatch logs.
