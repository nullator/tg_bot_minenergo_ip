package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"tg_bot_minenergo_ip/pkg/config"
	"tg_bot_minenergo_ip/pkg/databases"
)

// Bot — обработчик бизнес-логики Telegram-событий.
type Bot struct {
	client *Client
	base   databases.Database
	config *config.Config
}

// NewBot — создаёт обработчик бизнес-логики Telegram-событий.
func NewBot(client *Client, base databases.Database,
	config *config.Config) *Bot {
	return &Bot{client, base, config}
}

// HandleEvent — обрабатывает событие, полученное от telegram-server.
func (b *Bot) HandleEvent(ctx context.Context, event Event) error {
	if event.Type == "command" && event.Message != nil && event.Chat != nil {
		slog.Info("Пользователь ввёл команду",
			slog.String("user", event.Chat.Username),
			slog.String("command", event.Message.Command))
		return b.handleCommand(ctx, event.Message, event.Chat)
	}

	if event.Type == "message" && event.Message != nil && event.Chat != nil {
		slog.Info("Пользователь отправил сообщение:",
			slog.String("user", event.Chat.Username),
			slog.String("message", event.Message.Text))
		return nil
	}

	if event.Type != "callback" || event.Callback == nil || event.Chat == nil {
		return nil
	}

	q := event.Callback.Data
	switch q {
	case "start":
		return b.client.Send(ctx, SendCommand{
			ChatID:      event.Chat.ID,
			Text:        "Ты можешь подписаться или отписаться от рассылки на уведомления о размещении материалов проектов ИП:",
			ReplyMarkup: makeStartKeyboard(),
		})

	case "subscribe":
		return b.client.Send(ctx, SendCommand{
			ChatID:      event.Chat.ID,
			Text:        "Выбери проекты ИП для подписки:",
			ReplyMarkup: make_subscribe_kb(b, event.Chat.ID),
		})

	default:
		runes := []rune(q)
		if len(runes) < 5 {
			slog.Error("Получен некорректный callback",
				slog.String("callback", q))
			return nil
		}

		first_letter := string(runes[0])
		code := string(runes[1:5])
		if first_letter == "s" {
			status, err := b.base.Get(fmt.Sprintf("%d", event.Chat.ID), code)
			if err != nil {
				slog.Error("error getting status from db", slog.String("error", err.Error()))
			}
			if status == "subscride" {
				slog.Info("Пользователь запросил отписку",
					slog.String("user", event.Chat.Username),
					slog.String("ip", b.config.IP[code].Name))
				if err := b.unsubscribe(event.Chat.ID, code); err != nil {
					return err
				}
			} else {
				slog.Info("Пользователь запросил подписку",
					slog.String("user", event.Chat.Username),
					slog.String("ip", b.config.IP[code].Name))
				if err := b.subscribe(ctx, event.Chat.ID, code); err != nil {
					return err
				}
			}
			return b.client.Send(ctx, SendCommand{
				ChatID:      event.Chat.ID,
				Text:        "Настройки подписки обновлены:",
				ReplyMarkup: make_subscribe_kb(b, event.Chat.ID),
			})
		}
		if first_letter == "u" {
			slog.Info("Пользователь запросил отписку",
				slog.String("user", event.Chat.Username),
				slog.String("ip", b.config.IP[code].Name))
			if err := b.unsubscribe(event.Chat.ID, code); err != nil {
				return err
			}
			return b.client.Send(ctx, SendCommand{
				ChatID:      event.Chat.ID,
				Text:        "Настройки подписки обновлены:",
				ReplyMarkup: make_unsubscribe_kb(b, event.Chat.ID),
			})
		}
	}

	return nil
}
