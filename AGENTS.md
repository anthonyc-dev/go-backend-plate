# AGENTS.md - AI Agent Guidelines for This Repository

## Project Overview

This is a Go REST API project using Gin framework and GORM for database operations.

## Build Commands

### Local Development

```bash
go build -o app.exe
go run .
```

### Environment-Specific Run

```bash
# Default (.env)
go run .

# Production - set GO_ENV
GO_ENV=production go run .

# Development - set GO_ENV
GO_ENV=development go run .
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run single test
go test -v ./... -run TestName

# Test with coverage
go test -cover ./...
```

### Docker Commands (via Makefile)

```bash
# Development
make dev
make dev-down

# Production
make prod
make prod-down

# View logs
make logs

# Build and run
make up
make down
```

### Linting

```bash
go fmt ./...
go vet ./...
```

## Configuration

Environment variables are managed through `configs/env.go`:

- `.env` - Default/local development
- `.env.dev` - Development environment
- `.env.prod` - Production environment

Load config at app start:

```go
configs.LoadEnv()
```

Access config via `configs.AppEnv`:

```go
configs.AppEnv.Port
configs.AppEnv.DBHost
configs.AppEnv.DBUser
configs.AppEnv.DBPassword
configs.AppEnv.DBName
configs.AppEnv.DBPort
configs.AppEnv.DatabaseURL
configs.AppEnv.JWTSecret
configs.AppEnv.RedisHost
configs.AppEnv.RedisPort
```

## Code Style Guidelines

### Imports

Group imports: standard library first, then third-party, then local packages.

```go
import (
    "fmt"
    "net/http"
    "time"
    "rest-api/models"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)
```

### Formatting

- Use `gofmt` or IDE auto-formatter
- Keep lines under 100 characters
- Use tabs for indentation

### Types

- Use explicit types; avoid `var` when type can be inferred
- Use `:=` for short variable declarations
- Example: `users := []models.User{}`

### Naming Conventions

- **Files**: snake_case (e.g., `user_controller.go`, `database.go`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase
- **Structs**: PascalCase (e.g., `type User struct`)
- **Interfaces**: PascalCase with `er` suffix (e.g., `Reader`)

### Error Handling

- Always check errors and handle them appropriately
- Return meaningful error messages to clients
- Use `gin.H{"error": "message"}` for JSON error responses
- Log errors with context
- Never expose internal error details in production

### Database Operations

- Use GORM for all database operations
- Validate inputs before database operations
- Check for existing records before creating
- Use migrations: `db.AutoMigrate(&Model{})`

### API Design

- Use RESTful conventions: GET, POST, PUT, DELETE
- Group routes: `api.Group("/users")`
- Use proper HTTP status codes (200, 201, 400, 404, 409, 500)

### Security

- Never hardcode secrets in source code
- Use environment variables for sensitive data
- Validate all user inputs using `c.ShouldBindJSON()`
- Implement authentication and authorization

## Project Structure

```
rest-api/
├── main.go                    # Entry point
├── configs/
│   ├── env.go                 # Environment configuration
│   └── redis.go               # Redis configuration
├── database/
│   └── db.go                  # Database connection
├── controllers/
│   ├── auth_controller.go     # Auth HTTP handlers
│   └── user_controller.go    # User HTTP handlers
├── models/
│   ├── user.go               # User model
│   ├── refresh_token.go      # Refresh token model
│   └── token_blacklist.go    # Token blacklist model
├── routes/
│   ├── auth_routes.go        # Auth routes
│   └── user_routes.go        # User routes
├── middlewares/
│   ├── auth_middleware.go    # JWT authentication
│   └── rateLimiting_middleware.go
├── services/
│   └── auth_service.go       # Auth business logic
├── dto/
│   └── auth_request.go       # Request/Response DTOs
├── utils/
│   ├── jwt.go                # JWT utilities
│   ├── hash.go               # Password hashing
│   ├── validation.go        # Input validation
│   ├── cookies.go            # Cookie utilities
│   └── logger.go             # Logging utilities
├── .env                      # Default env
├── .env.dev                  # Dev env
├── .env.prod                 # Prod env
├── docker-compose.dev.yml   # Dev Docker compose
├── docker-compose.prod.yml  # Prod Docker compose
├── Dockerfile
└── Makefile
```

## Dependencies

- **Web Framework**: github.com/gin-gonic/gin
- **ORM**: gorm.io/gorm
- **Database Driver**: gorm.io/driver/postgres
- **Environment Variables**: github.com/joho/godotenv
- **JWT**: github.com/golang-jwt/jwt/v5
- **UUID**: github.com/google/uuid

## Environment Variables

```
PORT=8080
DB_CONNECTION=postgres
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=go-db
DB_PORT=5432
DATABASE_URL=postgresql://...
JWT_SECRET=your_secret_key
REDIS_HOST=localhost
REDIS_PORT=6379
```
