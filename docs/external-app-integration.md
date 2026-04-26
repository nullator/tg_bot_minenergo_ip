# Интеграция сторонних приложений с telegram-server

Документ предназначен для разработчиков сторонних приложений, в которых уже есть бизнес-логика Telegram-ботов.

`telegram-server` позволяет вынести работу с Telegram API в отдельный сервис. Стороннее приложение больше не должно само подключаться к Telegram, держать Telegram token, запускать polling или вызывать методы Telegram API. Вместо этого оно принимает нормализованные события от `telegram-server` и отправляет команды на отправку сообщений через HTTP API `telegram-server`.

## Что такое telegram-server

`telegram-server` — транспортный шлюз между Telegram и вашими приложениями.

Он делает следующее:

- подключается к Telegram по token бота;
- получает Telegram updates;
- преобразует updates в простой JSON-формат;
- отправляет этот JSON в ваше приложение;
- принимает от вашего приложения команды на отправку сообщений;
- ставит исходящие команды в очередь;
- отправляет сообщения, фото, документы и кнопки через Telegram API.

Он не содержит бизнес-логику конкретного бота. Бизнес-логика остаётся в вашем приложении.

## Как меняется архитектура приложения

До интеграции:

```text
Telegram
  -> ваше приложение
  -> бизнес-логика
  -> Telegram API
```

После интеграции:

```text
Telegram
  -> telegram-server
  -> ваше приложение
  -> telegram-server
  -> Telegram API
```

Ваше приложение становится HTTP-сервисом бизнес-логики:

- принимает входящие события от `telegram-server`;
- обрабатывает их своими существующими сценариями;
- отправляет ответы в `telegram-server` через HTTP.

## Что нужно получить от администратора telegram-server

Разработчику стороннего приложения не нужно управлять `telegram-server`, но ему нужны параметры подключения.

Попросите у администратора:

- `bot_id` — идентификатор вашего бота в `telegram-server`, например `support`;
- `telegram_server_base_url` — базовый URL `telegram-server`, например `http://10.10.0.5:8080`;
- `auth_token` — значение header `Authorization`, например `Bearer support-secret`;
- какой URL вашего приложения должен быть указан как endpoint входящих событий;
- максимальный размер запроса, если вы отправляете большие payload.

Telegram token вашему приложению не нужен. Он должен храниться только на стороне `telegram-server`.

## Что такое URL приложения для входящих событий

В конфигурации `telegram-server` для каждого бота есть поле `external.url`.

Это URL вашего приложения, на который `telegram-server` будет отправлять входящие события из Telegram.

Пример:

```json
{
  "external": {
    "url": "http://10.10.0.20:9001/telegram/events",
    "auth_token": "Bearer support-secret",
    "timeout_seconds": 10
  }
}
```

`external.url` не обязан быть публичным доменом. Это просто HTTP URL, доступный из процесса `telegram-server`.

Допустимые варианты:

- `http://10.10.0.20:9001/telegram/events` — IP и порт сервера, где работает ваше приложение;
- `http://support-app:9001/telegram/events` — DNS-имя или service name внутри Docker Compose, Kubernetes или другой внутренней сети;
- `http://support.internal:9001/telegram/events` — внутренний домен;
- `http://localhost:9001/telegram/events` — только если `telegram-server` и ваше приложение запущены на одной машине или в одном сетевом namespace.

Домен регистрировать не обязательно. Главное условие: `telegram-server` должен иметь возможность выполнить HTTP `POST` на этот URL.

Если ваше приложение сейчас не имеет HTTP endpoint, его нужно добавить. Это не Telegram webhook. Это обычный внутренний HTTP endpoint, который принимает события от `telegram-server`.

## Эндпоинты telegram-server

Стороннее приложение обычно использует только один endpoint `telegram-server`.

### `POST /telegram/send`

Отправка сообщения через Telegram-бота.

URL запроса строится из выданного базового URL:

```text
{telegram_server_base_url}/telegram/send
```

Если администратор выдал:

```text
http://10.10.0.5:8080
```

то полный URL:

```text
http://10.10.0.5:8080/telegram/send
```

Если в примерах встречается `http://telegram-server:8080`, это не специальное значение. Это пример DNS-имени сервиса во внутренней сети. В вашем окружении вместо него может быть IP, домен, Kubernetes service name, Docker Compose service name или `localhost`.

Headers:

```http
Content-Type: application/json
Authorization: Bearer support-secret
```

Успешный ответ:

```http
HTTP/1.1 202 Accepted
```

```json
{
  "status": "accepted"
}
```

`202 Accepted` означает, что команда принята в очередь `telegram-server`. Это не гарантия, что Telegram уже доставил сообщение.

### `GET /health`

Проверка, что процесс `telegram-server` жив.

