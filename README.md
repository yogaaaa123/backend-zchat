# Obrolan API — Social Forum Backend

**Obrolan** (Indonesian for "conversation") is a production-ready social forum backend API built with Go. Features thread discussions, nested comments, likes, real-time WebSocket chat, image uploads, JWT authentication with refresh token rotation, rate limiting, and full Swagger documentation.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Language** | Go 1.26.3 |
| **HTTP Framework** | Gin v1.12 |
| **Database** | PostgreSQL 16 + GORM v1.31 |
| **Auth** | JWT HS256 (golang-jwt v5) + bcrypt |
| **Real-time** | gorilla/websocket v1.5 (Hub pattern) |
| **Image Upload** | Cloudinary v2 |
| **Docs** | swaggo/swag v1.16 |
| **Testing** | stretchr/testify v1.11 |
| **Container** | Docker + docker-compose |
| **Validation** | go-playground/validator v10 |

## Architecture

Clean Architecture — strict 3-tier separation:

```
Client → Gin Router → Auth Middleware → Handler → Service → Repository → PostgreSQL
```

- **Handlers** — HTTP request/response, validation, Swagger annotations
- **Services** — Business logic, authorization checks, orchestration
- **Repositories** — Data access via GORM parameterized queries

## Features

### Auth
- Register (bcrypt, unique email/username)
- Login → JWT (15min) + refresh token (7d)
- Refresh token rotation (SHA256 hashed in DB)
- Refresh token reuse detection → auto-revoke all sessions
- Logout → revoke all refresh tokens
- Rate limiting: 20 req/min per IP on auth endpoints

### Threads
- CRUD with owner-only update/delete
- Pagination (`?page=&limit=`)
- Full-text search by title & content (`?q=`)
- `is_liked` per user on list & detail
- `like_count` + `comment_count` via subquery

### Comments
- Create top-level comment or nested reply (1 level)
- Paginated listing with preloaded replies + user
- Owner-only delete (cascade soft-deletes replies)

### Likes
- Toggle like/unlike per user+thread
- Unique constraint prevents duplicates
- Hard delete on unlike (avoids unique conflict)
- Accurate count via SQL COUNT

### Chat (WebSocket)
- Real-time per-thread chat rooms
- Hub pattern with register/unregister/broadcast
- Messages persisted to PostgreSQL
- Join/leave events broadcast to room
- History endpoint (last 50 messages)
- Ping/pong keepalive (60s timeout)

### Image Upload
- Cloudinary integration (configurable)
- Magic byte MIME validation (jpeg/png/gif/webp)
- Extension matching
- 5MB max file size
- Auto-delete from Cloudinary when thread deleted
- Graceful no-op fallback if Cloudinary not configured

## Database Schema

6 tables with UUID primary keys and soft deletes:

```
users              refresh_tokens         threads
├── id (UUID PK)   ├── id (UUID PK)       ├── id (UUID PK)
├── username (UK)  ├── user_id (FK)       ├── user_id (FK)
├── email (UK)     ├── token_hash (UK)    ├── title
├── password       ├── expires_at         ├── content
├── bio            └── created_at         ├── image_url
├── avatar_url                          ├── created_at (idx)
├── created_at                           ├── updated_at
├── updated_at                           └── deleted_at
└── deleted_at

comments             likes                messages
├── id (UUID PK)     ├── id (UUID PK)     ├── id (UUID PK)
├── user_id (FK)     ├── user_id (FK)     ├── thread_id (FK)
├── thread_id (FK)   ├── thread_id (FK)   ├── user_id (FK)
├── parent_id (FK)   ├── created_at       ├── content
├── content          └── UNIQUE           └── created_at (idx)
├── created_at (idx)    (user+thread)
├── updated_at
└── deleted_at
```

## API Endpoints

