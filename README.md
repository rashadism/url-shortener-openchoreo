# URL Shortener - Microservices Sample App

A distributed URL shortener application built for testing Root Cause Analysis (RCA) tools. This application features multiple services with deep integration chains, external API calls, caching layers, and various failure scenarios.

## Architecture

```mermaid
graph TB
    User[User/Browser]
    Frontend[Frontend]
    API[API Service]
    Analytics[Analytics Service]
    DB[(PostgreSQL)]
    Cache[(Redis)]
    External[Destination Websites]

    User -->|Create/View URLs| Frontend
    User -->|Click Short URL| API
    Frontend --> API
    Frontend --> Analytics
    API --> DB
    API --> Cache
    API -->|Fetch page metadata| External
    Analytics --> DB

    classDef service fill:#4a90e2,stroke:#333,stroke-width:2px,color:#fff
    classDef datastore fill:#336791,stroke:#333,stroke-width:2px,color:#fff

    class Frontend,API,Analytics service
    class DB,Cache datastore
```

### Services

1. **API Service** (Go - Port 7543)
   - REST API for URL shortening
   - Handles redirects
   - Rate limiting via Redis
   - Fetches link metadata (external HTTP calls)

2. **Analytics Service** (Python/FastAPI - Port 7544)
   - Aggregates click analytics
   - Complex database queries with joins
   - Direct PostgreSQL queries (no caching)
   - Time-series data analysis

3. **Frontend** (React - Port 7545 dev, Port 80 production)
   - Minimal UI for creating short URLs
   - Real-time analytics dashboard
   - Built with Vite
   - Nginx server for production

### Data Stores

- **PostgreSQL** (Port 5432) - Primary database
  - Tables: users, urls, clicks

- **Redis** (Port 6379) - Cache and rate limiting (API Service only)
  - URL mappings cache (2 minute TTL)
  - Rate limit counters
  - Click counters

## Integration Chains (RCA Testing Scenarios)

### Create Short URL Flow
```
Frontend → API Service → PostgreSQL (insert URL)
                      → External HTTP (fetch metadata - can timeout)
                      → PostgreSQL (update metadata)
                      → Redis (cache URL)
```

### Click/Redirect Flow
```
Browser → API Service → Redis (lookup - L1 cache)
                      → [cache miss] → PostgreSQL (fallback)
                      → PostgreSQL (insert click record)
                      → Redis (update counters)
                      → HTTP 302 redirect
```

### Analytics Flow
```
Frontend → Analytics Service → PostgreSQL (complex queries)
                             → Always fresh data (no caching)
```

## Setup Instructions

### Prerequisites

- Docker
- PostgreSQL 15
- Redis 7
- Go 1.21+ (for local development)
- Python 3.11+ (for local development)
- Node.js 18+ (for local development)

### Database Setup

#### Option 1: Using Custom Image (Recommended)

The custom database image includes the schema and initialization script:

```bash
# Build the custom database image
cd db-service
docker build -t rashadxyz/url-shortener-db .

# Run container (database is automatically initialized)
docker run -d \
  --name url-shortener-db \
  -e POSTGRES_USER=urlshortener \
  -e POSTGRES_PASSWORD=password123 \
  -e POSTGRES_DB=urlshortener \
  -p 5432:5432 \
  rashadxyz/url-shortener-db
```

#### Option 2: Manual Setup with Official Postgres Image

1. **Start PostgreSQL**:
```bash
docker run -d \
  --name url-shortener-db \
  -e POSTGRES_USER=urlshortener \
  -e POSTGRES_PASSWORD=password123 \
  -e POSTGRES_DB=urlshortener \
  -p 5432:5432 \
  postgres:15-alpine
```

2. **Initialize Database**:
```bash
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

### Redis Setup

```bash
docker run -d \
  --name url-shortener-redis \
  -p 6379:6379 \
  redis:7-alpine
```

### Build and Run Services

#### API Service (Go)

```bash
cd api-service

# Build Docker image
docker build -t rashadxyz/url-shortener-api .

