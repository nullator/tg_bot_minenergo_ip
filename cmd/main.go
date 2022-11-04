package main

import (
	"io"
	"log"
	"os"
	"tg_bot_minenergo_ip/pkg/config"
	boltdb "tg_bot_minenergo_ip/pkg/databases/boltDB"
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

	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	cfg, err := config.Init()
	if err != nil {
		log.Fatalln(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalln(err)
	}

	bot.Debug = false

	db, err := bolt.Open(cfg.DB_file, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	base := boltdb.NewDatabase(db)

	tg_bot := telegram.NewBot(bot, base, cfg)
	go tg_bot.LoadIP()
	tg_bot.Start()

}
