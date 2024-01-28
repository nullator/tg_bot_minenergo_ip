package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"log/slog"
)

type Log struct {
	App     string    `json:"app"`
	Message string    `json:"message"`
	Code    int       `json:"code"`
	Level   string    `json:"level"`
	Time    time.Time `json:"time"`
}

type CustomSlogHandlerInterface interface {
	Enabled(ctx context.Context, level slog.Level) bool
	Handle(ctx context.Context, r slog.Record) error
	WithAttrs(attrs []slog.Attr) slog.Handler
	WithGroup(name string) slog.Handler
	Handler() slog.Handler
}

type CustomSlogHandler struct {
	handler slog.Handler
}

var _ CustomSlogHandlerInterface = (*CustomSlogHandler)(nil)

func NewCustomSlogHandler(h slog.Handler) *CustomSlogHandler {
	return &CustomSlogHandler{h}
}

func (h *CustomSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *CustomSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	var msg = r.Message
	var lvl string
	var code int = 0

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "code" {
			if a.Value.Kind().String() == "Int64" {
				code = int(a.Value.Int64())
				return true
			}
		}
		msg = fmt.Sprintf("%s; %s: %s", msg, a.Key, a.Value)
		return true
	})

	switch r.Level.String() {
	case "DEBUG":
		return h.handler.Handle(ctx, r)
	case "INFO":
		lvl = "info"
	case "WARN":
		lvl = "warning"
	case "ERROR":
		lvl = "error"
	}

	go func(message string, code int, level string) {
		err := sendLogToServer(message, code, level)
		if err != nil {
			log.Printf("error sending log to server: %v", err)
		}
	}(msg, code, lvl)

	return h.handler.Handle(ctx, r)
}

func (h *CustomSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewCustomSlogHandler(h.handler.WithAttrs(attrs))
}

func (h *CustomSlogHandler) WithGroup(name string) slog.Handler {
	return NewCustomSlogHandler(h.handler.WithGroup(name))
}

func (h *CustomSlogHandler) Handler() slog.Handler {
	return h.handler
}

func sendLogToServer(message string, code int, level string) error {
	request := Log{
		App:     os.Getenv("APP_NAME"),
		Message: message,
		Code:    code,
		Level:   level,
		Time:    time.Now().UTC(),
	}

	json_request, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		os.Getenv("LOGGER_SERVER"),
		bytes.NewBuffer(json_request))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", os.Getenv("LOGGER_AUTH"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	return nil
}
