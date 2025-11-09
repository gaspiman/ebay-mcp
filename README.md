# OAuth Provider Platform

A full-stack web application that provides user authentication and acts as an OAuth 2.0 provider, allowing third-party applications to securely access user data.

## Features

### User Management
- **User Registration** - Create new accounts with email and password
- **User Login** - Secure authentication with JWT tokens
- **User Dashboard** - View profile information and manage connected apps

### OAuth 2.0 Provider
- **Authorization Code Flow** - Full OAuth 2.0 implementation
- **Consent Screen** - Users can approve/deny third-party app access
- **Token Management** - Access tokens and refresh tokens
- **UserInfo Endpoint** - Standard OAuth user information endpoint
- **Scope-based Permissions** - Granular access control

### Security
- **Password Hashing** - bcrypt with salt
- **JWT Authentication** - Secure session management
- **CORS Protection** - Configurable cross-origin policies
- **SQL Injection Protection** - GORM ORM with parameterized queries
- **Token Expiration** - Automatic token invalidation

## Architecture

```
┌─────────────┐         ┌─────────────┐         ┌──────────────┐
│             │         │             │         │              │
│   Frontend  │◄───────►│   Backend   │◄───────►│  PostgreSQL  │
│   (React)   │  REST   │    (Go)     │  GORM   │   Database   │
│             │  API    │             │         │              │
└─────────────┘         └─────────────┘         └──────────────┘
```

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt

### Frontend
- **Framework**: React 18 with TypeScript
- **Routing**: React Router v6
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Build Tool**: Create React App

## Quick Start

See [SETUP.md](SETUP.md) for detailed installation instructions.

### Prerequisites
- Go 1.21+
- Node.js 16+
- PostgreSQL 12+

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd ebay-mcp
```

2. **Set up the database**
```bash
createdb ebay_mcp_db
```

3. **Start the backend**
```bash
cd backend
cp .env.example .env
# Edit .env with your database credentials
go mod download
go run main.go
```

4. **Start the frontend**
```bash
cd frontend
npm install
npm start
```

5. **Access the application**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

## Project Structure

```
ebay-mcp/
├── backend/                    # Go backend server
│   ├── config/                # Configuration management
│   ├── controllers/           # Request handlers
│   ├── database/              # Database setup
│   ├── middleware/            # Middleware functions
│   ├── models/                # Data models
│   ├── routes/                # Route definitions
│   ├── utils/                 # Utility functions
│   ├── main.go               # Entry point
│   ├── go.mod                # Go dependencies
│   └── .env.example          # Environment template
│
├── frontend/                  # React frontend
│   ├── src/
│   │   ├── api/              # API client
│   │   ├── context/          # React Context
│   │   ├── pages/            # Page components
│   │   ├── App.tsx           # Main component
│   │   └── index.tsx         # Entry point
│   ├── public/               # Static files
│   └── package.json          # npm dependencies
│
├── main.go                   # Legacy eBay proxy (separate)
├── SETUP.md                  # Setup instructions
└── README.md                 # This file
```

## API Documentation

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

### OAuth Endpoints

#### Authorization
```http
GET /oauth/authorize?client_id={id}&redirect_uri={uri}&response_type=code&scope={scope}&state={state}
Authorization: Bearer {jwt_token}
```

#### Token Exchange
```http
POST /oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code={code}&redirect_uri={uri}&client_id={id}&client_secret={secret}
```

#### UserInfo
```http
GET /oauth/userinfo
Authorization: Bearer {access_token}
```

For complete API documentation, see [backend/README.md](backend/README.md).

## OAuth Flow Example

1. **Third-party app initiates OAuth**
```
https://yourapp.com/oauth/consent?
  client_id=app123&
  redirect_uri=https://thirdparty.com/callback&
  response_type=code&
  scope=read&
  state=random123
```

2. **User logs in** (if not already authenticated)

3. **User sees consent screen** with app details and permissions

4. **User approves** → Redirect to `https://thirdparty.com/callback?code=ABC123&state=random123`

5. **Third-party exchanges code for token**
```bash
curl -X POST http://localhost:8080/oauth/token \
  -d "grant_type=authorization_code" \
  -d "code=ABC123" \
  -d "redirect_uri=https://thirdparty.com/callback" \
  -d "client_id=app123" \
  -d "client_secret=secret456"
```

6. **Response contains access token**
```json
{
  "access_token": "token123",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh456",
  "scope": "read"
}
```

## Development

### Running Tests
```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm test
```

### Building for Production
```bash
# Backend
cd backend
go build -o backend-server main.go

# Frontend
cd frontend
npm run build
```

## Security Considerations

- Always use HTTPS in production
- Change `JWT_SECRET` to a strong random value
- Use environment variables for secrets
- Enable database SSL connections
- Implement rate limiting
- Add logging and monitoring
- Regular security updates

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT

## Support

For issues and questions:
- Check [SETUP.md](SETUP.md) for installation help
- Review component READMEs in [backend/](backend/) and [frontend/](frontend/)
- Check server logs for errors
- Ensure all prerequisites are installed
