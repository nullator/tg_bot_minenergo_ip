package telegram

import (
	"fmt"
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

		} else if update.CallbackQuery != nil {
			msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, tgbotapi.InlineKeyboardMarkup{InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0)})
			_, err := b.bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

			q := update.CallbackQuery.Data
			switch q {
			case "subscribe":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ДО ПАО \"Россети\"", "s_rosseti"),
						tgbotapi.NewInlineKeyboardButtonData("Прочие", "s_other"),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "unsubscribe":
				ip_list, err := b.base.GetAll(fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID))
				subscribe_ip_list := make(map[int]string)
				i := 0
				for v, k := range ip_list {
					if k == "subscride" {
						subscribe_ip_list[i] = v
						i += 1
					}
				}

				if err != nil {
					log.Printf("Ошибка чтения из БД данных о подписке")
				}

				if len(subscribe_ip_list) > 0 {
					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					for _, v := range subscribe_ip_list {
						numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Тест", "u"+v)))
					}

					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err = b.bot.Send(msg)
					if err != nil {
						log.Println(err)
					}

				}

				// var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
				// 	tgbotapi.NewInlineKeyboardRow(
				// 		tgbotapi.NewInlineKeyboardButtonData("ДО ПАО \"Россети\"", "u_rosseti"),
				// 		tgbotapi.NewInlineKeyboardButtonData("Прочие", "u_other"),
				// 	),
				// )

			case "s_rosseti":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"ФСК ЕЭС\"", "s"+fsk_ees),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Волга\"", "s"+rosseti_volga),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Юг\"", "s"+rosseti_yug),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Центр\"", "s"+rosseti_centr),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Сибири\"", "s"+rosseti_sibir),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети ЦиП\"", "s"+rosseti_cip),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"МРСК Урала\"", "s"+rosseti_ural),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Северо-Запада\"", "s"+rosseti_sev_zap),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "s_other":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"РусГидро\"", "s"+rusgydro),
						tgbotapi.NewInlineKeyboardButtonData("АО \"Концерн Росэнергоатом\"", "s"+krea),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "u_rosseti":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"ФСК ЕЭС\"", "u"+fsk_ees),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Волга\"", "u"+rosseti_volga),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Юг\"", "u"+rosseti_yug),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Центр\"", "u"+rosseti_centr),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Сибири\"", "u"+rosseti_sibir),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети ЦиП\"", "u"+rosseti_cip),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"МРСК Урала\"", "u"+rosseti_ural),
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"Россети Северо-Запада\"", "u"+rosseti_sev_zap),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "u_other":
				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("ПАО \"РусГидро\"", "u"+rusgydro),
						tgbotapi.NewInlineKeyboardButtonData("АО \"Концерн Росэнергоатом\"", "u"+krea),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err := b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			default:
				first_letter := string([]rune(q)[0])
				code := string([]rune(q)[1:5])
				if first_letter == "s" {
					b.subscribe(update.CallbackQuery.Message.Chat.ID, code)

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
				}
				if first_letter == "u" {
					b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)

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
				}
			}
		}
	}
}
