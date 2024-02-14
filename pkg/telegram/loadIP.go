package telegram

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"sync"
	"tg_bot_minenergo_ip/pkg/models"
	"tg_bot_minenergo_ip/pkg/parser"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var wg sync.WaitGroup

func (b *Bot) LoadIP(ctx context.Context) {
	c := make(chan string, len(b.config.IP))

	for {
		select {
		case <-ctx.Done():
			slog.Info("Завершение функции LoadIP")
			return
		default:
			start_time := time.Now().UnixMilli()
			wg.Add(len(b.config.IP))
			for ip := range b.config.IP {
				c <- ip
			}

			// Переменная для хранения предупреждений
			warn := models.LogCollector{}

			// Есть сомнения что Минэнерго предоставляет корретные данны в json формате, поэтому для отладки первое время планирую считать количество записей ИП
			err := b.base.Save("all", "0", "count")
			if err != nil {
				slog.Error("Ошибка сохранения в БД счетчика записей ИП",
					slog.String("error", err.Error()))
			}

			go b.startParse_v2(ctx, c, &warn)
			wg.Wait()

			// Проверка наличия warning
			if len(warn.Get()) > 0 {
				slog.Warn("Обнаружены предупреждения",
					slog.Any("warn", warn.Get()))
			}

			end_time := time.Now().UnixMilli()
			delta := end_time - start_time

			count_str, err := b.base.Get("all", "count")
			if err != nil {
				slog.Error("Ошибка чтения из БД счетчика записей ИП",
					slog.String("error", err.Error()))
			}
			count, err := strconv.Atoi(count_str)
			if err != nil {
				slog.Error("Ошибка преобразования счетчика записей ИП",
					slog.String("error", err.Error()))
			}

			format_delta := fmt.Sprintf("%v сек.", float64(delta)/1000)
			slog.Info("Выполнен парсинг сайта МЭ",
				slog.String("time", fmt.Sprintf("%v", format_delta)),
				slog.String("count", fmt.Sprintf("%v", count)))

			// sleep 20 min
			select {
			case <-ctx.Done():
				slog.Info("Остановка функции LoadIP")
				return
			case <-time.After(20 * time.Minute):
			}

		}
	}
}

func (b *Bot) startParse_v2(ctx context.Context, c chan string, w *models.LogCollector) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Остановка функции startParse_v2")
			return
		default:
			ip := <-c
			func(ip string) {
				defer wg.Done()

				ctx_to, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				code := b.config.IP[ip].Code
				data, err := parser.GetIP(ctx_to, code, w)
				if len(data) == 0 {
					w.Add(fmt.Sprintf("Получена пустая запись ИП %s", b.config.IP[ip].Name))
					return
				}
				last_report := data[0]

				new_report := data[0].Dsc
				new_report = html.UnescapeString(new_report)
				if err != nil {
					w.Add(fmt.Sprintf("Не удалось получить данные по api для ИП %s", b.config.IP[ip].Name))
				}

				// Проверяем изменилось ли количество записей ИП (для отладки)
				old_count, err := b.base.Get(ip, "count")
				if err != nil {
					slog.Error("Ошибка чтения из БД при парсинге новости",
						slog.String("error", err.Error()))
				}
				new_count := strconv.Itoa(len(data))
				if new_count != old_count {
					w.Add(fmt.Sprintf("Обнаружено изменение количества записей ИП %s, old: %s, new: %s", b.config.IP[ip].Name, old_count, new_count))
					err = b.base.Save(ip, new_count, "count")
					if err != nil {
						slog.Error("Ошибка сохранения в БД нового количества записей ИП",
							slog.String("error", err.Error()))
					}
				}

				count_str, err := b.base.Get("all", "count")
				if err != nil {
					slog.Error("Ошибка чтения из БД счетчика записей ИП", slog.String("error", err.Error()))
				}
				count, err := strconv.Atoi(count_str)
				if err != nil {
					slog.Error("Ошибка преобразования счетчика записей ИП", slog.String("error", err.Error()))
				}
				count += len(data)
				count_str = strconv.Itoa(count)
				err = b.base.Save("all", count_str, "count")
				if err != nil {
					slog.Error("Ошибка сохранения в БД счетчика записей ИП", slog.String("error", err.Error()))
				}

				old_report, err := b.base.Get(ip, ip)
				if err != nil {
					slog.Error("Ошибка чтения из БД старой записи по ИП", slog.String("error", err.Error()))
				}

				if (new_report != old_report) && (new_report != "ERROR") {
					slog.Info("Обнаружена новая запись ИП",
						slog.String("ip", b.config.IP[ip].Name),
						slog.String("new_report", new_report))

					// Игнорировать изменение последовательности записей ИП
					if new_count == old_count {
						w.Add(fmt.Sprintf("Количество записей ИП %s не поменялось, рассылка не выполняется", b.config.IP[ip].Name))
						return
					}

					b.make_notify(ip, b.config.IP[ip].Name, new_report, last_report.Src)
					err = b.base.Save(ip, new_report, ip)
					if err != nil {
						slog.Error("Ошибка сохранения в БД новой новости по ИП", slog.String("error", err.Error()))
					}
				}
			}(ip)
		}
	}

}

func (b *Bot) make_notify(ip_code string, ip_name string, news string, src string) {
	users, err := b.base.GetAll(ip_code)
	if err != nil {
		slog.Error("Ошибка чтения из БД данных о подписчиках", slog.String("error", err.Error()))
	}

	for id, key := range users {
		if key == "subscride" {
			msg_text := fmt.Sprintf("*%s*\nРазмещена новая запись:\n%s\n[https://minenergo.gov.ru%s]", ip_name, news, src)
			id_int64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				slog.Error("Ошибка преобразования id пользователя", slog.String("error", err.Error()))
			}
			msg := tgbotapi.NewMessage(id_int64, msg_text)
			msg.ParseMode = tgbotapi.ModeMarkdown
			_, err = b.bot.Send(msg)
			if err != nil {
				slog.Error("Не удалось отправить уведомление", slog.String("error", err.Error()))
			}

			time.Sleep(time.Millisecond * 300)

		}
	}
}
