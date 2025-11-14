# Setup Scripts

## 1. Create IAM User (Run First)

Creates an IAM user with minimal permissions for the application:

```bash
./create-iam-user.sh
```

**Permissions granted:**
- DynamoDB: Read/write access to lingopaste tables only
- S3: Read/write access to lingopaste-data bucket only
- No other AWS permissions

**Output:** Access key ID and secret access key (save these!)

## 2. Create DynamoDB Tables

```bash
./setup-dynamodb.sh
```

Creates three tables:
- `lingopaste-accounts` - User accounts
- `lingopaste-pastes` - Paste metadata
- `lingopaste-rate-limits` - Rate limiting data (with TTL)

## 3. Create S3 Bucket

```bash
./create-s3-bucket.sh
```

Creates bucket with:
- Versioning enabled
- AES-256 encryption
- Public access blocked

## Complete Setup Example

```bash
cd backend/scripts

# 1. Create IAM user and get credentials
./create-iam-user.sh
# Copy the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY

# 2. Export credentials for next scripts
export AWS_ACCESS_KEY_ID="your_key_from_step_1"
export AWS_SECRET_ACCESS_KEY="your_secret_from_step_1"
export AWS_REGION="ap-northeast-1"

# 3. Create infrastructure
./setup-dynamodb.sh
./create-s3-bucket.sh

# 4. Add credentials to .env
cd ..
cp .env.example .env
# Edit .env and add the credentials from step 1
```

## Security Notes

- The IAM user has **minimal permissions** - only what's needed for the app
- Access keys are saved to `.env.credentials` (git-ignored)
- Never commit credentials to version control
- Rotate access keys periodically
- Consider using AWS IAM roles instead for production (EC2/ECS)

## Cleanup

To delete the IAM user:

```bash
# List access keys
aws iam list-access-keys --user-name lingopaste-app

# Delete access keys (do this for each key)
aws iam delete-access-key --user-name lingopaste-app --access-key-id YOUR_KEY_ID

# Detach policy
aws iam detach-user-policy \
  --user-name lingopaste-app \
  --policy-arn arn:aws:iam::ACCOUNT_ID:policy/lingopaste-app-policy

# Delete user
aws iam delete-user --user-name lingopaste-app

# Delete policy (optional)
aws iam delete-policy --policy-arn arn:aws:iam::ACCOUNT_ID:policy/lingopaste-app-policy
```
