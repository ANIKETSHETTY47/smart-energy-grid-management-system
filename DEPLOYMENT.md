# Deployment Guide

This guide covers deploying the Smart Energy Grid Management System to AWS.

## Prerequisites

Before deploying, ensure you have:

- AWS Account with appropriate permissions
- AWS CLI installed and configured
- EB CLI installed: `pip install awsebcli`
- Go 1.22+ installed
- Git repository set up

## AWS IAM Setup

### Required IAM Roles

1. **aws-elasticbeanstalk-service-role**
   - Managed policies:
     - AWSElasticBeanstalkEnhancedHealth
     - AWSElasticBeanstalkManagedUpdatesCustomerRolePolicy

2. **aws-elasticbeanstalk-ec2-role**
   - Managed policies:
     - AWSElasticBeanstalkWebTier
     - AWSElasticBeanstalkWorkerTier
     - AWSElasticBeanstalkMulticontainerDocker

### Required AWS Services Access

Ensure your IAM user/role has permissions for:
- Elastic Beanstalk (full access)
- EC2 (launch instances, security groups)
- S3 (create buckets, upload objects)
- CloudWatch (logs and metrics)
- DynamoDB (if using cloud services)
- Lambda (if deploying functions)
- SNS (for notifications)

## GitHub Secrets Configuration

For CI/CD pipeline, add these secrets to your GitHub repository:

1. Go to Settings → Secrets and variables → Actions → New repository secret
2. Add the following secrets:

```
AWS_ACCESS_KEY_ID=<your-access-key>
AWS_SECRET_ACCESS_KEY=<your-secret-key>
```

## Automated Deployment (CI/CD)

The project uses GitHub Actions for automated deployments.

### Workflow Triggers

- **Push to main**: Automatically deploys to production
- **Push to develop**: Runs tests only
- **Pull Request**: Runs tests and builds
- **Manual**: Via workflow_dispatch

### Pipeline Stages

1. **Test & Build** (runs on all branches)
   - Downloads dependencies
   - Runs tests with coverage
   - Builds backend and frontend binaries
   - Uploads artifacts

2. **Deploy Backend** (main branch only)
   - Configures AWS credentials
   - Initializes Elastic Beanstalk
   - Deploys API server
   - Outputs backend URL

3. **Deploy Frontend** (main branch only)
   - Configures AWS credentials
   - Sets environment variables (API_URL)
   - Deploys dashboard
   - Outputs frontend URL

4. **Deploy Lambda** (main branch only, optional)
   - Builds Lambda functions
   - Updates function code
   - Handles analytics-processing and anomaly-detection

### Triggering Deployments

**Automatic (Recommended)**
```bash
git add .
git commit -m "feat: your feature description"
git push origin main
```

**Manual via GitHub UI**
1. Go to Actions tab
2. Select "CI/CD Pipeline"
3. Click "Run workflow"
4. Select branch (main)
5. Click "Run workflow"

## Manual Deployment

### Backend Deployment

1. **Initialize Elastic Beanstalk (First time only)**
   ```bash
   eb init smart-energy-grid \
     --platform "Go 1 running on 64bit Amazon Linux 2023" \
     --region us-east-1
   ```

2. **Create Environment (First time only)**
   ```bash
   eb create smart-energy-grid-env \
     --single \
     --instance_type t3.micro \
     --service-role aws-elasticbeanstalk-service-role \
     --instance_profile aws-elasticbeanstalk-ec2-role
   ```

3. **Deploy Updates**
   ```bash
   eb deploy smart-energy-grid-env
   ```

4. **Check Status**
   ```bash
   eb status
   eb health
   ```

5. **View Logs**
   ```bash
   eb logs
   eb logs --stream  # Real-time logs
   ```

### Frontend Deployment

1. **Navigate to Frontend Directory**
   ```bash
   cd web/energy-dashboard
   ```

2. **Initialize Elastic Beanstalk (First time only)**
   ```bash
   eb init energy-dashboard-frontend \
     --platform "Go 1 running on 64bit Amazon Linux 2023" \
     --region us-east-1
   ```

3. **Create Environment (First time only)**
   ```bash
   eb create energy-dashboard-frontend-env \
     --single \
     --instance_type t3.micro \
     --service-role aws-elasticbeanstalk-service-role \
     --instance_profile aws-elasticbeanstalk-ec2-role \
     --envvars API_URL=http://smart-energy-grid-env.us-east-1.elasticbeanstalk.com,PORT=3000
   ```

4. **Deploy Updates**
   ```bash
   eb deploy energy-dashboard-frontend-env
   ```

5. **Update Environment Variables**
   ```bash
   eb setenv API_URL=http://your-backend-url.com PORT=3000
   ```

### Lambda Functions Deployment

1. **Build Function**
   ```bash
   cd lambda-functions/analytics-processing
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
   zip function.zip bootstrap
   ```

