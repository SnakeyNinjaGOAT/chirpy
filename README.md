# Chirpy

A REST API for a social media platform built with Go, where users can create accounts, authenticate, and post short messages called "chirps".

## What is Chirpy?

Chirpy is a backend API that powers a Twitter-like social media application. Users can:

- Create accounts with email and password
- Authenticate using JWT tokens
- Post short messages ("chirps")
- View chirps from all users or specific users
- Delete their own chirps
- Upgrade to premium "Chirpy Red" status

The API is built with modern Go practices, featuring:
- Type-safe database queries with sqlc
- Secure authentication with JWT and refresh tokens
- Password hashing with Argon2
- Clean architecture with separated concerns
- Comprehensive test coverage

## Why Chirpy?

- **Learn Go**: Perfect example of a production-ready Go web application
- **API Design**: Demonstrates RESTful API design patterns
- **Authentication**: Complete JWT authentication flow with refresh tokens
- **Database**: PostgreSQL integration with type-safe queries
- **Testing**: Unit and integration tests showing Go testing best practices
- **DevOps**: Easy to deploy and scale

Whether you're learning Go, building a social media app, or studying web API patterns, Chirpy provides a solid foundation with clean, maintainable code.

## Installation & Setup

### Prerequisites

- Go 1.19 or later
- PostgreSQL 12 or later
- goose (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- sqlc (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/SnakeyNinjaGOAT/chirpy.git
   cd chirpy
   ```

2. **Set up environment:**
   ```bash
   cp .env.example .env  # Or create .env with required variables
   # Edit .env with your database URL and secrets
   ```

3. **Set up the database:**
   ```bash
   createdb chirpy
   # Run migrations with goose
   goose -dir sql/schema postgres "$DB_URL" up
   ```

4. **Generate database code:**
   ```bash
   sqlc generate
   ```

5. **Install dependencies:**
   ```bash
   go mod download
   ```

6. **Run the application:**
   ```bash
   go run .
   ```

The API will be available at `http://localhost:8080`.

### Environment Variables

Create a `.env` file with:

```env
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
JWT_SECRET="your-secure-jwt-secret-key"
POLKA_KEY="your-polka-webhook-key"
PLATFORM="dev"
```

## Usage

Once running, you can:

- Visit `http://localhost:8080` for the web interface
- Use API endpoints like `GET /api/healthz` for health checks
- See [API Documentation](docs/api-consumer.md) for full endpoint reference

## Database Management

Chirpy uses [goose](https://github.com/pressly/goose) for database migrations. Useful commands:

```bash
# Check migration status
goose -dir sql/schema postgres "$DB_URL" status

# Run all pending migrations
goose -dir sql/schema postgres "$DB_URL" up

# Rollback the last migration
goose -dir sql/schema postgres "$DB_URL" down

# Create a new migration
goose -dir sql/schema postgres "$DB_URL" create migration_name sql
```

## Documentation

- **[API Consumer Guide](docs/api-consumer.md)** - Complete API reference for developers using the Chirpy API
- **[Maintainer Guide](docs/maintainer.md)** - Development setup, architecture, and contribution guidelines

## Testing

Run the test suite:

```bash
# Unit tests
go test ./...

# Integration tests
go test -v main_test.go
```

## Contributing

See the [Maintainer Guide](docs/maintainer.md) for development setup and contribution guidelines.

## License

This project is open source and available under the [MIT License](LICENSE).