Стороннему приложению обычно не нужно вызывать этот endpoint в бизнес-логике. Он полезен для диагностики и мониторинга.

### `GET /ready`

Проверка, что `telegram-server` готов обслуживать запросы.

Стороннему приложению можно использовать этот endpoint при старте или в диагностике, но нельзя строить бизнес-логику только на нём.

## Что должно быть настроено в стороннем приложении

В конфигурации вашего приложения должны появиться параметры:

```json
{
  "telegram_server": {
    "base_url": "http://10.10.0.5:8080",
    "auth_token": "Bearer support-secret",
    "bot_id": "support",
    "timeout_seconds": 5
  },
  "http": {
    "addr": ":9001",
    "telegram_events_path": "/telegram/events"
  }
}
```

Названия полей могут быть любыми. Важно, чтобы приложение знало:

- куда отправлять команды на отправку сообщений;
- какой `Authorization` header использовать;
- какой `bot_id` указывать в командах;
- на каком адресе и пути принимать входящие события.

Если приложение обслуживает несколько ботов, храните настройки по каждому `bot_id`.

## Endpoint входящих событий в стороннем приложении

Ваше приложение должно поднять HTTP endpoint, который принимает `POST` от `telegram-server`.

Пример endpoint:

```text
POST /telegram/events
```

`telegram-server` будет отправлять туда JSON события.

Headers входящего запроса:

```http
Content-Type: application/json
Authorization: Bearer support-secret
```

Ваше приложение должно:

- проверить `Authorization`;
- прочитать JSON;
- обработать событие бизнес-логикой;
- вернуть HTTP status `2xx`, если событие принято.

Тело ответа сейчас игнорируется.

Если ваше приложение вернёт non-2xx или не ответит до таймаута, `telegram-server` залогирует ошибку. Повторная доставка входящих событий сейчас не выполняется.

## Формат входящего события

Базовая структура:

```json
{
  "bot_id": "support",
  "update_id": 10001,
  "type": "message",
  "from": {},
  "chat": {},
  "message": {},
  "callback": {},
  "received": "2026-04-26T01:00:01Z"
}
```

Поля:

- `bot_id` — бот, которому принадлежит событие;
- `update_id` — Telegram update ID;
- `type` — тип события: `message`, `command`, `callback`;
- `from` — пользователь Telegram;
- `chat` — чат Telegram;
- `message` — сообщение, если событие связано с сообщением;
- `callback` — callback, если событие пришло от inline-кнопки;
- `received` — время нормализации события на стороне `telegram-server`.

В своём приложении не обязательно описывать все поля. Можно описать только те, которые реально нужны бизнес-логике.

### Текстовое сообщение

```json
{
  "bot_id": "support",
  "update_id": 10001,
  "type": "message",
  "from": {
    "id": 123456789,
    "username": "ivan",
    "display_name": "Ivan Petrov"
  },
  "chat": {
    "id": 123456789,
    "type": "private",
    "username": "ivan"
  },
  "message": {
    "id": 55,
    "date": "2026-04-26T01:00:00Z",
    "text": "hello"
  },
  "received": "2026-04-26T01:00:01Z"
}
```

### Команда

```json
{
  "bot_id": "support",
  "update_id": 10002,
  "type": "command",
  "chat": {
    "id": 123456789,
    "type": "private"
  },
  "message": {
    "id": 56,
    "date": "2026-04-26T01:01:00Z",
    "text": "/start",
    "command": "start"
  },
  "received": "2026-04-26T01:01:01Z"
}
```

### Callback от inline-кнопки

```json
{
  "bot_id": "support",
  "update_id": 10003,
  "type": "callback",
  "from": {
    "id": 123456789,
    "username": "ivan"
  },
  "chat": {
    "id": 123456789,
    "type": "private"
  },
  "callback": {
    "id": "callback-query-id",
    "data": "ticket:123:close",
    "message_id": 57
  },
  "received": "2026-04-26T01:02:01Z"
}
```

### Фото

```json
{
  "bot_id": "support",
  "update_id": 10004,
  "type": "message",
  "chat": {
    "id": 123456789,
    "type": "private"
  },
  "message": {
    "id": 58,
    "date": "2026-04-26T01:03:00Z",
    "caption": "photo caption",
    "photo": [
      {
        "file_id": "photo-small-file-id",
        "file_unique_id": "photo-small-unique-id",
        "width": 90,
        "height": 90,
        "file_size": 1024
      },
      {
        "file_id": "photo-large-file-id",
        "file_unique_id": "photo-large-unique-id",
        "width": 1280,
        "height": 960,
        "file_size": 120000
      }
    ]
  },
  "received": "2026-04-26T01:03:01Z"
}
```

`telegram-server` не скачивает файлы из Telegram. Приложение получает только metadata и `file_id`.

## Отправка сообщений из Go-приложения

