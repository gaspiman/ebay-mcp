# Frontend - OAuth Provider Web Application

A modern React-based web application for user authentication and OAuth authorization.

## Features

- User registration and login
- Protected dashboard
- OAuth consent screen
- JWT-based authentication
- Responsive design with Tailwind CSS
- TypeScript for type safety

## Tech Stack

- **Framework**: React 18 with TypeScript
- **Routing**: React Router v6
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Build Tool**: Create React App

## Prerequisites

- Node.js 16+ and npm
- Backend API running on `http://localhost:8080`

## Installation

1. Install dependencies:
```bash
cd frontend
npm install
```

2. Create `.env` file:
```bash
echo "REACT_APP_API_URL=http://localhost:8080" > .env
```

## Running the Application

```bash
npm start
```

The application will start on `http://localhost:3000`

## Available Scripts

- `npm start` - Runs the app in development mode
- `npm test` - Launches the test runner
- `npm run build` - Builds the app for production
- `npm run eject` - Ejects from Create React App (one-way operation)

## Pages

### Home (`/`)
Landing page with application overview and features.

### Register (`/register`)
User registration form with validation:
- Email validation
- Password strength (minimum 8 characters)
- Password confirmation

### Login (`/login`)
User login form with email and password.

### Dashboard (`/dashboard`) - Protected
User dashboard showing:
- User profile information
- Account details
- OAuth application management

### OAuth Consent (`/oauth/consent`) - Protected
OAuth authorization screen where users can:
- Review application permissions
- Approve or deny access
- See requested scopes

## Authentication Flow

1. User registers or logs in
2. JWT token is stored in localStorage
3. Token is sent with all API requests via Authorization header
4. Protected routes redirect to login if not authenticated

## OAuth Flow

1. Third-party app redirects user to `/oauth/consent?client_id=...&redirect_uri=...&response_type=code&scope=...&state=...`
2. User must be logged in (redirected to login if not)
3. User sees consent screen with app details and permissions
4. User approves or denies
5. User is redirected back to app with authorization code or error

## Project Structure

```
frontend/
├── public/                # Static files
├── src/
│   ├── api/              # API client functions
│   │   └── auth.ts
│   ├── context/          # React Context providers
│   │   └── AuthContext.tsx
│   ├── pages/            # Page components
│   │   ├── Home.tsx
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Dashboard.tsx
│   │   └── OAuthConsent.tsx
│   ├── App.tsx           # Main app component with routing
│   ├── index.tsx         # Application entry point
│   └── index.css         # Global styles with Tailwind
├── tailwind.config.js    # Tailwind configuration
├── postcss.config.js     # PostCSS configuration
└── package.json          # Dependencies and scripts
```

## Environment Variables

- `REACT_APP_API_URL` - Backend API URL (default: http://localhost:8080)

## Styling

The application uses Tailwind CSS for styling with a custom color scheme:
- Primary: Indigo (indigo-600)
- Background: Gray gradients
- Accent: Cyan

## API Integration

The frontend communicates with the backend via REST API:

```typescript
// Authentication
authApi.register(data) // Register new user
authApi.login(data)    // Login existing user
authApi.getProfile(token) // Get user profile

// OAuth (via Axios)
GET /oauth/authorize      // Get consent info
POST /oauth/authorize/consent // Submit consent decision
```

## Building for Production

```bash
npm run build
```

This creates an optimized production build in the `build/` directory.

### Deployment

The production build can be served by any static file server:

```bash
# Using serve
npx serve -s build

# Using nginx
# Copy build/ contents to nginx html directory
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

MIT
