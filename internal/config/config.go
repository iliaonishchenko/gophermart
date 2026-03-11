package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	JwtSecret   string `env:"JWT_SECRET"`
	JwtExpire   int    `env:"JWT_EXPIRE" envDefault:"31556952"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	DatabaseURI string `env:"DATABASE_URI"`
	ServerAddr  string `env:"RUN_ADDRESS"`
	AccrualAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func loadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return cfg, nil
}

func parseFlags(cfg *Config) {
	var serverAddr string   // -a
	var databaseAddr string // -d
	var accrualAddr string  // -r

	flag.StringVar(&serverAddr, "a", "localhost:8080", "server address")
	flag.StringVar(&databaseAddr, "d", "localhost:27017", "database address")
	flag.StringVar(&accrualAddr, "r", "", "accrual address")

	flag.Parse()

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = serverAddr
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = databaseAddr
	}
	if cfg.AccrualAddr == "" {
		cfg.AccrualAddr = accrualAddr
	}
}
func LoadConfiguration() (*Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	parseFlags(cfg)

	return cfg, nil
}
