package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type LoggerInterface interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type Log struct {
	App     string    `json:"app"`
	Message string    `json:"message"`
	Code    int       `json:"code"`
	Level   string    `json:"level"`
	Time    time.Time `json:"time"`
}

type Logger struct {
	name   string
	logger *log.Logger
}

var _ LoggerInterface = (*Logger)(nil)

func New(name string, l *log.Logger) *Logger {
	return &Logger{name: name, logger: l}
}

func (l *Logger) Info(args ...interface{}) {
	output := fmt.Sprint(args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "info")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	output := fmt.Sprintf(format, args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "info")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Warn(args ...interface{}) {
	output := fmt.Sprint(args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "warning")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	output := fmt.Sprintf(format, args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "warning")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Error(args ...interface{}) {
	output := fmt.Sprint(args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "error")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	output := fmt.Sprintf(format, args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "error")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Fatal(args ...interface{}) {
	output := fmt.Sprint(args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "fatal")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	output := fmt.Sprintf(format, args...)
	l.logger.Print(output)
	err := sendLogToServer(output, 0, "fatal")
	if err != nil {
		l.logger.Printf("ERROR SEND TO LOG SERVER: %s\n", err)
	}
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

	req, err := http.NewRequest("POST", os.Getenv("LOGGER_SERVER"), bytes.NewBuffer(json_request))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", os.Getenv("LOGGER_AUTH"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
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
