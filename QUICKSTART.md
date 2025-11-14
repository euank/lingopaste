# Quick Start

## Setup (Already Done ✓)
- ✓ DynamoDB tables created
- ✓ S3 bucket created

## Run the Application

### 1. Configure Environment

```bash
cd backend
cp .env.example .env
```

Edit `.env` and add your credentials:
```bash
# Required
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
OPENAI_API_KEY=sk-your-openai-key
JWT_SECRET=$(openssl rand -base64 32)

# Optional (for auth/payments - can skip for now)
# GOOGLE_CLIENT_ID=...
# STRIPE_SECRET_KEY=...
```

### 2. Run Backend

```bash
cd backend

# Install Go dependencies
go mod download

# Run the server
go run ./cmd/server/main.go
```

Backend will start on **http://localhost:8080**

### 3. Run Frontend (in a new terminal)

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

Frontend will start on **http://localhost:5173**

### 4. Test It!

Open **http://localhost:5173** in your browser:

1. Type some text (e.g., "Hello world, this is a test!")
2. Select a tone
3. Click "Create Paste"
4. You'll be redirected to the paste view
5. Select different languages from the dropdown
6. Watch it translate in real-time!

## Testing via API

```bash
# Create a paste
curl -X POST http://localhost:8080/api/pastes \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, world!", "tone": "default"}'

# Response: {"paste_id":"abc12345",...}

# View the paste
curl http://localhost:8080/api/pastes/abc12345

# Translate to Spanish
curl http://localhost:8080/api/pastes/abc12345/translate?lang=es
```

## Quick Makefile Commands

```bash
cd backend

# Run the server
make run

# Build binary
make build

# Run tests
make test
```

## Troubleshooting

**Backend won't start?**
- Check `.env` has all required variables
- Verify AWS credentials: `aws sts get-caller-identity`
- Verify OpenAI key is valid

**Can't connect to backend?**
- Backend must be running on port 8080
- Check: `curl http://localhost:8080/health`

**Translations failing?**
- Check OpenAI API key
- Check OpenAI account has credits
- Look at backend logs

**CORS errors?**
- Make sure frontend is on port 5173
- Check backend FRONTEND_URL setting

## What's Working

✅ Create pastes  
✅ Auto language detection  
✅ On-demand translation  
✅ Multiple language support  
✅ Translation caching (S3 + memory)  
✅ Rate limiting (5 pastes/day per IP)  
✅ Responsive UI with 3 view modes

## What's Not Implemented Yet

⏳ Authentication (Google/Apple/Email OAuth)  
⏳ Paid subscriptions (Stripe)  
⏳ User accounts & paste history  
⏳ Account dashboard

These can be added later - the core paste/translate functionality is fully working!
