package telegram

import (
	"log"
	"tg_bot_minenergo_ip/pkg/databases"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot  *tgbotapi.BotAPI
	base databases.Database
}

func NewBot(bot *tgbotapi.BotAPI, base databases.Database) *Bot {
	return &Bot{bot, base}
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s\n", b.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	b.handleUpdates(updates)

	return nil
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

			switch update.Message.Text {
			case "Подписаться":
				b.handleSubscribeComand(update.Message)

			case "Отписаться":
				b.handleUnsubscribeComand(update.Message)

			case "ДО \"Россети\"":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
						tgbotapi.NewInlineKeyboardButtonData("2", "2"),
						tgbotapi.NewInlineKeyboardButtonData("3", "3"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("4", "4"),
						tgbotapi.NewInlineKeyboardButtonData("5", "5"),
						tgbotapi.NewInlineKeyboardButtonData("6", "6"),
					),
				)
				// var numericKeyboard = tgbotapi.NewReplyKeyboard(
				// 	tgbotapi.NewKeyboardButtonRow(
				// 		tgbotapi.NewKeyboardButton("ПАО \"ФСК ЕЭС\""),
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети Волга\""),
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети ЦиП\""),
				// 	),
				// 	tgbotapi.NewKeyboardButtonRow(
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети Юг\""),
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети Центр\""),
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети Сибири\""),
				// 	),
				// 	tgbotapi.NewKeyboardButtonRow(
				// 		tgbotapi.NewKeyboardButton("ПАО \"МРСК Урала\""),
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети Северо-Запада\""),
				// 		tgbotapi.NewKeyboardButton("ПАО \"Россети Северный Кавказ\""),
				// 	),
				// )
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери субъект электроэнергетики:")
				msg.ReplyMarkup = numericKeyboard
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Printf("Не удалось раскрыть список ИП ДО Россети: %s", err)
				}

			case "ДО \"РусГидро\"":
				var numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("ПАО \"РусГидро\""),
						tgbotapi.NewKeyboardButton("АО \"ДРСК\""),
					),
				)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери субъект электроэнергетики:")
				msg.ReplyMarkup = numericKeyboard
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Printf("Не удалось раскрыть список ИП ДО РусГидро: %s", err)
				}

			case "Прочие":
				var numericKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("АО \"Концерн Росэнергоатом\""),
						tgbotapi.NewKeyboardButton("ОАО \"Сетевая компания\""),
					),
				)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выбери субъект электроэнергетики:")
				msg.ReplyMarkup = numericKeyboard
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Printf("Не удалось раскрыть список прочих ИП: %s", err)
				}

			case "ПАО \"ФСК ЕЭС\"":
				b.subdcribe(update.Message, fsk_ees)

			case "ПАО \"Россети Волга\"":
				b.subdcribe(update.Message, rosseti_volga)

			case "ПАО \"Россети ЦиП\"":
				b.subdcribe(update.Message, rosseti_cip)

			case "ПАО \"Россети Юг\"":
				b.subdcribe(update.Message, rosseti_yug)

			case "ПАО \"Россети Центр\"":
				b.subdcribe(update.Message, rosseti_centr)

			case "ПАО \"Россети Сибири\"":
				b.subdcribe(update.Message, rosseti_sibir)

			case "ПАО \"МРСК Урала\"":
				b.subdcribe(update.Message, rosseti_ural)

			case "ПАО \"Россети Северо-Запада\"":
				b.subdcribe(update.Message, rosseti_sev_zap)

			case "ПАО \"Россети Северный Кавказ\"":
				b.subdcribe(update.Message, rosseti_sev_kav)

			case "ПАО \"РусГидро\"":
				b.subdcribe(update.Message, rusgydro)

			case "АО \"ДРСК\"":
				b.subdcribe(update.Message, drsk)

			case "АО \"Концерн Росэнергоатом\"":
				b.subdcribe(update.Message, krea)

			case "ОАО \"Сетевая компания\"":
				b.subdcribe(update.Message, setevaya_komp)

			}

		} else if update.CallbackQuery != nil {
			log.Printf(update.CallbackQuery.Data)

			msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)})

			_, err := b.bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

		}
	}
}
