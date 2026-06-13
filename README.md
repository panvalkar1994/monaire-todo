# monaire-todo

A production-style Go REST API for managing todo items. Built as a layered service with Cobra CLI commands, configurable persistence (SQLite or MySQL), database migrations, structured logging, and Docker Compose support.

**Base URL:** `http://localhost:8080/api/v1`

---

## What this project provides

| Area | Details |
|------|---------|
| **REST API** | Full CRUD for todos at `/api/v1/todos` |
| **List behaviour** | Sorted by `due_date` ascending; completed items hidden by default |
| **Filtering** | `include_completed` query param with strict validation |
| **Persistence** | SQLite (local dev) or MySQL (Docker Compose) via GORM |
| **Migrations** | Versioned SQL migrations with [golang-migrate](https://github.com/golang-migrate/migrate) |
| **Configuration** | YAML config + `MONAIRE_*` environment variable overrides |
| **Logging** | Structured `log/slog` (text or JSON) |
| **Graceful shutdown** | SIGINT/SIGTERM with 15s drain timeout |
| **API contract** | OpenAPI 3.0 spec, Postman collection, k6 smoke script |
| **Tests** | Unit, integration, and benchmark suites |

---

## Assumptions

- Todos are identified by UUID.
- Due dates use `YYYY-MM-DD`.
- Completed todos are hidden by default; use `include_completed=true` (or `1`, `yes`, etc.) to include them.
- Invalid `include_completed` values return HTTP 400.
- PUT performs full replacement; if the ID does not exist, the todo is created (upsert).
- PUT with identical data returns 200 and `X-No-Changes: true`.
- PATCH performs partial updates.
- Todos are permanently deleted.

---

## Tech stack

- **Go 1.25** — language & toolchain
- **Cobra + Viper** — CLI and configuration (`config/config.yaml`, `MONAIRE_*` env)
- **Gin** — HTTP router (no access-log middleware; errors logged via slog)
- **GORM** — ORM with repository interface pattern
- **golang-migrate** — schema migrations
- **Docker Compose** — MySQL 8.4 + API container

---

## Project layout

```
monaire-todo/
├── main.go                 # Entry point → cmd.Execute()
├── cmd/                    # Cobra commands (thin; no business logic)
│   ├── root.go             # Config flags, Viper init
│   ├── server.go           # server subcommand
│   └── migrate.go          # migrate up | down | version
├── config/
│   ├── config.example.yaml # Local SQLite template
│   └── config.docker.yaml  # Docker / MySQL template
├── internal/
│   ├── config/             # Typed config load & validation
│   ├── domain/             # Entities and domain errors
│   ├── repository/         # Repository interface + GORM impl
│   ├── service/            # Business logic (transport-agnostic)
│   ├── handler/            # HTTP/JSON adapter (Gin)
│   ├── server/             # HTTP server lifecycle
│   ├── database/           # DB connection
│   ├── migrate/            # Migration runner
│   └── logging/            # slog setup
├── migrations/             # SQL migration files
├── openapi/openapi.yaml    # OpenAPI 3.0 contract
├── postman/                # Postman collection + environment
├── scripts/k6/smoke.js     # k6 load/smoke test
├── test/integration/       # HTTP integration tests
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

---

## Prerequisites

**Local development**

- Go **1.25+**
- Make (optional, recommended)

**Docker Compose**

- Docker Engine **24+**
- Docker Compose v2 (`docker compose`)

**Optional tooling**

- [Postman](https://www.postman.com/) — import collection
- [k6](https://k6.io/docs/get-started/installation/) — run smoke script

---

## Quick start (local — SQLite)

The fastest way to run the API on your machine without Docker.

### 1. Clone and enter the repo

```bash
git clone <repository-url>
cd monaire-todo
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure (optional)

The Makefile uses `config/config.example.yaml` directly. For a personal config file:

```bash
cp config/config.example.yaml config/config.yaml
```

The SQLite database file is created automatically under `data/` — no manual directory setup required.

### 4. Run migrations

```bash
make migrate-up
```

Equivalent:

```bash
go run . --config config/config.example.yaml migrate up
```

### 5. Start the server

```bash
make run
```

Equivalent:

```bash
go run . --config config/config.example.yaml server
```

Expected startup output (abbreviated):

```
level=INFO msg="starting server" addr=:8080 gin_mode=debug
level=INFO msg="database connected" driver=sqlite
level=INFO msg=listening addr=:8080
```

Stop the server with **Ctrl+C** — it shuts down gracefully (15s timeout for in-flight requests).

### 6. Try the API

```bash
# Create a todo
curl -s -X POST http://localhost:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -d '{"description":"Buy groceries","due_date":"2026-06-15"}'

# List incomplete todos (sorted by due_date)
curl -s http://localhost:8080/api/v1/todos

# List all todos including completed
curl -s 'http://localhost:8080/api/v1/todos?include_completed=true'
```

---

## Docker Compose (MySQL)

Runs **MySQL 8.4** and the API in containers. **Migrations are manual** — run them once before starting the API (or after schema changes).

### 1. Start MySQL and apply migrations

```bash
# Start MySQL (or start everything in detached mode)
docker compose up -d mysql

# Apply migrations (one-off container; exits when done)
make docker-migrate-up
```

Equivalent:

```bash
docker compose run --rm api migrate up
```

### 2. Start the API

```bash
make docker-up
```

Equivalent:

```bash
docker compose up --build
```

Or start only the API if MySQL is already running:

```bash
docker compose up --build api
```

What happens:

1. **mysql** service starts and waits until healthy
2. **api** service builds from the Dockerfile and runs **`server` only** (no auto-migrate on restart)
3. MySQL DSN is injected via `MONAIRE_DATABASE_DSN`

Re-run `make docker-migrate-up` when you add new migration files.

### Verify

```bash
curl -s http://localhost:8080/api/v1/todos
```

### Stop and clean up

```bash
# Stop containers
docker compose down

# Stop and remove persisted MySQL volume
docker compose down -v
```

### Docker services

| Service | Image / build | Port | Purpose |
|---------|---------------|------|---------|
| `mysql` | `mysql:8.4` | 3306 (internal) | Primary database |
| `api` | Built from `Dockerfile` | **8080** | REST API |

Default MySQL credentials (development only):

| Setting | Value |
|---------|-------|
| Database | `todos` |
| User | `todo` / `todo` |
| Root password | `root` |

---

## CLI reference

```bash
# Start HTTP server
go run . server [--config path]

# Apply migrations
go run . migrate up [--config path]

# Roll back last migration
go run . migrate down [--config path]

# Show migration version
go run . migrate version [--config path]
```

**Makefile shortcuts**

| Target | Command |
|--------|---------|
| `make run` | Start server with example config |
| `make migrate-up` | Apply migrations (local SQLite) |
| `make docker-migrate-up` | Apply migrations in Docker (`docker compose run --rm api migrate up`) |
| `make docker-up` | Start Docker Compose stack |
| `make docker-down` | Stop Docker Compose stack |
| `make build` | Build binary to `bin/monaire-todo` |
| `make test-unit` | Unit tests (`-short`) |
| `make test-integration` | Integration tests |
| `make bench` | Benchmarks |

---

## API reference

All endpoints return JSON. Errors use `{ "error": "message" }`.

### Endpoints

| Method | Path | Status | Description |
|--------|------|--------|-------------|
| `POST` | `/api/v1/todos` | 201 | Create todo |
| `GET` | `/api/v1/todos` | 200 | List todos (incomplete by default) |
| `GET` | `/api/v1/todos/{id}` | 200 / 404 | Get by UUID |
| `PUT` | `/api/v1/todos/{id}` | 200 | Full replace; **creates** if ID missing (upsert) |
| `PATCH` | `/api/v1/todos/{id}` | 200 / 404 | Partial update |
| `DELETE` | `/api/v1/todos/{id}` | 204 / 404 | Hard delete |

### Todo object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "description": "Buy groceries",
  "due_date": "2026-06-15",
  "completed": false
}
```

| Field | Type | Notes |
|-------|------|-------|
| `id` | UUID string | Server-generated on create |
| `description` | string | Required, non-empty task details |
| `due_date` | string | ISO date `YYYY-MM-DD` |
| `completed` | boolean | Defaults to `false` on create |

### List — `include_completed`

| Value | Result |
|-------|--------|
| omitted, empty, `false`, `0`, `no` | Incomplete todos only |
| `true`, `1`, `yes` | All todos |
| anything else | **400** — `include_completed allowed values: true\|false\|empty` |

List results are always sorted by **`due_date` ascending**.

### PUT behaviour

- Replaces the entire todo (all fields required in body)
- If the ID does not exist, the todo is **created** with that ID (upsert)
- If the payload is identical to the stored todo, response is **200** with header `X-No-Changes: true`

### Example requests

```bash
# Create
curl -X POST http://localhost:8080/api/v1/todos \
  -H 'Content-Type: application/json' \
  -d '{"description":"Write README","due_date":"2026-06-20","completed":false}'

