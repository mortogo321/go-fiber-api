# Go Fiber REST API

A production-ready REST API proof-of-concept built with [Go Fiber](https://gofiber.io/), demonstrating JWT authentication, PostgreSQL persistence via GORM, Redis cache-aside caching, and a Docker multi-stage build.

## Project Structure

```
.
├── config/          # Environment-based configuration
│   └── config.go
├── database/        # Database and cache client initialization
│   ├── database.go
│   └── redis.go
├── handlers/        # HTTP request handlers
│   ├── auth.go
│   └── product.go
├── middleware/      # Fiber middleware (JWT guard, request logger)
│   ├── auth.go
│   └── logger.go
├── models/          # GORM models
│   ├── user.go
│   └── product.go
├── services/        # Shared services (cache helper)
│   └── cache.go
├── utils/           # Helpers (response builders, JWT utilities)
│   ├── jwt.go
│   └── response.go
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── main.go
```

## API Endpoints

| Method | Path               | Auth     | Description             |
|--------|--------------------|----------|-------------------------|
| POST   | `/api/auth/register` | Public   | Register a new user     |
| POST   | `/api/auth/login`    | Public   | Login and receive JWT   |
| GET    | `/api/auth/profile`  | Bearer   | Get current user profile|
| GET    | `/api/products`      | Bearer   | List products (cached)  |
| GET    | `/api/products/:id`  | Bearer   | Get product by ID (cached) |
| POST   | `/api/products`      | Bearer   | Create a product        |
| PUT    | `/api/products/:id`  | Bearer   | Update a product        |
| DELETE | `/api/products/:id`  | Bearer   | Delete a product        |

## Quick Start

### With Docker Compose (recommended)

```bash
cp .env.example .env
docker-compose up --build
```

The API will be available at `http://localhost:3000`.

### Local Development

```bash
# Ensure PostgreSQL and Redis are running
cp .env.example .env
# Edit .env with your local connection details
go mod tidy
go run main.go
```

## Architecture Decisions

### Why Fiber?

Fiber is built on top of **fasthttp**, the fastest HTTP engine for Go. It provides an Express-like API that is intuitive for developers coming from Node.js while delivering significantly better throughput than net/http-based frameworks. For a REST API POC where raw performance and developer ergonomics both matter, Fiber strikes the right balance.

### Why GORM?

GORM is the most mature ORM in the Go ecosystem. It handles migrations, associations, hooks, and query building with a clean chainable API. For a POC that needs to demonstrate relational data (users owning products) without hand-writing SQL, GORM reduces boilerplate while remaining transparent about the queries it generates.

### Cache-Aside Pattern

The product endpoints implement a **cache-aside** (lazy-loading) strategy:

1. **Read path** -- check Redis first; on a cache miss, query PostgreSQL, then populate Redis with a TTL.
2. **Write path** -- after any mutation (create/update/delete), invalidate the relevant cache keys so subsequent reads fetch fresh data.

This keeps reads fast without risking stale data on writes. It is the simplest caching pattern to reason about and is appropriate for workloads where reads vastly outnumber writes.

### Graceful Shutdown

The application listens for `SIGINT` and `SIGTERM`, then calls `app.ShutdownWithTimeout` to drain in-flight requests before closing database and Redis connections. This ensures zero dropped requests during deployments.

### Docker Multi-Stage Build

The Dockerfile uses a two-stage build:
- **Builder stage** -- compiles a statically-linked binary with `CGO_ENABLED=0` and strips debug symbols (`-ldflags="-s -w"`).
- **Production stage** -- copies the binary into a minimal `alpine:3.20` image running as a non-root user, producing a final image under 20 MB.
