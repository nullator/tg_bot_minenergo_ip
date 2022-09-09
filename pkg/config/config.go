package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	IP            map[string]IP
	TelegramToken string
	IP_file       string
}

type IP struct {
	Name        string `json:"name"`
	ID          int    `json:"id"`
	First_entry string `json:"first_entry"`
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
		log.Printf("Ошибка открытия json файла с ИП %s\n", err)
		return nil, err
	}

	ip_data, err = ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Ошибка чтения json файла с ИП %s\n", err)
		return nil, err
	}
	f.Close()

	err = json.Unmarshal([]byte(ip_data), &ip_list)
	if err != nil {
		log.Printf("Ошибка распаковки json в структуру гороскопа %s\n", err)
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
	return nil
}
