package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	JwtSecret   string `env:"JWT_SECRET"`
	JwtExpire   int    `env:"JWT_EXPIRE" envDefault:"31556952"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
