# Phase 3: Domain Services & Database Integration

Phase 3 implements PostgreSQL database connectivity, storage layers, business logic services, and integrates persistent user data with the authentication system.

## Architecture

### Database Layer
```
┌─────────────────┐
│   Handlers      │  ← HTTP request handlers
└────────┬────────┘
         │
┌────────▼────────┐
│   Services      │  ← Business logic layer
└────────┬────────┘
         │
┌────────▼────────┐
│   Storage       │  ← Data access interfaces
└────────┬────────┘
         │
┌────────▼────────┐
│   PostgreSQL    │  ← Database with pgx driver
└─────────────────┘
```

### Components

**1. Database Connection Pool** ([internal/storage/db.go](api/internal/storage/db.go))
- Connection pool management with pgxpool
- Configurable pool size and timeouts
- Health checks and statistics
- Graceful shutdown support

**2. Storage Interfaces** ([internal/storage/interfaces.go](api/internal/storage/interfaces.go))
- `UserStorage`: CRUD operations for users
- `SessionStorage`: Session management operations
- Clean interface for multiple implementations (PostgreSQL, in-memory, etc.)

**3. PostgreSQL Implementations**
- [internal/storage/user_postgres.go](api/internal/storage/user_postgres.go): User data persistence
- [internal/storage/session_postgres.go](api/internal/storage/session_postgres.go): Session data persistence
- Proper error handling with custom error types
- Soft delete support for users
- Efficient indexing strategy

**4. Service Layer** ([internal/service/user_service.go](api/internal/service/user_service.go))
- `UserService`: Business logic for user and session management
- Get or create user pattern for OAuth flows
- Session creation and cleanup
- Structured logging for audit trails

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,          -- UUID
    discord_id VARCHAR(32) UNIQUE,       -- Discord user ID
    username VARCHAR(255),               -- Discord username
    email VARCHAR(255),                  -- Discord email
    avatar VARCHAR(512),                 -- Avatar URL
    roles TEXT[],                        -- Application roles array
    created_at TIMESTAMP,                -- Record creation time
    updated_at TIMESTAMP,                -- Last update time
    deleted_at TIMESTAMP                 -- Soft delete timestamp
);

-- Indexes
CREATE INDEX idx_users_discord_id ON users(discord_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC) WHERE deleted_at IS NULL;
```

### Sessions Table
```sql
CREATE TABLE sessions (
    id VARCHAR(36) PRIMARY KEY,          -- UUID
    user_id VARCHAR(36) REFERENCES users(id) ON DELETE CASCADE,
    token TEXT,                          -- JWT token
    expires_at TIMESTAMP,                -- Expiration time
    created_at TIMESTAMP                 -- Creation time
);

-- Indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_token ON sessions(token) WHERE expires_at > NOW();
```

## Migrations

### Migration System
- Uses golang-migrate/migrate v4
- File-based migrations in `api/migrations/`
- Up and down migration support
- Dirty state detection and recovery

### Migration Commands
```bash
# Apply all migrations
make migrate-up

# Rollback one migration
make migrate-down

