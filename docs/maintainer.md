# Chirpy API - Maintainer Guide

## Overview

This guide is for developers who want to contribute to or maintain the Chirpy API project.

## Architecture

Chirpy is built with Go and uses:
- **PostgreSQL** for data persistence
- **sqlc** for type-safe database queries
- **JWT** for authentication
- **Argon2** for password hashing
- Standard Go HTTP server

### Project Structure

```
chirpy/
├── internal/
│   ├── auth/          # Authentication utilities
│   ├── config/        # Configuration and environment loading
│   ├── database/      # Generated database code (sqlc)
│   ├── handlers/      # HTTP request handlers
│   ├── models/        # Data models and validation
│   └── utils/         # Utility functions
├── sql/
│   ├── schema/        # Database schema files
│   └── queries/       # SQL query files
├── static/            # Static web assets
├── templates/         # HTML templates
├── docs/              # Documentation
├── main.go            # Application entry point
├── main_test.go       # Integration tests
└── sqlc.yaml          # sqlc configuration
```

## Development Setup

### Prerequisites

- Go 1.19 or later
- PostgreSQL 12 or later
- goose (install with `go install github.com/pressly/goose/v3/cmd/goose@latest`)
- sqlc (install with `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### Environment Setup

1. Clone the repository
2. Copy `.env` and update the values:
   ```
   DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
   JWT_SECRET="your-jwt-secret-key"
   POLKA_KEY="your-polka-webhook-key"
   PLATFORM="dev"
   ```

3. Create the database:
   ```sql
   CREATE DATABASE chirpy;
   ```

4. Run database migrations:
   ```bash
   # Run migrations with goose
   goose -dir sql/schema postgres "$DB_URL" up
   ```

5. Generate database code:
   ```bash
   sqlc generate
   ```

6. Install dependencies:
   ```bash
   go mod download
   ```

7. Run the application:
   ```bash
   go run .
   ```

## Development Workflow

### Adding New Features

1. **Database Changes:**
   - Add schema changes to `sql/schema/`
   - Add queries to `sql/queries/`
   - Run `sqlc generate` to update Go code

2. **API Changes:**
   - Add handlers in `internal/handlers/`
   - Update routes in `main.go`
   - Add tests in corresponding `_test.go` files

3. **Testing:**
   - Run unit tests: `go test ./...`
   - Run integration tests: `go test -v main_test.go`

### Code Style

- Follow standard Go formatting (`go fmt`)
- Use `gofmt` and `goimports`
- Run `go vet` to check for common errors
- Use descriptive variable names
- Add comments for exported functions

### Database Migrations

Database migrations are managed using [goose](https://github.com/pressly/goose). All migration files are located in `sql/schema/` and follow the goose format with `-- +goose Up` and `-- +goose Down` directives.

#### Creating New Migrations

To create a new migration:

```bash
# Create a new migration file
goose -dir sql/schema postgres "$DB_URL" create migration_name sql
```

This will create a new file like `sql/schema/006_migration_name.sql` with the proper goose headers.

#### Running Migrations

```bash
# Run all pending migrations
goose -dir sql/schema postgres "$DB_URL" up

# Run a specific number of migrations
goose -dir sql/schema postgres "$DB_URL" up 2

# Rollback the last migration
goose -dir sql/schema postgres "$DB_URL" down

# Check migration status
goose -dir sql/schema postgres "$DB_URL" status
```

#### Migration Workflow

When making database schema changes:

1. Create a new migration: `goose -dir sql/schema postgres "$DB_URL" create description sql`
2. Add your schema changes in the `-- +goose Up` section
3. Add the rollback logic in the `-- +goose Down` section
4. Test the migration: `goose -dir sql/schema postgres "$DB_URL" up`
5. Update existing queries in `sql/queries/` if needed
6. Run `sqlc generate` to update the generated code

### Authentication Flow

The API uses JWT tokens with refresh tokens:

1. User logs in with `/api/login` → receives access + refresh tokens
2. Access token used for API calls (short-lived)
3. Refresh token used to get new access tokens via `/api/refresh`
4. Refresh tokens can be revoked via `/api/revoke`

### Testing

#### Unit Tests
Run specific package tests:
```bash
go test ./internal/handlers/
go test ./internal/auth/
```

#### Integration Tests
```bash
go test -v main_test.go
```

#### Database Tests
Tests that require database access are in `*_test.go` files and use a test database.

## Deployment

### Environment Variables

Production environment variables:
- `DB_URL`: Production PostgreSQL connection string
- `JWT_SECRET`: Strong random secret for JWT signing
- `POLKA_KEY`: Webhook secret for Polka service
- `PLATFORM`: Set to "prod" for production mode

### Building

```bash
go build -o chirpy .
```

### Docker (if applicable)

Add Dockerfile and docker-compose.yml for containerized deployment.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Ensure all tests pass
6. Submit a pull request

## Troubleshooting

### Common Issues

1. **Database connection errors:**
   - Check `DB_URL` in `.env`
   - Ensure PostgreSQL is running
   - Verify database exists

2. **sqlc generation fails:**
   - Check SQL syntax in schema/query files
   - Ensure sqlc is installed: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

3. **Authentication issues:**
   - Verify `JWT_SECRET` is set
   - Check token expiration (access tokens expire in 1 hour)

4. **Port conflicts:**
   - Default port is 8080, change in `main.go` if needed

### Logs

The application logs to stdout. Use `go run . 2>&1 | tee log.txt` to save logs.

## API Versioning

Currently at v1. Future versions will use URL prefixes like `/api/v2/`.

## Security Considerations

- JWT secrets must be strong and rotated regularly
- Passwords are hashed with Argon2
- Input validation on all endpoints
- CORS headers configured for web clients
- Rate limiting should be added for production use