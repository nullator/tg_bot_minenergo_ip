package telegram

import (
	"fmt"
	"log/slog"
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

	slog.Info("Выполнена команда Start")
	return err
}

func (b *Bot) subscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "subscride", ip)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о подписке", slog.String("error", err.Error()))
		return err
	}
	err = b.base.Save(ip, "subscride", fmt.Sprintf("%d", chatID))
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о подписке", slog.String("error", err.Error()))
		return err
	}

	msg_text := fmt.Sprintf("Выполнена подписка на %s", b.config.IP[ip].Name)
	msg := tgbotapi.NewMessage(113053945, msg_text)
	_, err = b.bot.Send(msg)
	if err != nil {
		slog.Error("Не удалось отправить обратную связь", slog.String("error", err.Error()))
	}

	slog.Info("Выполнена подписка",
		slog.String("user_id", fmt.Sprintf("%d", chatID)),
		slog.String("ip", b.config.IP[ip].Name))
	return err
}

func (b *Bot) unsubscribe(chatID int64, ip string) error {
	err := b.base.Save(fmt.Sprintf("%d", chatID), "unsubscride", ip)
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о подписке", slog.String("error", err.Error()))
		return err
	}
	err = b.base.Save(ip, "unsubscride", fmt.Sprintf("%d", chatID))
	if err != nil {
		slog.Error("Ошибка сохранения в БД данных о подписке", slog.String("error", err.Error()))
		return err
	}

	slog.Info("Выполнена отписка",
		slog.String("user_id", fmt.Sprintf("%d", chatID)),
		slog.String("ip", b.config.IP[ip].Name))
	return err
}
