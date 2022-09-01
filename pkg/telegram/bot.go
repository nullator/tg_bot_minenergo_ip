package telegram

import (
	"fmt"
	"log"
	"strconv"
	"tg_bot_minenergo_ip/pkg/databases"
	parser_ip "tg_bot_minenergo_ip/pkg/parser"
	"time"

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
				if err != nil {
					log.Printf("Ошибка чтения из БД данных о подписке")
				}

				subscribe_ip_list := make(map[int]string)
				i := 0
				for v, k := range ip_list {
					if k == "subscride" {
						subscribe_ip_list[i] = v
						i += 1
					}
				}

				if len(subscribe_ip_list) > 0 {
					var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
					i = 0
					n := len(subscribe_ip_list) / 2
					for n > 0 {
						numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("⛔ "+getIPname(subscribe_ip_list[i]), "u"+subscribe_ip_list[i]),
							tgbotapi.NewInlineKeyboardButtonData("⛔ "+getIPname(subscribe_ip_list[i+1]), "u"+subscribe_ip_list[i+1]),
						),
						)
						i += 2
						n -= 1
					}

					if len(subscribe_ip_list)%2 == 1 {
						numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("⛔ "+getIPname(subscribe_ip_list[len(subscribe_ip_list)-1]), "u"+subscribe_ip_list[len(subscribe_ip_list)-1]),
						),
						)
					}

					numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅️ Отмена", "start"),
					),
					)

					msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
					_, err = b.bot.Send(msg)
					if err != nil {
						log.Println(err)
					}

				}

			case "s_rosseti":
				ip_list, err := b.base.GetAll(fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID))
				if err != nil {
					log.Printf("Ошибка чтения из БД данных о подписках")
				}

				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(getICO(fsk_ees, ip_list)+"ПАО \"ФСК ЕЭС\"", "s"+fsk_ees),
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_volga, ip_list)+"ПАО \"Россети Волга\"", "s"+rosseti_volga),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_yug, ip_list)+"ПАО \"Россети Юг\"", "s"+rosseti_yug),
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_centr, ip_list)+"ПАО \"Россети Центр\"", "s"+rosseti_centr),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_sibir, ip_list)+"ПАО \"Россети Сибири\"", "s"+rosseti_sibir),
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_cip, ip_list)+"ПАО \"Россети ЦиП\"", "s"+rosseti_cip),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_ural, ip_list)+"ПАО \"МРСК Урала\"", "s"+rosseti_ural),
						tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_sev_zap, ip_list)+"ПАО \"Россети Сев-Зап\"", "s"+rosseti_sev_zap),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅️ Отмена", "start"),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err = b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			case "s_other":
				ip_list, err := b.base.GetAll(fmt.Sprintf("%d", update.CallbackQuery.Message.Chat.ID))
				if err != nil {
					log.Printf("Ошибка чтения из БД данных о подписках")
				}

				var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(getICO(rusgydro, ip_list)+"ПАО \"РусГидро\"", "s"+rusgydro),
						tgbotapi.NewInlineKeyboardButtonData(getICO(krea, ip_list)+"АО \"Концерн Росэнергоатом\"", "s"+krea),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⬅️ Отмена", "start"),
					),
				)
				msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
				_, err = b.bot.Send(msg)
				if err != nil {
					log.Println(err)
				}

			// case "u_rosseti":
			// 	ip_list, err := b.base.GetAll(string(update.CallbackQuery.Message.Chat.ID))
			// 	if err != nil {
			// 		log.Printf("Ошибка чтения из БД данных о подписках")
			// 	}
			// 	log.Println(ip_list)
			// 	log.Println(getICO(fsk_ees, ip_list))
			// 	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			// 		tgbotapi.NewInlineKeyboardRow(
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(fsk_ees, ip_list)+"ПАО \"ФСК ЕЭС\"", "u"+fsk_ees),
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_volga, ip_list)+"ПАО \"Россети Волга\"", "u"+rosseti_volga),
			// 		),
			// 		tgbotapi.NewInlineKeyboardRow(
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_yug, ip_list)+"ПАО \"Россети Юг\"", "u"+rosseti_yug),
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_centr, ip_list)+"ПАО \"Россети Центр\"", "u"+rosseti_centr),
			// 		),
			// 		tgbotapi.NewInlineKeyboardRow(
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_sibir, ip_list)+"ПАО \"Россети Сибири\"", "u"+rosseti_sibir),
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_cip, ip_list)+"ПАО \"Россети ЦиП\"", "u"+rosseti_cip),
			// 		),
			// 		tgbotapi.NewInlineKeyboardRow(
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_ural, ip_list)+"ПАО \"МРСК Урала\"", "u"+rosseti_ural),
			// 			tgbotapi.NewInlineKeyboardButtonData(getICO(rosseti_sev_zap, ip_list)+"ПАО \"Россети Северо-Запада\"", "u"+rosseti_sev_zap),
			// 		),
			// 	)
			// 	msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
			// 	_, err = b.bot.Send(msg)
			// 	if err != nil {
			// 		log.Println(err)
			// 	}

			// case "u_other":
			// 	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			// 		tgbotapi.NewInlineKeyboardRow(
			// 			tgbotapi.NewInlineKeyboardButtonData("ПАО \"РусГидро\"", "u"+rusgydro),
			// 			tgbotapi.NewInlineKeyboardButtonData("АО \"Концерн Росэнергоатом\"", "u"+krea),
			// 		),
			// 	)
			// 	msg := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, numericKeyboard)
			// 	_, err := b.bot.Send(msg)
			// 	if err != nil {
			// 		log.Println(err)
			// 	}

			default:
				first_letter := string([]rune(q)[0])
				code := string([]rune(q)[1:5])
				if first_letter == "s" {
					log.Printf("Пользователь %s запросил подписку на %s", update.CallbackQuery.Message.Chat.UserName, getIPname(code))
					b.subscribe(update.CallbackQuery.Message.Chat.ID, code)
				}
				if first_letter == "u" {
					log.Printf("Пользователь %s запросил отписку от %s", update.CallbackQuery.Message.Chat.UserName, getIPname(code))
					b.unsubscribe(update.CallbackQuery.Message.Chat.ID, code)
				}
			}
		}
	}
}

