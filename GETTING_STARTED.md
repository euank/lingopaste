# Getting Started with Lingopaste

## Quick Start

### Prerequisites

1. **Go 1.21+** - Backend language
2. **Node.js 20+** - Frontend development
3. **AWS Account** with:
   - DynamoDB access
   - S3 access
   - IAM credentials configured
4. **OpenAI API Key** - For translations
5. **Docker** (optional) - For containerized deployment

### Initial Setup

#### 1. AWS Infrastructure Setup

```bash
cd backend/scripts

# Set your AWS region
export AWS_REGION=ap-northeast-1

# Create DynamoDB tables
./setup-dynamodb.sh

# Create S3 bucket
./create-s3-bucket.sh
```

#### 2. Backend Setup

```bash
cd backend

# Create environment file
cp .env.example .env

# Edit .env with your credentials
# Required:
#   - AWS credentials
#   - OpenAI API key
#   - JWT secret (generate with: openssl rand -base64 32)

# Install dependencies
go mod download

# Run the server
make run
```

The backend will start on http://localhost:8080

#### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will start on http://localhost:5173

### Using Docker Compose (Alternative)

```bash
# Create .env file in project root
cp backend/.env.example .env
# Edit .env with your credentials

# Start all services
docker-compose up

# Backend: http://localhost:8080
# Frontend: http://localhost:5173
```

## Testing the Application

### 1. Create a Paste

```bash
curl -X POST http://localhost:8080/api/pastes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Hello, world! This is a test paste.",
    "tone": "default"
  }'

# Response:
# {
#   "paste_id": "abc12345",
#   "original_language": "en",
#   "available_languages": ["en"]
# }
```

### 2. View a Paste

```bash
curl http://localhost:8080/api/pastes/abc12345

# Response includes:
# - Original text
# - All cached translations
# - Metadata
```

### 3. Translate to Another Language

```bash
curl http://localhost:8080/api/pastes/abc12345/translate?lang=es

# Response:
# {
#   "language": "es",
#   "translation": "¡Hola, mundo! Esto es una pasta de prueba."
# }
```

### 4. Test via Browser

1. Open http://localhost:5173
2. Enter some text
3. Select a tone (default, professional, friendly, brusque)
4. Click "Create Paste"
5. You'll be redirected to the paste view
6. Select different languages from the dropdown
7. Toggle between translation/original/side-by-side views

## Rate Limiting

The application has built-in rate limiting:

- **Anonymous users**: 5 pastes per day (per IP)
- **Logged-in users** (not implemented yet): 5 pastes per day + 50 per IP
- **Paid users** (not implemented yet): 1000 pastes per day

Test rate limiting:
```bash
# Create 6 pastes quickly - the 6th should fail with 429 error
for i in {1..6}; do
  curl -X POST http://localhost:8080/api/pastes \
    -H "Content-Type: application/json" \
    -d "{\"content\": \"Test $i\", \"tone\": \"default\"}"
  echo
done
```

## Deployment to Kubernetes

### Prerequisites

- Kubernetes cluster configured
- `kubectl` installed and configured
- Docker images built and pushed to registry

### Build and Push Images

```bash
# Build backend
cd backend
docker build -t your-registry/lingopaste-backend:v1.0 .
docker push your-registry/lingopaste-backend:v1.0

# Build frontend
cd ../frontend
docker build -t your-registry/lingopaste-frontend:v1.0 .
docker push your-registry/lingopaste-frontend:v1.0
```

### Update Kubernetes Manifests

Edit `k8s/deployment.yaml` and update image references:

```yaml
image: your-registry/lingopaste-backend:v1.0
image: your-registry/lingopaste-frontend:v1.0
```

### Create Secrets

```bash
cd k8s
cp secrets.yaml.example secrets.yaml
# Edit secrets.yaml with your actual credentials
```

### Deploy

```bash
cd k8s
./deploy.sh

# Check status
kubectl get pods -n lingopaste
kubectl get svc -n lingopaste
```

## Architecture Overview

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       ▼
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Frontend  │──────▶│   Backend   │──────▶│  DynamoDB   │
│   (React)   │       │    (Go)     │       │  (Metadata) │
└─────────────┘       └──────┬──────┘       └─────────────┘
                             │
                    ┌────────┼────────┐
                    ▼        ▼        ▼
              ┌─────────┬─────────┬─────────┐
              │   S3    │ OpenAI  │  Cache  │
              │ (Pastes)│ (Trans) │  (LRU)  │
              └─────────┴─────────┴─────────┘
```

## Cost Estimation

### Per Translation
- OpenAI gpt-4o-mini: ~$0.0002 per translation
- DynamoDB: ~$0.000001 per read/write
- S3: ~$0.000001 per request
- **Total: ~$0.0002 per translation**

### Monthly (1M translations)
- OpenAI: ~$200
- DynamoDB: ~$25 (5 RCU/WCU)
- S3: ~$25 (storage + requests)
- **Total: ~$250/month**

### Revenue (100 paid users @ $5/mo)
- **$500/month**
- **Profit: $250/month**

## Next Steps

1. **Implement Authentication**
   - Google OAuth
   - Apple OAuth
   - Email authentication
   - JWT token management

2. **Implement Stripe Payments**
   - Subscription checkout
   - Webhook handlers
   - Billing portal

3. **Add Monitoring**
   - Prometheus metrics
   - Grafana dashboards
   - Error tracking (Sentry)

4. **Performance Optimization**
   - CDN for frontend
   - Database indexing
   - Cache warming strategies

5. **Additional Features**
   - Paste expiration
   - Private pastes (password protected)
   - Custom URLs
   - Paste editing history
   - Export to PDF/DOCX

## Troubleshooting

### Backend won't start
- Check `.env` file has all required variables
- Verify AWS credentials are correct
- Ensure DynamoDB tables exist
- Check S3 bucket exists and is accessible

### Frontend can't connect to backend
- Verify backend is running on port 8080
- Check CORS settings in backend
- Ensure `FRONTEND_URL` is set correctly

### Translations failing
- Verify OpenAI API key is valid
- Check API quota/billing
- Look at backend logs for errors

### Rate limiting not working
- Check DynamoDB table has TTL enabled
- Verify IP extraction middleware is active
- Check system date/time is correct

## Support

For issues or questions:
1. Check logs: `docker-compose logs` or `kubectl logs -n lingopaste`
2. Verify AWS services are accessible
3. Check OpenAI API status
4. Review environment variables

## License

Proprietary - All rights reserved
