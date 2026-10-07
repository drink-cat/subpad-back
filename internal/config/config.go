package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Domain DomainConfig `mapstructure:"domain"`
	Eth    EthConfig    `mapstructure:"eth"`
	Cron   CronConfig   `mapstructure:"cron"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expireHours"`
}

type DomainConfig struct {
	Suffix string `mapstructure:"suffix"`
}

type EthConfig struct {
	RPC     string `mapstructure:"rpc"`
	ChainID int64  `mapstructure:"chainId"`
}

type CronConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func (s ServerConfig) Addr() string {
	host := s.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := s.Port
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("SUBPAD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.JWT.ExpireHours <= 0 {
		cfg.JWT.ExpireHours = 72
	}
	if cfg.Domain.Suffix == "" {
		cfg.Domain.Suffix = "launch.o1.local"
	}
	return &cfg, nil
}
