# 🎯 DEPLOYMENT COMMAND REFERENCE

## Single Command Deployment

```bash
cd "/Users/shetty/Desktop/Sem 1 Projects/Cloud Progm/smart-energy-grid-management-system"
./deploy-all.sh
```

---

## What Gets Created

```
┌─────────────────────────────────────────────────────────────┐
│                     AWS INFRASTRUCTURE                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  📦 S3 BUCKET                                               │
│     └─ smart-energy-grid-reports-nci                        │
│        └─ Stores analytics reports                          │
│                                                              │
│  🗄️  DYNAMODB TABLES (3)                                    │
│     ├─ energy-grid-readings      → Energy consumption data  │
│     ├─ energy-grid-alerts        → System alerts           │
│     └─ energy-grid-equipment     → Equipment metadata      │
│                                                              │
│  ⚡ LAMBDA FUNCTIONS (2)                                    │
│     ├─ anomaly-detection         → Detects abnormal usage  │
│     └─ analytics-processing      → Generates reports       │
│                                                              │
│  📢 SNS TOPIC                                               │
│     └─ energy-grid-alerts        → Sends notifications     │
│                                                              │
│  🚀 ELASTIC BEANSTALK                                       │
│     ├─ Application: smart-energy-grid                      │
│     ├─ Environment: smart-energy-grid-env                  │
│     └─ Instance: t3.micro (free tier)                      │
│                                                              │
│  📊 CLOUDWATCH                                              │
│     ├─ Application logs                                     │
│     ├─ Lambda logs                                          │
│     └─ Metrics & monitoring                                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Service Count: 6/5 Required ✅

1. ✅ **DynamoDB** - NoSQL Database
2. ✅ **Lambda** - Serverless Computing
3. ✅ **S3** - Object Storage
4. ✅ **SNS** - Notifications
5. ✅ **Elastic Beanstalk** - PaaS Hosting
6. ✅ **CloudWatch** - Monitoring

---

## API Endpoints After Deployment

```
Base URL: http://[your-app-url].eu-north-1.elasticbeanstalk.com

GET  /health              → Service health check
GET  /api/metrics         → Current energy metrics
GET  /api/alerts          → Active alerts
GET  /api/equipment       → Equipment status
POST /api/readings        → Submit energy reading
GET  /api/analytics       → Analytics dashboard data
```

---

## Testing Commands

```bash
# Set your URL (get from deployment output)
export API_URL="http://smart-energy-grid-env.eba-xyz.eu-north-1.elasticbeanstalk.com"

# Health check
curl $API_URL/health

# Expected: {"status":"healthy","timestamp":"2024-11-23T..."}

# Get metrics
curl $API_URL/api/metrics

# Get alerts
curl $API_URL/api/alerts

# Get equipment
curl $API_URL/api/equipment
```

---

## AWS Console Links (eu-north-1)

```
Elastic Beanstalk:
https://eu-north-1.console.aws.amazon.com/elasticbeanstalk

DynamoDB:
https://eu-north-1.console.aws.amazon.com/dynamodbv2

Lambda:
https://eu-north-1.console.aws.amazon.com/lambda

S3:
https://s3.console.aws.amazon.com/s3/buckets/smart-energy-grid-reports-nci

CloudWatch:
https://eu-north-1.console.aws.amazon.com/cloudwatch
```

---

## Deployment Timeline

```
┌─────────────────────────────────────────────────────────────┐
│ PHASE 1: Infrastructure Setup (5-7 min)                     │
├─────────────────────────────────────────────────────────────┤
│ ├─ Create S3 bucket                          [30s]          │
│ ├─ Create DynamoDB tables (3)                [2min]         │
│ ├─ Create SNS topic                          [20s]          │
│ ├─ Create IAM roles                          [30s]          │
│ └─ Deploy Lambda functions (2)               [2min]         │
├─────────────────────────────────────────────────────────────┤
│ PHASE 2: Application Deployment (5-8 min)                   │
├─────────────────────────────────────────────────────────────┤
│ ├─ Build Go application                      [30s]          │
│ ├─ Create EB application                     [20s]          │
│ ├─ Create EB environment                     [5min]         │
│ └─ Deploy application                        [2min]         │
└─────────────────────────────────────────────────────────────┘

Total Time: ~10-15 minutes
```

---

## Status Indicators

```
After deployment completes, you should see:

✅ S3 Bucket Created
✅ DynamoDB Tables Active (3/3)
✅ SNS Topic Created
✅ Lambda Functions Deployed (2/2)
✅ IAM Roles Configured
✅ Elastic Beanstalk Environment: Green (Healthy)
✅ Public URL: http://[app-url].elasticbeanstalk.com
✅ CloudWatch Logs: Active
```

---

## Troubleshooting Quick Commands

```bash
# Check AWS identity
aws sts get-caller-identity

