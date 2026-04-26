package telegram

import (
	"fmt"
	"log/slog"
)

func makeStartKeyboard() *ReplyMarkup {
	return &ReplyMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         "Настроить подписку",
					CallbackData: "subscribe",
				},
			},
		},
	}
}

func make_subscribe_kb(b *Bot, id_chat int64) *ReplyMarkup {
	ip_list, err := b.base.GetAll(fmt.Sprintf("%d", id_chat))
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о подписке", slog.String("error", err.Error()))
	}

	full_ip_list := make(map[int]string)
	for key, value := range b.config.IP {
		full_ip_list[value.ID-1] = key
	}

	numericKeyboard := &ReplyMarkup{}
	i := 0
	n := len(full_ip_list) / 2
	for n > 0 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, []InlineKeyboardButton{
			{
				Text:         getICO(full_ip_list[i], ip_list) + b.config.IP[full_ip_list[i]].Name,
				CallbackData: "s" + full_ip_list[i],
			},
			{
				Text:         getICO(full_ip_list[i+1], ip_list) + b.config.IP[full_ip_list[i+1]].Name,
				CallbackData: "s" + full_ip_list[i+1],
			},
		})
		i += 2
		n -= 1
	}

	if len(full_ip_list)%2 == 1 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, []InlineKeyboardButton{
			{
				Text:         getICO(full_ip_list[len(full_ip_list)-1], ip_list) + b.config.IP[full_ip_list[len(full_ip_list)-1]].Name,
				CallbackData: "s" + full_ip_list[len(full_ip_list)-1],
			},
		})
	}

	numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, []InlineKeyboardButton{
		{
			Text:         "⬅️ Сохранить",
			CallbackData: "start",
		},
	})

	return numericKeyboard
}

func make_unsubscribe_kb(b *Bot, id_chat int64) *ReplyMarkup {

	ip_list, err := b.base.GetAll(fmt.Sprintf("%d", id_chat))
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о подписке", slog.String("error", err.Error()))
	}

	subscribe_ip_list := make(map[int]string)
	i := 0
	for v, k := range ip_list {
		if k == "subscride" {
			subscribe_ip_list[i] = v
			i += 1
		}
	}

	numericKeyboard := &ReplyMarkup{}
	i = 0
	n := len(subscribe_ip_list) / 2
	for n > 0 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, []InlineKeyboardButton{
			{
				Text:         "⛔ " + b.config.IP[subscribe_ip_list[i]].Name,
				CallbackData: "u" + subscribe_ip_list[i],
			},
			{
				Text:         "⛔ " + b.config.IP[subscribe_ip_list[i+1]].Name,
				CallbackData: "u" + subscribe_ip_list[i+1],
			},
		})
		i += 2
		n -= 1
	}

	if len(subscribe_ip_list)%2 == 1 {
		numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, []InlineKeyboardButton{
			{
				Text:         "⛔ " + b.config.IP[subscribe_ip_list[len(subscribe_ip_list)-1]].Name,
				CallbackData: "u" + subscribe_ip_list[len(subscribe_ip_list)-1],
			},
		})
	}

	numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, []InlineKeyboardButton{
		{
			Text:         "⬅️ Сохранить",
			CallbackData: "start",
		},
	})

	return numericKeyboard
}

func getICO(ip string, ip_list map[string]string) string {
	if ip_list[ip] == "subscride" {
		return "✅ "
	} else {
		return "⬜ "
	}
}
