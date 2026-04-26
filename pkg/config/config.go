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
	IP                           map[string]IP
	TelegramServerBaseURL        string
	TelegramServerAuthToken      string
	TelegramServerBotID          string
	TelegramServerTimeoutSeconds int
	HTTPAddr                     string
	TelegramEventsPath           string
	IP_file                      string
	DB_file                      string
	LogServer                    string
	LogAuthToken                 string
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

	cfg.TelegramServerBaseURL = viper.GetString("TELEGRAM_SERVER_BASE_URL")
	cfg.TelegramServerAuthToken = viper.GetString("TELEGRAM_SERVER_AUTH_TOKEN")
	cfg.TelegramServerBotID = viper.GetString("TELEGRAM_SERVER_BOT_ID")
	cfg.TelegramServerTimeoutSeconds = viper.GetInt("TELEGRAM_SERVER_TIMEOUT_SECONDS")
	cfg.HTTPAddr = viper.GetString("HTTP_ADDR")
	cfg.TelegramEventsPath = viper.GetString("TELEGRAM_EVENTS_PATH")
	cfg.IP_file = viper.GetString("IP_file")
	cfg.DB_file = viper.GetString("DB_file")
	cfg.LogServer = viper.GetString("LOGGER_SERVER")
	cfg.LogAuthToken = viper.GetString("LOGGER_AUTH")
	if cfg.TelegramServerTimeoutSeconds == 0 {
		cfg.TelegramServerTimeoutSeconds = 5
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":9001"
	}
	if cfg.TelegramEventsPath == "" {
		cfg.TelegramEventsPath = "/telegram/events"
	}
	return nil
}
