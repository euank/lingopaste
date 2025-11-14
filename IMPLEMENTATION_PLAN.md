# Lingopaste.com - Implementation Plan

## Tech Stack Recommendation

### Backend: **Go**
- ✅ Excellent for responsive APIs
- ✅ Best-in-class AWS SDK (DynamoDB, S3)
- ✅ Strong LLM code generation support
- ✅ Built-in concurrency for handling translations
- ✅ Easy HTTP server with great performance

### Frontend: **React + TypeScript**
- Modern, responsive SPA
- Easy language switching without page reloads
- Strong typing for API contracts

### Infrastructure
- **Storage**: AWS S3 (pastes + translations)
- **Database**: AWS DynamoDB (accounts, metadata, rate limits)
- **LLM**: OpenAI API (gpt-4o-mini for cost efficiency)
- **Payments**: Stripe
- **Deployment**: Docker → Kubernetes

---

## 1. Database Schema (DynamoDB)

### Table: `accounts`
```
PK: email/oauth_id (String)
SK: "ACCOUNT" (String)
---
Attributes:
- account_id (String, UUID)
- auth_provider (String: "google" | "apple" | "email")
- email (String)
- is_paid (Boolean)
- stripe_customer_id (String, nullable)
- stripe_subscription_id (String, nullable)
- created_at (Number, Unix timestamp)
- updated_at (Number, Unix timestamp)

GSI: account_id-index (for lookups by UUID)
```

### Table: `pastes`
```
PK: paste_id (String, short alphanumeric like "abc123")
SK: "META" (String)
---
Attributes:
- paste_id (String)
- original_language (String, detected)
- tone (String: "default" | "professional" | "friendly" | "brusque")
- creator_ip (String, hashed for privacy)
- creator_account_id (String, nullable)
- created_at (Number, Unix timestamp)
- character_count (Number)
- available_translations (StringSet) // e.g., ["en", "es", "fr"]

GSI: creator_account_id-created_at-index (for user paste history)
```

### Table: `rate_limits`
```
PK: identifier (String) // IP hash or account_id
SK: date (String) // YYYY-MM-DD
---
Attributes:
- paste_count (Number)
- limit_type (String: "anonymous_ip" | "account" | "ip_total")
- ttl (Number, Unix timestamp, 48 hours from date for auto-cleanup)
```

---

## 2. S3 Storage Structure

```
bucket: lingopaste-data/

pastes/
  {paste_id}/
    original.txt           // Original paste content
    meta.json             // Redundant metadata backup
    translations/
      en.txt
      es.txt
      fr.txt
      ...
```

**S3 Settings:**
- Versioning: Enabled (for safety)
- Lifecycle: None (keep forever, cheap storage)
- Encryption: AES-256

---

## 3. API Endpoints

### Backend API (Go)

```
POST   /api/pastes                  # Create new paste
GET    /api/pastes/:id              # Get paste metadata + all cached translations
GET    /api/pastes/:id/translate    # Translate to specific language (on-demand)
POST   /api/auth/google             # Google OAuth callback
POST   /api/auth/apple              # Apple OAuth callback
POST   /api/auth/email              # Email login/signup
POST   /api/auth/logout             # Logout
GET    /api/account/me              # Get current user info
POST   /api/payment/create-checkout # Create Stripe checkout session
POST   /api/payment/webhook         # Stripe webhook handler
GET    /health                      # Health check
```

---

## 4. Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
- [x] Set up Go project structure
- [x] Configure AWS SDK (S3 + DynamoDB clients)
- [x] Create DynamoDB tables with schemas
- [x] Create S3 bucket with proper permissions
- [x] Implement basic HTTP server with routing
- [x] Add CORS middleware for frontend
- [x] Set up environment variable configuration

### Phase 2: Paste Creation & Translation (Week 2)
- [x] Implement OpenAI API client wrapper
- [x] Create paste creation endpoint with character limit (20k)
- [x] Implement language detection (OpenAI or dedicated API)
- [x] Build translation service with tone support
- [x] Implement S3 upload for original + translations
- [x] Add DynamoDB metadata storage
- [x] Build read-through cache (LRU, 100k items)
- [x] Create paste retrieval endpoint

