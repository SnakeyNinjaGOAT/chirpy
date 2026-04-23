# Chirpy API Documentation

## Overview

Chirpy is a REST API for a social media platform where users can create accounts, authenticate, and post short messages called "chirps".

## Base URL

```
http://localhost:8080
```

## Authentication

Most endpoints require authentication using JWT tokens. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Health Check

#### GET /api/healthz
Returns the health status of the API.

**Response:**
```json
200 OK
```

### User Management

#### POST /api/users
Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "id": "uuid",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

#### PUT /api/users
Update user information. Requires authentication.

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "password": "newpassword123"
}
```

**Response:**
```json
{
  "id": "uuid",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "email": "newemail@example.com",
  "is_chirpy_red": false
}
```

#### POST /api/login
Authenticate a user and receive JWT tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "id": "uuid",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "email": "user@example.com",
  "token": "jwt-access-token",
  "refresh_token": "jwt-refresh-token",
  "is_chirpy_red": false
}
```

### Token Management

#### POST /api/refresh
Refresh an access token using a refresh token.

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
```json
{
  "token": "new-jwt-access-token"
}
```

#### POST /api/revoke
Revoke a refresh token.

**Headers:**
```
Authorization: Bearer <refresh-token>
```

**Response:**
```
204 No Content
```

### Chirps

#### GET /api/chirps
Get all chirps. Optional query parameter `author_id` to filter by user.

**Query Parameters:**
- `author_id` (optional): UUID of the author
- `sort` (optional): "asc" or "desc" (default: "asc")

**Response:**
```json
[
  {
    "id": "uuid",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "body": "This is a chirp!",
    "user_id": "uuid"
  }
]
```

#### GET /api/chirps/{chirpId}
Get a specific chirp by ID.

**Response:**
```json
{
  "id": "uuid",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "body": "This is a chirp!",
  "user_id": "uuid"
}
```

#### POST /api/chirps
Create a new chirp. Requires authentication.

**Request Body:**
```json
{
  "body": "This is my new chirp!"
}
```

**Response:**
```json
{
  "id": "uuid",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "body": "This is my new chirp!",
  "user_id": "uuid"
}
```

#### DELETE /api/chirps/{chirpId}
Delete a chirp. Requires authentication and ownership of the chirp.

**Response:**
```
204 No Content
```

### Webhooks

#### POST /api/polka/webhooks
Handle webhooks from Polka service for user upgrades.

**Headers:**
```
Authorization: Bearer <polka-key>
```

**Request Body:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid"
  }
}
```

**Response:**
```
204 No Content
```

### Admin Endpoints

#### GET /admin/metrics
Get server metrics (requires no authentication in dev mode).

**Response:**
```
Hits: 123
```

#### POST /admin/reset
Reset server state (development only).

**Response:**
```
200 OK
```

### Static Files

#### GET /app/*
Serve static files from the `/app/` path.

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error