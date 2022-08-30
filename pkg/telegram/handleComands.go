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
	}

	// // msg_text := fmt.Sprintf("Выполнена подписка на %s", message.Text)
	// news_txt, err := parser_ip.Parse(ip)
	// msg_text := fmt.Sprintf("Последняя запись: %s", news_txt)

	// msg := tgbotapi.NewMessage(chatID, msg_text)
	// _, err = b.bot.Send(msg)

	return err
}

func (b *Bot) unsubscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "", ip)
	if err != nil {
		log.Printf("Ошибка сохранения в БД данных о подписке")
	}

	return err
}
