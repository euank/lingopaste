#!/usr/bin/env bash

# Script to create IAM user with minimal permissions for Lingopaste
# Usage: ./create-iam-user.sh

set -e

IAM_USER_NAME="lingopaste-app"
POLICY_NAME="lingopaste-app-policy"
AWS_REGION=${AWS_REGION:-ap-northeast-1}
S3_BUCKET_NAME=${S3_BUCKET_NAME:-lingopaste-data}

echo "Creating IAM user and policy for Lingopaste..."
echo "Region: $AWS_REGION"
echo "S3 Bucket: $S3_BUCKET_NAME"
echo ""

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo "AWS Account ID: $ACCOUNT_ID"

# Create IAM user
echo ""
echo "Creating IAM user: $IAM_USER_NAME"
if aws iam get-user --user-name $IAM_USER_NAME 2>/dev/null; then
    echo "User already exists, skipping creation..."
else
    aws iam create-user --user-name $IAM_USER_NAME
    echo "User created successfully"
fi

# Create IAM policy with minimal permissions
echo ""
echo "Creating IAM policy: $POLICY_NAME"

POLICY_DOCUMENT=$(cat <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "DynamoDBTableAccess",
      "Effect": "Allow",
      "Action": [
        "dynamodb:PutItem",
        "dynamodb:GetItem",
        "dynamodb:UpdateItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:DescribeTable",
        "dynamodb:DescribeTimeToLive"
      ],
      "Resource": [
        "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/lingopaste-accounts",
        "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/lingopaste-accounts/index/*",
        "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/lingopaste-pastes",
        "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/lingopaste-pastes/index/*",
        "arn:aws:dynamodb:${AWS_REGION}:${ACCOUNT_ID}:table/lingopaste-rate-limits"
      ]
    },
    {
      "Sid": "S3BucketAccess",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::${S3_BUCKET_NAME}",
        "arn:aws:s3:::${S3_BUCKET_NAME}/*"
      ]
    }
  ]
}
EOF
)

# Check if policy already exists
POLICY_ARN="arn:aws:iam::${ACCOUNT_ID}:policy/${POLICY_NAME}"
if aws iam get-policy --policy-arn $POLICY_ARN 2>/dev/null; then
    echo "Policy already exists, updating to latest version..."
    
    # Create a new policy version
    aws iam create-policy-version \
        --policy-arn $POLICY_ARN \
        --policy-document "$POLICY_DOCUMENT" \
        --set-as-default
    
    echo "Policy updated successfully"
else
    # Create new policy
    POLICY_ARN=$(aws iam create-policy \
        --policy-name $POLICY_NAME \
        --policy-document "$POLICY_DOCUMENT" \
        --description "Minimal permissions for Lingopaste application" \
        --query 'Policy.Arn' \
        --output text)
    
    echo "Policy created: $POLICY_ARN"
fi

# Attach policy to user
echo ""
echo "Attaching policy to user..."
aws iam attach-user-policy \
    --user-name $IAM_USER_NAME \
    --policy-arn $POLICY_ARN 2>/dev/null || echo "Policy already attached"

echo "Policy attached successfully"

# Create access keys
echo ""
echo "Creating access keys..."

# Check if user already has access keys
EXISTING_KEYS=$(aws iam list-access-keys --user-name $IAM_USER_NAME --query 'AccessKeyMetadata[].AccessKeyId' --output text)

if [ -n "$EXISTING_KEYS" ]; then
    echo "WARNING: User already has access keys:"
    echo "$EXISTING_KEYS"
    echo ""
    read -p "Do you want to create new access keys? (existing keys will remain active) [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping access key creation"
        echo ""
        echo "Setup complete!"
        echo "Use existing access keys for AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY"
        exit 0
    fi
fi

# Create new access key
KEY_OUTPUT=$(aws iam create-access-key --user-name $IAM_USER_NAME)
ACCESS_KEY_ID=$(echo $KEY_OUTPUT | jq -r '.AccessKey.AccessKeyId')
SECRET_ACCESS_KEY=$(echo $KEY_OUTPUT | jq -r '.AccessKey.SecretAccessKey')

echo ""
echo "=========================================="
echo "ACCESS KEYS CREATED SUCCESSFULLY"
echo "=========================================="
echo ""
echo "Add these to your backend/.env file:"
echo ""
echo "AWS_ACCESS_KEY_ID=$ACCESS_KEY_ID"
echo "AWS_SECRET_ACCESS_KEY=$SECRET_ACCESS_KEY"
echo "AWS_REGION=$AWS_REGION"
echo ""
echo "⚠️  IMPORTANT: Save these credentials now!"
echo "The secret access key will not be shown again."
echo "=========================================="
echo ""

# Save to a file (optional)
read -p "Save credentials to backend/.env.credentials file? [Y/n]: " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    cat > ../.env.credentials <<EOF
# IAM credentials for Lingopaste
# Created: $(date)
# User: $IAM_USER_NAME

AWS_ACCESS_KEY_ID=$ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY=$SECRET_ACCESS_KEY
AWS_REGION=$AWS_REGION
EOF
    echo "Credentials saved to backend/.env.credentials"
    echo "Remember to add this file to .gitignore!"
fi

echo ""
echo "Setup complete!"
