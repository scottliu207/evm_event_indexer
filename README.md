# Simple EVM Event Indexer

Go service that indexes EVM logs from configured contracts, stores them in MySQL, tracks session/auth state in Redis, and exposes query/admin APIs via Gin.

## Features

- Event indexing by contract address + topics (`scanner*.json` driven).
- Background workers: API server, metrics server, scanner, subscription, reorg consumer.
- Reorg handling for removed logs with checkpoint rollback.
- JWT access token + opaque refresh token + CSRF protection for refresh endpoints.
- Admin and user auth flows with session revocation in Redis.
- Swagger/OpenAPI docs via `swaggo`.
- Prometheus metrics at `/metrics`.

## Requirements

- Go `1.25.5`
- Docker + Docker Compose (recommended)
- For optional tasks:
  - Swagger generation: `swag`
  - Contract flow: `forge`, `jq`, `abigen`

## Project Layout

```text
.
├── api/                 # HTTP handlers, middleware, routes
├── background/          # workers (api, metrics, scanner, subscription, reorg)
├── cmd/indexer/         # entrypoint + global Swagger annotations
├── config/              # config.yaml + scanner JSON files
├── docker/              # compose + DB bootstrap scripts/schema
├── docs/                # generated Swagger docs
├── internal/            # config/storage/session/decoder/eth/metrics
├── service/             # business logic + repositories
└── utils/               # shared utilities
```

## Quick Start (Docker)

Start full stack (MySQL, Redis, Anvil, indexer):

```bash
./run.sh up
```

Useful flags:

```bash
./run.sh up --infra   # only mysql, redis, anvil
./run.sh up --purge   # reset volumes before start
```

Stop:

```bash
./run.sh down
./run.sh down --purge
```

Default local URLs (compose):

- API: `http://localhost:8080`
- Swagger (enabled in compose): `http://localhost:8080/swagger/index.html`
- Metrics: `http://localhost:9090/metrics`
- MySQL: `127.0.0.1:3306`
- Redis: `127.0.0.1:6379`
- Anvil: `127.0.0.1:8545`

## Run Locally (Non-Docker)

1. Bring up infra (MySQL, Redis, Anvil), for example:

```bash
./run.sh up --infra
```

2. Run indexer:

```bash
make run
```

`make run` executes `go run cmd/indexer/main.go` and loads `./config/config.yaml`.

## Configuration

Main config: `config/config.yaml`

- App settings: `start_block`, `log_scanner_interval`, `reorg_window`, retry/backoff, timeouts.
- API settings under `api`: `port`, `timeout`, `enable_swagger`.
- DB settings under `mysql`/`redis`.
- Scanner file path: `scanner_path`.

Environment variables override YAML via Viper (`.` -> `_`), for example:

- `API_PORT`
- `API_ENABLE_SWAGGER`
- `SCANNER_PATH`
- `SESSION_JWT_SECRET`
- `SESSION_CSRF_SECRET`
- `MYSQL_DATABASES_*`
- `REDIS_DATABASES_*`

Scanner config files:

- `config/scanner.json` (local)
- `config/scanner.docker.json` (compose)

`topics[]` in scanner config are event signatures (e.g. `Transfer(address,address,uint256)`); the service hashes them to topic0 via `keccak256` internally.

## Indexing and Reorg Behavior

- Scanner worker periodically fetches logs in batches and upserts:
  - `event_db.event_log`
  - `event_db.block_sync`
- Subscription worker listens for removed logs (`log.Removed == true`) and pushes reorg tasks.
- Reorg consumer rolls back from a discovered checkpoint and deletes logs after that checkpoint.
- If no checkpoint is found within the fallback window, it falls back to `start_block`.

## API Overview

Base path: `/api`

Common response envelope:

```json
{
  "code": 0,
  "message": "success",
  "result": {}
}
```

### Endpoints

Health:

- `GET /api/status`

User auth:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh` (requires `refresh_token` cookie + `X-CSRF-Token`)
- `POST /api/v1/auth/logout` (requires `Authorization`)

User:

- `GET /api/v1/user/me` (requires `Authorization`)
- `PUT /api/v1/user/me` (requires `Authorization`)

Admin auth:

- `POST /api/v1/admin/auth/login`
- `POST /api/v1/admin/auth/refresh` (requires `admin_refresh_token` cookie + `X-CSRF-Token`)
- `POST /api/v1/admin/auth/logout` (requires `Authorization`)

Admin users:

- `POST /api/v1/admin/users`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/:user_id`
- `PUT /api/v1/admin/users/:user_id`
- `DELETE /api/v1/admin/users/:user_id`

Transactions:

- `GET /api/v1/txn/logs` (`start_time`, `end_time`, `page`, `size` are required query params)

### Auth Notes

- `Authorization` header accepts `Bearer <token>` (and also tolerates raw token).
- Cookie names:
  - user refresh token: `refresh_token`
  - admin refresh token: `admin_refresh_token`
- CSRF header: `X-CSRF-Token`
- Refresh cookies are set `HttpOnly` + `Secure`.

## Swagger

Swagger UI is served only when enabled:

- Config key: `api.enable_swagger`
- Env override: `API_ENABLE_SWAGGER`

Current defaults:

- `config/config.yaml`: `api.enable_swagger: false`
- `docker/docker-compose.yml`: `API_ENABLE_SWAGGER=true` for `indexer`

Regenerate docs after annotation changes:

```bash
make swagger
```

(`make swagger` runs `swag init -g cmd/indexer/main.go -o docs --parseInternal -d .`)

## Database

Schema files:

- `docker/db/schema/event_db.sql`
  - `event_log`
  - `block_sync`
- `docker/db/schema/account_db.sql`
  - `user`
  - `admin`

DB init order in MySQL container is handled by `docker/db/00-run.sh` (runs all `schema/*` then `init/*`).

Seed data:

- `docker/db/init/local.sql` inserts an initial admin row into `account_db.admin`.

## Development

Run tests:

```bash
go test ./...
```

Test notes:

- API integration tests use `config/config.yaml` and expect MySQL/Redis reachable per config.
- Tests create their own temporary user records; do not rely on a fixed `root/root` test account.

Generate contract bindings:

```bash
make gen
```

Deploy sample contract / send transfer on Anvil (Docker network `indexer-network`):

```bash
make deploy
make transfer
```
