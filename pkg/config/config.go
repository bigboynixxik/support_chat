package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string `env:"APP_ENV"`
	HTTPPort string `env:"HTTP_PORT"`
	PGDsn    string `env:"PG_DSN"`
}

func LoadConfig(path string) (*Config, error) {
	_ = godotenv.Load(path)
	//if err != nil {
	//	return nil, fmt.Errorf("config.LoadConfig: %w", err)
	//}
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("config.LoadConfig failed to parse config: %w", err)
	}
	return &cfg, nil
}
