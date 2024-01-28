package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"tg_bot_minenergo_ip/pkg/config"
	"tg_bot_minenergo_ip/pkg/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot    *tgbotapi.BotAPI
	base   databases.Database
	config *config.Config
}

func NewBot(bot *tgbotapi.BotAPI, base databases.Database,
	config *config.Config) *Bot {
	return &Bot{bot, base, config}
}

func (b *Bot) Start(ctx context.Context) {
	slog.Info("Authorized on account %s",
		slog.String("account", b.bot.Self.UserName))
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	b.handleUpdates(ctx, updates)
}

func (b *Bot) handleUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Остановка telegram бота")
			return

		case update, ok := <-updates:
			if !ok {
				slog.Error("Telegram update chan closed")
				return
			}

			if update.Message != nil {
				if update.Message.IsCommand() {
					slog.Info("Пользователь ввёл команду",
						slog.String("user", update.Message.From.UserName),
						slog.String("command", update.Message.Command()))
					if err := b.handleCommand(update.Message); err != nil {
						slog.Error("При обработке команды произошла ошибка",
							slog.String("command", update.Message.Command()),
							slog.String("error", err.Error()))
					}
					continue
				}

				slog.Info("Пользователь отправил сообщение:",
					slog.String("user", update.Message.From.UserName),
					slog.String("message", update.Message.Text))
			} else if update.CallbackQuery != nil {
				q := update.CallbackQuery.Data
				switch q {
				case "start":
					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Настроить подписку", "subscribe"),
							// tgbotapi.NewInlineKeyboardButtonData("Отписаться", "unsubscribe"),
						),
					)
					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("error sending message", slog.String("error", err.Error()))
					}

				case "subscribe":
					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					numericKeyboard = make_subscribe_kb(b, update.CallbackQuery.Message.Chat.ID)

					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err := b.bot.Send(msg)
					if err != nil {
						slog.Error("error sending message", slog.String("error", err.Error()))
					}

				default:
					first_letter := string([]rune(q)[0])
					code := string([]rune(q)[1:5])
					if first_letter == "s" {
						status, err := b.base.Get(fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID), code)
						if err != nil {
							slog.Error("error getting status from db", slog.String("error", err.Error()))
						}
						if status == "subscride" {
							slog.Info("Пользователь запросил отписку",
								slog.String("user", update.CallbackQuery.Message.Chat.UserName),
								slog.String("ip", b.config.IP[code].Name))
							b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)

						} else {
							slog.Info("Пользователь запросил подписку",
								slog.String("user", update.CallbackQuery.Message.Chat.UserName),
								slog.String("ip", b.config.IP[code].Name))
							b.subscribe(update.CallbackQuery.Message.Chat.ID, code)
						}
						var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
						numericKeyboard = make_subscribe_kb(b, update.CallbackQuery.Message.Chat.ID)
						msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
						_, err = b.bot.Send(msg)
						if err != nil {
							slog.Error("error sending message", slog.String("error", err.Error()))
						}

					}
					if first_letter == "u" {
						slog.Info("Пользователь запросил отписку",
							slog.String("user", update.CallbackQuery.Message.Chat.UserName),
							slog.String("ip", b.config.IP[code].Name))
						b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)

						var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
						numericKeyboard = make_unsubscribe_kb(b, update.CallbackQuery.Message.Chat.ID)
						msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
						_, err := b.bot.Send(msg)
						if err != nil {
							slog.Error("error sending message", slog.String("error", err.Error()))
						}
					}

				}
			}
		}
	}
}
