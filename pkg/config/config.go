package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken string
}

func Init() (*Config, error) {
	var cfg Config

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseEnv(cfg *Config) error {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	cfg.TelegramToken = viper.GetString("TOKEN")
	return nil
}
