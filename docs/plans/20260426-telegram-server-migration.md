# Миграция на telegram-server

## Цель

Полностью убрать прямую работу приложения с Telegram API через `tgbotapi` и заменить её на интеграцию с `telegram-server`.

После миграции приложение должно:

- принимать входящие команды и callback через HTTP endpoint;
- отправлять исходящие сообщения через `POST /telegram/send`;
- сохранять бизнес-логику подписок, парсинга, выявления новых записей и рассылки внутри текущего приложения;
- не хранить и не использовать Telegram token.

## Принятое допущение

UX inline-кнопок можно изменить: вместо редактирования существующей клавиатуры приложение будет отправлять новое сообщение с актуальной клавиатурой.

Это нужно, потому что в текущей документации `telegram-server` описана только отправка сообщений через `POST /telegram/send`, без команды для `editMessageReplyMarkup`.

## План реализации

### 1. Обновить конфигурацию приложения

Файл:

- `pkg/config/config.go`

Изменения:

- удалить использование `TelegramToken`;
- добавить настройки `telegram-server`:
  - `TelegramServerBaseURL`;
  - `TelegramServerAuthToken`;
  - `TelegramServerBotID`;
  - `TelegramServerTimeoutSeconds`;
- добавить настройки HTTP-сервера приложения:
  - `HTTPAddr`;
  - `TelegramEventsPath`.

Критерий готовности:

- приложение больше не читает `TOKEN`;
- все параметры интеграции берутся из `.env`;
- Telegram token остаётся только на стороне `telegram-server`.

### 2. Добавить локальные модели для telegram-server

Вероятный файл:

- `pkg/telegram/types.go`

Изменения:

- описать входящее событие от `telegram-server`:
  - `Event`;
  - `Chat`;
  - `Message`;
  - `Callback`;
- описать исходящую команду:
  - `SendCommand`;
  - `ReplyMarkup`;
  - `InlineKeyboardButton`.

Критерий готовности:

- код больше не зависит от типов `tgbotapi.Message`, `tgbotapi.InlineKeyboardMarkup`, `tgbotapi.BotAPI`.

### 3. Добавить HTTP-клиент для отправки сообщений через telegram-server

Вероятный файл:

- `pkg/telegram/client.go`

Изменения:

- реализовать клиент с методом `Send(ctx context.Context, command SendCommand) error`;
- отправлять `POST {base_url}/telegram/send`;
- выставлять headers:
  - `Content-Type: application/json`;
  - `Authorization: <auth_token>`;
- автоматически подставлять `bot_id`;
- считать успешным ответом `202 Accepted`;
- при ошибках читать ограниченный body ответа и возвращать понятную ошибку.

Критерий готовности:

- все исходящие сообщения идут только через `telegram-server`;
- в коде не остаётся `b.bot.Send(...)`.

### 4. Переделать структуру `Bot`

Файл:

- `pkg/telegram/bot.go`

Изменения:

- убрать поле `bot *tgbotapi.BotAPI`;
- заменить его на HTTP-клиент `telegram-server`;
- изменить конструктор `NewBot`;
- удалить polling через `GetUpdatesChan`;
- оставить методы обработки событий бизнес-логики.

Критерий готовности:

- `Bot` перестаёт быть Telegram-клиентом;
- `Bot` становится обработчиком бизнес-сценариев Telegram-событий.

### 5. Добавить HTTP endpoint для событий от telegram-server

Вероятный файл:

- `pkg/telegram/server.go`

Изменения:

- добавить handler для `POST /telegram/events`;
- проверять метод запроса;
- проверять `Authorization`;
- декодировать JSON события;
- проверять `bot_id`;
- вызывать бизнес-обработку события;
- возвращать:
  - `202 Accepted` при успехе;
  - `401 Unauthorized` при неверной авторизации;
  - `400 Bad Request` при неверном JSON;
  - `405 Method Not Allowed` при неверном методе;
  - `500 Internal Server Error` при ошибке бизнес-логики.

Критерий готовности:

- приложение принимает нормализованные события от `telegram-server`;
- polling Telegram updates полностью удалён.

### 6. Перенести обработку команд и callback на новые модели

Файлы:

- `pkg/telegram/bot.go`
- `pkg/telegram/handleComands.go`

Изменения:

