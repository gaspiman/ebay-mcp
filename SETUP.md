# Complete Setup Guide

This guide will help you set up and run the complete OAuth Provider application (backend + frontend).

## Prerequisites

Before starting, make sure you have the following installed:

- **Go** 1.21 or higher ([Download](https://go.dev/dl/))
- **Node.js** 16+ and npm ([Download](https://nodejs.org/))
- **PostgreSQL** 12+ ([Download](https://www.postgresql.org/download/))
- **Git** ([Download](https://git-scm.com/downloads))

## Step 1: Database Setup

1. Install PostgreSQL if you haven't already

2. Create a new database:
```bash
# On macOS/Linux
createdb ebay_mcp_db

# Or using psql
psql -U postgres
CREATE DATABASE ebay_mcp_db;
\q
```

3. Verify the database exists:
```bash
psql -U postgres -l | grep ebay_mcp_db
```

## Step 2: Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install Go dependencies:
```bash
go mod download
```

3. Create environment configuration:
```bash
cp .env.example .env
```

4. Edit `.env` and update with your settings:
```env
PORT=8080
FRONTEND_URL=http://localhost:3000

# Update these with your PostgreSQL credentials
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_postgres_password
DB_NAME=ebay_mcp_db

# Generate a secure random string for JWT_SECRET
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
OAUTH_ISSUER=http://localhost:8080
```

5. Start the backend server:
```bash
go run main.go
```

You should see:
```
Database connection established
Database migration completed
Starting server on port 8080...
```

The backend is now running at `http://localhost:8080`

## Step 3: Frontend Setup

1. Open a new terminal window and navigate to the frontend directory:
```bash
cd frontend
```

2. Install npm dependencies:
```bash
npm install
```

3. Verify the `.env` file exists with:
```env
REACT_APP_API_URL=http://localhost:8080
```

4. Start the frontend development server:
```bash
npm start
```

The frontend will automatically open in your browser at `http://localhost:3000`

## Step 4: Verify Installation

1. Open `http://localhost:3000` in your browser
2. You should see the home page with "Welcome to OAuth Provider"
3. Click "Get Started" to test user registration
4. Create a test account
5. You should be redirected to the dashboard

## Step 5: Create an OAuth Client (Optional)

To test OAuth functionality, you need to register a client application:

1. Connect to your database:
```bash
psql -U postgres -d ebay_mcp_db
```

2. Insert a test OAuth client:
```sql
INSERT INTO oauth_clients (id, client_secret, name, redirect_uris, created_at, updated_at)
VALUES (
  'test-client',
  'test-secret',
  'Test Application',
  '["http://localhost:4000/callback"]',
  NOW(),
  NOW()
);
```

3. Test the OAuth flow by visiting:
```
http://localhost:3000/oauth/consent?client_id=test-client&redirect_uri=http://localhost:4000/callback&response_type=code&scope=read&state=random123
```

## Common Issues and Solutions

### Backend won't start

**Issue**: `Failed to connect to database`
**Solution**:
- Verify PostgreSQL is running: `pg_isready`
- Check your DB credentials in `.env`
- Ensure the database exists: `psql -U postgres -l`

**Issue**: `Address already in use`
**Solution**: Another process is using port 8080
```bash
# Find and kill the process
lsof -ti:8080 | xargs kill -9
```

### Frontend won't start

**Issue**: `Port 3000 is already in use`
**Solution**:
- Kill the process using port 3000, or
- Start on a different port: `PORT=3001 npm start`

**Issue**: `Module not found`
**Solution**: Delete node_modules and reinstall
```bash
rm -rf node_modules package-lock.json
npm install
```

### CORS errors in browser

**Issue**: `Access-Control-Allow-Origin` errors
**Solution**:
- Verify FRONTEND_URL in backend `.env` matches your frontend URL
- Restart the backend after changing `.env`

### Database migration issues

**Issue**: Tables not created
**Solution**: GORM auto-migrates on startup. Check logs for errors.

To manually reset the database:
```bash
dropdb ebay_mcp_db
createdb ebay_mcp_db
# Restart backend to trigger migration
```

## Development Workflow

### Running both servers

Use two terminal windows:

**Terminal 1 (Backend)**:
```bash
cd backend
go run main.go
```

**Terminal 2 (Frontend)**:
```bash
cd frontend
npm start
```

### Making changes

- **Backend**: Changes require restart (Ctrl+C and `go run main.go`)
- **Frontend**: Changes auto-reload (hot module replacement)

## Production Deployment

### Backend

1. Build the binary:
```bash
cd backend
go build -o backend-server main.go
```

2. Run with production environment:
```bash
export DB_PASSWORD="secure_password"
export JWT_SECRET="secure_random_string"
./backend-server
```

### Frontend

1. Build for production:
```bash
cd frontend
npm run build
```

2. Serve the `build/` directory with any static file server:
```bash
npx serve -s build
```

Or deploy to:
- Netlify
- Vercel
- AWS S3 + CloudFront
- Any static hosting service

## Security Checklist for Production

- [ ] Change JWT_SECRET to a strong random value
- [ ] Use strong database password
- [ ] Enable HTTPS/TLS
- [ ] Set secure CORS origins
- [ ] Use environment variables (never commit `.env`)
- [ ] Enable database SSL connections
- [ ] Set up proper firewall rules
- [ ] Regular security updates
- [ ] Enable rate limiting
- [ ] Add logging and monitoring

## Testing the Application

### Test User Registration
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'
```

### Test Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

## Next Steps

- Customize the UI styling
- Add email verification
- Implement password reset
- Add OAuth scope management
- Create admin panel for managing OAuth clients
- Add rate limiting
- Implement logging
- Add monitoring and analytics

## Support

For issues or questions:
- Check the backend logs in the terminal
- Check browser console for frontend errors
- Review the API endpoints in the README files
- Ensure all prerequisites are installed

## License

MIT
