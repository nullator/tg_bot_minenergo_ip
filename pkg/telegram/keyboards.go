package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func make_subscribe_kb(b *Bot, id_chat int64) tgbotapi.InlineKeyboardMarkup {
	ip_list, err := b.base.GetAll(fmt.Sprintf("%d", id_chat))
	if err != nil {
		b.logger.Errorf("Ошибка чтения из БД данных о подписках - %s", err.Error())
	}

	full_ip_list := make(map[int]string)
	for key, value := range b.config.IP {
		full_ip_list[value.ID-1] = key
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	i := 0
	n := len(full_ip_list) / 2
	for n > 0 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getICO(full_ip_list[i], ip_list)+b.config.IP[full_ip_list[i]].Name, "s"+full_ip_list[i]),
			tgbotapi.NewInlineKeyboardButtonData(getICO(full_ip_list[i+1], ip_list)+b.config.IP[full_ip_list[i+1]].Name, "s"+full_ip_list[i+1]),
		),
		)
		i += 2
		n -= 1
	}

	if len(full_ip_list)%2 == 1 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(getICO(full_ip_list[len(full_ip_list)-1], ip_list)+b.config.IP[full_ip_list[len(full_ip_list)-1]].Name, "s"+full_ip_list[len(full_ip_list)-1]),
		),
		)
	}

	numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "start"),
	),
	)

	return numericKeyboard
}

func make_unsubscribe_kb(b *Bot, id_chat int64) tgbotapi.InlineKeyboardMarkup {

	ip_list, err := b.base.GetAll(fmt.Sprintf("%d", id_chat))
	if err != nil {
		b.logger.Errorf("Ошибка чтения из БД данных о подписке - %s", err.Error())
	}

	subscribe_ip_list := make(map[int]string)
	i := 0
	for v, k := range ip_list {
		if k == "subscride" {
			subscribe_ip_list[i] = v
			i += 1
		}
	}

	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	i = 0
	n := len(subscribe_ip_list) / 2
	for n > 0 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⛔ "+b.config.IP[subscribe_ip_list[i]].Name, "u"+subscribe_ip_list[i]),
			tgbotapi.NewInlineKeyboardButtonData("⛔ "+b.config.IP[subscribe_ip_list[i+1]].Name, "u"+subscribe_ip_list[i+1]),
		),
		)
		i += 2
		n -= 1
	}

	if len(subscribe_ip_list)%2 == 1 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⛔ "+b.config.IP[subscribe_ip_list[len(subscribe_ip_list)-1]].Name, "u"+subscribe_ip_list[len(subscribe_ip_list)-1]),
		),
		)
	}

	numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "start"),
	),
	)

	return numericKeyboard
}

func getICO(ip string, ip_list map[string]string) string {
	if ip_list[ip] == "subscride" {
		return "✅ "
	} else {
		return "⬜ "
	}
}
