package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ContractsAddress   []string      `yaml:"contracts_address"`
	EthRpcHTTP         string        `yaml:"eth_rpc_http"`
	EthRpcWS           string        `yaml:"eth_rpc_ws"`
	LogScannerInterval time.Duration `yaml:"log_scanner_interval"`
	EventDB            string        `yaml:"event_db"`
	ReorgWindow        int           `yaml:"reorg_window"`
	LogLevel           string        `yaml:"log_level"`
	Timeout            time.Duration `yaml:"timeout"`
	Retry              int           `yaml:"retry"`
	Backoff            time.Duration `yaml:"backoff"`
	MaxBackoff         time.Duration `yaml:"max_backoff"`
	API                struct {
		Port    string        `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	}
	MySQL struct {
		MaxOpenConns    int           `yaml:"max_open_conns"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time"`
		Retry           int           `yaml:"retry"`
		WaitDuration    time.Duration `yaml:"wait_duration"`
		Timeout         time.Duration `yaml:"timeout"`
		EventDBM        struct {
			DBName   string `yaml:"db_name"`
			Account  string `yaml:"account"`
			Password string `yaml:"password"`
			IP       string `yaml:"ip"`
			Port     string `yaml:"port"`
		} `yaml:"event_dbm"`
		EventDBS struct {
			DBName   string `yaml:"db_name"`
			Account  string `yaml:"account"`
			Password string `yaml:"password"`
			IP       string `yaml:"ip"`
			Port     string `yaml:"port"`
		} `yaml:"event_dbs"`
	} `yaml:"mysql"`
}

var config = new(Config)

func LoadConfig(path string) {
	if path == "" {
		path = "./config/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(err)
	}
}

func Get() *Config {
	return config
}
