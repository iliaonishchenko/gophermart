package config

import (
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

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
