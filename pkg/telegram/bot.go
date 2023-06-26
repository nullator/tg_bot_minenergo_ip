package telegram

import (
	"fmt"
	"tg_bot_minenergo_ip/pkg/config"
	"tg_bot_minenergo_ip/pkg/databases"
	"tg_bot_minenergo_ip/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot    *tgbotapi.BotAPI
	base   databases.Database
	config *config.Config
	logger *logger.Logger
}

func NewBot(bot *tgbotapi.BotAPI, base databases.Database,
	config *config.Config, logger *logger.Logger) *Bot {
	return &Bot{bot, base, config, logger}
}

func (b *Bot) Start() {
	b.logger.Infof("Authorized on account %s", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	b.handleUpdates(updates)
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				b.logger.Infof("[%s] ввёл команду %s", update.Message.From.UserName, update.Message.Text)
				if err := b.handleCommand(update.Message); err != nil {
					b.logger.Errorf("При обработке команды %s произошла ошибка %s", update.Message.Command(), err)
				}
				continue
			}

			b.logger.Infof("[%s] отправил сообщение: %s", update.Message.From.UserName, update.Message.Text)

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
					b.logger.Error(err)
				}

			case "subscribe":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
				numericKeyboard = make_subscribe_kb(b, update.CallbackQuery.Message.Chat.ID)

				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					b.logger.Error(err)
				}

			default:
				first_letter := string([]rune(q)[0])
				code := string([]rune(q)[1:5])
				if first_letter == "s" {
					status, err := b.base.Get(fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID), code)
					if err != nil {
						b.logger.Error(err)
					}
					if status == "subscride" {
						b.logger.Infof("Пользователь %s запросил отписку от %s", update.CallbackQuery.Message.Chat.UserName, b.config.IP[code].Name)
						b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)

					} else {
						b.logger.Infof("Пользователь %s запросил подписку на %s", update.CallbackQuery.Message.Chat.UserName, b.config.IP[code].Name)
						b.subscribe(update.CallbackQuery.Message.Chat.ID, code)
					}
					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					numericKeyboard = make_subscribe_kb(b, update.CallbackQuery.Message.Chat.ID)
					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err = b.bot.Send(msg)
					if err != nil {
						b.logger.Error(err)
					}

				}
				if first_letter == "u" {
					b.logger.Infof("Пользователь %s запросил отписку от %s", update.CallbackQuery.Message.Chat.UserName, b.config.IP[code].Name)
					b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)

					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					numericKeyboard = make_unsubscribe_kb(b, update.CallbackQuery.Message.Chat.ID)
					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err := b.bot.Send(msg)
					if err != nil {
						b.logger.Error(err)
					}
				}

			}
		}
	}
}