func (b *Bot) LoadIP() {
	full_ip_list := []string{rosseti_volga, fsk_ees, rosseti_cip, rosseti_yug, rosseti_centr,
		rosseti_sibir, rosseti_ural, rosseti_sev_zap, rosseti_sev_kav, rusgydro, drsk, krea}

	for {
		for _, ip := range full_ip_list {
			new_report, err := parser_ip.Parse(ip)
			if err != nil {
				log.Printf("Ошибка парсинга: %s", err)
			}
			old_report, err := b.base.Get(ip, ip)
			if err != nil {
				log.Printf("Ошибка чтения из БД при парсинге новости: %s", err)
			}
			if (new_report != old_report) && (new_report != "ERROR") {
				log.Printf("Обнаружена новая запись ИП %s: %s", getIPname(ip), new_report)
				b.make_notify(ip, new_report)
				err = b.base.Save(ip, new_report, ip)
				if err != nil {
					log.Printf("Ошибка сохранения в БД новой новости по ИП: %s", err)
				}
			}
		}
		time.Sleep(time.Minute * 10)
	}
}

func (b *Bot) make_notify(ip string, news string) {
	users, err := b.base.GetAll(ip)
	if err != nil {
		log.Printf("Ошибка чтения из БД данных о подписчиках")
	}

	for id, key := range users {
		if key == "subscride" {
			msg_text := "Размещены материалы ИП **" + getIPname(ip) + "**:\n" + news
			id_int64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Println(err)
			}
			msg := tgbotapi.NewMessage(id_int64, msg_text)
			_, err = b.bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

		}
	}
}

func getICO(ip string, ip_list map[string]string) string {
	if ip_list[ip] == "subscride" {
		return "✅ "
	} else {
		return "⬜ "
	}
}
