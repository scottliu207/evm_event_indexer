package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql 實際連線實例化套件
)

var (
	pool  = make(map[string]*sql.DB)
	mutex sync.Mutex
)

const EVENT_DB = "event_db"
const MAX_OPEN_CONNS = 10
const MAX_IDLE_CONNS = 10
const CONN_MAX_LIFE_TIME = time.Minute * 5
const RETRY = 10
const WAIT_DURATION = time.Second * 10
const TIMEOUT = time.Second * 30

type mysql struct {
	DB       string
	Account  string
	Password string
	IP       string
	Port     string
}

// GetMysql : 取得 Mysql 物件
func GetMysql(dbName string) *sql.DB {
	mutex.Lock()
	defer mutex.Unlock()

	var db *sql.DB
	var ok bool
	var err error
	if db, ok = pool[dbName]; !ok {
		db, err = initMysql(dbName)
		if err != nil {
			slog.Error(
				"Mysql Ping 失敗",
				slog.Any("db name", dbName),
				slog.Any("err", err.Error()),
			)
			panic(err)
		}

		pool[dbName] = db
	}

	return db
}

func initMysql(dbName string) (*sql.DB, error) {

	// TODO: config
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s&charset=utf8mb4,utf8&timeout=%s",
			"root",          //
			"root",          //
			"127.0.0.1",     //
			"3306",          //
			EVENT_DB,        //
			"Asia%2FTaipei", //
			TIMEOUT.String(),
		),
	)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(MAX_OPEN_CONNS)
	db.SetMaxIdleConns(MAX_IDLE_CONNS)
	db.SetConnMaxLifetime(CONN_MAX_LIFE_TIME)

	for i := 0; i < RETRY; i++ {
		err = func() error {
			ctxTimeout, cancel := context.WithTimeout(
				context.Background(),
				TIMEOUT,
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
		slog.Warn(
			"failed to connect to mysql",
			slog.Any("DB Name", dbName),
			slog.Any("err", err),
			slog.Any("waiting for retry", WAIT_DURATION.String()),
			slog.Any("retry", i),
		)
		time.Sleep(WAIT_DURATION)
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}
