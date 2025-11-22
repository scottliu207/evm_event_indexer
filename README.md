# EVM Event Indexer

A Go application for indexing and storing EVM blockchain event logs.

## Features

- **Auto Sync**: Periodically fetches smart contract event logs from EVM chains
- **Persistent Storage**: Stores event logs in MySQL database
- **Sync Status Tracking**: Records synchronization progress for each contract address
- **Resume Support**: Continues indexing from the last sync position
- **Error Retry**: Built-in retry mechanism for improved stability
- **Comprehensive Testing**: Includes unit tests and integration tests

## Architecture

### Tech Stack

- **Language**: Go 1.24.2
- **Blockchain Interaction**: go-ethereum (geth) v1.16.7
- **Database**: MySQL 8.0
- **Environment Config**: godotenv
- **Testing Framework**: testify

### Project Structure

```
.
├── background/          # Background tasks (event log fetching)
├── contracts/           # Smart contract related files
├── docker/              # Docker configuration
│   ├── db/
│   │   └── schema/      # Database schema
│   └── docker-compose.yml
├── env/                 # Environment configuration
├── internal/            # Internal packages
│   ├── erc20/          # ERC20 contract interaction
│   └── eth/            # Ethereum client wrapper
├── service/
│   ├── db/             # Database connection management
│   ├── model/          # Data models
│   └── repo/           # Data access layer
│       ├── blocksync/  # Block sync status
│       └── eventlog/   # Event logs
├── utils/              # Utility functions
├── main.go             # Application entry point
└── run.sh              # Docker management script
```

## Getting Started

### Prerequisites

- Go 1.24.2 or higher
- Docker and Docker Compose
- Accessible EVM node (e.g., Hardhat, Ganache, or public chain node)

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/scottliu207/evm_event_indexer.git
cd evm_event_indexer
```

2. **Install dependencies**

```bash
go mod download
```

3. **Configure environment variables**

Create a `.env` file:

```bash
cp .env.example .env
```

Edit `.env` to configure necessary environment variables (e.g., database connection info).

4. **Start MySQL database**

```bash
./run.sh up
```

This will start the MySQL container and automatically execute database initialization scripts.

5. **Run the application**

```bash
go run main.go
```

## Usage

### Docker Management

```bash
# Start services
./run.sh up

# Stop services
./run.sh down

# View help
./run.sh --help
```

### Configure Contract Address

Modify the contract address to monitor in `main.go`:

```go
const CONTRACT_ADDRESS = "0x5FbDB2315678afecb367f032d93F642f64180aa3"
```

### View Logs

```bash
# View application logs
go run main.go

# View MySQL container logs
docker logs evm-event-mysql
```

## Database Schema

### event_log Table

Stores event log data:

| Column | Type | Description |
|--------|------|-------------|
| id | bigint unsigned | Primary key, auto-increment |
| address | varchar(128) | Contract address |
| block_hash | varchar(128) | Block hash |
| block_number | bigint unsigned | Block number |
| tx_hash | varchar(128) | Transaction hash |
| tx_index | bigint unsigned | Transaction index |
| log_index | bigint unsigned | Log index |
| data | blob | Event data |
| block_timestamp | timestamp | Block timestamp |
| topics | json | Event topics |

**Unique Index**: `(block_number, tx_index, log_index)`

### block_sync Table

Tracks synchronization status:

| Column | Type | Description |
|--------|------|-------------|
| address | varchar(128) | Contract address (primary key) |
| last_sync_number | bigint unsigned | Last synced block number |
| last_sync_timestamp | timestamp | Last sync time |
| last_finalized_number | bigint unsigned | Last finalized block number |
| last_finalized_timestamp | timestamp | Last finalized time |
| updated_at | timestamp | Update time |

## Development Guide

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./service/repo/eventlog
go test ./service/repo/blocksync

# Run tests with coverage
go test -cover ./...
```

### Code Structure

#### Background Worker

`background/get_logs.go` implements event log fetching logic:

- Polls for new events every 5 seconds
- Fetches events for 100 blocks per iteration
- Uses transactions to ensure data consistency
- Built-in retry mechanism (up to 10 retries)

#### Repository Layer

- `eventlog`: Handles CRUD operations for event logs
- `blocksync`: Manages block synchronization status

All database operations are wrapped in transactions to ensure data consistency.

### Adding New Features

1. **Add new event handlers**: Create new workers in the `background/` directory
2. **Add new data models**: Define in `service/model/`
3. **Add new data access layer**: Implement in `service/repo/`
4. **Add tests**: Write corresponding test files for new features

## Configuration

### Environment Variables

Configure the following variables in the `.env` file:

```env
# MySQL configuration
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=root
MYSQL_DATABASE=event_db

# EVM node configuration
ETH_RPC_URL=http://localhost:8545
```

### Adjust Sync Parameters

You can adjust parameters in `background/get_logs.go`:

```go
size := uint64(100)                        // Number of blocks to fetch per iteration
ticker := time.NewTicker(time.Second * 5)  // Polling interval
const RETRY = 10                           // Maximum retry attempts
```

## Troubleshooting

### MySQL Connection Failed

```bash
# Check MySQL container status
docker ps -a | grep mysql

# View MySQL logs
docker logs evm-event-mysql

# Restart MySQL
./run.sh down
./run.sh up
```

### Event Fetching Failed

1. Verify EVM node is accessible
2. Check if contract address is correct
3. Review error messages in application logs

### Database Initialization Failed

```bash
# Clean up old data and reinitialize
docker compose -p evm-event-indexer -f docker/docker-compose.yml down -v
./run.sh up
```
