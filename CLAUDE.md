# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A distributed URL shortener application built for testing Root Cause Analysis (RCA) tools. Features multiple microservices with deep integration chains, external API calls, caching layers, and various failure scenarios.

## Services Architecture

### 1. API Service (Go - Port 7543)
- **Location**: `api-service/`
- REST API for URL shortening and redirects
- Rate limiting via Redis middleware
- External HTTP calls for metadata fetching (async, 5s timeout)
- Framework: Gin
- Package structure: handlers, middleware, services, models, database, cache, config

### 2. Analytics Service (Python/FastAPI - Port 7544)
- **Location**: `analytics-service/`
- Aggregates click analytics
- Complex database queries with joins and time-series analysis
- Direct PostgreSQL queries (no caching layer)
- Framework: FastAPI with Uvicorn
- Module structure: routes, services, models, database, auth

### 3. Frontend (React - Port 7545)
- **Location**: `frontend/src/`
- Minimal UI for creating short URLs
- Real-time analytics dashboard
- Built with Vite
- Production: Nginx server on port 80

## Data Flow & Integration Chains

### Create Short URL Flow
```
Frontend → API Service → PostgreSQL (insert URL)
                      → External HTTP (fetch metadata - can timeout)
                      → PostgreSQL (update metadata)
                      → Redis (cache URL for 2 minutes)
```

### Click/Redirect Flow
```
Browser → API Service → Redis (L1 cache lookup - 2 minute TTL)
                      → PostgreSQL (fallback on cache miss)
                      → PostgreSQL (insert click record)
                      → Redis (increment click counter)
                      → HTTP 302 redirect
```

### Analytics Flow
```
Frontend → Analytics Service → PostgreSQL (direct complex queries)
                              → No caching (always fresh data)
```

**Key Architecture Notes**:
- API service metadata fetching is asynchronous (goroutine) - doesn't block URL creation
- Click recording in redirect flow is asynchronous - doesn't block redirect
- Analytics service uses RealDictCursor for JSON serialization
- Rate limiting middleware wraps only the `/api/urls` POST endpoint

## Development Commands

### Local Development Setup

**Database Setup**:
```bash
# Option 1: Use custom image with built-in schema (recommended)
docker run -d \
  --name url-shortener-db \
  -e POSTGRES_USER=urlshortener \
  -e POSTGRES_PASSWORD=password123 \
  -e POSTGRES_DB=urlshortener \
  -p 5432:5432 \
  rashadxyz/url-shortener-db:latest

# Option 2: Start PostgreSQL and manually initialize
docker run -d \
  --name url-shortener-db \
  -e POSTGRES_USER=urlshortener \
  -e POSTGRES_PASSWORD=password123 \
  -e POSTGRES_DB=urlshortener \
  -p 5432:5432 \
  postgres:15-alpine

# Initialize database with schema (only needed for Option 2)
docker exec -i url-shortener-db psql -U urlshortener -d urlshortener << 'EOF'
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    title VARCHAR(500),
    description TEXT,
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS clicks (
    id SERIAL PRIMARY KEY,
    url_id INTEGER REFERENCES urls(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(100),
    city VARCHAR(100),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
CREATE INDEX IF NOT EXISTS idx_clicks_url_id ON clicks(url_id);
CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks(clicked_at);
CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

INSERT INTO users (username, api_key)
VALUES ('testuser', 'test-api-key-12345')
ON CONFLICT (username) DO NOTHING;
EOF
```

**Redis Setup**:
```bash
docker run -d \
  --name url-shortener-redis \
  -p 6379:6379 \
  redis:7-alpine
```

### Database Service (PostgreSQL)

The database service uses a custom Docker image that extends `postgres:15-alpine` with the application schema and seed data.

**Build Docker Image**:
```bash
cd db-service
docker build -t rashadxyz/url-shortener-db .
```

**Run Container**:
```bash
docker run -d \
  --name url-shortener-db \
  -p 5432:5432 \
  -e POSTGRES_USER=urlshortener \
  -e POSTGRES_PASSWORD=password123 \
  -e POSTGRES_DB=urlshortener \
  rashadxyz/url-shortener-db
```

**Note**: The database schema and seed data (including the test user) are automatically initialized on first startup via the `init.sql` script in the image.

### API Service (Go)

**Local Development**:
```bash
cd api-service
go mod download
DATABASE_URL="postgres://urlshortener:password123@localhost:5432/urlshortener?sslmode=disable" \
REDIS_URL="localhost:6379" \
OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318" \
go run main.go
```

**Note**: `OTEL_EXPORTER_OTLP_ENDPOINT` is optional for local development. Omit if you don't have an OpenTelemetry collector running.

