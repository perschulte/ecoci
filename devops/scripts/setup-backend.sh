#!/bin/bash

# EcoCI Infrastructure Backend Setup Script
# This script creates the S3 bucket and DynamoDB table for Terraform state management

set -e

# Configuration
ENVIRONMENT=${1:-staging}
AWS_REGION=${2:-us-west-2}
PROJECT_NAME="ecoci"

BUCKET_NAME="${PROJECT_NAME}-${ENVIRONMENT}-terraform-state"
DYNAMODB_TABLE="${PROJECT_NAME}-${ENVIRONMENT}-terraform-locks"

echo "Setting up Terraform backend for EcoCI ${ENVIRONMENT} environment..."
echo "AWS Region: ${AWS_REGION}"
echo "S3 Bucket: ${BUCKET_NAME}"
echo "DynamoDB Table: ${DYNAMODB_TABLE}"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo "Error: AWS CLI is not installed. Please install it first."
    exit 1
fi

# Check AWS credentials
if ! aws sts get-caller-identity &> /dev/null; then
    echo "Error: AWS credentials not configured. Please run 'aws configure' first."
    exit 1
fi

# Create S3 bucket for Terraform state
echo "Creating S3 bucket: ${BUCKET_NAME}"
if aws s3 ls "s3://${BUCKET_NAME}" 2>&1 | grep -q 'NoSuchBucket'; then
    aws s3 mb "s3://${BUCKET_NAME}" --region "${AWS_REGION}"
    echo "✅ S3 bucket created successfully"
else
    echo "ℹ️  S3 bucket already exists"
fi

# Enable versioning on S3 bucket
echo "Enabling versioning on S3 bucket..."
aws s3api put-bucket-versioning \\
    --bucket "${BUCKET_NAME}" \\
    --versioning-configuration Status=Enabled

# Enable server-side encryption
echo "Enabling server-side encryption on S3 bucket..."
aws s3api put-bucket-encryption \\
    --bucket "${BUCKET_NAME}" \\
    --server-side-encryption-configuration '{
        "Rules": [
            {
                "ApplyServerSideEncryptionByDefault": {
                    "SSEAlgorithm": "AES256"
                },
                "BucketKeyEnabled": true
            }
        ]
    }'

# Block public access
echo "Blocking public access on S3 bucket..."
aws s3api put-public-access-block \\
    --bucket "${BUCKET_NAME}" \\
    --public-access-block-configuration \\
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true

# Create DynamoDB table for state locking
echo "Creating DynamoDB table: ${DYNAMODB_TABLE}"
if ! aws dynamodb describe-table --table-name "${DYNAMODB_TABLE}" --region "${AWS_REGION}" &> /dev/null; then
    aws dynamodb create-table \\
        --table-name "${DYNAMODB_TABLE}" \\
        --attribute-definitions AttributeName=LockID,AttributeType=S \\
        --key-schema AttributeName=LockID,KeyType=HASH \\
        --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \\
        --region "${AWS_REGION}" \\
        --tags Key=Project,Value=EcoCI Key=Environment,Value="${ENVIRONMENT}" Key=ManagedBy,Value=Script
    
    echo "Waiting for DynamoDB table to become active..."
    aws dynamodb wait table-exists --table-name "${DYNAMODB_TABLE}" --region "${AWS_REGION}"
    echo "✅ DynamoDB table created successfully"
else
    echo "ℹ️  DynamoDB table already exists"
fi

echo ""
echo "🎉 Terraform backend setup completed successfully!"
echo ""
echo "Backend configuration:"
echo "  Bucket: ${BUCKET_NAME}"
echo "  DynamoDB Table: ${DYNAMODB_TABLE}"
echo "  Region: ${AWS_REGION}"
echo ""
echo "You can now run 'terraform init' in the environment directory."