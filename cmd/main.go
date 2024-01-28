package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tg_bot_minenergo_ip/pkg/config"
	boltdb "tg_bot_minenergo_ip/pkg/databases/boltDB"
	"tg_bot_minenergo_ip/pkg/logger"
	"tg_bot_minenergo_ip/pkg/telegram"

	"github.com/boltdb/bolt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env: %v", err)
	}

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
	l.SetOutput(wrt)

	// setup slog
	log, err := setupLogger(os.Getenv("ENV"))
	if err != nil {
		l.Fatal(err)
	}
	slog.SetDefault(log)

	log.Info("start app", slog.String("env", os.Getenv("ENV")))
	log.Debug("debug level is enabled")

	cfg, err := config.Init()
	if err != nil {
		l.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		l.Fatal(err)
	}

	bot.Debug = false

	db, err := bolt.Open(cfg.DB_file, 0600, nil)
	if err != nil {
		l.Fatal(err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			l.Fatal(err)
		}
	}()
	base := boltdb.NewDatabase(db)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Info("stop app")
		cancel()
	}()

	tg_bot := telegram.NewBot(bot, base, cfg)
	go tg_bot.LoadIP(ctx)
	tg_bot.Start(ctx)
}

func setupLogger(env string) (*slog.Logger, error) {
	var log *slog.Logger

	switch env {
	// для локальной разработки и отладки используется дефолтный json handler
	case "local":
		h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: false,
		})
		log = slog.New(h)
	// для продакшена используется кастомный handler, который отправляет логи на сервер
	case "prod":
		h := logger.NewCustomSlogHandler(slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{
				Level:     slog.LevelInfo,
				AddSource: false,
			}))
		log = slog.New(h)
	default:
		return nil, fmt.Errorf("incorrect error level: %s", env)
	}

	return log, nil
}