**Build Docker Image**:
```bash
cd api-service
docker build -t rashadxyz/url-shortener-api .
```

**Run Container**:
```bash
docker run -d \
  --name url-shortener-api \
  -p 7543:7543 \
  -e DATABASE_URL="postgres://urlshortener:password123@host.docker.internal:5432/urlshortener?sslmode=disable" \
  -e REDIS_URL="host.docker.internal:6379" \
  -e PORT=7543 \
  -e RATE_LIMIT_REQUESTS=100 \
  -e RATE_LIMIT_WINDOW=60 \
  -e OTEL_EXPORTER_OTLP_ENDPOINT="http://host.docker.internal:4318" \
  rashadxyz/url-shortener-api
```

### Analytics Service (Python)

**Local Development**:
```bash
cd analytics-service
pip install -r requirements.txt
DATABASE_URL="postgresql://urlshortener:password123@localhost:5432/urlshortener" \
PORT=7544 \
OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318" \
python main.py
```

**Note**: `OTEL_EXPORTER_OTLP_ENDPOINT` is optional for local development. Omit if you don't have an OpenTelemetry collector running.

**Build Docker Image**:
```bash
cd analytics-service
docker build -t rashadxyz/url-shortener-analytics .
```

**Run Container**:
```bash
docker run -d \
  --name url-shortener-analytics \
  -p 7544:7544 \
  -e DATABASE_URL="postgresql://urlshortener:password123@host.docker.internal:5432/urlshortener" \
  -e PORT=7544 \
  -e OTEL_EXPORTER_OTLP_ENDPOINT="http://host.docker.internal:4318" \
  rashadxyz/url-shortener-analytics
```

### Frontend (React)

**Local Development**:
```bash
cd frontend
npm install
npm run dev
```

**Build for Production**:
```bash
cd frontend
npm run build
```

**Preview Production Build**:
```bash
npm run preview
```

**Build Docker Image**:
```bash
cd frontend
docker build -t rashadxyz/url-shortener-frontend .

# Frontend uses relative URLs with nginx reverse proxy
# nginx.conf proxies:
#   /api/ → api-service:80
#   /api/analytics/ → analytics-service:80
# This works across all deployment environments (local, k8s, OpenChoreo)
```

**Run Container**:
```bash
docker run -d \
  --name url-shortener-frontend \
  -p 7545:80 \
  rashadxyz/url-shortener-frontend
```

## Docker Hub Images

Images are published to Docker Hub at:
- `rashadxyz/url-shortener-db:latest`
- `rashadxyz/url-shortener-api:latest`
- `rashadxyz/url-shortener-analytics:latest`
- `rashadxyz/url-shortener-frontend:latest`

### Build and Push All Images
```bash
# Login to Docker Hub
docker login

# Build and push all services
./build-and-push.sh

# Build with specific version
VERSION=v1.0.0 ./build-and-push.sh

# Build without cache (clean build)
./build-and-push.sh --no-cache
```

## Kubernetes Deployment

### Pull Images from Docker Hub
```bash
docker pull rashadxyz/url-shortener-db:latest
docker pull rashadxyz/url-shortener-api:latest
docker pull rashadxyz/url-shortener-analytics:latest
docker pull rashadxyz/url-shortener-frontend:latest
```

### For Minikube (Local Development)
```bash
# Use minikube's Docker daemon
eval $(minikube docker-env)

# Pull or build images in minikube's Docker daemon
docker pull rashadxyz/url-shortener-db:latest
docker pull rashadxyz/url-shortener-api:latest
docker pull rashadxyz/url-shortener-analytics:latest
docker pull rashadxyz/url-shortener-frontend:latest
```

### Deploy to Kubernetes
```bash
# Deploy all resources using Kustomize
kubectl apply -k manifests/

# Check status
kubectl get pods

# Access services (with port-forward)
kubectl port-forward svc/frontend 7545:80
kubectl port-forward svc/api-service 7543:80
kubectl port-forward svc/analytics-service 7544:80
```

### OpenChoreo Deployment

**IMPORTANT: Port Configuration**
- In OpenChoreo, `componentType: deployment/service` exposes services on **port 80** by default
- The `spec.parameters.port` value is the **targetPort** (container port)
- Inter-service communication uses `service-name:80` (e.g., `postgres:80`, `redis:80`, `api-service:80`)
- Frontend uses nginx reverse proxy to handle service communication (no build-time configuration needed)

