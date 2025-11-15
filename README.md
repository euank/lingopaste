# Lingopaste.com

A web application for sharing text snippets with built-in LLM-based translation capabilities.

Note, this project is almost entirely vibe coded. A human is reviewing the code.

## Features

- Share text snippets via links
- Automatic translation to any language using AI
- Rate limiting (5 pastes/day free, 1000 pastes/day for $5/mo)
- Multiple authentication methods (Google, Apple, Email)
- Translation tone control (Default, Professional, Friendly, Brusque)
- Responsive UI with side-by-side viewing

## Tech Stack

### Backend
- **Language**: Go 1.21
- **Framework**: Gorilla Mux
- **Database**: AWS DynamoDB
- **Storage**: AWS S3
- **LLM**: OpenAI API (gpt-4o-mini)
- **Payments**: Stripe
- **Auth**: JWT + OAuth2

### Frontend
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Routing**: React Router
- **Styling**: CSS Modules

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Cloud**: AWS

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 20+
- AWS Account with credentials
- OpenAI API key
- Stripe account (for payments)

### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env
# Edit .env with your credentials

# Run the server
make run
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

### Using Docker Compose

```bash
# Create .env file in project root with required variables
cp backend/.env.example .env

# Start all services
docker-compose up

# Backend will be available at http://localhost:8080
# Frontend will be available at http://localhost:5173
```

## Project Structure

```
lingopaste/
├── backend/
│   ├── cmd/server/          # Main application entry point
│   ├── internal/
│   │   ├── auth/            # Authentication logic
│   │   ├── cache/           # LRU cache implementation
│   │   ├── config/          # Configuration management
│   │   ├── db/              # DynamoDB operations
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/      # HTTP middleware
│   │   ├── models/          # Data models
│   │   ├── payments/        # Stripe integration
│   │   ├── storage/         # S3 operations
│   │   └── translate/       # OpenAI translation
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/      # Reusable components
│   │   ├── pages/           # Page components
│   │   └── api/             # API client
│   └── Dockerfile
├── k8s/                     # Kubernetes manifests
└── docker-compose.yml
```

## API Endpoints

- `POST /api/pastes` - Create new paste
- `GET /api/pastes/:id` - Get paste with translations
- `GET /api/pastes/:id/translate?lang=:lang` - Translate to specific language
- `POST /api/auth/google` - Google OAuth
- `POST /api/auth/apple` - Apple OAuth
- `POST /api/auth/email` - Email auth
- `GET /api/account/me` - Get account info
- `POST /api/payment/create-checkout` - Create Stripe checkout
- `POST /api/payment/webhook` - Stripe webhook handler
- `GET /health` - Health check

## Environment Variables

See `backend/.env.example` for all required environment variables.

## Deployment

The application is containerized and ready for Kubernetes deployment.

```bash
# Build Docker images
docker build -t lingopaste-backend ./backend
docker build -t lingopaste-frontend ./frontend

# Apply Kubernetes manifests
kubectl apply -f k8s/
```

## License

AGPL
