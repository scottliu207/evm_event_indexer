package mysql

import (
	"context"
	"database/sql"
	internalCnf "evm_event_indexer/internal/config"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	pool = make(map[string]*sql.DB)
	mu   sync.RWMutex
)

type MySQL struct {
	Account  string
	Password string
	IP       string
	Port     string
	DBName   string
}

func GetMysql(name string) *sql.DB {
	mu.RLock()
	db := pool[name]
	mu.RUnlock()
	if db == nil {
		panic(fmt.Sprintf("mysql %s not initialized", name))
	}
	return db
}

func InitMysql(data *MySQL) error {
	if data == nil {
		return fmt.Errorf("mysql config is required")
	}

	cnf := internalCnf.Get().MySQL

	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s&charset=utf8mb4,utf8&timeout=%s",
			data.Account,
			data.Password,
			data.IP,
			data.Port,
			data.DBName,
			"Asia%2FTaipei",
			cnf.Timeout.String(),
		),
	)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(cnf.MaxOpenConns)
	db.SetMaxIdleConns(cnf.MaxIdleConns)
	db.SetConnMaxLifetime(cnf.ConnMaxLifeTime)

	for i := 0; i < cnf.Retry; i++ {
		err = func() error {
			ctxTimeout, cancel := context.WithTimeout(
				context.Background(),
				cnf.Timeout,
			)
			defer cancel()

			if err := db.PingContext(ctxTimeout); err != nil {
				return err
			}
			return nil
		}()
		if err == nil {
			break
		}
		slog.Error(
			"failed to connect to mysql",
			slog.Any("DB Name", data.DBName),
			slog.Any("err", err),
			slog.Any("waiting for retry", cnf.WaitDuration.String()),
			slog.Any("retry", i),
		)
		time.Sleep(cnf.WaitDuration)
	}
	if err != nil {
		return err
	}

	mu.Lock()
	pool[data.DBName] = db
	mu.Unlock()

	return nil
}