**Deploy OpenChoreo Manifests**:
```bash
# Recommended: Deploy using Kustomize (see Kustomize Deployment section below)
kubectl apply -k manifests/

# Or deploy base manifests directly (without Kustomize)
kubectl apply -f manifests/base/url-shortener-demo-project.yaml
kubectl apply -f manifests/base/postgres-component.yaml
kubectl apply -f manifests/base/redis-component.yaml
kubectl apply -f manifests/base/api-service-component.yaml
kubectl apply -f manifests/base/analytics-service-component.yaml
kubectl apply -f manifests/base/frontend-component.yaml
```

**Port Mapping in OpenChoreo**:
| Service | Service Port | Container Port (targetPort) |
|---------|-------------|----------------------------|
| postgres | 80 | 5432 |
| redis | 80 | 6379 |
| api-service | 80 | 7543 |
| analytics-service | 80 | 7544 |
| frontend | 80 | 80 |

### Kustomize Deployment (Recommended)

The manifests directory uses Kustomize to manage all configuration from a single place. This eliminates hardcoded values and makes it easy to customize deployments.

**Quick Start**:
```bash
# Deploy with default configuration
kubectl apply -k manifests/

# Preview generated manifests first
kubectl kustomize manifests/

# Or build to a file
kubectl kustomize manifests/ > deployment.yaml
kubectl apply -f deployment.yaml
```

**Customize Configuration**:
All variables are defined in `manifests/vars.yaml`. Edit this file to change:
- Image tags: `IMAGE_TAG: "v1.0.0"`
- Database credentials: `DB_PASSWORD: "your-password"`
- Resource limits: `API_MEMORY_LIMIT: "512Mi"`
- Replica counts: `API_SERVICE_REPLICAS: "3"`
- Ports, environment variables, and more

**Example - Update Image Tag**:
```bash
# Edit manifests/vars.yaml
# Change: IMAGE_TAG: "demo"
# To:     IMAGE_TAG: "v2.0.0"

# Apply changes
kubectl apply -k manifests/
```

**Example - Change Namespace**:
```bash
# Edit manifests/vars.yaml
# Change: NAMESPACE: "default"
# To:     NAMESPACE: "production"

# Apply to new namespace
kubectl apply -k manifests/
```

See `manifests/README.md` for detailed documentation on all configurable variables and advanced usage.

### View Logs
```bash
kubectl logs -n url-shortener -l app=api-service -f
kubectl logs -n url-shortener -l app=analytics-service -f
kubectl logs -n url-shortener -l app=frontend -f
```

### Cleanup
```bash
kubectl delete namespace url-shortener
```

## Database Schema

### Tables
- **users**: User accounts with API keys
- **urls**: Short URLs with metadata (title, description)
- **clicks**: Click tracking (ip_address, user_agent, referer)

### Key Indexes
- `idx_urls_short_code` on `urls(short_code)` - Critical for redirect performance
- `idx_clicks_url_id` on `clicks(url_id)` - For analytics queries
- `idx_clicks_clicked_at` on `clicks(clicked_at)` - For time-series queries
- `idx_users_api_key` on `users(api_key)` - For authentication

## API Endpoints

### API Service (Port 7543)
- `POST /api/urls` - Create short URL (requires username in body)
- `GET /api/urls?username={username}` - List URLs for user
- `GET /{code}` - Redirect to long URL (tracks click)
- `GET /health` - Health check (DB + Redis)

### Analytics Service (Port 7544)
- `GET /api/analytics/summary?username={username}` - Overall stats
- `GET /api/analytics/top-urls?username={username}&limit=10` - Top URLs by clicks
- `GET /api/analytics/time-series?username={username}&days=7` - Time series data
- `GET /api/analytics/url/{url_id}?username={username}` - Detailed URL analytics
- `GET /health` - Health check (DB only)

**Default Username**: `testuser`

## Caching Strategy

### Redis Cache Keys (API Service Only)
- `url:{short_code}` - Cached URL lookups (2 minute TTL)
- `rate_limit:{api_key}` - Rate limit counters with TTL
- `clicks:{short_code}` - Click counters (no expiration)

### Cache Behavior
- API service: Falls back to PostgreSQL on Redis miss for URL lookups
- API service: Stores URL cache for 2 minutes (120 seconds)
- Analytics service: No caching - direct PostgreSQL queries
- Rate limiting: Degrades gracefully on Redis failure

## Service Dependencies

### API Service
- **Required**: PostgreSQL (hard dependency)
- **Optional**: Redis (graceful degradation for caching and rate limiting)
- **External**: Metadata fetching API (timeouts handled)

### Analytics Service
- **Required**: PostgreSQL (hard dependency)

### Frontend
- **Required**: API Service (for URL creation/listing)
- **Required**: Analytics Service (for dashboard)

