package config

import (
	"fmt"
	"time"
)

type Config struct {
	ScannerPath string     `yaml:"scanner_path"`
	Scanners    []struct { // json file, should located in the same directory as the config file
		RpcHTTP   string `json:"rpc_http"`
		RpcWS     string `json:"rpc_ws"`
		BatchSize int32  `json:"batch_size"`
		Addresses []struct {
			Address string   `json:"address"`
			Topics  []string `json:"topics"`
		} `json:"addresses"`
	}
	Decoders []struct {
		Name      string `yaml:"name"`
		Signature string `yaml:"signature"`
	} `yaml:"decoders"`
	LogScannerInterval time.Duration `yaml:"log_scanner_interval"`
	ReorgWindow        int32         `yaml:"reorg_window"`
	LogLevel           string        `yaml:"log_level"`
	Timeout            time.Duration `yaml:"timeout"`
	Retry              int           `yaml:"retry"`
	Backoff            time.Duration `yaml:"backoff"`
	MaxBackoff         time.Duration `yaml:"max_backoff"`
	API                struct {
		Port    string        `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	}
	Argon2 struct {
		Time    uint32 `yaml:"time"`
		Memory  uint32 `yaml:"memory"`
		Threads uint8  `yaml:"threads"`
		KeyLen  uint32 `yaml:"key_len"`
	}
	MySQL struct {
		MaxOpenConns    int           `yaml:"max_open_conns"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time"`
		Retry           int           `yaml:"retry"`
		WaitDuration    time.Duration `yaml:"wait_duration"`
		Timeout         time.Duration `yaml:"timeout"`
		EventDBM        mysql         `yaml:"event_dbm"`
		EventDBS        mysql         `yaml:"event_dbs"`
	} `yaml:"mysql"`
}

type mysql struct {
	DBName   string `yaml:"db_name"`
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
}

var config = new(Config)

func (c Config) Validate() error {
	if len(c.Scanners) == 0 {
		return fmt.Errorf("scanner is required")
	}

	for _, scanner := range c.Scanners {
		if scanner.RpcHTTP == "" {
			return fmt.Errorf("scanner.rpc_http is required")
		}
		if scanner.RpcWS == "" {
			return fmt.Errorf("scanner.rpc_ws is required")
		}
		if len(scanner.Addresses) == 0 {
			return fmt.Errorf("scanner.address is required")
		}

		for _, address := range scanner.Addresses {
			if address.Address == "" {
				return fmt.Errorf("scanner.address is required")
			}
		}

		if scanner.BatchSize == 0 {
			return fmt.Errorf("scanner.batch_size is required")
		}
	}

	if c.LogScannerInterval == 0 {
		return fmt.Errorf("log_scanner_interval is required")
	}

	if c.ReorgWindow == 0 {
		return fmt.Errorf("reorg_window is required")
	}

	if c.LogLevel == "" {
		return fmt.Errorf("log_level is required")
	}

	if c.Timeout == 0 {
		return fmt.Errorf("timeout is required")
	}

	if c.Retry == 0 {
		return fmt.Errorf("retry is required")
	}

	if c.Backoff == 0 {
		return fmt.Errorf("backoff is required")
	}

	if c.MaxBackoff == 0 {
		return fmt.Errorf("max_backoff is required")
	}

	if c.API.Port == "" {
		return fmt.Errorf("api.port is required")
	}

	if c.API.Timeout == 0 {
		return fmt.Errorf("api.timeout is required")
	}

	if c.MySQL.MaxOpenConns == 0 {
		return fmt.Errorf("mysql.max_open_conns is required")
	}

	if c.MySQL.MaxIdleConns == 0 {
		return fmt.Errorf("mysql.max_idle_conns is required")
	}

	if c.MySQL.ConnMaxLifeTime == 0 {
		return fmt.Errorf("mysql.conn_max_life_time is required")
	}

	if c.MySQL.Retry == 0 {
		return fmt.Errorf("mysql.retry is required")
	}

	if c.MySQL.WaitDuration == 0 {
		return fmt.Errorf("mysql.wait_duration is required")
	}

	if c.MySQL.Timeout == 0 {
		return fmt.Errorf("mysql.timeout is required")
	}

	if c.MySQL.EventDBM.DBName == "" {
		return fmt.Errorf("mysql.event_dbm.db_name is required")
	}

	if c.MySQL.EventDBM.Account == "" {
		return fmt.Errorf("mysql.event_dbm.account is required")
	}

	if c.MySQL.EventDBM.Password == "" {
		return fmt.Errorf("mysql.event_dbm.password is required")
	}

	if c.MySQL.EventDBM.IP == "" {
		return fmt.Errorf("mysql.event_dbm.ip is required")
	}

	if c.MySQL.EventDBM.Port == "" {
		return fmt.Errorf("mysql.event_dbm.port is required")
	}

	if c.MySQL.EventDBS.DBName == "" {
		return fmt.Errorf("mysql.event_dbs.db_name is required")
	}

	if c.MySQL.EventDBS.Account == "" {
		return fmt.Errorf("mysql.event_dbs.account is required")
	}

	if c.MySQL.EventDBS.Password == "" {
		return fmt.Errorf("mysql.event_dbs.password is required")
	}

	if c.MySQL.EventDBS.IP == "" {
		return fmt.Errorf("mysql.event_dbs.ip is required")
	}

	if c.MySQL.EventDBS.Port == "" {
		return fmt.Errorf("mysql.event_dbs.port is required")
	}

	return nil
}
