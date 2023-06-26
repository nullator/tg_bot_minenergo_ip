package main

import (
	"io"
	"log"
	"os"
	"tg_bot_minenergo_ip/pkg/config"
	boltdb "tg_bot_minenergo_ip/pkg/databases/boltDB"
	"tg_bot_minenergo_ip/pkg/logger"
	"tg_bot_minenergo_ip/pkg/telegram"

	"github.com/boltdb/bolt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	err := os.MkdirAll("log", os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	f, err := os.OpenFile("log/all.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	l := log.Default()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
	l.SetOutput(wrt)
	logger := logger.New("minenergo", l)

	cfg, err := config.Init(logger)
	if err != nil {
		logger.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		logger.Fatal(err)
	}

	bot.Debug = false

	db, err := bolt.Open(cfg.DB_file, 0600, nil)
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Fatal(err)
		}
	}()
	base := boltdb.NewDatabase(db)

	tg_bot := telegram.NewBot(bot, base, cfg, logger)
	go tg_bot.LoadIP()
	tg_bot.Start()

}
