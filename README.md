# Sol Armada Dashboard

A dashboard and admin panel for Sol Armada member management, built with Vue 3 + Vuetify for the frontend and Go + PostgreSQL for the backend.

## Project Structure

```
.
├── api/                          # Go backend API
│   ├── cmd/server/              # Server entrypoint
│   ├── internal/                # Internal packages
│   │   ├── auth/                # Discord OAuth and session management
│   │   ├── middleware/          # HTTP middleware (auth, RBAC, CSRF, logging)
│   │   ├── handlers/            # HTTP request handlers
│   │   ├── models/              # Domain entities (imported from sol-bot)
│   │   ├── storage/             # Database repositories
│   │   ├── service/             # Business logic services
│   │   ├── dto/                 # Request/response DTOs
│   │   └── errors/              # Custom error types
│   ├── config/                  # Configuration files (TOML)
│   ├── migrations/              # Database migrations
│   └── go.mod                   # Go module definition
├── web/                          # Vue 3 + Vuetify frontend
│   ├── src/
│   │   ├── pages/               # Page components (auth, member, admin)
│   │   ├── layouts/             # Layout components
│   │   ├── components/          # Reusable components
│   │   ├── router/              # Vue Router config
│   │   ├── stores/              # Pinia stores
│   │   ├── composables/         # Vue composables
│   │   ├── plugins/             # Vuetify and other plugins
│   │   └── utils/               # Utility functions
│   ├── package.json             # Node dependencies
│   └── vite.config.ts           # Vite configuration
├── makefile                      # Build targets
└── README.md                     # This file
```

## Setup & Development

### Backend

1. Ensure Go 1.26.3+ is installed
2. Copy `.env.example` to `.env.local` and fill in Discord OAuth credentials:
   ```bash
   cp .env.example .env.local
   ```
3. Configure PostgreSQL connection in `api/config/config.local.toml`
4. Run migrations (TODO: implement)
5. Start the backend:
   ```bash
   make dev-server
   ```

### Frontend

1. Install Node dependencies:
   ```bash
   cd web && yarn install
   ```
2. Start the dev server:
   ```bash
   yarn dev
   ```
   Frontend runs on `http://localhost:5173`, proxies `/api` to backend.

### Both (Development)

```bash
make dev-web    # Terminal 1: Frontend with hot reload
make dev-server # Terminal 2: Backend with auto-reload
```

## Build

```bash
make build              # Production build (default)
make build-production   # Production: web + server
make build-beta         # Beta: web + server
```

Output: `./bin/api` (backend binary)

## Testing

### Backend unit tests

```bash
cd api && go test ./...
```

### Frontend unit tests

```bash
cd web && npm run test:run
```

### Run all tests

```bash
make test
```

## Configuration

### Environment Variables

See `.env.example`. Key variables:
- `ENVIRONMENT`: local | beta | production
- `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`: Discord OAuth app credentials
- `JWT_SECRET`: Secret for signing session JWTs
- `DATABASE_DSN`: PostgreSQL connection string

### Config Files

- `api/config/config.local.toml`: Local development
- `api/config/config.beta.toml`: Staging/beta environment
- `api/config/config.production.toml`: Production environment

Role IDs are configured in the `[roles]` section of each config file.

## Deployment

Backend is deployed as a systemd service or container. Frontend is built as static assets and served by the backend or a CDN.

## Implementation Roadmap

- **Phase 1**: Foundation ✅
- **Phase 2**: Auth & session core (Discord OAuth)
- **Phase 3**: Backend domain APIs & RBAC
- **Phase 4**: Frontend shell & design system
- **Phase 5**: Member pages (dashboard, profile)
- **Phase 6**: Admin pages (overview, attendance, token ledger, members)
- **Phase 7**: Hardening & observability ✅
- **Phase 8**: Deployment & DevOps ✅
- **Phase 9**: Testing & documentation ✅

## Tech Stack

- **Frontend**: Vue 3, TypeScript, Vuetify 3, Vue Router, Pinia, Axios
- **Backend**: Go 1.26.3, PostgreSQL, gorilla/mux, pgx, golang.org/x/oauth2
- **DevOps**: Docker (optional), systemd, Make