# Create new migration
make migrate-create
```

### Migration Tool ([cmd/migrate/main.go](api/cmd/migrate/main.go))
- Standalone CLI tool for database migrations
- Reads DSN from `SA_DATABASE_DSN` environment variable
- Supports step-wise migrations
- Tracks migration version and dirty state

## Integration

### Auth Flow with Database

1. **User Authenticates via Discord**
   - OAuth callback receives Discord user data
   - Handler calls `UserService.GetOrCreateUser()`
   
2. **Get or Create User**
   - Check if user exists by Discord ID
   - If exists: Update username, email, avatar, roles
   - If new: Create user record with UUID
   
3. **Generate Session**
   - Create JWT token with user claims
   - Store session in database with expiry
   - Set secure httpOnly cookie
   
4. **Subsequent Requests**
   - Middleware validates JWT from cookie
   - User context available in handlers
   - Session tracked in database

### Service Methods

**UserService:**
- `GetOrCreateUser()` - OAuth user provisioning
- `GetUserByID()` - Fetch user by ID
- `GetUserByDiscordID()` - Fetch by Discord ID
- `UpdateUserRoles()` - Update role assignments
- `ListUsers()` - Paginated user listing
- `CreateSession()` - Create new session
- `DeleteSession()` - Logout single session
- `DeleteUserSessions()` - Logout all devices
- `CleanupExpiredSessions()` - Remove stale sessions

## Configuration

### Database Connection
```bash
SA_DATABASE_DSN=postgres://user:password@localhost:5432/solarmada
SA_DATABASE_MAX_CONNECTIONS=50
SA_DATABASE_IDLE_TIMEOUT_SECONDS=300
```

### Connection Pool Settings
- Max connections: Configurable (default: 50)
- Max connection lifetime: 1 hour
- Health check period: 1 minute
- Idle timeout: Configurable (default: 5 minutes)

## Error Handling

### Custom Errors
- `ErrUserNotFound` - User lookup failed
- `ErrUserAlreadyExists` - Duplicate user creation
- `ErrSessionNotFound` - Session lookup failed

### Error Propagation
```go
// Storage layer returns typed errors
user, err := storage.GetByID(ctx, id)
if errors.Is(err, storage.ErrUserNotFound) {
    return nil, http.StatusNotFound
}
```

## Security Features

### Data Protection
- **Soft Deletes**: Users are marked deleted, not removed
- **Cascading Deletes**: Sessions auto-delete when user is deleted
- **Parameterized Queries**: All queries use prepared statements
- **Connection Pooling**: Prevents connection exhaustion attacks

### Session Management
- JWT tokens stored in database for revocation support
- Expired sessions automatically excluded from queries
- Periodic cleanup of expired sessions
- Per-user session tracking (logout all devices support)

## Testing

### Database Setup
1. **Create PostgreSQL database:**
   ```bash
   createdb solarmada
   ```

2. **Run migrations:**
   ```bash
   make migrate-up
   ```

3. **Start server:**
   ```bash
   make dev-server
   ```

### Manual Testing
```bash
# Check database connection
psql $SA_DATABASE_DSN -c "SELECT version();"

# View users
psql $SA_DATABASE_DSN -c "SELECT id, username, email, roles FROM users;"

# View sessions
psql $SA_DATABASE_DSN -c "SELECT id, user_id, expires_at FROM sessions;"

# Clean expired sessions manually
psql $SA_DATABASE_DSN -c "DELETE FROM sessions WHERE expires_at <= NOW();"
```

## Performance Considerations

### Indexing Strategy
- Discord ID indexed for fast OAuth lookups
- Session token indexed for validation (partial index on active sessions)
- User ID indexed in sessions for fast user session queries
- Created_at indexed for sorted user listings

### Connection Pooling
- Pool size based on expected concurrency
- Idle connections released after timeout
- Health checks prevent stale connections
- Connection lifetime prevents memory leaks

### Query Optimization
- Partial indexes reduce index size
- WHERE clauses leverage indexes
- Soft delete filter in all user queries
- Expired session filter in all session queries

## Dependencies Added

```go
require (
    github.com/jackc/pgx/v5 v5.5.5          // PostgreSQL driver
    github.com/golang-migrate/migrate/v4     // Database migrations
    github.com/google/uuid v1.6.0            // UUID generation
)
```

## File Structure

```
api/
├── cmd/
│   ├── migrate/
│   │   └── main.go              # Migration CLI tool
│   └── server/
│       └── main.go              # Server with DB initialization
├── internal/
│   ├── service/
│   │   └── user_service.go      # Business logic layer
│   └── storage/
│       ├── db.go                # Connection pool manager
│       ├── interfaces.go        # Storage interfaces
│       ├── user_postgres.go     # User storage implementation
│       └── session_postgres.go  # Session storage implementation
└── migrations/
    ├── 000001_create_users_and_sessions.up.sql
    └── 000001_create_users_and_sessions.down.sql
```

## Next Steps (Phase 4+)

- [ ] **Frontend Integration**: Connect Vue frontend to auth endpoints
- [ ] **Session Cleanup Job**: Periodic background task for expired sessions
- [ ] **User Profile Pages**: View and edit user profiles
- [ ] **Admin User Management**: List, view, edit users (admin only)
- [ ] **Attendance Tracking**: Leverage sol-bot attendance entities
- [ ] **Token Ledger**: Integrate token tracking system
- [ ] **Rank System**: Display and manage member ranks

## Logging

All database operations include structured logging:
- User creation/update events
- Session creation/deletion events
- Connection pool statistics
- Migration operations
- Error details with context

Example log output:
```json
{
  "level": "info",
  "msg": "User created",
  "user_id": "abc-123",
  "time": "2026-05-31T10:30:00Z"
}
```

## Monitoring

### Health Checks
- Database ping on startup
- Connection pool stats available via `db.Stats()`
- Migration version tracking

### Metrics to Monitor
- Active database connections
- Idle connections
- Query latency
- Failed login attempts
- Session creation rate
- Expired session cleanup count
