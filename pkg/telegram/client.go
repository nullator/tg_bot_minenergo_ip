package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"tg_bot_minenergo_ip/pkg/config"
)

// Client — HTTP-клиент для команд telegram-server.
type Client struct {
	baseURL   string
	authToken string
	botID     string
	client    *http.Client
}

// NewClient — создаёт HTTP-клиент для telegram-server.
func NewClient(cfg *config.Config) *Client {
	timeout := time.Duration(cfg.TelegramServerTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		baseURL:   strings.TrimRight(cfg.TelegramServerBaseURL, "/"),
		authToken: cfg.TelegramServerAuthToken,
		botID:     cfg.TelegramServerBotID,
		client:    &http.Client{Timeout: timeout},
	}
}

// Send — отправляет команду на отправку сообщения через telegram-server.
func (c *Client) Send(ctx context.Context, command SendCommand) error {
	command.BotID = c.botID

	body, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("marshal telegram command: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/telegram/send",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create telegram-server request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram-server request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("telegram-server returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}
