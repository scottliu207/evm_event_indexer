# Simple EVM Event Indexer

A Go-based EVM event (log) indexer: scan/subscribe contract events based on configuration, persist to MySQL, and expose a query API.

## Features

- **Indexing**: Scan/subscribe event logs by contract address + topics
- **Reorg**:  Handles chain reorganizations with a configurable reorg window and reprocesses affected logs
- **Storage**: MySQL for logs + sync state; Redis for storing token
- **API**: Gin HTTP API (logs query)
- **Auth**: Access Token (JWT) + Refresh Token (cookie) + CSRF protection

## Requirements

- Go `1.25.4`
- Docker + Docker Compose (recommended for local MySQL / Redis / Anvil)

## Project Layout

```
.
├── api/                 # HTTP server + routes + middleware
├── background/          # background workers (scanner/subscription/api server)
├── cmd/indexer/         # application entrypoint
├── config/              # config.yaml + scanner*.json
├── contracts/           # foundry contracts (example ERC20)
├── docker/              # docker-compose + MySQL init schema
├── internal/            # internal packages (config/session/storage/decoder/...)
├── service/             # service layer + repos
└── utils/               # helpers
```

## Quick Start (Docker)

1. Start MySQL / Redis / Anvil / indexer

```bash
./run.sh up
```

Note: `run.sh up` will also reset data.

2. Stop services

```bash
./run.sh down
```

## Run Locally (Non-Docker)
```bash
make run
```

## Configuration

### Main config (`config/config.yaml`)

- Main config file: `config/config.yaml`
- Scanner JSON path: `scanner_path` (e.g. `./config/scanner.json`)
- Viper reads environment variables to override YAML (e.g. `SESSION_JWT_SECRET`)

Common environment variables (see `docker/docker-compose.yml` for the full set):

- `API_PORT`
- `SCANNER_PATH`
- `SESSION_JWT_SECRET`
- `SESSION_CSRF_SECRET`
- `MYSQL_DATABASES_*`
- `REDIS_DATABASES_*`

### Scanner config (`config/scanner*.json`)

- `config/scanner.json`: scanner config for local runs
- `config/scanner.docker.json`: scanner config used by Docker compose (via `SCANNER_PATH`)

Scanner JSON key fields:

- `rpc_http` / `rpc_ws`
- `batch_size`
- `addresses[]`:
  - `address`: contract address
  - `topics[]`: event signatures (strings); the indexer hashes them with `keccak256` as topic0

## API

- `GET /api/status`: health check
- `POST /api/v1/auth/login`: login, returns `access_token` and sets cookies (`refresh_token` / `csrf_token`)
- `POST /api/v1/auth/refresh`: rotate access/refresh token (cookie-based; requires CSRF)
- `POST /api/v1/auth/logout`: logout, deletes refresh token (cookie-based; requires CSRF)
- `GET /api/v1/txn/logs`: query event logs (requires `Authorization: Bearer <access_token>`)

## Auth & CSRF

This project stores the refresh token in a cookie (`refresh_token`), so endpoints that rely on this cookie are protected with CSRF checks (double-submit cookie + header).

### Cookies

- `refresh_token`: `HttpOnly` + `Secure` (not readable by browser JS)
- `csrf_token`: `Secure`, **not** `HttpOnly` (readable by browser JS)

### Flow

1. `POST /api/v1/auth/login`
   - Response JSON: `access_token`, `csrf_token`
   - Response cookies: `refresh_token`, `csrf_token`
2. `POST /api/v1/auth/refresh` / `POST /api/v1/auth/logout`
   - Browser must send cookies: `fetch(..., { credentials: "include" })` (or axios `withCredentials: true`)
   - Must include header: `X-CSRF-Token: <csrf_token>`
     - `<csrf_token>` can be read from the `csrf_token` cookie or from the login/refresh JSON response

Notes:

- The CSRF middleware only requires `X-CSRF-Token` when the request includes the `refresh_token` cookie; missing refresh token behavior is handled by each endpoint (e.g. refresh returns invalid credentials; logout remains idempotent).

## Database Schema

MySQL schema is initialized by `docker/db/schema/*.sql`:

- `docker/db/schema/event_db.sql`:
  - `event_log`: event logs (`chain_id`, `topic_0..3`, `decoded_event`, `block_timestamp`)
  - `block_sync`: sync state (primary key: `(chain_id, address)`)
- `docker/db/schema/account_db.sql`:
  - `user`: login accounts (argon2 hash + `auth_meta`)

## Development

### Tests

```bash
go test ./...
```

Local test api credentials:
- account: `root`
- password: `root`

### Contract bindings

Requires `forge`, `jq`, and `abigen`:

```bash
make gen
```

### Deploy / Transfer (Anvil in Docker)

```bash
make deploy
make transfer
```
