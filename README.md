# Go Fiber REST API

A production-ready REST API built with Go Fiber, featuring JWT authentication, PostgreSQL persistence via GORM, and Redis cache-aside pattern.

## Project Structure

```
.
├── config/
│   └── config.go          # Environment-based configuration
├── database/
│   ├── database.go        # PostgreSQL connection & migrations
│   └── redis.go           # Redis client initialization
├── handlers/
│   ├── auth.go            # Register, Login, GetProfile
│   └── product.go         # CRUD with cache-aside pattern
├── middleware/
│   ├── auth.go            # JWT Bearer token validation
│   └── logger.go          # Request logging with latency
├── models/
│   ├── user.go            # User entity
│   └── product.go         # Product entity
├── services/
│   └── cache.go           # Redis cache abstraction
├── utils/
│   ├── jwt.go             # Token generation & validation
│   └── response.go        # Standardized API responses
├── docker-compose.yml     # App + PostgreSQL + Redis
├── Dockerfile             # Multi-stage build
├── main.go                # Application entrypoint
├── go.mod                 # Go module definition
└── .env.example           # Environment variable template
```

## API Endpoints

| Method | Path                  | Auth     | Description          |
|--------|-----------------------|----------|----------------------|
| POST   | `/api/auth/register`  | Public   | Create new account   |
| POST   | `/api/auth/login`     | Public   | Authenticate & get JWT |
| GET    | `/api/users/profile`  | Bearer   | Get current user profile |
| GET    | `/api/products`       | Bearer   | List all products    |
| GET    | `/api/products/:id`   | Bearer   | Get product by ID    |
| POST   | `/api/products`       | Bearer   | Create product       |
| PUT    | `/api/products/:id`   | Bearer   | Update product       |
| DELETE | `/api/products/:id`   | Bearer   | Delete product       |

## Architecture: Cache-Aside Pattern

```
┌────────┐     ┌──────────┐     ┌───────────┐
│ Client │────>│  Fiber   │────>│  Handler  │
└────────┘     │  Router  │     └─────┬─────┘
               └──────────┘           │
                                      │  1. Check cache
                                      v
                                ┌───────────┐
                           ┌───>│   Redis    │
                           │    │  (Cache)   │
                           │    └───────────┘
                           │          │
                           │    2. Cache miss?
                           │          v
                           │    ┌───────────┐
                           │    │ PostgreSQL │
                           │    │   (GORM)  │
                           │    └─────┬─────┘
                           │          │
                           │    3. Store in cache
                           └──────────┘
                              (TTL: 5min)

Write operations:
  Create/Update/Delete ──> DB write ──> Invalidate related cache keys
```

**Key decisions:**

- **Cache-aside (lazy-loading):** Data is loaded into cache only on read misses, keeping the cache lean.
- **TTL-based expiry (5 min):** Balances freshness with performance; stale reads are bounded.
- **Write-through invalidation:** Mutations immediately delete affected cache keys, ensuring the next read fetches fresh data.

## Quick Start

### Prerequisites

- Docker & Docker Compose

### Run

```bash
# Copy environment variables
cp .env.example .env

# Start all services
docker-compose up --build

# API is available at http://localhost:3000
```

### Example Requests

```bash
# Register
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123","name":"John"}'

# Login
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'

# Use the returned token for protected routes
TOKEN="<jwt_token_from_login>"

# Get profile
curl http://localhost:3000/api/users/profile \
  -H "Authorization: Bearer $TOKEN"

# Create product
curl -X POST http://localhost:3000/api/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget","description":"A fine widget","price":29.99,"sku":"WDG-001"}'

# List products
curl http://localhost:3000/api/products \
  -H "Authorization: Bearer $TOKEN"
```

### Development (without Docker)

```bash
# Requires Go 1.23+, running PostgreSQL, and running Redis

export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=fiber_api
export REDIS_URL=redis://localhost:6379
export JWT_SECRET=your-secret-key
export PORT=3000

go run main.go
```

## License

MIT