2. **Deploy to Lambda**
   ```bash
   aws lambda update-function-code \
     --function-name analytics-processing \
     --zip-file fileb://function.zip \
     --region us-east-1
   ```

3. **Repeat for anomaly-detection**
   ```bash
   cd ../anomaly-detection
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
   zip function.zip bootstrap
   aws lambda update-function-code \
     --function-name anomaly-detection \
     --zip-file fileb://function.zip \
     --region us-east-1
   ```

## Environment Configuration

### Backend Environment Variables

Set these in Elastic Beanstalk console or via CLI:

```bash
eb setenv \
  USE_CLOUD_SERVICES=true \
  AWS_REGION=us-east-1 \
  DYNAMODB_SENSORS_TABLE=energy-sensors \
  DYNAMODB_READINGS_TABLE=energy-readings \
  S3_BUCKET=energy-grid-data \
  SNS_ALERT_TOPIC_ARN=arn:aws:sns:us-east-1:123456789012:alerts
```

### Frontend Environment Variables

```bash
cd web/energy-dashboard
eb setenv \
  API_URL=http://smart-energy-grid-env.us-east-1.elasticbeanstalk.com \
  PORT=3000
```

## Health Checks and Monitoring

### Health Check Endpoints

- Backend: `http://your-backend-url/health`
- Frontend: `http://your-frontend-url/healthz`

### Monitoring

1. **Elastic Beanstalk Console**
   - Go to AWS Console → Elastic Beanstalk
   - Select your environment
   - View Monitoring tab for metrics

2. **CloudWatch Logs**
   ```bash
   eb logs --stream  # Real-time logs
   ```

3. **Application Performance**
   - Monitor response times
   - Check error rates
   - View CPU/Memory usage

## Scaling Configuration

### Horizontal Scaling (Load Balancing)

To convert from single instance to load balanced:

```bash
eb scale 2  # Scale to 2 instances
```

Or configure auto-scaling in `.ebextensions/scaling.config`:

```yaml
option_settings:
  aws:autoscaling:asg:
    MinSize: 1
    MaxSize: 4
  aws:autoscaling:trigger:
    MeasureName: CPUUtilization
    Statistic: Average
    Unit: Percent
    UpperThreshold: 80
    LowerThreshold: 20
```

### Vertical Scaling (Instance Type)

```bash
eb scale --instance_type t3.small
```

## Rollback Procedure

If a deployment fails or causes issues:

1. **Check Recent Deployments**
   ```bash
   aws elasticbeanstalk describe-application-versions \
     --application-name smart-energy-grid
   ```

2. **Rollback to Previous Version**
   ```bash
   eb deploy --version <previous-version-label>
   ```

3. **Via Console**
   - Go to Application Versions
   - Select previous version
   - Click "Deploy"

## Troubleshooting

### Deployment Fails

**Check logs:**
```bash
eb logs
```

**Common issues:**
- IAM permissions insufficient
- Build fails (check Buildfile)
- Port conflicts
- Environment variable missing

### Application Not Starting

1. **Check health status:**
   ```bash
   eb health --refresh
   ```

2. **Check application logs:**
   ```bash
   eb logs --stream
   ```

3. **Verify Procfile and Buildfile:**
   - Buildfile: `go build -o application ./cmd/api`
   - Procfile: `web: ./application`

### Database Connection Issues

If using PostgreSQL:
- Ensure RDS is in same VPC
- Check security groups
- Verify connection string

If using DynamoDB:
- Verify table names in environment variables
- Check IAM permissions
- Ensure tables exist in correct region

## Cost Optimization

### Development Environment

- Use single instance (no load balancer)
- Use t3.micro instances
- Terminate when not needed: `eb terminate`

### Production Environment

- Use appropriate instance types
- Enable auto-scaling
- Use reserved instances for consistent workload
- Monitor CloudWatch costs

## Security Best Practices

1. **Never commit secrets to Git**
   - Use environment variables
   - Store sensitive data in AWS Secrets Manager

2. **Enable HTTPS**
   - Use AWS Certificate Manager
   - Configure Load Balancer with SSL certificate

3. **Restrict Access**
   - Use security groups properly
   - Enable AWS WAF if needed
   - Implement API authentication

4. **Regular Updates**
   - Keep dependencies updated
   - Apply security patches
   - Monitor AWS security bulletins

## Useful Commands

```bash
# View environment info
eb status

# SSH into instance
eb ssh

# Open in browser
eb open

# List environments
eb list

# Terminate environment
eb terminate <env-name>

# Change instance type
eb scale --instance_type t3.small

# View recent events
eb events --follow
```

## Additional Resources

- [AWS Elastic Beanstalk Documentation](https://docs.aws.amazon.com/elasticbeanstalk/)
- [EB CLI Reference](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/eb-cli3.html)
- [Go on Elastic Beanstalk](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/go-environment.html)
