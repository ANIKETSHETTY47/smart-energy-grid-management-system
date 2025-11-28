#!/usr/bin/env python3
"""
Complete AWS Infrastructure Setup for Smart Energy Grid Management System
This script creates all required AWS resources for the project
"""

import boto3
import json
import time
import sys
from botocore.exceptions import ClientError

# AWS Configuration
AWS_REGION = 'eu-north-1'
AWS_ACCESS_KEY = 'AKIAVK3PK4IJAZ2HKIPA'
AWS_SECRET_KEY = '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'

# Resource Names
S3_BUCKET_NAME = 'smart-energy-grid-reports-nci-2025'
DYNAMODB_READINGS_TABLE = 'energy-grid-readings'
DYNAMODB_ALERTS_TABLE = 'energy-grid-alerts'
DYNAMODB_EQUIPMENT_TABLE = 'energy-grid-equipment'
SNS_TOPIC_NAME = 'energy-grid-alerts'
LAMBDA_ANOMALY_FUNCTION = 'energy-anomaly-detection'
LAMBDA_ANALYTICS_FUNCTION = 'energy-analytics-processing'

class AWSSetup:
    def __init__(self):
        """Initialize AWS clients"""
        self.s3 = boto3.client(
            's3',
            region_name=AWS_REGION,
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        )
        self.dynamodb = boto3.client(
            'dynamodb',
            region_name=AWS_REGION,
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        )
        self.sns = boto3.client(
            'sns',
            region_name=AWS_REGION,
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        )
        self.lambda_client = boto3.client(
            'lambda',
            region_name=AWS_REGION,
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        )
        self.iam = boto3.client(
            'iam',
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        )
        self.logs = boto3.client(
            'logs',
            region_name=AWS_REGION,
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        )
        
        self.account_id = boto3.client(
            'sts',
            aws_access_key_id=AWS_ACCESS_KEY,
            aws_secret_access_key=AWS_SECRET_KEY
        ).get_caller_identity()['Account']
        
    def create_s3_bucket(self):
        """Create S3 bucket for reports and analytics"""
        print(f"\n[1/7] Creating S3 Bucket: {S3_BUCKET_NAME}")
        try:
            # For eu-north-1, we need to specify LocationConstraint
            self.s3.create_bucket(
                Bucket=S3_BUCKET_NAME,
                CreateBucketConfiguration={'LocationConstraint': AWS_REGION}
            )
            
            # Enable versioning
            self.s3.put_bucket_versioning(
                Bucket=S3_BUCKET_NAME,
                VersioningConfiguration={'Status': 'Enabled'}
            )
            
            print(f"✅ S3 Bucket created successfully: {S3_BUCKET_NAME}")
            return True
        except ClientError as e:
            if e.response['Error']['Code'] == 'BucketAlreadyOwnedByYou':
                print(f"✅ S3 Bucket already exists: {S3_BUCKET_NAME}")
                return True
            else:
                print(f"❌ Error creating S3 bucket: {e}")
                return False
    
    def create_dynamodb_tables(self):
        """Create DynamoDB tables for readings, alerts, and equipment"""
        print(f"\n[2/7] Creating DynamoDB Tables")
        
        tables = [
            {
                'name': DYNAMODB_READINGS_TABLE,
                'attributes': [
                    {'AttributeName': 'device_id', 'AttributeType': 'S'},
                    {'AttributeName': 'timestamp', 'AttributeType': 'N'}
                ],
                'key_schema': [
                    {'AttributeName': 'device_id', 'KeyType': 'HASH'},
                    {'AttributeName': 'timestamp', 'KeyType': 'RANGE'}
                ]
            },
            {
                'name': DYNAMODB_ALERTS_TABLE,
                'attributes': [
                    {'AttributeName': 'alert_id', 'AttributeType': 'S'},
                    {'AttributeName': 'created_at', 'AttributeType': 'N'}
                ],
                'key_schema': [
                    {'AttributeName': 'alert_id', 'KeyType': 'HASH'},
                    {'AttributeName': 'created_at', 'KeyType': 'RANGE'}
                ]
            },
            {
                'name': DYNAMODB_EQUIPMENT_TABLE,
                'attributes': [
                    {'AttributeName': 'equipment_id', 'AttributeType': 'S'}
                ],
                'key_schema': [
                    {'AttributeName': 'equipment_id', 'KeyType': 'HASH'}
                ]
            }
        ]
        
        for table in tables:
            try:
                self.dynamodb.create_table(
                    TableName=table['name'],
                    KeySchema=table['key_schema'],
                    AttributeDefinitions=table['attributes'],
                    BillingMode='PAY_PER_REQUEST',  # On-demand pricing
                    StreamSpecification={
                        'StreamEnabled': True,
                        'StreamViewType': 'NEW_AND_OLD_IMAGES'
                    }
                )
                print(f"  ⏳ Creating table: {table['name']}...")
                
                # Wait for table to be active
                waiter = self.dynamodb.get_waiter('table_exists')
                waiter.wait(TableName=table['name'])
                
                print(f"  ✅ Table created: {table['name']}")
            except ClientError as e:
                if e.response['Error']['Code'] == 'ResourceInUseException':
                    print(f"  ✅ Table already exists: {table['name']}")
                else:
                    print(f"  ❌ Error creating table {table['name']}: {e}")
                    return False
        
        return True
    
    def create_sns_topic(self):
        """Create SNS topic for alerts"""
        print(f"\n[3/7] Creating SNS Topic: {SNS_TOPIC_NAME}")
        try:
            response = self.sns.create_topic(Name=SNS_TOPIC_NAME)
            topic_arn = response['TopicArn']
            print(f"✅ SNS Topic created: {topic_arn}")
            
            # Update .env file with topic ARN
            env_file = '/Users/shetty/Desktop/Sem 1 Projects/Cloud Progm/smart-energy-grid-management-system/.env'
            with open(env_file, 'r') as f:
                content = f.read()
            
            content = content.replace('AWS_SNS_TOPIC_ARN=', f'AWS_SNS_TOPIC_ARN={topic_arn}')
            
            with open(env_file, 'w') as f:
                f.write(content)
            
            return topic_arn
        except ClientError as e:
            if 'already exists' in str(e):
                # Get existing topic ARN
                topics = self.sns.list_topics()
                for topic in topics['Topics']:
                    if SNS_TOPIC_NAME in topic['TopicArn']:
                        print(f"✅ SNS Topic already exists: {topic['TopicArn']}")
                        return topic['TopicArn']
            print(f"❌ Error creating SNS topic: {e}")
            return None
    
    def create_lambda_execution_role(self):
        """Create IAM role for Lambda functions"""
        print(f"\n[4/7] Creating Lambda Execution Role")
        role_name = 'EnergyGridLambdaExecutionRole'
        
        trust_policy = {
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Principal": {"Service": "lambda.amazonaws.com"},
                "Action": "sts:AssumeRole"
            }]
        }
        
        try:
            response = self.iam.create_role(
                RoleName=role_name,
                AssumeRolePolicyDocument=json.dumps(trust_policy),
                Description='Execution role for Energy Grid Lambda functions'
            )
            role_arn = response['Role']['Arn']
            print(f"  ✅ Role created: {role_arn}")
            
            # Attach managed policies
            policies = [
                'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole',
                'arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess',
                'arn:aws:iam::aws:policy/AmazonS3FullAccess',
                'arn:aws:iam::aws:policy/AmazonSNSFullAccess'
            ]
            
            for policy in policies:
                self.iam.attach_role_policy(RoleName=role_name, PolicyArn=policy)
                print(f"  ✅ Attached policy: {policy.split('/')[-1]}")
            
            # Wait for role to propagate
            print("  ⏳ Waiting for role to propagate...")
            time.sleep(10)
            
            return role_arn
        except ClientError as e:
            if e.response['Error']['Code'] == 'EntityAlreadyExists':
                response = self.iam.get_role(RoleName=role_name)
                role_arn = response['Role']['Arn']
                print(f"  ✅ Role already exists: {role_arn}")
                return role_arn
            else:
                print(f"  ❌ Error creating role: {e}")
                return None
    
    def create_cloudwatch_log_groups(self):
        """Create CloudWatch log groups for the application"""
        print(f"\n[5/7] Creating CloudWatch Log Groups")
        
        log_groups = [
            '/aws/lambda/energy-anomaly-detection',
            '/aws/lambda/energy-analytics-processing',
            '/energy-grid/api',
            '/energy-grid/ingestor'
        ]
        
        for log_group in log_groups:
            try:
                self.logs.create_log_group(logGroupName=log_group)
                
                # Set retention policy (7 days)
                self.logs.put_retention_policy(
                    logGroupName=log_group,
                    retentionInDays=7
                )
                
                print(f"  ✅ Log group created: {log_group}")
            except ClientError as e:
                if e.response['Error']['Code'] == 'ResourceAlreadyExistsException':
                    print(f"  ✅ Log group already exists: {log_group}")
                else:
                    print(f"  ❌ Error creating log group: {e}")
        
        return True
    
    def deploy_lambda_functions(self, role_arn):
        """Deploy Lambda functions"""
        print(f"\n[6/7] Deploying Lambda Functions")
        
        lambda_functions = [
            {
                'name': LAMBDA_ANOMALY_FUNCTION,
                'zip_path': 'lambda-functions/anomaly-detection/function.zip',
                'handler': 'bootstrap',
                'description': 'Detects anomalies in energy consumption data'
            },
            {
                'name': LAMBDA_ANALYTICS_FUNCTION,
                'zip_path': 'lambda-functions/analytics-processing/function.zip',
                'handler': 'bootstrap',
                'description': 'Processes energy analytics and generates reports'
            }
        ]
        
        for func in lambda_functions:
            try:
                # Read the zip file
                zip_file_path = f"/Users/shetty/Desktop/Sem 1 Projects/Cloud Progm/smart-energy-grid-management-system/{func['zip_path']}"
                
                try:
                    with open(zip_file_path, 'rb') as f:
                        zip_content = f.read()
                except FileNotFoundError:
                    print(f"  ⚠️  Zip file not found: {zip_file_path}")
                    print(f"  ℹ️  Please build the Lambda function first using: cd {func['zip_path'].rsplit('/', 1)[0]} && make build")
                    continue
                
                # Create or update function
                try:
                    response = self.lambda_client.create_function(
                        FunctionName=func['name'],
                        Runtime='provided.al2023',  # Custom runtime for Go
                        Role=role_arn,
                        Handler=func['handler'],
                        Code={'ZipFile': zip_content},
                        Description=func['description'],
                        Timeout=60,
                        MemorySize=256,
                        Environment={
                            'Variables': {
                                'AWS_REGION': AWS_REGION,
                                'DYNAMODB_READINGS_TABLE': DYNAMODB_READINGS_TABLE,
                                'DYNAMODB_ALERTS_TABLE': DYNAMODB_ALERTS_TABLE,
                                'S3_BUCKET': S3_BUCKET_NAME,
                                'SNS_TOPIC_ARN': self.sns_topic_arn
                            }
                        }
                    )
                    print(f"  ✅ Lambda function created: {func['name']}")
                except ClientError as e:
                    if e.response['Error']['Code'] == 'ResourceConflictException':
                        # Update existing function
                        self.lambda_client.update_function_code(
                            FunctionName=func['name'],
                            ZipFile=zip_content
                        )
                        
                        self.lambda_client.update_function_configuration(
                            FunctionName=func['name'],
                            Runtime='provided.al2023',
                            Role=role_arn,
                            Handler=func['handler'],
                            Timeout=60,
                            MemorySize=256,
                            Environment={
                                'Variables': {
                                    'AWS_REGION': AWS_REGION,
                                    'DYNAMODB_READINGS_TABLE': DYNAMODB_READINGS_TABLE,
                                    'DYNAMODB_ALERTS_TABLE': DYNAMODB_ALERTS_TABLE,
                                    'S3_BUCKET': S3_BUCKET_NAME,
                                    'SNS_TOPIC_ARN': self.sns_topic_arn
                                }
                            }
                        )
                        print(f"  ✅ Lambda function updated: {func['name']}")
                    else:
                        print(f"  ❌ Error creating Lambda function {func['name']}: {e}")
            except Exception as e:
                print(f"  ❌ Error with Lambda function {func['name']}: {e}")
        
        return True
    
    def print_summary(self):
        """Print setup summary"""
        print("\n" + "="*70)
        print("🎉 AWS INFRASTRUCTURE SETUP COMPLETE!")
        print("="*70)
        print(f"\n📦 Resources Created in Region: {AWS_REGION}")
        print(f"\n1. S3 Bucket:")
        print(f"   - Name: {S3_BUCKET_NAME}")
        print(f"   - URL: https://s3.{AWS_REGION}.amazonaws.com/{S3_BUCKET_NAME}")
        
        print(f"\n2. DynamoDB Tables:")
        print(f"   - {DYNAMODB_READINGS_TABLE}")
        print(f"   - {DYNAMODB_ALERTS_TABLE}")
        print(f"   - {DYNAMODB_EQUIPMENT_TABLE}")
        
        print(f"\n3. SNS Topic:")
        print(f"   - ARN: {self.sns_topic_arn}")
        
        print(f"\n4. Lambda Functions:")
        print(f"   - {LAMBDA_ANOMALY_FUNCTION}")
        print(f"   - {LAMBDA_ANALYTICS_FUNCTION}")
        
        print(f"\n5. CloudWatch Logs:")
        print(f"   - Multiple log groups created for monitoring")
        
        print(f"\n6. IAM Role:")
        print(f"   - EnergyGridLambdaExecutionRole")
        
        print("\n" + "="*70)
        print("📝 Next Steps:")
        print("="*70)
        print("1. Build Lambda functions:")
        print("   cd lambda-functions/anomaly-detection && make build")
        print("   cd lambda-functions/analytics-processing && make build")
        print("\n2. Deploy the main application to Elastic Beanstalk")
        print("\n3. Test all endpoints and verify cloud service integration")
        print("\n4. Check CloudWatch logs for any issues")
        print("="*70 + "\n")
    
    def run(self):
        """Run complete setup"""
        print("="*70)
        print("🚀 AWS INFRASTRUCTURE SETUP FOR SMART ENERGY GRID")
        print("="*70)
        print(f"Region: {AWS_REGION}")
        print(f"Account ID: {self.account_id}")
        print("="*70)
        
        # Step 1: S3
        if not self.create_s3_bucket():
            sys.exit(1)
        
        # Step 2: DynamoDB
        if not self.create_dynamodb_tables():
            sys.exit(1)
        
        # Step 3: SNS
        self.sns_topic_arn = self.create_sns_topic()
        if not self.sns_topic_arn:
            sys.exit(1)
        
        # Step 4: Lambda Role
        role_arn = self.create_lambda_execution_role()
        if not role_arn:
            sys.exit(1)
        
        # Step 5: CloudWatch
        if not self.create_cloudwatch_log_groups():
            sys.exit(1)
        
        # Step 6: Lambda Functions
        print("\n[7/7] Lambda Functions")
        print("  ℹ️  Lambda functions need to be built first")
        print("  ℹ️  Run the setup again after building: make build in each lambda directory")
        
        # Summary
        self.print_summary()

if __name__ == '__main__':
    try:
        setup = AWSSetup()
        setup.run()
    except KeyboardInterrupt:
        print("\n\n⚠️  Setup interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n\n❌ Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