### Public
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/auth/register` | Register |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh access token |

### Protected (JWT)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/me` | Get current user |
| POST | `/api/v1/auth/logout` | Revoke all refresh tokens |
| GET | `/api/v1/users/me` | Get own profile |
| PUT | `/api/v1/users/me` | Update bio & avatar |
| PUT | `/api/v1/users/me/password` | Change password |
| GET | `/api/v1/users/:id` | Get public profile |
| POST | `/api/v1/upload` | Upload image |
| GET | `/api/v1/threads` | List threads (paginated) |
| GET | `/api/v1/threads/search?q=` | Search threads |
| POST | `/api/v1/threads` | Create thread |
| GET | `/api/v1/threads/:id` | Get thread detail |
| PUT | `/api/v1/threads/:id` | Update thread (owner) |
| DELETE | `/api/v1/threads/:id` | Delete thread (owner) |
| POST | `/api/v1/threads/:id/comments` | Create comment/reply |
| GET | `/api/v1/threads/:id/comments` | List comments (paginated) |
| DELETE | `/api/v1/comments/:id` | Delete comment (owner) |
| POST | `/api/v1/threads/:id/like` | Toggle like/unlike |
| GET | `/api/v1/threads/:id/messages` | Get chat history |

### WebSocket
| Path | Description |
|------|-------------|
| `GET /api/v1/threads/:id/chat?token=` | Real-time chat per thread |

### Swagger UI
```
/swagger/index.html
```

## Quick Start

### Prerequisites
- Go 1.26+
- PostgreSQL 16
- (Optional) Cloudinary account for image upload

### 1. Clone & Configure
```bash
git clone https://github.com/yogaaaa123/backend-zchat.git
cd backend-zchat
cp .env.example .env
# Edit .env with your credentials
```

### 2. Run with Docker
```bash
make docker-up
```

### 3. Run Locally
```bash
make run
```

### 4. Run Tests
```bash
make test        # all tests
make cover       # service coverage
make vet         # static analysis
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | Server port |
| `APP_ENV` | `development` | `development`, `production`, `staging` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `social_forum` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | *(required in prod)* | HMAC signing key |
| `JWT_EXPIRES_IN` | `15m` | Access token TTL |
| `REFRESH_TOKEN_EXPIRES_IN` | `168h` | Refresh token TTL (7d) |
| `CORS_ORIGINS` | `http://localhost:3000,http://localhost:5173` | Allowed origins |
| `CLOUDINARY_CLOUD_NAME` | — | Cloudinary cloud name |
| `CLOUDINARY_API_KEY` | — | Cloudinary API key |
| `CLOUDINARY_API_SECRET` | — | Cloudinary API secret |

## Testing

**128 tests** across 4 packages — all passing:

| Package | Tests | Coverage |
|---------|:-----:|:--------:|
| `internal/services` | 80 | ~42% |
| `internal/handlers` | 9 | — |
| `internal/middleware` | 20 | — |
| `internal/utils` | 19 | — |

- **DB:** SQLite in-memory, cleaned per test
- **WS:** Real gorilla/websocket connections via httptest
- **Framework:** testify (assert + require)
- **Handler:** Mock services + httptest recorder

```bash
go test ./... -count=1          # 128 tests
go test -v ./internal/services/ -cover
```

## Security

- **Passwords:** bcrypt cost 10, never exposed in JSON (`json:"-"`)
- **JWT:** HS256, validated on every protected request, "alg:none" rejected
- **Authorization:** Owner-only update/delete (IDOR prevention)
- **SQL Injection:** All queries via GORM parameterized methods
- **File Upload:** Magic byte validation + extension check + 5MB max
- **Refresh Token:** SHA256 hashed in DB, rotated on every use, reuse detection revokes all sessions
- **Rate Limiting:** 20 req/min on auth endpoints
- **WebSocket:** Origin whitelist from CORS config
- **Cascade Delete:** Thread soft-delete cascades to comments, likes, messages

## Makefile Commands

```bash
make run          # Start server
make build        # Build binary to bin/server
make test         # Run all tests
make cover        # Service coverage
make vet          # Static analysis
make swagger      # Regenerate Swagger docs
make docker-up    # docker compose up -d
make docker-down  # docker compose down
```

## License

MIT
