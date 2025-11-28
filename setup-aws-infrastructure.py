#!/usr/bin/env python3
"""
AWS Infrastructure Setup Script for Smart Energy Grid Management System
Creates all necessary AWS resources in eu-north-1 region
"""

import subprocess
import json
import time
import sys

# Configuration
REGION = "eu-north-1"
AWS_ACCOUNT_ID = "366916330002"
S3_BUCKET = "smart-energy-grid-reports-nci"
DYNAMODB_TABLE_READINGS = "energy-grid-readings"
DYNAMODB_TABLE_ALERTS = "energy-grid-alerts"
DYNAMODB_TABLE_EQUIPMENT = "energy-grid-equipment"
SNS_TOPIC_NAME = "energy-grid-alerts"
LAMBDA_ROLE_NAME = "EnergyGridLambdaRole"

# ANSI color codes
class Colors:
    GREEN = '\033[0;32m'
    BLUE = '\033[0;34m'
    YELLOW = '\033[1;33m'
    RED = '\033[0;31m'
    NC = '\033[0m'  # No Color

def run_aws_command(command, capture_output=True):
    """Run AWS CLI command"""
    try:
        result = subprocess.run(
            command, 
            shell=True,
            capture_output=capture_output,
            text=True,
            check=False
        )
        return result
    except Exception as e:
        print(f"{Colors.RED}Error running command: {e}{Colors.NC}")
        return None

def print_status(message):
    print(f"{Colors.GREEN}✓{Colors.NC} {message}")

def print_info(message):
    print(f"{Colors.BLUE}ℹ{Colors.NC} {message}")

def print_warning(message):
    print(f"{Colors.YELLOW}⚠{Colors.NC} {message}")

def print_error(message):
    print(f"{Colors.RED}✗{Colors.NC} {message}")

print(f"{Colors.BLUE}========================================{Colors.NC}")
print(f"{Colors.BLUE}Smart Energy Grid AWS Infrastructure Setup{Colors.NC}")
print(f"{Colors.BLUE}========================================{Colors.NC}\n")

# Step 1: Create S3 Bucket
print(f"\n{Colors.BLUE}Step 1: Creating S3 Bucket{Colors.NC}")
result = run_aws_command(f"aws s3 ls s3://{S3_BUCKET} --region {REGION}")
if result and result.returncode != 0:
    run_aws_command(f"aws s3 mb s3://{S3_BUCKET} --region {REGION}", capture_output=False)
    run_aws_command(f"aws s3api put-bucket-versioning --bucket {S3_BUCKET} --versioning-configuration Status=Enabled --region {REGION}")
    print_status(f"S3 Bucket created: {S3_BUCKET}")
else:
    print_warning(f"S3 Bucket already exists: {S3_BUCKET}")

# Step 2: Create DynamoDB Tables
print(f"\n{Colors.BLUE}Step 2: Creating DynamoDB Tables{Colors.NC}")

# Create readings table
print_info(f"Creating DynamoDB table: {DYNAMODB_TABLE_READINGS}")
result = run_aws_command(f"aws dynamodb describe-table --table-name {DYNAMODB_TABLE_READINGS} --region {REGION}")
if result and result.returncode != 0:
    run_aws_command(f"""aws dynamodb create-table \
        --table-name {DYNAMODB_TABLE_READINGS} \
        --attribute-definitions AttributeName=node_id,AttributeType=S AttributeName=timestamp,AttributeType=N \
        --key-schema AttributeName=node_id,KeyType=HASH AttributeName=timestamp,KeyType=RANGE \
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region {REGION}""", capture_output=False)
    print_info("Waiting for table to be created...")
    run_aws_command(f"aws dynamodb wait table-exists --table-name {DYNAMODB_TABLE_READINGS} --region {REGION}")
    print_status(f"Table created: {DYNAMODB_TABLE_READINGS}")
else:
    print_warning(f"Table already exists: {DYNAMODB_TABLE_READINGS}")

# Create alerts table
print_info(f"Creating DynamoDB table: {DYNAMODB_TABLE_ALERTS}")
result = run_aws_command(f"aws dynamodb describe-table --table-name {DYNAMODB_TABLE_ALERTS} --region {REGION}")
if result and result.returncode != 0:
    run_aws_command(f"""aws dynamodb create-table \
        --table-name {DYNAMODB_TABLE_ALERTS} \
        --attribute-definitions AttributeName=alert_id,AttributeType=S AttributeName=timestamp,AttributeType=N \
        --key-schema AttributeName=alert_id,KeyType=HASH AttributeName=timestamp,KeyType=RANGE \
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region {REGION}""", capture_output=False)
    print_info("Waiting for table to be created...")
    run_aws_command(f"aws dynamodb wait table-exists --table-name {DYNAMODB_TABLE_ALERTS} --region {REGION}")
    print_status(f"Table created: {DYNAMODB_TABLE_ALERTS}")
else:
    print_warning(f"Table already exists: {DYNAMODB_TABLE_ALERTS}")

