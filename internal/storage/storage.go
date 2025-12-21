package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"

	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/storage/mysql"
	"evm_event_indexer/internal/storage/redis"

	redisClient "github.com/redis/go-redis/v9"
)

var (
	dbManager *DBManager
	once      sync.Once
)

type DBManager struct {
	mysqlManager *mysql.MySQLManager
	redisManager *redis.RedisManager
}

func Forge() *DBManager {
	once.Do(func() {
		dbManager = newDBManager()
	})
	return dbManager
}

func newDBManager() *DBManager {
	return &DBManager{
		mysqlManager: mysql.NewMySQLManager(),
		redisManager: redis.NewRedisManager(),
	}
}

func (m *DBManager) Init() error {
	// init mysql
	for name, db := range config.Get().MySQL.DBs {
		if err := m.mysqlManager.InitMySQL(name, db); err != nil {
			return fmt.Errorf("failed to init mysql: %w", err)
		}
	}

	// init redis
	for name, db := range config.Get().Redis.DBs {
		if err := m.redisManager.InitRedis(name, db); err != nil {
			return fmt.Errorf("failed to init redis: %w", err)
		}
	}

	return nil
}

func (m *DBManager) Shutdown() {
	for name := range config.Get().MySQL.DBs {
		if err := m.mysqlManager.Shutdown(name); err != nil {
			slog.Error("failed to shutdown mysql", slog.Any("error", err))
		}
	}
	for name := range config.Get().Redis.DBs {
		if err := m.redisManager.Shutdown(name); err != nil {
			slog.Error("failed to shutdown redis", slog.Any("error", err))
		}
	}
}

func GetMySQL(name string) (*sql.DB, error) {
	return Forge().mysqlManager.GetMySQL(name)
}

func GetRedis(name string) (*redisClient.Client, error) {
	return Forge().redisManager.GetRedis(name)
}