`curl` может использоваться только для ручной проверки API. В приложениях на Go нужно использовать обычный HTTP-клиент.

### Минимальные структуры

```go
type TelegramServerConfig struct {
    BaseURL        string
    AuthToken      string
    BotID          string
    TimeoutSeconds int
}

type SendCommand struct {
    BotID            string       `json:"bot_id"`
    ChatID           int64        `json:"chat_id"`
    Text             string       `json:"text,omitempty"`
    Photo            string       `json:"photo,omitempty"`
    Document         string       `json:"document,omitempty"`
    Caption          string       `json:"caption,omitempty"`
    ParseMode        string       `json:"parse_mode,omitempty"`
    ReplyToMessageID int          `json:"reply_to_message_id,omitempty"`
    ReplyMarkup      *ReplyMarkup `json:"reply_markup,omitempty"`
}

type ReplyMarkup struct {
    InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
    ReplyKeyboard  [][]ReplyKeyboardButton  `json:"reply_keyboard,omitempty"`
    ResizeKeyboard bool                     `json:"resize_keyboard,omitempty"`
    OneTimeKeyboard bool                    `json:"one_time_keyboard,omitempty"`
}

type InlineKeyboardButton struct {
    Text         string `json:"text"`
    CallbackData string `json:"callback_data,omitempty"`
    URL          string `json:"url,omitempty"`
}

type ReplyKeyboardButton struct {
    Text string `json:"text"`
}
```

### HTTP-клиент отправки

```go
type TelegramServerClient struct {
    baseURL   string
    authToken string
    botID     string
    client    *http.Client
}

func NewTelegramServerClient(cfg TelegramServerConfig) *TelegramServerClient {
    timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
    if timeout <= 0 {
        timeout = 5 * time.Second
    }

    return &TelegramServerClient{
        baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
        authToken: cfg.AuthToken,
        botID:     cfg.BotID,
        client:    &http.Client{Timeout: timeout},
    }
}

func (c *TelegramServerClient) Send(ctx context.Context, command SendCommand) error {
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
```

### Отправка текста

```go
err := telegramClient.Send(ctx, SendCommand{
    ChatID: chatID,
    Text:   "Ваш запрос принят",
})
```

### Ответ на конкретное сообщение

```go
err := telegramClient.Send(ctx, SendCommand{
    ChatID:           event.Chat.ID,
    Text:             "Ответ на ваше сообщение",
    ReplyToMessageID: event.Message.ID,
})
```

### Отправка фото

```go
err := telegramClient.Send(ctx, SendCommand{
    ChatID:  chatID,
    Photo:   "photo-file-id",
    Caption: "Фото по вашему запросу",
})
```

`Photo` может быть Telegram `file_id` или `http/https` URL. Локальный путь к файлу не поддерживается.

### Отправка документа

```go
err := telegramClient.Send(ctx, SendCommand{
    ChatID:   chatID,
    Document: "https://example.com/report.pdf",
    Caption:  "Отчёт",
})
```

`Document` может быть Telegram `file_id` или `http/https` URL. Multipart upload локальных файлов не поддерживается.

### Inline-кнопки

```go
err := telegramClient.Send(ctx, SendCommand{
    ChatID: chatID,
    Text:   "Выберите действие",
    ReplyMarkup: &ReplyMarkup{
        InlineKeyboard: [][]InlineKeyboardButton{
            {
                {
                    Text:         "Закрыть заявку",
                    CallbackData: "ticket:123:close",
                },
                {
                    Text: "Открыть",
                    URL:  "https://example.com/tickets/123",
                },
            },
        },
    },
})
```

## Правила исходящей команды

Команда должна содержать:

- `bot_id`;
- `chat_id`;
- ровно один payload: `text`, `photo` или `document`.

В Go-клиенте выше `bot_id` подставляется автоматически из конфигурации.

Нельзя отправлять одновременно, например, `text` и `photo`.

Если указан `reply_markup`, можно использовать только один тип клавиатуры:

- `inline_keyboard`;
- `reply_keyboard`.

Для inline-кнопки поддерживается только один из вариантов:

- `callback_data`;
- `url`.

## Ошибки отправки

`telegram-server` возвращает JSON-ошибки:

```json
{
  "error": "invalid_command"
}
```

Основные статусы:

- `400 invalid_json` — некорректный JSON.
- `400 invalid_command` — команда не прошла валидацию.
- `401 unauthorized` — неверный `Authorization`.
- `404 unknown_bot` — неизвестный `bot_id`.
- `409 bot_not_started` — бот есть в конфигурации, но не запущен.
- `413 request_too_large` — тело запроса слишком большое.
- `503 send_queue_closed` — очередь бота закрыта.
- `503 send_queue_full` — очередь бота заполнена.
- `502 telegram_send_failed` — ошибка отправки в Telegram.