### Phase 3: Authentication (Week 3)
- [x] Implement Google OAuth flow
- [x] Implement Apple OAuth flow
- [x] Implement email-based auth (with verification)
- [x] Create JWT session management
- [x] Build account management endpoints
- [x] Add middleware for auth checking

### Phase 4: Rate Limiting (Week 4)
- [x] Implement IP extraction middleware
- [x] Build rate limit checker service
  - Anonymous: 5/day per IP
  - Logged in free: 5/day per account + 50/day per IP
  - Paid: 1000/day per account
- [x] Add DynamoDB-based counter with atomic increments
- [x] Implement rate limit response headers
- [x] Add cleanup job for expired rate limit records

### Phase 5: Payments (Week 5)
- [x] Set up Stripe account and API keys
- [x] Implement Stripe checkout session creation
- [x] Build webhook handler for subscription events
- [x] Update account status on successful payment
- [x] Handle subscription cancellations
- [x] Add billing portal integration

### Phase 6: Frontend (Week 6-7)
- [x] Set up React + TypeScript + Vite
- [x] Create paste creation page
  - Text editor (20k char limit, live counter)
  - Tone selector dropdown
  - Anonymous/authenticated state handling
- [x] Create paste viewing page
  - Display translated text based on Accept-Language
  - Language selector dropdown (responsive, no reload)
  - Original/Translation/Side-by-side tabs
  - "Machine translation" disclaimer header
- [x] Build authentication UI
  - Login modal (Google, Apple, Email)
  - Account dashboard
  - Upgrade to paid button
- [x] Add rate limit UI feedback
- [x] Implement client-side caching

### Phase 7: Deployment (Week 8)
- [x] Create optimized Dockerfile (multi-stage build)
- [x] Add docker-compose for local development
- [x] Create Kubernetes manifests
  - Deployment
  - Service
  - ConfigMap for env vars
  - Secret for API keys
  - HPA (Horizontal Pod Autoscaler)
- [x] Set up health checks
- [x] Configure logging (structured JSON logs)
- [x] Add monitoring hooks (Prometheus metrics)

### Phase 8: Polish & Testing (Week 9)
- [x] End-to-end testing
- [x] Load testing for rate limits
- [x] Security audit (OWASP top 10)
- [x] Performance optimization
- [x] Error handling polish
- [x] Documentation (API docs, deployment guide)

---

## 5. Key Technical Decisions

### Translation Flow
1. User creates paste → Store original in S3
2. Detect original language via OpenAI
3. User views paste → Check Accept-Language header
4. If translation exists in cache/S3 → Return immediately
5. If not → Trigger translation, store in S3, update DynamoDB, return to user
6. Subsequent requests → Serve from cache

### Caching Strategy
- **In-memory LRU cache** (100k items) for paste metadata + translations
- Cache key: `{paste_id}:{language_code}`
- Eviction: Least recently used
- Preload: On-demand only (no warming)

### Rate Limiting Logic
```go
func checkRateLimit(ip, accountID string, isPaid bool) error {
    if isPaid {
        return checkAccountLimit(accountID, 1000)
    }
    
    if accountID != "" {
        // Logged in, free account
        if err := checkAccountLimit(accountID, 5); err != nil {
            return err
        }
        return checkIPLimit(ip, 50) // IP-wide limit
    }
    
    // Anonymous
    return checkIPLimit(ip, 5)
}
```

### OpenAI Prompt Template
```
System: You are a professional translator. Translate the following text to {target_language}.
Tone: {tone_instruction}
- Default: Natural and accurate
- Professional: Formal business language
- Friendly: Warm and conversational
- Brusque: Direct and concise

Preserve formatting, but translate all content.

User: {original_text}
```

