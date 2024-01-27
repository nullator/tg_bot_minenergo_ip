package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	commandStart = "start"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	command := strings.ToLower(message.Command())
	switch command {
	case commandStart:
		err := b.handleStartComand(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) handleStartComand(message *tgbotapi.Message) error {
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Настроить подписку", "subscribe"),
			// tgbotapi.NewInlineKeyboardButtonData("Отписаться", "unsubscribe"),
		),
	)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Ты можешь подписаться или отписаться от рассылки на уведомления о размещении материалов проектов ИП:")
	msg.ReplyMarkup = numericKeyboard
	_, err := b.bot.Send(msg)

	b.logger.Info("Выполнена команда Start")
	return err
}

func (b *Bot) subscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "subscride", ip)
	if err != nil {
		b.logger.Errorf("Ошибка сохранения в БД данных о подписке - %s", err.Error())
		return err
	}
	err = b.base.Save(ip, "subscride", fmt.Sprintf("%d", chatID))
	if err != nil {
		b.logger.Errorf("Ошибка сохранения в БД данных о подписке - %s", err.Error())
		return err
	}

	msg_text := fmt.Sprintf("Выполнена подписка на %s", b.config.IP[ip].Name)
	msg := tgbotapi.NewMessage(113053945, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		b.logger.Errorf("Не удалось отправить обратную связь: %s", err.Error())
	}

	b.logger.Infof("Успешно выполнена подписка на %s", b.config.IP[ip].Name)
	return err
}

func (b *Bot) unsubscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "unsubscride", ip)
	if err != nil {
		b.logger.Errorf("Ошибка сохранения в БД данных о подписке - %s", err.Error())
		return err
	}
	err = b.base.Save(ip, "unsubscride", fmt.Sprintf("%d", chatID))
	if err != nil {
		b.logger.Errorf("Ошибка сохранения в БД данных о подписке - %s", err.Error())
		return err
	}

	b.logger.Infof("Успешно выполнена отписка от %s", b.config.IP[ip].Name)
	return err
}
