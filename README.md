# ChattyCathy

A full-stack application with a Go API backend and a Next.js frontend built with Bun.

## Tech Stack

### Frontend (App)

- **Next.js 16** with App Router
- **React 19** with React Compiler
- **Bun** as package manager and runtime
- **TypeScript**
- **Tailwind CSS** with shadcn/ui components
- **Zustand** for state management
- **TanStack Query (React Query)** for data fetching
- **React Hook Form** with Zod validation
- **next-intl** for internationalization (i18n)

### Backend (API)

- **Go** with Gin framework
- **GORM** for database ORM
- **PostgreSQL** database
- **Redis** for session/token storage
- **JWT RS256** authentication with refresh tokens
- **Google OAuth** for social login
- **RBAC** with permission-based access control
- **OpenAPI/Swagger** documentation
- **Zerolog** for structured logging

### Infrastructure

- **Docker** with Docker Compose
- **Traefik** reverse proxy with automatic service discovery

## Prerequisites

### Docker (Recommended)

The easiest way to run the project is with Docker. Install Docker Desktop for your platform:

- **Windows**: [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/)
- **macOS**: [Docker Desktop for Mac](https://docs.docker.com/desktop/install/mac-install/)
- **Linux**: [Docker Engine](https://docs.docker.com/engine/install/)

### Local Development (Optional)

If you prefer to run services locally without Docker:

#### Go (for API)

- **Windows**: Download from [go.dev/dl](https://go.dev/dl/) or use `winget install GoLang.Go`
- **macOS**: `brew install go`
- **Linux**: `sudo apt install golang` or download from [go.dev/dl](https://go.dev/dl/)

#### Bun (for App)

- **Windows**:

  ```powershell
  powershell -c "irm bun.sh/install.ps1 | iex"
  ```

- **macOS**:

  ```bash
  brew install oven-sh/bun/bun
  # or
  curl -fsSL https://bun.sh/install | bash
  ```

- **Linux**:
  ```bash
  curl -fsSL https://bun.sh/install | bash
  ```

#### PostgreSQL (for local development without Docker)

- **Windows**: Download from [postgresql.org](https://www.postgresql.org/download/windows/) or use `winget install PostgreSQL.PostgreSQL`
- **macOS**: `brew install postgresql@16`
- **Linux**: `sudo apt install postgresql`

---

## Quick Start (Docker)

1. **Clone the repository**

   ```bash
   git clone https://github.com/your-org/chattycathy.git
   cd chattycathy
   ```

2. **Copy environment file**

   ```bash
   cp .env.example .env
   ```

3. **Start all services**

   ```bash
   make up
   ```

4. **Verify it's running**

   ```bash
   curl http://localhost:8080/api/v1/ping
   # Returns: {"message":"pong"}
   ```

5. **View API documentation**

   Open http://localhost:8080/api/docs/ in your browser.

6. **View Traefik dashboard**

   Open http://localhost:8081 or http://traefik.localhost in your browser.

---

## Development

### Using Docker (Recommended)

```bash
# Start all services
make up

# View logs
make logs

# Stop all services
make down

# Rebuild after code changes
make up-build

# Reset database
make db-reset

# Connect to database shell
make db-shell
```

### Local Development (Without Docker)

#### API

```bash
cd api
cp ../.env.example .env
go mod download
go run cmd/server/main.go
```

#### App

```bash
cd app
bun install
bun run dev
```

---

## Project Structure

```
chattycathy/
├── api/                    # Go API server
│   ├── cmd/server/         # Entry point
│   ├── config/             # Configuration
│   ├── db/                 # Database connection & models
│   ├── docs/               # OpenAPI documentation
│   └── internal/           # Internal packages
├── app/                    # Frontend application (Bun)
├── traefik/                # Traefik reverse proxy config
│   ├── traefik.yml         # Static configuration
│   └── dynamic.yml         # Dynamic configuration (middlewares)
├── scripts/                # Database init scripts
├── docker-compose.yml      # Docker orchestration
├── Makefile                # Development commands
└── README.md
```

---

## Available Make Commands

| Command             | Description                                     |
| ------------------- | ----------------------------------------------- |
| `make up`           | Start all services in Docker                    |
| `make down`         | Stop all services                               |
| `make logs`         | View logs (follow mode)                         |
| `make logs-api`     | View API logs only                              |
| `make logs-app`     | View App logs only                              |
| `make logs-traefik` | View Traefik logs only                          |
| `make db-seed`      | Seed database with dummy data                   |
| `make db-reset`     | Reset database (removes all data)               |
| `make db-shell`     | Connect to PostgreSQL shell                     |
| `make build`        | Build all Docker images                         |
| `make clean`        | Remove containers, volumes, and build artifacts |
| `make help`         | Show all available commands                     |

---

## API Endpoints

### Health & Status

| Method | Endpoint            | Description                                      |
| ------ | ------------------- | ------------------------------------------------ |
| GET    | `/api/v1/ping`      | Ping endpoint (requires auth + `ping:read` perm) |
| GET    | `/api/v1/health`    | Liveness probe                                   |
| GET    | `/api/v1/ready`     | Readiness probe (checks DB & Redis)              |
| GET    | `/api/docs/`        | Swagger UI documentation                         |
| GET    | `/api/openapi.yaml` | OpenAPI specification                            |

### Authentication

| Method | Endpoint                     | Description                              |
| ------ | ---------------------------- | ---------------------------------------- |
| GET    | `/api/v1/auth/google/config` | Get Google OAuth client configuration    |
| POST   | `/api/v1/auth/google`        | Authenticate with Google OAuth token     |
| POST   | `/api/v1/auth/login`         | Login with username/password (demo only) |
| POST   | `/api/v1/auth/refresh`       | Refresh tokens (rotates refresh token)   |
| POST   | `/api/v1/auth/logout`        | Logout current session                   |
| POST   | `/api/v1/auth/logout-all`    | Logout all sessions (requires auth)      |
| GET    | `/api/v1/auth/sessions`      | List all active sessions (requires auth) |

### Protected Routes

| Method | Endpoint                            | Description                        |
| ------ | ----------------------------------- | ---------------------------------- |
| GET    | `/api/v1/protected/secret`          | Protected endpoint (requires auth) |
| GET    | `/api/v1/protected/profile`         | Get current user profile           |
| GET    | `/api/v1/protected/admin/dashboard` | Admin only endpoint                |

### Admin Routes (requires admin role)

| Method | Endpoint                            | Description                      |
| ------ | ----------------------------------- | -------------------------------- |
| GET    | `/api/v1/admin/permissions`         | List all available permissions   |
| GET    | `/api/v1/admin/roles`               | List all roles with permissions  |
| GET    | `/api/v1/admin/roles/:id`           | Get a specific role              |
| POST   | `/api/v1/admin/roles`               | Create a new role                |
| PUT    | `/api/v1/admin/roles/:id`           | Update a role                    |
| DELETE | `/api/v1/admin/roles/:id`           | Delete a role (non-system only)  |
| PUT    | `/api/v1/admin/roles/:id/permissions` | Set permissions for a role     |

---

## Service URLs

| Service  | URL                                                   |
| -------- | ----------------------------------------------------- |
| App      | http://chattycathy.localhost or http://localhost:3000 |
| API      | http://api.localhost or http://localhost:8080         |
| API Docs | http://localhost:8080/api/docs/                       |
| Traefik  | http://traefik.localhost or http://localhost:8081     |
| Database | localhost:5432                                        |
| Redis    | localhost:6379                                        |

> **Note**: The `*.localhost` domains work automatically in most browsers. If not, add them to your hosts file.

---

## Environment Variables

### Database

| Variable      | Default       | Description                      |
| ------------- | ------------- | -------------------------------- |
| `DB_HOST`     | `localhost`   | Database host                    |
| `DB_PORT`     | `5432`        | Database port                    |
| `DB_USER`     | `postgres`    | Database user                    |
| `DB_PASSWORD` | `postgres`    | Database password                |
| `DB_NAME`     | `chattycathy` | Database name                    |
| `DB_SSLMODE`  | `disable`     | SSL mode for database connection |

### Redis

| Variable         | Default     | Description    |
| ---------------- | ----------- | -------------- |
| `REDIS_HOST`     | `localhost` | Redis host     |
| `REDIS_PORT`     | `6379`      | Redis port     |
| `REDIS_PASSWORD` | (empty)     | Redis password |
| `REDIS_DB`       | `0`         | Redis database |

### JWT Authentication

| Variable                  | Default             | Description                    |
| ------------------------- | ------------------- | ------------------------------ |
| `JWT_PRIVATE_KEY_PATH`    | `./jwt/private.pem` | Path to RSA private key        |
| `JWT_PUBLIC_KEY_PATH`     | `./jwt/public.pem`  | Path to RSA public key         |
| `JWT_ISSUER`              | `chattycathy`       | JWT issuer claim               |
| `JWT_ACCESS_EXPIRY_MINS`  | `15`                | Access token expiry in minutes |
| `JWT_REFRESH_EXPIRY_DAYS` | `7`                 | Refresh token expiry in days   |

### Google OAuth

| Variable                       | Default | Description                   |
| ------------------------------ | ------- | ----------------------------- |
| `GOOGLE_CLIENT_ID`             | -       | Google OAuth client ID        |
| `GOOGLE_CLIENT_SECRET`         | -       | Google OAuth client secret    |
| `GOOGLE_REDIRECT_URI`          | -       | OAuth redirect URI            |
| `NEXT_PUBLIC_GOOGLE_CLIENT_ID` | -       | Google client ID for frontend |

### Server

| Variable | Default | Description     |
| -------- | ------- | --------------- |
| `PORT`   | `8080`  | API server port |

---

## Authentication

The API uses JWT with RSA-256 asymmetric encryption. This provides:

- **Short-lived access tokens** (15 minutes by default) for API requests
- **Long-lived refresh tokens** (7 days by default) stored in Redis
- **Token rotation** on refresh for enhanced security
- **Session management** to view and revoke active sessions

### Google OAuth (Recommended)

The primary authentication method is Google OAuth:

1. User clicks "Sign in with Google" on the login page
2. Google authenticates the user and returns an access token
3. Frontend sends the token to `/api/v1/auth/google`
4. Backend verifies the token, creates/finds the user, and returns JWT tokens
5. User object includes their permissions based on assigned roles

```bash
# Authenticate with Google access token
curl -X POST http://localhost:8080/api/v1/auth/google \
  -H "Content-Type: application/json" \
  -d '{"access_token": "<google_access_token>"}'
```

### Role-Based Access Control (RBAC)

The API implements permission-based access control:

**Default Roles:**
| Role | Permissions | Notes |
| ------ | -------------------------------------------------------- | ------------------------------ |
| admin | All permissions | Must be manually assigned |
| user | `ping:read`, `news:read` | Auto-assigned to new users |
| editor | `ping:read`, `news:read`, `news:create`, `news:update` | Must be manually assigned |

**Automatic Role Assignment:**

- New users registered via Google OAuth are automatically assigned the **user** role
- This grants `ping:read` and `news:read` permissions by default
- Existing users without roles are also assigned the **user** role during migrations

**Available Permissions:**

- `ping:read` - Access the ping endpoint
- `news:read`, `news:create`, `news:update`, `news:delete`
- `users:read`, `users:update`, `users:delete`, `users:manage_roles`
- `roles:read`, `roles:create`, `roles:update`, `roles:delete`

Permissions are included in JWT token claims and returned in the user object on login.

**Admin UI:**

Admins with `roles:update` permission can manage role permissions through the web UI:

1. Log in with an admin account
2. Click your profile picture in the navbar
3. Select "Role & Permission Management"
4. Expand a role to view/modify its permissions
5. Toggle permissions on/off and click "Save Changes"

> **Note:** Permissions are baked into the JWT at login time. After changing role permissions, affected users must log out and log back in to get their updated permissions.

**Assigning Additional Roles:**

```sql
-- Assign admin role to a user
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.email = 'user@example.com' AND r.name = 'admin';
```

### Web Clients

For web browsers, tokens are automatically handled via HTTP-only cookies:

```bash
# Login - sets cookies automatically
curl -X POST http://api.localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  -c cookies.txt

# Access protected routes with cookies
curl http://api.localhost/api/v1/protected/profile \
  -b cookies.txt

# Refresh tokens - reads from cookie, sets new cookies
curl -X POST http://api.localhost/api/v1/auth/refresh \
  -b cookies.txt -c cookies.txt

# Logout
curl -X POST http://api.localhost/api/v1/auth/logout \
  -b cookies.txt
```

### Mobile/API Clients

For mobile apps or API clients, use the Authorization header:

```bash
# Login - get tokens in response body
curl -X POST http://api.localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'
# Response: {"access_token":"...","refresh_token":"...","token_type":"Bearer",...}

# Access protected routes with Bearer token
curl http://api.localhost/api/v1/protected/profile \
  -H "Authorization: Bearer <access_token>"

# Refresh tokens - send refresh token in body or header
curl -X POST http://api.localhost/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
# Or use header:
curl -X POST http://api.localhost/api/v1/auth/refresh \
  -H "X-Refresh-Token: <refresh_token>"

# View active sessions
curl http://api.localhost/api/v1/auth/sessions \
  -H "Authorization: Bearer <access_token>"

# Logout all sessions
curl -X POST http://api.localhost/api/v1/auth/logout-all \
  -H "Authorization: Bearer <access_token>"
```

### Demo Credentials

| Email               | Password   | Role  |
| ------------------- | ---------- | ----- |
| `admin@example.com` | `admin123` | admin |
| `user@example.com`  | `user123`  | user  |

---

## Troubleshooting

### Port 5432 already in use

If you have a local PostgreSQL running:

```bash
# Linux
sudo systemctl stop postgresql

# macOS
brew services stop postgresql

# Windows (PowerShell as Admin)
Stop-Service postgresql*
```

### Docker containers not starting

```bash
# Check container status
make ps

# View logs for errors
make logs

# Reset everything
make clean
make up
```

### API returns 404

Make sure you're using the correct path with `/api/v1/` prefix:

```bash
curl http://localhost:8080/api/v1/ping
```

---

## License

MIT