# Run container
docker run -d \
  --name url-shortener-api \
  -p 7543:7543 \
  -e DATABASE_URL="postgres://urlshortener:password123@host.docker.internal:5432/urlshortener?sslmode=disable" \
  -e REDIS_URL="host.docker.internal:6379" \
  -e PORT=7543 \
  -e RATE_LIMIT_REQUESTS=100 \
  -e RATE_LIMIT_WINDOW=60 \
  -e CACHE_TTL=120 \
  rashadxyz/url-shortener-api

# Note: Code defaults are RATE_LIMIT_REQUESTS=5 and CACHE_TTL=120 (2 minutes)
```

**Local Development**:
```bash
cd api-service
go mod download
DATABASE_URL="postgres://urlshortener:password123@localhost:5432/urlshortener?sslmode=disable" \
REDIS_URL="localhost:6379" \
go run main.go
```

#### Analytics Service (Python)

```bash
cd analytics-service

# Build Docker image
docker build -t rashadxyz/url-shortener-analytics .

# Run container
docker run -d \
  --name url-shortener-analytics \
  -p 7544:7544 \
  -e DATABASE_URL="postgresql://urlshortener:password123@host.docker.internal:5432/urlshortener" \
  -e PORT=7544 \
  rashadxyz/url-shortener-analytics
```

**Local Development**:
```bash
cd analytics-service
pip install -r requirements.txt
DATABASE_URL="postgresql://urlshortener:password123@localhost:5432/urlshortener" \
PORT=7544 \
python main.py
```

#### Frontend (React)

```bash
cd frontend

# Build Docker image
docker build -t rashadxyz/url-shortener-frontend .

# Run container
docker run -d \
  --name url-shortener-frontend \
  -p 7545:80 \
  rashadxyz/url-shortener-frontend
```

**Local Development**:
```bash
cd frontend
npm install
npm run dev
```

## Usage

1. **Access the frontend**: http://localhost:7545
2. **Create a short URL**:
   - Enter a long URL
   - Optionally provide a custom short code
   - Click "Shorten URL"

3. **Test redirect**: http://localhost:7543/{short_code}

4. **View analytics**: Displayed on the frontend dashboard

### API Endpoints

**API Service (Port 7543)**:
- `POST /api/urls` - Create short URL (requires username in body)
- `GET /api/urls?username={username}` - List URLs for user
- `GET /{code}` - Redirect to long URL
- `GET /health` - Health check

**Analytics Service (Port 7544)**:
- `GET /api/analytics/summary?username={username}` - Overall stats
- `GET /api/analytics/top-urls?username={username}` - Top URLs by clicks
- `GET /api/analytics/time-series?username={username}&days=7` - Time series data
- `GET /api/analytics/url/{url_id}?username={username}` - Detailed URL analytics
- `GET /health` - Health check

**Default Username**: `testuser`

## Failure Scenarios for RCA Testing

This application is designed to test RCA tools with various failure modes:

### 1. Cache Failures
- Stop Redis: `docker stop url-shortener-redis`
- Observe: Slower redirects, fallback to PostgreSQL, increased DB load

### 2. Database Connection Issues
- Simulate connection pool exhaustion
- Heavy analytics queries blocking API writes
- Query timeouts

### 3. External API Timeouts
- Link metadata fetching (in URL creation) can timeout
- Network delays cascading through the system

### 4. Rate Limiting
- Exceed rate limits to trigger 429 responses
- Test Redis-based rate limiting failures

### 5. Cache Inconsistency
- Stale data in Redis vs PostgreSQL
- Cache invalidation issues
- Counter drift between Redis and PostgreSQL

### 6. Resource Exhaustion
- Database connection pool depletion
- Memory leaks in long-running services
- Redis memory limits

### 7. Service Dependencies
- API Service depends on: PostgreSQL (required), Redis (optional - graceful degradation), External metadata API (optional)
- Analytics Service depends on: PostgreSQL (required)
- Frontend depends on: API Service, Analytics Service

## Monitoring & Observability

Each service exposes:
- Health check endpoints (`/health`)
- Structured logging to stdout
- Database query performance can be monitored
- Redis cache hit/miss rates in logs

## Project Structure

```
sample-app/
├── api-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── analytics-service/
│   ├── Dockerfile
│   ├── requirements.txt
│   └── main.py
├── frontend/
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   └── src/
│       ├── main.jsx
│       ├── App.jsx
│       └── index.css
├── manifests/
│   ├── kustomization.yaml    # Kustomize configuration
│   ├── vars.yaml             # All configurable variables
│   ├── README.md             # Kustomize documentation
│   └── base/                 # OpenChoreo manifests
│       ├── url-shortener-demo-project.yaml
│       ├── api-service-component.yaml
│       ├── analytics-service-component.yaml
│       ├── frontend-component.yaml
│       ├── postgres-component.yaml
│       └── redis-component.yaml
├── build-and-push.sh
└── README.md
```

## Docker Hub Images

This project uses Docker Hub for image distribution. Images are available at:
- `rashadxyz/url-shortener-db:latest`
- `rashadxyz/url-shortener-api:latest`
- `rashadxyz/url-shortener-analytics:latest`
- `rashadxyz/url-shortener-frontend:latest`

### Build and Push Images

Use the provided build script to build and push all images to Docker Hub:

```bash
# Login to Docker Hub
docker login

