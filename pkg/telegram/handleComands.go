package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	commandStart       = "start"
	commandSubscribe   = "Подписаться"
	commandUnsubscribe = "Отписаться"
)

const (
	rosseti_volga   = "4190"
	fsk_ees         = "4174"
	rosseti_cip     = "4178"
	rosseti_yug     = "4191"
	rosseti_centr   = "4192"
	rosseti_sibir   = "4185"
	rosseti_ural    = "4189"
	rosseti_sev_zap = "4193"
	rosseti_sev_kav = "4177"
	rusgydro        = "4195"
	drsk            = "4217"
	krea            = "4224"
	setevaya_komp   = "4361"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	command := strings.ToLower(message.Command())
	switch command {
	case commandStart:
		err := b.handleStartComand(message)
		if err != nil {
			return err
		}
	case commandSubscribe:
		err := b.handleSubscribeComand(message.Chat.ID)
		if err != nil {
			return err
		}
	case commandUnsubscribe:
		err := b.handleUnsubscribeComand(message.Chat.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) handleStartComand(message *tgbotapi.Message) error {
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подписаться", "subscribe"),
			tgbotapi.NewInlineKeyboardButtonData("Отписаться", "unsubscribe"),
		),
	)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Ты можешь подписаться или отписаться от рассылки на уведомления о размещении материалов проектов ИП:")
	msg.ReplyMarkup = numericKeyboard
	_, err := b.bot.Send(msg)

	log.Println("Выполнена команда Start")
	return err
}

func (b *Bot) handleSubscribeComand(chatID int64) error {
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ДО ПАО \"Россети\"", "rosseti"),
			tgbotapi.NewInlineKeyboardButtonData("Прочие", "other"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Выбери группу субъектов:")
	msg.ReplyMarkup = numericKeyboard
	_, err := b.bot.Send(msg)

	log.Println("Выполнена команда Subscribe")
	return err
}

func (b *Bot) handleUnsubscribeComand(chatID int64) error {
	b.base.GetAll(rosseti_volga)

	log.Println("Выполнена команда Unsubscribe")
	return nil
}

func (b *Bot) subscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "subscride", ip)
	if err != nil {
		log.Printf("Ошибка сохранения в БД данных о подписке")
		return err
	}
	err = b.base.Save(ip, "subscride", fmt.Sprintf("%d", chatID))
	if err != nil {
		log.Printf("Ошибка сохранения в БД данных о подписке")
		return err
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подписаться", "subscribe"),
			tgbotapi.NewInlineKeyboardButtonData("Отписаться", "unsubscribe"),
		),
	)
	msg_text := "Выполнена подписка на " + getIPname(ip)
	msg := tgbotapi.NewMessage(chatID, msg_text)
	msg.ReplyMarkup = numericKeyboard
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (b *Bot) unsubscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "unsubscride", ip)
	if err != nil {
		log.Printf("Ошибка сохранения в БД данных о подписке")
		return err
	}
	err = b.base.Save(ip, "unsubscride", fmt.Sprintf("%d", chatID))
	if err != nil {
		log.Printf("Ошибка сохранения в БД данных о подписке")
		return err
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подписаться", "subscribe"),
			tgbotapi.NewInlineKeyboardButtonData("Отписаться", "unsubscribe"),
		),
	)
	msg_text := "Выполнена отписка от " + getIPname(ip)
	msg := tgbotapi.NewMessage(chatID, msg_text)
	msg.ReplyMarkup = numericKeyboard
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	return err
}

func getIPname(ip string) string {
	switch ip {
	case rosseti_volga:
		return "ПАО \"Россети Волга\""
	case fsk_ees:
		return "ПАО \"ФСК ЕЭС\""
	case rosseti_cip:
		return "ПАО \"Россети Центр и Приволжье\""
	case rosseti_yug:
		return "ПАО \"Россети Юг\""
	case rosseti_centr:
		return "ПАО \"Россети Центр\""
	case rosseti_sibir:
		return "ПАО \"Россети Сибирь\""
	case rosseti_ural:
		return "ОАО \"МРСК Урала\""
	case rosseti_sev_zap:
		return "ПАО \"Россети Северо-Запад\""
	case rosseti_sev_kav:
		return "ПАО \"Россети Северный Кавказ\""
	case rusgydro:
		return "ПАО \"РусГидро\""
	case drsk:
		return "ПАО \"РусГидро\""
	case krea:
		return "АО \"Концерн Росэнергоатом\""
	}
	return ""
}