- заменить `handleCommand(message *tgbotapi.Message)` на обработку `Event`;
- команду `/start` читать из `event.Message.Command`;
- `chat_id` брать из `event.Chat.ID`;
- callback data брать из `event.Callback.Data`;
- текущую логику подписки/отписки сохранить:
  - `subscribe`;
  - `unsubscribe`;
  - работа с БД;
  - логирование.

Критерий готовности:

- сценарии `/start`, `subscribe`, `sXXXX`, `uXXXX`, `start` продолжают работать;
- обработка больше не зависит от типов Telegram SDK.

### 7. Переделать клавиатуры на JSON-структуры telegram-server

Файл:

- `pkg/telegram/keyboards.go`

Изменения:

- заменить возвращаемый тип `tgbotapi.InlineKeyboardMarkup` на локальный `ReplyMarkup`;
- заменить `tgbotapi.NewInlineKeyboardButtonData(...)` на `InlineKeyboardButton`;
- сохранить текущие callback values:
  - `subscribe`;
  - `start`;
  - `s<код>`;
  - `u<код>`.

Критерий готовности:

- inline-клавиатуры отправляются через `telegram-server`;
- внешний формат callback остаётся совместимым с текущей бизнес-логикой.

### 8. Заменить рассылку уведомлений о новых записях

Файл:

- `pkg/telegram/loadIP.go`

Изменения:

- убрать `tgbotapi.NewMessage`;
- отправлять уведомления через `telegramClient.Send`;
- сохранить `ParseMode: "Markdown"`;
- сохранить задержку `300ms` между отправками.

Критерий готовности:

- регулярный парсинг и выявление новых записей остаются без изменения;
- рассылка подписчикам идёт через `telegram-server`.

### 9. Переделать запуск приложения

Файл:

- `cmd/main.go`

Изменения:

- удалить создание `tgbotapi.NewBotAPI`;
- создать HTTP-клиент `telegram-server`;
- создать `telegram.Bot`;
- запустить `LoadIP(ctx)` в goroutine;
- поднять `http.Server` на `cfg.HTTPAddr`;
- зарегистрировать endpoint `cfg.TelegramEventsPath`;
- при `SIGTERM`/`Interrupt` корректно останавливать:
  - контекст парсинга;
  - HTTP-сервер через `Shutdown`.

Критерий готовности:

- приложение становится HTTP-сервисом;
- основной процесс одновременно выполняет регулярный парсинг и принимает события от `telegram-server`.

### 10. Удалить зависимость `tgbotapi`

Файлы:

- `go.mod`;
- `go.sum`.

Изменения:

- убрать импорт `github.com/go-telegram-bot-api/telegram-bot-api/v5`;
- выполнить `go mod tidy`;
- проверить, что `rg "tgbotapi|telegram-bot-api|NewBotAPI|GetUpdatesChan|BotAPI"` ничего не находит.

Критерий готовности:

- в проекте не остаётся прямой зависимости от Telegram SDK.

### 11. Проверка

Команды:

```bash
gofmt -w cmd/main.go pkg/config/config.go pkg/telegram/*.go
go test ./...
go build ./...
```

Ручная проверка:

- отправить тестовое событие `/start` на `POST /telegram/events`;
- проверить, что приложение делает запрос в `telegram-server`;
- отправить callback `subscribe`;
- проверить запись подписки в БД;
- имитировать новую запись ИП или локально вызвать соответствующий сценарий рассылки;
- проверить, что уведомление уходит через `POST /telegram/send`.

## Основные риски

1. `telegram-server` принимает только отправку новых сообщений.
   Поэтому текущее редактирование inline-клавиатуры будет заменено отправкой нового сообщения.

2. Доставка входящих событий без retry.
   Если приложение вернёт non-2xx или не ответит вовремя, повторной доставки сейчас нет.

3. `202 Accepted` не означает доставку в Telegram.
   Приложение будет знать только то, что `telegram-server` принял команду в очередь.

4. Пакет остаётся `pkg/telegram`.
   Переименование не выполняется, чтобы не делать структурный рефакторинг.

## Критерии готовности миграции

- в проекте нет импортов `tgbotapi`;
- в `go.mod` нет `github.com/go-telegram-bot-api/telegram-bot-api/v5`;
- приложение не читает Telegram token;
- приложение поднимает HTTP endpoint для событий;
- `/start` работает через событие от `telegram-server`;
- подписка/отписка работают через callback events;
- рассылка новых записей отправляется через `POST /telegram/send`;
- `go test ./...` и `go build ./...` проходят.
