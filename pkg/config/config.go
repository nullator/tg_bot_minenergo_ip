package config

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	IP            map[string]IP
	TelegramToken string
	IP_file       string
	DB_file       string
	LogServer     string
	LogAuthToken  string
}

type IP struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Code string `json:"code"`
}

func Init() (*Config, error) {
	var cfg Config
	ip_list := make(map[string]IP)
	var ip_data []byte

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}

	f, err := os.Open(cfg.IP_file)
	if err != nil {
		slog.Error("Ошибка открытия json файла с ИП - %s", slog.String("error", err.Error()))
		return nil, err
	}

	ip_data, err = io.ReadAll(f)
	if err != nil {
		slog.Error("Ошибка чтения json файла с ИП - %s", slog.String("error", err.Error()))
		return nil, err
	}
	f.Close()

	err = json.Unmarshal([]byte(ip_data), &ip_list)
	if err != nil {
		slog.Error("Ошибка распаковки json в структуру ИП - %s",
			slog.String("error", err.Error()))
		return nil, err
	}

	cfg.IP = ip_list
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
	cfg.IP_file = viper.GetString("IP_file")
	cfg.DB_file = viper.GetString("DB_file")
	cfg.LogServer = viper.GetString("LOGGER_SERVER")
	cfg.LogAuthToken = viper.GetString("LOGGER_AUTH")
	return nil
}
