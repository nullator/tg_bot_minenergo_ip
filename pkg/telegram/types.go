package telegram

// Event — событие Telegram, нормализованное telegram-server.
type Event struct {
	BotID    string    `json:"bot_id"`
	Type     string    `json:"type"`
	Chat     *Chat     `json:"chat,omitempty"`
	Message  *Message  `json:"message,omitempty"`
	Callback *Callback `json:"callback,omitempty"`
}

// Chat — чат Telegram из события telegram-server.
type Chat struct {
	ID       int64  `json:"id"`
	Username string `json:"username,omitempty"`
}

// Message — сообщение Telegram из события telegram-server.
type Message struct {
	ID      int    `json:"id"`
	Text    string `json:"text,omitempty"`
	Command string `json:"command,omitempty"`
}

// Callback — callback от inline-кнопки из события telegram-server.
type Callback struct {
	Data      string `json:"data,omitempty"`
	MessageID int    `json:"message_id,omitempty"`
}

// SendCommand — команда отправки сообщения через telegram-server.
type SendCommand struct {
	BotID       string       `json:"bot_id"`
	ChatID      int64        `json:"chat_id"`
	Text        string       `json:"text,omitempty"`
	ParseMode   string       `json:"parse_mode,omitempty"`
	ReplyMarkup *ReplyMarkup `json:"reply_markup,omitempty"`
}

// ReplyMarkup — клавиатура сообщения для telegram-server.
type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

// InlineKeyboardButton — кнопка inline-клавиатуры для telegram-server.
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}
