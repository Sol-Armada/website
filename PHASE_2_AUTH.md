# Phase 2: Authentication & Session Management

Phase 2 implements Discord OAuth2 authentication with secure session management using JWT tokens and httpOnly cookies.

## Architecture

### Auth Flow
1. **Login Initiation** (`GET /auth/login`)
   - Generates OAuth state for CSRF protection
   - Redirects to Discord authorization page
   - Stores state in temporary cookie (5 min expiry)

2. **OAuth Callback** (`GET /auth/callback`)
   - Validates state parameter against cookie
   - Exchanges authorization code for access token
   - Fetches user info from Discord API
   - Fetches guild member roles
   - Maps Discord roles to app roles (admin, moderator, member)
   - Generates JWT token with user claims
   - Sets secure httpOnly session cookie (7 day expiry)
   - Sets CSRF cookie for state-changing requests
   - Returns user info and CSRF token to frontend

3. **Protected Routes**
   - Middleware validates session cookie JWT
   - Extracts user claims and stores in Echo context
   - Routes can require specific roles

4. **Logout** (`POST /auth/logout`)
   - Clears session and CSRF cookies
   - Frontend redirects to login

## Components

### JWT Token Service (`internal/auth/jwt.go`)
- **TokenService**: Creates and validates JWT tokens
- Claims include: userID, discordID, username, email, roles
- Uses HS256 signing algorithm
- Configurable expiry duration (default: 168 hours / 7 days)
- Token refresh capability (generates new token from existing claims)

### Cookie Service (`internal/auth/cookie.go`)
- **CookieService**: Manages secure cookie operations
- Session cookie: httpOnly, Secure (in production), SameSite=Lax
- CSRF cookie: JavaScript-readable, Secure (in production), SameSite=Strict
- Domain configurable for cross-subdomain auth
- Automatic secure flag based on environment

### Auth Middleware (`internal/middleware/auth.go`)
- **RequireAuth**: Validates session, extracts user claims to context
- **RequireRole**: Checks user has one of allowed roles (admin, moderator, etc.)
- **OptionalAuth**: Validates session if present but doesn't require it
- **CSRFMiddleware**: Validates CSRF tokens for POST/PUT/DELETE requests

### Auth Handler (`internal/handlers/auth.go`)
- Discord OAuth2 client configuration
- Login/callback/logout/me endpoints
- Fetches user and guild member info from Discord API
- Maps Discord role IDs to application roles
- Secure state validation

## Configuration

Required environment variables:
```bash
SA_DISCORD_CLIENT_ID=<your-client-id>
SA_DISCORD_CLIENT_SECRET=<your-client-secret>
SA_DISCORD_REDIRECT_URI=http://localhost:3000/auth/callback
SA_DISCORD_GUILD_ID=<your-guild-id>
SA_AUTH_JWT_SECRET=<secure-random-secret>
SA_ROLES_ADMIN_ROLE_ID=<discord-admin-role-id>
SA_ROLES_MODERATOR_ROLE_ID=<discord-moderator-role-id>
```

## API Endpoints

### Public Endpoints
- `GET /health` - Health check (always available)
- `GET /version` - API version info
- `GET /auth/login` - Initiate Discord OAuth flow
- `GET /auth/callback` - OAuth callback handler

### Protected Endpoints
- `POST /auth/logout` - Clear session (requires auth)
- `GET /auth/me` - Get current user info (requires auth)
- `GET /api/*` - All API routes require authentication

## Security Features

1. **CSRF Protection**
   - OAuth state parameter validation
   - CSRF tokens for state-changing operations
   - Double-submit cookie pattern

2. **Secure Cookies**
   - httpOnly flag prevents XSS access to session tokens
   - Secure flag (production only) ensures HTTPS-only transmission
   - SameSite prevents CSRF attacks
   - Short-lived CSRF cookies (1 hour)
   - Longer session cookies (7 days)

3. **Token Security**
   - JWT signed with HS256
   - Configurable secret key
   - Expiry validation
   - Refresh capability without re-authentication

4. **Guild Verification**
   - Requires Discord guild membership
   - Fetches real-time role information
   - Denies access if not guild member

## Role Mapping

Discord roles map to application roles:
- **Admin Role ID** → `admin` (full access)
- **Moderator Role ID** → `moderator` (limited admin access)
- All guild members → `member` (basic access)

Routes can require specific roles:
```go
adminGroup := e.Group("/admin")
adminGroup.Use(authMiddleware.RequireRole("admin", "moderator"))
```

## Models & DTOs

### User Model (`internal/models/user.go`)
```go
type User struct {
    ID        string
    DiscordID string
    Username  string
    Email     string
    Avatar    string
    Roles     []string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Auth Response (`internal/dto/auth.go`)
```go
type AuthResponse struct {
    User  UserDTO
    CSRF  string
}
```

## Testing

**Start development server:**
```bash
make dev-server
```

**Test auth flow manually:**
1. Navigate to `http://localhost:8080/auth/login`
2. Authorize with Discord
3. Check callback redirects and sets cookies
4. Test `/auth/me` endpoint with session cookie
5. Test `/auth/logout` clears cookies

## Next Steps (Phase 3+)

- [ ] Database integration for user persistence
- [ ] Redis session store (optional, for multi-instance deployments)
- [ ] Token refresh endpoint
- [ ] Rate limiting on auth endpoints
- [ ] Audit logging for auth events
- [ ] 2FA support (optional future enhancement)

## Dependencies Added

- `github.com/golang-jwt/jwt/v5` - JWT token generation and validation
- `golang.org/x/oauth2` - OAuth2 client implementation
- Echo middleware for CORS support
