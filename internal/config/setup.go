package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

func Get() *Config {
	return config
}

func LoadConfig(path string) {
	if path == "" {
		path = "./config/config.yaml"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("no default configuration file loaded, because %v", err.Error()))
	}

	// parse config file
	if err := viper.Unmarshal(config,
		viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			func(f reflect.Kind, t reflect.Kind, data any) (any, error) {
				if f != reflect.String || t != reflect.Slice {
					return data, nil
				}

				raw := data.(string)
				if raw == "" {
					return []string{}, nil
				}

				return strings.Fields(raw), nil
			},
			func(f reflect.Kind, t reflect.Kind, data any) (any, error) {
				if t != reflect.TypeOf(decimal.Zero).Kind() {
					return data, nil
				}
				switch t := data.(type) {
				case string:
					return decimal.NewFromString(t)
				case int:
					return decimal.NewFromInt(int64(t)), nil
				case float64:
					return decimal.NewFromFloat(t), nil
				default:
					return data, nil
				}
			},
		)),
		// read config file with yaml tag
		func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "yaml"
		},
	); err != nil {
		panic(fmt.Errorf("unable to decode config into struct, because %v", err.Error()))
	}

	// load scanner from json file
	scanner, err := os.ReadFile(config.ScannerPath)
	if err != nil {
		panic(fmt.Errorf("unable to read scanner file, because %v", err.Error()))
	}

	// store into config
	if err := json.Unmarshal(scanner, &config.Scanners); err != nil {
		panic(fmt.Errorf("unable to unmarshal scanner file, because %v", err.Error()))
	}

}