# Build and push all images
./build-and-push.sh

# Build with a specific version
VERSION=v1.0.0 ./build-and-push.sh

# Build without cache (clean build)
./build-and-push.sh --no-cache
```

## Deployment Options

### OpenChoreo Deployment (with Kustomize)

OpenChoreo manifests are available in the `manifests/` directory with **Kustomize support** for easy configuration management.

The manifests use Kustomize to eliminate hardcoded values. All configuration is centralized in `manifests/vars.yaml`:
- Image tags and registries
- Database credentials
- Port numbers
- Resource limits (CPU, memory)
- Replica counts
- Environment variables

**Quick Deploy** (recommended):
```bash
# Deploy with Kustomize (uses manifests/vars.yaml for configuration)
kubectl apply -k manifests/

# Preview what will be deployed
kubectl kustomize manifests/
```

**Customize Deployment**:
Edit `manifests/vars.yaml` to change any configuration:
```yaml
# Example: Change image tag and database password
data:
  IMAGE_TAG: "v2.0.0"
  DB_PASSWORD: "your-secure-password"
  API_SERVICE_REPLICAS: "3"
```

Then apply:
```bash
kubectl apply -k manifests/
```

**Without Kustomize** (direct deployment):
```bash
# Apply base manifests directly
kubectl apply -f manifests/base/url-shortener-demo-project.yaml
kubectl apply -f manifests/base/postgres-component.yaml
kubectl apply -f manifests/base/redis-component.yaml
kubectl apply -f manifests/base/api-service-component.yaml
kubectl apply -f manifests/base/analytics-service-component.yaml
kubectl apply -f manifests/base/frontend-component.yaml
```

See `manifests/README.md` for detailed Kustomize documentation and all configurable variables.

### Kubernetes Deployment

Use the OpenChoreo manifests with Kustomize for Kubernetes deployment:

```bash
# Pull images from Docker Hub (or build and push using ./build-and-push.sh)
docker pull rashadxyz/url-shortener-api:latest
docker pull rashadxyz/url-shortener-analytics:latest
docker pull rashadxyz/url-shortener-frontend:latest

# Deploy to Kubernetes using Kustomize
kubectl apply -k manifests/

# Check status
kubectl get pods

# Access services (with port-forward)
kubectl port-forward svc/frontend 7545:80
kubectl port-forward svc/api-service 7543:80
kubectl port-forward svc/analytics-service 7544:80
```

## Testing RCA Scenarios

Example test scenarios:

1. **Slow Redirect Response**:
   - Root cause: Redis down → Database query slow
   - Trace: Browser → API → Redis (failed) → PostgreSQL (slow)

2. **Failed URL Creation**:
   - Root cause: External metadata fetch timeout
   - Trace: Frontend → API → PostgreSQL → External HTTP (timeout)

3. **Stale Analytics**:
   - Root cause: Cache not invalidated after new clicks
   - Trace: Frontend → Analytics → Redis (stale cache)

4. **Rate Limit Errors**:
   - Root cause: Redis rate limit counter exceeded
   - Trace: Frontend → API → Redis (rate check) → 429 response

## License

MIT