# List DynamoDB tables
aws dynamodb list-tables --region eu-north-1

# List Lambda functions
aws lambda list-functions --region eu-north-1

# Check S3 bucket
aws s3 ls s3://smart-energy-grid-reports-nci

# View EB logs
aws logs tail /aws/elasticbeanstalk/smart-energy-grid-env/var/log/web.stdout.log \
  --follow --region eu-north-1

# EB environment status
aws elasticbeanstalk describe-environments \
  --application-name smart-energy-grid \
  --environment-names smart-energy-grid-env \
  --region eu-north-1
```

---

## File Structure Reference

```
smart-energy-grid-management-system/
│
├── 🚀 DEPLOYMENT SCRIPTS
│   ├── deploy-all.sh                    ← RUN THIS ONE
│   ├── setup-aws-infrastructure.sh      ← Creates AWS resources
│   └── deploy-elastic-beanstalk.sh      ← Deploys application
│
├── 📚 DOCUMENTATION
│   ├── PROJECT-STATUS.md                ← Current status
│   ├── DEPLOYMENT-GUIDE.md              ← Complete guide
│   ├── QUICK-START.md                   ← Quick reference
│   └── COMMANDS.md                      ← This file
│
├── 🔧 CONFIGURATION
│   ├── .env                             ← AWS credentials
│   ├── Procfile                         ← EB startup
│   └── .ebextensions/                   ← EB config
│
├── 💻 APPLICATION CODE
│   ├── cmd/api/main.go                  ← Entry point
│   ├── internal/cloud/                  ← AWS integrations
│   ├── internal/service/                ← Business logic
│   └── lambda-functions/                ← Lambda code
│
└── 📦 DEPENDENCIES
    ├── go.mod                           ← Go dependencies
    └── energy-grid-analytics            ← Custom library
```

---

## Success Checklist

After running `./deploy-all.sh`, verify:

- [ ] Script completes without errors
- [ ] Public URL is provided
- [ ] Health check returns 200 OK
- [ ] AWS Console shows all resources
- [ ] DynamoDB tables are created (3)
- [ ] Lambda functions are active (2)
- [ ] S3 bucket exists
- [ ] CloudWatch logs show activity
- [ ] EB environment is Green/Healthy

---

## Cost Summary (Free Tier)

```
Service              Free Tier Limit           Expected Usage
─────────────────────────────────────────────────────────────
EC2 (t3.micro)       750 hrs/month            ~730 hrs/month  ✅
DynamoDB             25 GB storage            ~1 GB           ✅
Lambda               1M requests/month        ~10K requests   ✅
S3                   5 GB storage             ~500 MB         ✅
CloudWatch           5 GB logs                ~1 GB           ✅
SNS                  1000 notifications       ~100            ✅

Estimated Cost: $0-5/month (staying within free tier)
```

---

## Next Actions After Deployment

```
1. ✅ Verify deployment completed
   └─ Check terminal output for "DEPLOYMENT SUCCESSFUL"

2. 🧪 Test API endpoints
   └─ Run curl commands listed above

3. 🖼️ Take screenshots
   └─ AWS Console (all 6 services)
   └─ Browser (application running)
   └─ Terminal (successful tests)

4. 📝 Document in report
   └─ Public URL
   └─ AWS services used
   └─ Architecture diagram
   └─ Screenshots

5. 🎓 Submit project
   └─ Report with screenshots
   └─ GitHub repository links
   └─ Public URL for testing
```

---

## Emergency Rollback

If something goes wrong:

```bash
# Terminate Elastic Beanstalk environment
eb terminate smart-energy-grid-env

# Delete other resources
aws cloudformation delete-stack --stack-name smart-energy-grid-infra
```

---

## Repository Links

```
Main Application:
https://github.com/ANIKETSHETTY47/smart-energy-grid-management-system

Custom Library:
https://github.com/ANIKETSHETTY47/energy-grid-analytics
```

---

## Support

If deployment fails:
1. Check `DEPLOYMENT-GUIDE.md` troubleshooting section
2. Review AWS CloudWatch logs
3. Verify AWS credentials are correct
4. Ensure all prerequisites are installed

---

**Ready? Deploy now:**

```bash
./deploy-all.sh
```

---

*Quick Reference Guide*  
*Smart Energy Grid Management System*  
*AWS Region: eu-north-1*  
*Last Updated: November 23, 2024*