## Environment Variables

### API Service
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `PORT` - Server port (default: 7543)
- `RATE_LIMIT_REQUESTS` - Max requests per window (default: 5)
- `RATE_LIMIT_WINDOW` - Rate limit window in seconds (default: 60)
- `CACHE_TTL` - Cache TTL in seconds (default: 120)
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector endpoint (default: http://opentelemetry-collector.openchoreo-observability-plane.svc.cluster.local:4318)

### Analytics Service
- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - Server port (default: 7544)
- `OTEL_EXPORTER_OTLP_ENDPOINT` - OpenTelemetry collector endpoint (default: http://opentelemetry-collector.openchoreo-observability-plane.svc.cluster.local:4318)

### Frontend
- No environment variables needed
- Uses nginx reverse proxy with relative URLs for backend communication

## OpenTelemetry Instrumentation

Both API and Analytics services are instrumented with OpenTelemetry for distributed tracing.

### API Service (Go)
- Uses `otelgin` middleware for automatic Gin instrumentation
- Traces all HTTP requests including handlers and middleware
- Exports traces via OTLP HTTP to the collector endpoint

### Analytics Service (Python)
- Uses `opentelemetry-instrumentation-fastapi` for automatic FastAPI instrumentation
- Traces all HTTP requests and responses
- Exports traces via OTLP HTTP to the collector endpoint

### Configuration
- Set `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable to point to your OpenTelemetry collector
- Default: `http://opentelemetry-collector.openchoreo-observability-plane.svc.cluster.local:4318`
- For cross-namespace access, use fully qualified DNS: `<service>.<namespace>.svc.cluster.local`
- Traces are exported automatically with service names: `api-service` and `analytics-service`
- Context propagation is enabled for distributed tracing across services

## Failure Scenarios (RCA Testing)

This application is designed to test RCA tools with these failure modes:

1. **Cache Failures**: Stop Redis to observe slower redirects, DB fallback
2. **Database Connection Issues**: Connection pool exhaustion, query timeouts
3. **External API Timeouts**: Metadata fetching timeouts
4. **Rate Limiting**: Trigger 429 responses by exceeding limits
5. **Cache Inconsistency**: Stale data between Redis and PostgreSQL (2 minute TTL)
6. **Resource Exhaustion**: Connection pool depletion, memory leaks
7. **Service Dependencies**: Cascading failures through the stack

## Code Architecture Notes

### API Service (Go)
**Structure**:
- `main.go` - Entry point, router setup, middleware registration
- `handlers/` - HTTP request handlers (url.go, health.go)
  - `url.go:CreateURL` - Creates short URL, async metadata fetch
  - `url.go:Redirect` - Cache-first redirect with async click recording
  - `url.go:ListURLs` - Lists user's URLs
- `middleware/` - Gin middleware (cors.go, ratelimit.go)
- `services/` - Business logic (shortcode.go, metadata.go)
  - `metadata.go:FetchAndUpdateMetadata` - 5s timeout HTTP client
- `database/` - Database connection and queries
- `cache/` - Redis operations with TTL management
- `models/` - Request/response structs
- `config/` - Environment variable loading

**Key Patterns**:
- Async operations via goroutines (metadata fetch, click recording)
- Cache-first pattern with database fallback
- Rate limiting only on POST /api/urls
- Health check verifies both DB and Redis

### Analytics Service (Python)
**Structure**:
- `main.py` - FastAPI app, lifespan management, CORS
- `routes/` - API endpoints (analytics.py, health.py)
- `services/analytics_service.py` - Business logic with complex SQL queries
  - Time-series aggregation queries
  - JOIN queries across urls and clicks tables
- `database.py` - Connection pooling with context managers
- `models.py` - Pydantic models for validation
- `auth.py` - API key validation

**Key Patterns**:
- RealDictCursor for automatic JSON serialization
- Context managers for database connections
- No caching - all queries hit PostgreSQL directly
- Lifespan context for startup/shutdown

### Frontend (React + Vite)
**Structure**:
- `src/main.jsx` - React app entry point
- `src/App.jsx` - Main component with all UI logic
- `vite.config.js` - Build configuration
- `nginx.conf` - Production server configuration
- `Dockerfile` - Multi-stage build (build → nginx)

**Key Patterns**:
- No state management library (useState, useEffect only)
- Direct fetch() calls using relative URLs (e.g., '/api/urls')
- Nginx reverse proxy handles backend routing
- nginx.conf proxies /api/ to api-service:80 and /api/analytics/ to analytics-service:80
- Nginx serves production build on port 80
