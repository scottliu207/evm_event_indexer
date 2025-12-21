package mysql

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLManager struct {
	pool map[string]*sql.DB
	mu   sync.Mutex
}

func NewMySQLManager() *MySQLManager {
	return &MySQLManager{
		pool: make(map[string]*sql.DB),
	}
}

func (m *MySQLManager) GetMySQL(name string) (*sql.DB, error) {
	db, ok := m.pool[name]
	if !ok {
		return nil, fmt.Errorf("mysql %s not initialized", name)
	}
	return db, nil
}

func (m *MySQLManager) InitMySQL(name string, dbCnf config.MySQL) error {

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s&charset=utf8mb4,utf8&timeout=%s",
			dbCnf.Account,
			dbCnf.Password,
			dbCnf.IP,
			dbCnf.Port,
			dbCnf.DBName,
			"Asia%2FTaipei",
			config.Get().MySQL.Timeout.String(),
		),
	)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(config.Get().MySQL.MaxOpenConns)
	db.SetMaxIdleConns(config.Get().MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(config.Get().MySQL.ConnMaxLifeTime)

	for i := 0; i < config.Get().MySQL.Retry; i++ {
		err = func() error {
			ctxTimeout, cancel := context.WithTimeout(
				context.Background(),
				config.Get().MySQL.Timeout,
			)
			defer cancel()

			if err := db.PingContext(ctxTimeout); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			slog.Error(
				"mysql initialize failure, failed to connect to mysql",
				slog.Any("DB Name", dbCnf.DBName),
				slog.Any("err", err),
				slog.Any("waiting for retry", config.Get().MySQL.WaitDuration.String()),
				slog.Any("retry", i),
			)
			time.Sleep(config.Get().MySQL.WaitDuration)
			continue
		}

		m.mu.Lock()
		defer m.mu.Unlock()
		m.pool[name] = db
		return nil
	}

	db.Close()
	return fmt.Errorf("mysql initialize failure, retry limit exceed")
}

func (m *MySQLManager) Shutdown(name string) error {
	db, err := m.GetMySQL(name)
	if err != nil {
		return fmt.Errorf("failed to get mysql: %w", err)
	}
	db.Close()
	delete(m.pool, name)
	return nil
}
