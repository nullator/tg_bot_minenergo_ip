package telegram

import (
	"log"
	"tg_bot_minenergo_ip/pkg/config"
	"tg_bot_minenergo_ip/pkg/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot    *tgbotapi.BotAPI
	base   databases.Database
	config *config.Config
}

func NewBot(bot *tgbotapi.BotAPI, base databases.Database, config *config.Config) *Bot {
	return &Bot{bot, base, config}
}

func (b *Bot) Start() {
	log.Printf("Authorized on account %s\n", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	b.handleUpdates(updates)
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message != nil { // If we got a message
			if update.Message.IsCommand() {
				log.Printf("[%s] ввёл команду %s", update.Message.From.UserName, update.Message.Text)
				if err := b.handleCommand(update.Message); err != nil {
					log.Printf("При обработке команды %s произошла ошибка %s", update.Message.Command(), err)
				}
				continue
			}

			log.Printf("[%s] отправил сообщение: %s", update.Message.From.UserName, update.Message.Text)

		} else if update.CallbackQuery != nil {
			// msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)})
			// _, err := b.bot.Send(msg)
			// if err != nil {
			// 	log.Println(err)
			// }

			q := update.CallbackQuery.Data
			switch q {
			case "start":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Подписаться", "subscribe"),
						tgbotapi.NewInlineKeyboardButtonData("Отписаться", "unsubscribe"),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "subscribe":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
				numericKeyboard = make_subscribe_kb(b, update.CallbackQuery.Message.Chat.ID)

				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "unsubscribe":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
				numericKeyboard = make_unsubscribe_kb(b, update.CallbackQuery.Message.Chat.ID)

				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			default:
				first_letter := string([]rune(q)[0])
				code := string([]rune(q)[1:5])
				if first_letter == "s" {
					log.Printf("Пользователь %s запросил подписку на %s", update.CallbackQuery.Message.Chat.UserName, b.config.IP[code].Name)
					b.subscribe(update.CallbackQuery.Message.Chat.ID, code)

					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					numericKeyboard = make_subscribe_kb(b, update.CallbackQuery.Message.Chat.ID)
					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err := b.bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
				}
				if first_letter == "u" {
					log.Printf("Пользователь %s запросил отписку от %s", update.CallbackQuery.Message.Chat.UserName, b.config.IP[code].Name)
					b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)

					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					numericKeyboard = make_unsubscribe_kb(b, update.CallbackQuery.Message.Chat.ID)
					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err := b.bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}
}