# Create equipment table
print_info(f"Creating DynamoDB table: {DYNAMODB_TABLE_EQUIPMENT}")
result = run_aws_command(f"aws dynamodb describe-table --table-name {DYNAMODB_TABLE_EQUIPMENT} --region {REGION}")
if result and result.returncode != 0:
    run_aws_command(f"""aws dynamodb create-table \
        --table-name {DYNAMODB_TABLE_EQUIPMENT} \
        --attribute-definitions AttributeName=equipment_id,AttributeType=S \
        --key-schema AttributeName=equipment_id,KeyType=HASH \
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
        --region {REGION}""", capture_output=False)
    print_info("Waiting for table to be created...")
    run_aws_command(f"aws dynamodb wait table-exists --table-name {DYNAMODB_TABLE_EQUIPMENT} --region {REGION}")
    print_status(f"Table created: {DYNAMODB_TABLE_EQUIPMENT}")
else:
    print_warning(f"Table already exists: {DYNAMODB_TABLE_EQUIPMENT}")

# Step 3: Create SNS Topic
print(f"\n{Colors.BLUE}Step 3: Creating SNS Topic{Colors.NC}")
result = run_aws_command(f"aws sns create-topic --name {SNS_TOPIC_NAME} --region {REGION} --query 'TopicArn' --output text")
if result:
    SNS_TOPIC_ARN = result.stdout.strip()
    if not SNS_TOPIC_ARN or SNS_TOPIC_ARN == "":
        result = run_aws_command(f"aws sns list-topics --region {REGION} --query \"Topics[?contains(TopicArn, '{SNS_TOPIC_NAME}')].TopicArn\" --output text")
        SNS_TOPIC_ARN = result.stdout.strip()
    print_status(f"SNS Topic ARN: {SNS_TOPIC_ARN}")

# Step 4: Create IAM Role
print(f"\n{Colors.BLUE}Step 4: Creating IAM Role for Lambda Functions{Colors.NC}")
result = run_aws_command(f"aws iam get-role --role-name {LAMBDA_ROLE_NAME}")
if result and result.returncode != 0:
    trust_policy = {
        "Version": "2012-10-17",
        "Statement": [{
            "Effect": "Allow",
            "Principal": {"Service": "lambda.amazonaws.com"},
            "Action": "sts:AssumeRole"
        }]
    }
    
    with open('/tmp/lambda-trust-policy.json', 'w') as f:
        json.dump(trust_policy, f)
    
    run_aws_command(f"aws iam create-role --role-name {LAMBDA_ROLE_NAME} --assume-role-policy-document file:///tmp/lambda-trust-policy.json", capture_output=False)
    
    # Attach policies
    policies = [
        "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
        "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
        "arn:aws:iam::aws:policy/AmazonS3FullAccess",
        "arn:aws:iam::aws:policy/AmazonSNSFullAccess"
    ]
    
    for policy in policies:
        run_aws_command(f"aws iam attach-role-policy --role-name {LAMBDA_ROLE_NAME} --policy-arn {policy}")
    
    print_status(f"IAM Role created: {LAMBDA_ROLE_NAME}")
    print_info("Waiting 10 seconds for IAM role to propagate...")
    time.sleep(10)
else:
    print_warning(f"IAM Role already exists: {LAMBDA_ROLE_NAME}")

LAMBDA_ROLE_ARN = f"arn:aws:iam::{AWS_ACCOUNT_ID}:role/{LAMBDA_ROLE_NAME}"

print(f"\n{Colors.GREEN}========================================{Colors.NC}")
print(f"{Colors.GREEN}AWS Infrastructure Setup Complete!{Colors.NC}")
print(f"{Colors.GREEN}========================================{Colors.NC}\n")

print(f"{Colors.BLUE}Resource Summary:{Colors.NC}")
print(f"  S3 Bucket: {Colors.GREEN}{S3_BUCKET}{Colors.NC}")
print(f"  DynamoDB Tables:")
print(f"    - {Colors.GREEN}{DYNAMODB_TABLE_READINGS}{Colors.NC}")
print(f"    - {Colors.GREEN}{DYNAMODB_TABLE_ALERTS}{Colors.NC}")
print(f"    - {Colors.GREEN}{DYNAMODB_TABLE_EQUIPMENT}{Colors.NC}")
print(f"  SNS Topic: {Colors.GREEN}{SNS_TOPIC_ARN}{Colors.NC}")
print(f"  IAM Role: {Colors.GREEN}{LAMBDA_ROLE_NAME}{Colors.NC}\n")

print(f"{Colors.BLUE}Next Steps:{Colors.NC}")
print(f"  1. Update your .env file with:")
print(f"     {Colors.YELLOW}AWS_SNS_TOPIC_ARN={SNS_TOPIC_ARN}{Colors.NC}")
print(f"  2. Build and deploy Lambda functions")
print(f"  3. Deploy your main application to Elastic Beanstalk")

print_status("Setup script completed successfully!")