# Get (replace {id})
curl http://localhost:8080/api/v1/todos/{id}

# Partial update
curl -X PATCH http://localhost:8080/api/v1/todos/{id} \
  -H 'Content-Type: application/json' \
  -d '{"completed":true}'

# Full replace
curl -X PUT http://localhost:8080/api/v1/todos/{id} \
  -H 'Content-Type: application/json' \
  -d '{"description":"Write README","due_date":"2026-06-22","completed":true}'

# Delete
curl -X DELETE http://localhost:8080/api/v1/todos/{id}
```

Full contract: [`openapi/openapi.yaml`](openapi/openapi.yaml)

---

## Postman

1. Start the API (`make run` or Docker Compose).
2. **Import** → `postman/monaire-todo.postman_collection.json`
3. Optional environment: `postman/monaire-todo.postman_environment.json` (`baseUrl` = `http://localhost:8080`)
4. Run **Create Todo** first — it saves `todoId` for subsequent requests.
5. The **List** folder covers default, `include_completed`, and invalid-query (400) cases.

You can also import `openapi/openapi.yaml` directly as OpenAPI 3.0.

---

## k6 smoke test

Requires [k6](https://k6.io/docs/get-started/installation/). With the server running:

```bash
k6 run scripts/k6/smoke.js
```

Override host:

```bash
BASE_URL=http://localhost:8080 k6 run scripts/k6/smoke.js
```

The script covers create → get → list → patch → put → delete → 404, including `include_completed` and 400 validation.

---

## Configuration

Config files live in `config/`. Precedence: **`--config` file → `MONAIRE_*` env vars → `./config/config.yaml` → defaults**.

### Example (`config/config.example.yaml`)

```yaml
server:
  addr: ":8080"
  gin_mode: "debug"

database:
  driver: "sqlite"
  dsn: "file:data/todos.db?cache=shared"

migrations:
  path: "./migrations"

logging:
  level: "info"
  format: "text"
```

### Environment variables

| Config key | Environment variable | Example |
|------------|---------------------|---------|
| `server.addr` | `MONAIRE_SERVER_ADDR` | `:8080` |
| `server.gin_mode` | `MONAIRE_SERVER_GIN_MODE` | `release` |
| `database.driver` | `MONAIRE_DATABASE_DRIVER` | `mysql` |
| `database.dsn` | `MONAIRE_DATABASE_DSN` | `todo:todo@tcp(localhost:3306)/todos?parseTime=true` |
| `migrations.path` | `MONAIRE_MIGRATIONS_PATH` | `./migrations` |
| `logging.level` | `MONAIRE_LOGGING_LEVEL` | `info` |
| `logging.format` | `MONAIRE_LOGGING_FORMAT` | `json` |

Docker Compose sets `MONAIRE_DATABASE_DSN` and `MONAIRE_SERVER_GIN_MODE=release` on the API service. The container uses `config/config.docker.yaml` baked in at build time (JSON logging, MySQL driver).

---

## Logging

Structured logging via Go **`log/slog`**.

| Setting | Values | Default (local) |
|---------|--------|-----------------|
| `logging.level` | `debug`, `info`, `warn`, `error` | `info` |
| `logging.format` | `text`, `json` | `text` |

Logged events include server lifecycle, database connection, migration runs, HTTP requests (method, path, status, latency), HTTP 500s, and panics.

---

## Tests

```bash
make test-unit          # Fast unit tests (service, handler, etc.)
make test-integration   # Full HTTP flow against SQLite temp DB
make bench              # Handler benchmarks
```

Integration tests use an isolated SQLite file per test run — **no Docker or MySQL required**. Docker Compose MySQL is for local/production-like development only.

---

## Architecture notes

The codebase follows a **thin transport, thick service** layout:

```
HTTP (Gin handler) → Service (business rules) → Repository (GORM) → Database
```

- **Handlers** translate JSON/HTTP status codes; they do not contain business rules.
- **Services** are transport-agnostic and reusable (e.g. a future gRPC adapter could call the same service layer).
- **Repository interface** enables testing with mocks and swapping persistence details.

This separation keeps the REST layer replaceable while preserving domain logic and data access patterns.

---

## License

See repository license file (if applicable).