Практическая обработка:

- `400` — ошибка в коде приложения или данных команды, нужно исправлять payload.
- `401` — неверная настройка auth token.
- `404` — неверный `bot_id` или бот не подключён к `telegram-server`.
- `409` — инфраструктурная проблема на стороне `telegram-server`.
- `503` — временная проблема очереди, можно повторить позже с осторожностью.
- `502` — Telegram API вернул ошибку или недоступен.

## Как переделать существующее Go-приложение

### Шаг 1. Оставить бизнес-логику

Код сценариев, проверок, работы с БД и принятия решений должен остаться в приложении.

Например, если раньше было:

```go
func handleStart(chatID int64) {
    // бизнес-логика
}
```

этот обработчик не нужно переносить в `telegram-server`.

### Шаг 2. Убрать Telegram polling или webhook

Ваше приложение больше не должно получать updates напрямую из Telegram.

Уберите или отключите:

- создание Telegram client по token;
- polling loop;
- Telegram webhook endpoint;
- хранение Telegram token.

### Шаг 3. Добавить endpoint для событий от telegram-server

Упрощённый пример:

```go
func TelegramEventsHandler(authToken string, telegramClient *TelegramServerClient) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }
        if r.Header.Get("Authorization") != authToken {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }

        var event Event
        if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        if err := handleTelegramEvent(r.Context(), event, telegramClient); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusAccepted)
    }
}
```

`Event` можно описать в вашем приложении минимально:

```go
type Event struct {
    BotID    string   `json:"bot_id"`
    Type     string   `json:"type"`
    Chat     *Chat    `json:"chat,omitempty"`
    Message  *Message `json:"message,omitempty"`
    Callback *Callback `json:"callback,omitempty"`
}

type Chat struct {
    ID int64 `json:"id"`
}

type Message struct {
    ID      int    `json:"id"`
    Text    string `json:"text,omitempty"`
    Command string `json:"command,omitempty"`
}

type Callback struct {
    Data      string `json:"data,omitempty"`
    MessageID int    `json:"message_id,omitempty"`
}
```

### Шаг 4. Заменить прямые вызовы Telegram API

Было:

```go
msg := tgbotapi.NewMessage(chatID, "Ваш запрос принят")
_, err := bot.Send(msg)
```

Стало:

```go
err := telegramClient.Send(ctx, SendCommand{
    ChatID: chatID,
    Text:   "Ваш запрос принят",
})
```

### Шаг 5. Проверить сетевую доступность

Нужно проверить два направления:

1. `telegram-server` должен достучаться до endpoint вашего приложения.
2. Ваше приложение должно достучаться до `{telegram_server_base_url}/telegram/send`.

Если приложения находятся:

- на одной машине — можно использовать `localhost`, но только если оба процесса видят один и тот же localhost;
- на разных серверах — используйте IP, внутренний DNS или домен;
- в Docker Compose — используйте service name из compose-сети;
- в Kubernetes — используйте Kubernetes Service DNS;
- в закрытой сети — используйте VPN, private IP или внутренний балансировщик.

Публичный домен нужен только если между сервисами нет общей приватной сети и вы сознательно открываете endpoint наружу.

## Ручная проверка через curl

`curl` не является способом интеграции приложения. Это только инструмент для ручной диагностики.

Пример проверки отправки:

```bash
curl -X POST http://10.10.0.5:8080/telegram/send \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer support-secret' \
  -d '{
    "bot_id": "support",
    "chat_id": 123456789,
    "text": "Проверка отправки"
  }'
```

В коде Go-приложения используйте `net/http`, как показано выше.

## Эксплуатационные ограничения

- Успешный ответ `202 Accepted` означает постановку в очередь, а не доставку в Telegram.
- Очередь исходящих сообщений хранится в памяти `telegram-server`.
- При аварийном рестарте `telegram-server` часть принятых команд может быть потеряна.
- Входящие события от Telegram во внешнее приложение сейчас доставляются без retry.
- Файлы из Telegram не скачиваются и не передаются во внешнее приложение.
- Для отправки фото и документов поддерживаются только Telegram `file_id` и `http/https` URL.
- Локальный multipart upload файлов не поддерживается.
- Поддержаны только text, photo, document, inline keyboard и простая reply keyboard.

## Краткий чеклист интеграции

1. Получить у администратора `bot_id`, `telegram_server_base_url`, `auth_token`.
2. Добавить в приложение конфигурацию подключения к `telegram-server`.
3. Поднять HTTP endpoint для входящих событий.
4. Сообщить администратору URL этого endpoint.
5. Убрать прямое получение Telegram updates.
6. Заменить прямые вызовы Telegram API на вызовы `POST /telegram/send`.
7. Проверить сетевую доступность в обе стороны.
8. Протестировать сценарии: message, command, callback, отправка ответа.