### Cost Optimization
- Use **gpt-4o-mini** ($0.15/1M input, $0.60/1M output tokens)
- Average paste: ~500 chars = ~125 tokens
- Average translation: ~125 input + ~125 output = 250 tokens total
- Cost per translation: ~$0.0002
- Monthly with 1M translations: ~$200
- Revenue with 100 paid users: $500/mo (profitable)

---

## 6. Environment Variables

```env
# AWS
AWS_REGION=ap-northeast-1
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
S3_BUCKET_NAME=lingopaste-data
DYNAMODB_ACCOUNTS_TABLE=accounts
DYNAMODB_PASTES_TABLE=pastes
DYNAMODB_RATE_LIMITS_TABLE=rate_limits

# OpenAI
OPENAI_API_KEY=...
OPENAI_MODEL=gpt-4o-mini

# Auth
JWT_SECRET=...
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
APPLE_CLIENT_ID=...
APPLE_CLIENT_SECRET=...
FRONTEND_URL=https://lingopaste.com

# Stripe
STRIPE_SECRET_KEY=...
STRIPE_WEBHOOK_SECRET=...
STRIPE_PRICE_ID=... # For $5/mo subscription

# Server
PORT=8080
CACHE_SIZE=100000
MAX_PASTE_LENGTH=20000
```

---

## 7. File Structure

```
lingopaste/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   ├── google.go
│   │   │   ├── apple.go
│   │   │   └── email.go
│   │   ├── cache/
│   │   │   └── lru.go
│   │   ├── db/
│   │   │   ├── dynamodb.go
│   │   │   ├── accounts.go
│   │   │   ├── pastes.go
│   │   │   └── rate_limits.go
│   │   ├── handlers/
│   │   │   ├── pastes.go
│   │   │   ├── auth.go
│   │   │   └── payments.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── ratelimit.go
│   │   │   └── cors.go
│   │   ├── storage/
│   │   │   └── s3.go
│   │   ├── translate/
│   │   │   ├── openai.go
│   │   │   └── language.go
│   │   └── payments/
│   │       └── stripe.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── PasteCreator.tsx
│   │   │   ├── PasteViewer.tsx
│   │   │   ├── LanguageSelector.tsx
│   │   │   ├── AuthModal.tsx
│   │   │   └── Header.tsx
│   │   ├── pages/
│   │   │   ├── Home.tsx
│   │   │   ├── View.tsx
│   │   │   └── Account.tsx
│   │   ├── api/
│   │   │   └── client.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── k8s/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   └── hpa.yaml
├── docker-compose.yml
└── README.md
```

---

## 8. Security Considerations

1. **Input Validation**: Sanitize all user input (XSS prevention)
2. **Rate Limiting**: Prevent abuse via DynamoDB atomic counters
3. **Authentication**: Secure JWT with short expiry + refresh tokens
4. **API Keys**: Never expose in frontend, use env vars
5. **CORS**: Restrict to frontend domain only
6. **S3 Permissions**: Private buckets, signed URLs if needed
7. **SQL Injection**: N/A (DynamoDB, but validate all inputs)
8. **Stripe Webhooks**: Verify signatures
9. **IP Hashing**: Hash IPs before storing for privacy
10. **HTTPS Only**: Enforce in production

---

## 9. Monitoring & Observability

- **Metrics**: Request count, latency, error rates, cache hit ratio
- **Logging**: Structured JSON logs (request ID, user ID, errors)
- **Alerts**: High error rate, translation failures, payment issues
- **Tracing**: OpenTelemetry for distributed tracing
- **Health Checks**: `/health` endpoint (DB + S3 connectivity)

---

## 10. Next Steps

1. **Choose**: Confirm tech stack (Go + React)
2. **Scaffold**: Initialize project structure
3. **Develop**: Follow phases 1-8 sequentially
4. **Deploy**: Push to K8s cluster
5. **Monitor**: Set up observability
6. **Iterate**: Gather feedback, optimize costs

**Estimated Timeline**: 9 weeks to MVP
**Team Size**: 1-2 full-stack engineers
