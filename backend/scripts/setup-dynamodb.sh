#!/usr/bin/env bash

# Script to create DynamoDB tables for Lingopaste
# Usage: ./setup-dynamodb.sh

set -e

AWS_REGION=${AWS_REGION:-ap-northeast-1}

echo "Creating DynamoDB tables in region: $AWS_REGION"

# Helper function to check if table exists
table_exists() {
    aws dynamodb describe-table --table-name "$1" --region "$AWS_REGION" &>/dev/null
    return $?
}

# Create Accounts table
if table_exists lingopaste-accounts; then
    echo "Accounts table already exists, skipping..."
else
    echo "Creating accounts table..."
    aws dynamodb create-table \
    --table-name lingopaste-accounts \
    --attribute-definitions \
        AttributeName=email,AttributeType=S \
        AttributeName=account_id,AttributeType=S \
    --key-schema \
        AttributeName=email,KeyType=HASH \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"account_id-index\",
                \"KeySchema\": [{\"AttributeName\":\"account_id\",\"KeyType\":\"HASH\"}],
                \"Projection\":{\"ProjectionType\":\"ALL\"},
                \"ProvisionedThroughput\":{\"ReadCapacityUnits\":5,\"WriteCapacityUnits\":5}
            }
        ]" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region $AWS_REGION
fi

# Create Pastes table
if table_exists lingopaste-pastes; then
    echo "Pastes table already exists, skipping..."
else
    echo "Creating pastes table..."
    aws dynamodb create-table \
    --table-name lingopaste-pastes \
    --attribute-definitions \
        AttributeName=paste_id,AttributeType=S \
        AttributeName=creator_account_id,AttributeType=S \
        AttributeName=created_at,AttributeType=N \
    --key-schema \
        AttributeName=paste_id,KeyType=HASH \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"creator_account_id-created_at-index\",
                \"KeySchema\": [
                    {\"AttributeName\":\"creator_account_id\",\"KeyType\":\"HASH\"},
                    {\"AttributeName\":\"created_at\",\"KeyType\":\"RANGE\"}
                ],
                \"Projection\":{\"ProjectionType\":\"ALL\"},
                \"ProvisionedThroughput\":{\"ReadCapacityUnits\":5,\"WriteCapacityUnits\":5}
            }
        ]" \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region $AWS_REGION
fi

# Create Rate Limits table
if table_exists lingopaste-rate-limits; then
    echo "Rate limits table already exists, skipping..."
else
    echo "Creating rate limits table..."
    aws dynamodb create-table \
    --table-name lingopaste-rate-limits \
    --attribute-definitions \
        AttributeName=identifier,AttributeType=S \
        AttributeName=date,AttributeType=S \
    --key-schema \
        AttributeName=identifier,KeyType=HASH \
        AttributeName=date,KeyType=RANGE \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --region $AWS_REGION
fi

# Enable TTL on rate limits table (if not already enabled)
if table_exists lingopaste-rate-limits; then
    echo "Checking TTL status on rate limits table..."
    TTL_STATUS=$(aws dynamodb describe-time-to-live \
        --table-name lingopaste-rate-limits \
        --region $AWS_REGION \
        --query 'TimeToLiveDescription.TimeToLiveStatus' \
        --output text 2>/dev/null || echo "")
    
    if [ "$TTL_STATUS" != "ENABLED" ]; then
        echo "Enabling TTL on rate limits table..."
        aws dynamodb update-time-to-live \
            --table-name lingopaste-rate-limits \
            --time-to-live-specification \
                "Enabled=true,AttributeName=ttl" \
            --region $AWS_REGION
    else
        echo "TTL already enabled on rate limits table"
    fi
fi

echo "All tables created successfully!"
echo ""
echo "Table names:"
echo "  - lingopaste-accounts"
echo "  - lingopaste-pastes"
echo "  - lingopaste-rate-limits"
