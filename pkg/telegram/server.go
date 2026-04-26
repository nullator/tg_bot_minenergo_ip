package telegram

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// EventsHandler — принимает события от telegram-server.
func (b *Bot) EventsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Authorization") != b.config.TelegramServerAuthToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			slog.Error("Ошибка декодирования события telegram-server",
				slog.String("error", err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if event.BotID != b.config.TelegramServerBotID {
			slog.Error("Получено событие неизвестного бота",
				slog.String("bot_id", event.BotID))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := b.HandleEvent(r.Context(), event); err != nil {
			slog.Error("Ошибка обработки события telegram-server",
				slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
