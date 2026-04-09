# Go Fiber REST API

A production-ready REST API proof of concept built with Go Fiber v2, featuring JWT authentication, PostgreSQL with GORM, Redis caching, and Docker multi-stage builds.

## Architecture

```
+---------------------------------------------------------+
|                      Client                             |
+----------------------+----------------------------------+
                       | HTTP
+----------------------v----------------------------------+
|                   Go Fiber v2                           |
|  +----------+  +----------+  +-----------------------+  |
|  |  Logger  |  |   JWT    |  |     Validator         |  |
|  |Middleware |  |Middleware|  |                       |  |
|  +----------+  +----------+  +-----------------------+  |
+---------------------------------------------------------+
|                    Handlers                             |
|  +--------------+  +--------------------------------+   |
|  |  Auth Handler |  |       Product Handler          |   |
|  +------+-------+  +----------+---------------------+   |
+---------+----------------------+-------------------------+
|         |        Services      |                         |
|         |  +-------------------+                         |
|         |  |   Cache Service   |                         |
|         |  +---------+---------+                         |
+---------+------------+-----------------------------------+
|         |            |         Data Layer                |
|  +------v-------+  +-v--------------+                   |
|  |  PostgreSQL  |  |     Redis      |                   |
|  |   (GORM)     |  |   (Cache)      |                   |
|  +--------------+  +----------------+                   |
+---------------------------------------------------------+
```

## Project Structure

```
.
├── main.go                 # Application entry point, route setup, graceful shutdown
├── go.mod                  # Go module and dependencies
├── config/
│   └── config.go           # Environment-based configuration
├── database/
│   ├── database.go         # GORM PostgreSQL connection + AutoMigrate
│   └── redis.go            # Redis client initialization
├── models/
│   ├── user.go             # User model with GORM tags
│   └── product.go          # Product model with GORM tags
├── handlers/
│   ├── auth.go             # Register, Login, GetProfile
│   └── product.go          # CRUD with Redis cache-aside
├── middleware/
│   ├── auth.go             # JWT Bearer token validation
│   └── logger.go           # Request logging (method, path, status, duration)
├── services/
│   └── cache.go            # Redis cache wrapper (Get, Set, Delete, InvalidatePattern)
├── utils/
│   ├── response.go         # Standardized JSON response helpers
│   └── validator.go        # Struct validation with go-playground/validator
├── Dockerfile              # Multi-stage build (golang:1.23-alpine -> alpine:3.19)
├── docker-compose.yml      # App + PostgreSQL 17 + Redis 7
├── .env.example            # Environment variable template
└── .gitignore
```

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 17 (or use Docker)
- Redis 7 (or use Docker)

### Run with Docker Compose

```bash
cp .env.example .env
docker-compose up --build
```

The API will be available at `http://localhost:3000`.

### Run Locally

```bash
cp .env.example .env
# Edit .env with your local DB and Redis settings

go mod download
go run main.go
```

## API Endpoints

### Public

| Method | Path              | Description          |
|--------|-------------------|----------------------|
| POST   | `/api/auth/register` | Register a new user  |
| POST   | `/api/auth/login`    | Login and get JWT    |
| GET    | `/api/health`        | Health check         |

### Protected (Bearer Token Required)

| Method | Path                   | Description               |
|--------|------------------------|---------------------------|
| GET    | `/api/auth/profile`    | Get current user profile  |
| GET    | `/api/products`        | List products (cached)    |
| GET    | `/api/products/:id`    | Get product by ID (cached)|
| POST   | `/api/products`        | Create a product          |
| PUT    | `/api/products/:id`    | Update a product          |
| DELETE | `/api/products/:id`    | Delete a product          |

## Key Design Patterns

- **Cache-aside**: Check Redis first, fallback to PostgreSQL, then populate cache
- **JWT authentication**: Stateless auth with user ID and role in claims
- **Input validation**: Struct tag validation via `go-playground/validator`
- **Standardized responses**: Consistent JSON envelope for success, error, and pagination
- **Graceful shutdown**: Signal-based (`SIGINT`, `SIGTERM`) with connection cleanup
- **Multi-stage Docker build**: Minimal production image (~15MB)

## Environment Variables

| Variable      | Description             | Default          |
|---------------|-------------------------|------------------|
| `PORT`        | Server port             | `3000`           |
| `DB_HOST`     | PostgreSQL host         | `localhost`      |
| `DB_PORT`     | PostgreSQL port         | `5432`           |
| `DB_USER`     | PostgreSQL user         | `postgres`       |
| `DB_PASSWORD` | PostgreSQL password     | `postgres`       |
| `DB_NAME`     | PostgreSQL database     | `gofiber`        |
| `REDIS_URL`   | Redis connection string | `localhost:6379` |
| `JWT_SECRET`  | JWT signing key         | `changeme`       |
