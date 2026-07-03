# Elite4Print Go Backend

A **modular monolith** backend for Elite4Print built with Go 1.26, following
**Domain-Driven Design (DDD)** and **Hexagonal Architecture**.

This starter set contains:

- `identity` module: user registration, profiles, roles.
- `auth` module: JWT + session-based authentication, Redis-backed session
  cache, token blacklisting, multi-session management, session revocation.
- Platform infrastructure: PostgreSQL via `sqlx`/`pgx`, Redis cache, NATS
  JetStream scaffolding for events and background workers.

---

## Quick Start

```bash
# 1. Copy environment file
cp .env.example .env

# 2. Start everything (Postgres + Redis + NATS + API)
make up

# 3. The API is available at http://localhost:8080
```

For hot-reload development:

```bash
make dev
```

---

## Architecture

### Hexagonal / Ports & Adapters

```text
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Layer                           │
│  ┌─────────────┐  ┌──────────────────────────────────────┐  │
│  │  chi router │  │ request/response DTOs + middleware   │  │
│  └──────┬──────┘  └──────────────────────────────────────┘  │
├─────────┼────────────────────────────────────────────────────┤
│         │ Application Layer (use cases / command handlers)   │
│         │  - RegisterUser, Login, Refresh, RevokeSession     │
├─────────┼────────────────────────────────────────────────────┤
│         │ Domain Layer (entities, value objects, ports)      │
│         │  - User, Session, Repository interfaces, events    │
├─────────┼────────────────────────────────────────────────────┤
│         ▼ Infrastructure Layer (adapters)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ PostgresRepo │  │ RedisCache   │  │ NATS JetStream   │   │
│  └──────────────┘  └──────────────┘  └──────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

Dependencies point inward. The domain knows nothing about HTTP, SQL, Redis, or
NATS.

### Module Layout

```text
internal/modules/<module>/
├── domain/              # Entities, value objects, repository ports, events
├── application/
│   ├── commands/        # Write use cases
│   ├── queries/         # Read use cases
│   └── dto.go           # Input/output DTOs
├── infrastructure/
│   ├── persistence/     # SQL repository implementations
│   └── cache/           # Redis adapters
└── presentation/http/   # chi handlers, requests, responses
```

### Dependency Injection

Go prefers **explicit constructor injection** over runtime DI containers. The
`cmd/api/main.go` file is the composition root: it builds every adapter and
injects it into use cases. If wiring grows, we can introduce `google/wire` for
code generation without changing module code.

---

## Authentication Strategy

We support **both JWT access tokens and server-side sessions**:

1. **Login** creates a `Session` row in PostgreSQL and returns:
   - Short-lived **access token** (JWT, default 15 min).
   - Long-lived **refresh token** (JWT containing the session ID, default 7 days).
2. **Access token** is stateless and used for every API call.
3. **Refresh token** is validated against the `sessions` table.
4. **Logout / revoke** marks the session as revoked and adds the access token
   JTI to a Redis blacklist.
5. **Multi-session**: a user can have up to `SESSION_MAX_CONCURRENT` active
   sessions. Oldest session is evicted when the limit is reached.

---

## Money Handling with `shopspring/decimal`

Financial fields use `shopspring/decimal.Decimal` instead of `float64` to avoid
binary floating-point rounding errors. PostgreSQL stores these as `NUMERIC`.

Example:

```go
price := decimal.NewFromFloat(19.99)
total := price.Mul(decimal.NewFromInt(quantity))
```

This directly addresses the Django backend's `FloatField` precision bug.

---

## Event Bus & Background Workers

The starter set uses an **in-memory event bus** for simplicity. NATS JetStream
adapters are provided under `internal/platform/nats/` and can be swapped in when
you are ready to scale horizontally or run background workers in separate
processes.

To switch to NATS:

```go
bus, err := platformnats.NewEventBus(cfg.NATSURL, "elite4print")
```

Background jobs are published with `WorkerEnqueue` and consumed with
`WorkerRegister`.

---

## Swagger / OpenAPI

We use **swaggo/swag** for auto-generating Swagger docs from Go comments.

```bash
make swagger
```

This scans `cmd/api/main.go` and handler comments and writes `docs/swagger.json`
and `docs/swagger.yaml`. You can then serve them with `http-swagger`.

---

## Commands

```bash
make help              # Show all commands
make up                # Start production stack
make dev               # Start development stack with hot reload
make down              # Stop all services
make test              # Run tests
make lint              # Run go vet
make migrate-up        # Apply Goose migrations
make migrate-down      # Rollback last migration
make migrate-create NAME=foo  # Create new migration
make swagger           # Generate Swagger docs
```

---

## Environment Variables

See `.env.example` for all variables. Key ones:

| Variable | Description |
|----------|-------------|
| `ENVIRONMENT` | `production` or `development` |
| `POSTGRES_*` | PostgreSQL connection |
| `REDIS_*` | Redis connection |
| `NATS_URL` | NATS connection |
| `JWT_SECRET_KEY` | HS256 secret (min 32 chars) |
| `JWT_ACCESS_TTL` | Access token lifetime |
| `JWT_REFRESH_TTL` | Refresh token / session lifetime |
| `SESSION_MAX_CONCURRENT` | Max active sessions per user |
| `RATE_LIMIT_RPS` / `RATE_LIMIT_BURST` | Rate limiting |

---

## Future Extraction to Microservices

Each module is self-contained. To extract a module into a service:

1. Move `internal/modules/<module>` to the new service.
2. Create `cmd/<module>-service/main.go`.
3. Replace the in-memory event bus with NATS JetStream.
4. Keep domain/application code unchanged.

---

## License

Proprietary — Elite4Print.
