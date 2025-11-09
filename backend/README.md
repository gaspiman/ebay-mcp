# Backend - OAuth Provider API

A secure authentication and OAuth 2.0 provider service built with Go, Gin, and PostgreSQL.

## Features

- User registration and authentication
- JWT-based session management
- OAuth 2.0 authorization server
  - Authorization code flow
  - Refresh token support
  - UserInfo endpoint
- Secure password hashing with bcrypt
- PostgreSQL database with GORM
- CORS support for frontend integration

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

## Installation

1. Install dependencies:
```bash
cd backend
go mod download
```

2. Set up PostgreSQL database:
```bash
createdb ebay_mcp_db
```

3. Create `.env` file (copy from `.env.example`):
```bash
cp .env.example .env
```

4. Update `.env` with your configuration:
```env
PORT=8080
FRONTEND_URL=http://localhost:3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ebay_mcp_db

JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
OAUTH_ISSUER=http://localhost:8080
```

## Running the Server

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication Endpoints

#### Register
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Get Profile (Protected)
```http
GET /api/auth/profile
Authorization: Bearer <jwt_token>
```

### OAuth 2.0 Endpoints

#### Authorization Endpoint
```http
GET /oauth/authorize?client_id=xxx&redirect_uri=xxx&response_type=code&scope=read&state=xyz
Authorization: Bearer <jwt_token>
```

#### Consent Endpoint
```http
POST /oauth/authorize/consent
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "client_id": "client_id",
  "redirect_uri": "https://app.example.com/callback",
  "scope": "read write",
  "state": "random_state",
  "approved": true
}
```

#### Token Endpoint
```http
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=AUTH_CODE&
redirect_uri=REDIRECT_URI&
client_id=CLIENT_ID&
client_secret=CLIENT_SECRET
```

#### Refresh Token
```http
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&
refresh_token=REFRESH_TOKEN&
client_id=CLIENT_ID&
client_secret=CLIENT_SECRET
```

#### UserInfo Endpoint
```http
GET /oauth/userinfo
Authorization: Bearer <access_token>
```

## Database Schema

The application uses the following tables:

- **users**: User accounts
- **oauth_clients**: Registered OAuth applications
- **oauth_authorization_codes**: Temporary authorization codes
- **oauth_access_tokens**: Access tokens for API access
- **oauth_refresh_tokens**: Refresh tokens for obtaining new access tokens

## Creating an OAuth Client

To register a third-party application, insert a client into the database:

```sql
INSERT INTO oauth_clients (id, client_secret, name, redirect_uris, created_at, updated_at)
VALUES (
  'my-app-id',
  'my-app-secret',
  'My Application',
  '["http://localhost:4000/callback", "https://myapp.com/callback"]',
  NOW(),
  NOW()
);
```

## Security Features

- Password hashing with bcrypt (cost factor 10)
- JWT tokens with expiration
- Authorization codes expire in 10 minutes
- Access tokens expire in 1 hour
- Refresh tokens expire in 30 days
- CORS protection
- SQL injection protection via GORM

## Development

### Run tests
```bash
go test ./...
```

### Build binary
```bash
go build -o backend main.go
```

## Project Structure

```
backend/
├── main.go                 # Application entry point
├── config/                 # Configuration management
│   └── config.go
├── database/              # Database connection and migration
│   └── database.go
├── models/                # Data models
│   ├── user.go
│   └── oauth.go
├── controllers/           # Request handlers
│   ├── auth_controller.go
│   └── oauth_controller.go
├── middleware/            # Middleware functions
│   └── auth.go
├── routes/                # Route definitions
│   └── routes.go
└── utils/                 # Utility functions
    ├── jwt.go
    └── oauth.go
```

## License

MIT
