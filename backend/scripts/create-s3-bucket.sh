#!/usr/bin/env bash

# Script to create S3 bucket for Lingopaste
# Usage: ./create-s3-bucket.sh

set -e

BUCKET_NAME=${S3_BUCKET_NAME:-lingopaste-data}
export AWS_REGION=${AWS_REGION:-ap-northeast-1}

echo "Creating S3 bucket: $BUCKET_NAME in region: $AWS_REGION"

# Check if bucket already exists
if aws s3api head-bucket --bucket $BUCKET_NAME 2>/dev/null; then
    echo "Bucket $BUCKET_NAME already exists, skipping creation..."
else
    echo "Creating bucket..."
    # Create bucket (us-east-1 doesn't need LocationConstraint)
    if [ "$AWS_REGION" == "us-east-1" ]; then
        aws s3api create-bucket \
            --bucket $BUCKET_NAME \
            --region $AWS_REGION
    else
        aws s3api create-bucket \
            --bucket $BUCKET_NAME \
            --region $AWS_REGION \
            --create-bucket-configuration LocationConstraint=$AWS_REGION
    fi
fi

# Enable versioning
echo "Enabling versioning..."
aws s3api put-bucket-versioning \
    --bucket $BUCKET_NAME \
    --versioning-configuration Status=Enabled

# Enable encryption
echo "Enabling encryption..."
aws s3api put-bucket-encryption \
    --bucket $BUCKET_NAME \
    --server-side-encryption-configuration '{
        "Rules": [{
            "ApplyServerSideEncryptionByDefault": {
                "SSEAlgorithm": "AES256"
            }
        }]
    }'

# Block public access
echo "Blocking public access..."
aws s3api put-public-access-block \
    --bucket $BUCKET_NAME \
    --public-access-block-configuration \
        "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

echo "S3 bucket created successfully: $BUCKET_NAME"
