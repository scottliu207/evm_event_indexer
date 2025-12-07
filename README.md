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



2025/11/22 18:35:45 INFO new block "block number"=9 
new="&{ParentHash:0xdac2057923e45dba03a3bd57bddcbea2b292607bdc45550e4cb0aa5672ae3a32 
UncleHash:0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347 
Coinbase:0x0000000000000000000000000000000000000000 
Root:0x3b7ff8e2417b2affb72d93ec18e0c8451cbad6891b97551bc95447c387779e9c 

TxHash:0x583ee2160576f168724ba3143b6c8c3bc3f65ede6020add722035a5fca0ae84d
 ReceiptHash:0x6cc1b348af46d7bd3ade4e8d491c0469058616123a19db1321db0318ab04dfcf 
 Bloom:[0 0 0 0 0 0 0 0 4 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 16 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 8 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 20 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 2 0 0 0 32 0 0 0 0 0 0 16 0 0 0 0 0 32 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 128 0 0 0 0 0 0 0 0] 
 Difficulty:+0 
 Number:+9 
 GasLimit:30000000 
 GasUsed:35254 
 Time:1763807745 
 Extra:[] 
 MixDigest:0x0ea042854d60d98cfc2a9d5388704e43ea0df996a791e1a17ec8595fca7d7b45 
 Nonce:[0 0 0 0 0 0 0 0] 
 BaseFee:+349990303 
 WithdrawalsHash:0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421 
 BlobGasUsed:0xc00028ed28 
 ExcessBlobGas:0xc00028ed30 
 ParentBeaconRoot:0x0000000000000000000000000000000000000000000000000000000000000000
 RequestsHash:0xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855}"



 2025/11/22 18:57:37 INFO event log info log="{
 Address:0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512 
 Topics:[0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef 0x000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266 
 0x000000000000000000000000e7f1725e7734ce288f8367e1bb143e90bb3f0512] 
 Data:[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1] 
 BlockNumber:9
TxHash:0x8764c445266ca549f972e69fd42104e53a37dc6108e65cc8a258c298233bf25b 
TxIndex:0 BlockHash:0x4e9a0d6d97d4d0579c2c30b822b869b69132c0efa5c6641751c781b09170e325 
BlockTimestamp:1763807745 
Index:0 
 Removed:false}"