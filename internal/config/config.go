package config

import (
	"fmt"
	"time"
)

type Config struct {
	ScannerPath  string        `yaml:"scanner_path"`
	StartBlock   uint64        `yaml:"start_block"`
	WaitForStart time.Duration `yaml:"wait_for_start"`
	Scanners     []struct {    // json file, should located in the same directory as the config file
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
	Session struct {
		JWTSecret    string        `yaml:"jwt_secret"`
		CSRFSecret   string        `yaml:"csrf_secret"`
		ATExpiration time.Duration `yaml:"at_expiration"`
		RTExpiration time.Duration `yaml:"rt_expiration"`
	} `yaml:"session"`

	MySQL struct {
		MaxOpenConns    int              `yaml:"max_open_conns"`
		MaxIdleConns    int              `yaml:"max_idle_conns"`
		ConnMaxLifeTime time.Duration    `yaml:"conn_max_life_time"`
		Retry           int              `yaml:"retry"`
		WaitDuration    time.Duration    `yaml:"wait_duration"`
		Timeout         time.Duration    `yaml:"timeout"`
		DBs             map[string]MySQL `yaml:"databases"`
	} `yaml:"mysql"`

	Redis struct {
		Retry        int              `yaml:"retry"`
		WaitDuration time.Duration    `yaml:"wait_duration"`
		PingTimeout  time.Duration    `yaml:"ping_timeout"`
		DBs          map[string]Redis `yaml:"databases"`
	}
}

type MySQL struct {
	Name     string `yaml:"name"`
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
}

type Redis struct {
	ReadTimeout        time.Duration `yaml:"read_timeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout"`
	MaxRetries         int           `yaml:"max_retries"`
	DialTimeout        time.Duration `yaml:"dial_timeout"`
	PoolSize           int           `yaml:"pool_size"`
	PoolTimeout        time.Duration `yaml:"pool_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout"`
	IdleCheckFrequency time.Duration `yaml:"idle_check_frequency"`
	IP                 string        `yaml:"ip"`
	Port               int           `yaml:"port"`
	DB                 int           `yaml:"db"`
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

	for _, db := range c.MySQL.DBs {
		if db.Name == "" {
			return fmt.Errorf("mysql.db_name is required")
		}
		if db.Account == "" {
			return fmt.Errorf("mysql.account is required")
		}
		if db.Password == "" {
			return fmt.Errorf("mysql.password is required")
		}
		if db.IP == "" {
			return fmt.Errorf("mysql.ip is required")
		}
		if db.Port == "" {
			return fmt.Errorf("mysql.port is required")
		}
	}

	for _, db := range c.Redis.DBs {
		if db.IP == "" {
			return fmt.Errorf("redis.ip is required")
		}
		if db.Port == 0 {
			return fmt.Errorf("redis.port is required")
		}
		if db.ReadTimeout == 0 {
			return fmt.Errorf("redis.read_timeout is required")
		}
		if db.WriteTimeout == 0 {
			return fmt.Errorf("redis.write_timeout is required")
		}
		if db.MaxRetries == 0 {
			return fmt.Errorf("redis.max_retries is required")
		}
		if db.DialTimeout == 0 {
			return fmt.Errorf("redis.dial_timeout is required")
		}
		if db.PoolSize == 0 {
			return fmt.Errorf("redis.pool_size is required")
		}
		if db.PoolTimeout == 0 {
			return fmt.Errorf("redis.pool_timeout is required")
		}
		if db.IdleTimeout == 0 {
			return fmt.Errorf("redis.idle_timeout is required")
		}
		if db.IdleCheckFrequency == 0 {
			return fmt.Errorf("redis.idle_check_frequency is required")
		}
	}

	return nil
}

func (c *Config) GetRedis(name string) (*Redis, error) {
	if db, ok := c.Redis.DBs[name]; ok {
		return &db, nil
	}

	return nil, fmt.Errorf("redis %s not found", name)
}

func (c *Config) GetMySQL(name string) (*MySQL, error) {
	if db, ok := c.MySQL.DBs[name]; ok {
		return &db, nil
	}

	return nil, fmt.Errorf("mysql %s not found", name)
}
