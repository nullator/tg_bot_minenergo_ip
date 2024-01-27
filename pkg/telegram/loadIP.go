package telegram

import (
	"context"
	"fmt"
	"strconv"
	"sync"
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
			b.logger.Info("Завершение функции LoadIP")
			return
		default:
			start_time := time.Now().UnixMilli()
			wg.Add(len(b.config.IP))
			for ip := range b.config.IP {
				c <- ip
			}

			go b.startParse_v2(ctx, c)

			wg.Wait()
			end_time := time.Now().UnixMilli()
			delta := end_time - start_time
			b.logger.Infof("Выполнен парсинг сайта МЭ за время %v милисекунд", delta)

			// sleep 20 min
			select {
			case <-ctx.Done():
				b.logger.Info("Завершение функции LoadIP")
				return
			case <-time.After(20 * time.Minute):
			}

		}
	}
}

// func (b *Bot) startParse(ctx_c context.Context, c chan string) {
// 	for {
// 		select {
// 		case <-ctx_c.Done():
// 			return
// 		default:
// 			ip := <-c
// 			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 			defer cancel()
// 			new_report, err := parser.Start(ctx, b.config.IP[ip].First_entry, ip, b.logger)
// 			if err != nil {
// 				b.logger.Warnf("Не удалось распарсить страницу: %s", err.Error())
// 			}
// 			old_report, err := b.base.Get(ip, ip)
// 			if err != nil {
// 				b.logger.Errorf("Ошибка чтения из БД при парсинге новости: %s", err.Error())
// 			}
// 			if (new_report != old_report) && (new_report != "ERROR") {
// 				b.logger.Infof("Обнаружена новая запись ИП %s: %s", b.config.IP[ip].Name, new_report)
// 				b.make_notify(ip, new_report)
// 				err = b.base.Save(ip, new_report, ip)
// 				if err != nil {
// 					b.logger.Errorf("Ошибка сохранения в БД новой новости по ИП: %s", err.Error())
// 				}
// 			}
// 			wg.Done()
// 		}
// 	}

// }

func (b *Bot) startParse_v2(ctx context.Context, c chan string) {
	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Остановка функции startParse_v2")
			return
		default:
			ip := <-c
			ctx_to, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			data, err := parser.GetIP(ctx_to, ip, b.logger)

			new_report := data.Dsc
			if err != nil {
				b.logger.Warnf("Не удалось распарсить запись: %s", err.Error())
			}

			ip_db_code := b.config.IP[ip].Old_code
			old_report, err := b.base.Get(ip_db_code, ip_db_code)
			if err != nil {
				b.logger.Errorf("Ошибка чтения из БД при парсинге новости: %s", err.Error())
			}

			if (new_report != old_report) && (new_report != "ERROR") {
				b.logger.Infof("Обнаружена новая запись ИП %s: %s", b.config.IP[ip].Name, new_report)
				b.make_notify(ip_db_code, b.config.IP[ip].Name, new_report, data.Src)
				err = b.base.Save(ip_db_code, new_report, ip_db_code)
				if err != nil {
					b.logger.Errorf("Ошибка сохранения в БД новой новости по ИП: %s", err.Error())
				}
			}

			wg.Done()
		}
	}

}

func (b *Bot) make_notify(ip_code string, ip_name string, news string, src string) {
	users, err := b.base.GetAll(ip_code)
	if err != nil {
		b.logger.Errorf("Ошибка чтения из БД данных о подписчиках - %s", err.Error())
	}

	for id, key := range users {
		if key == "subscride" {
			msg_text := fmt.Sprintf("*%s*\nРазмещена новая запись:\n%s\n[https://minenergo.gov.ru%s]", ip_name, news, src)
			id_int64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				b.logger.Error(err)
			}
			msg := tgbotapi.NewMessage(id_int64, msg_text)
			msg.ParseMode = tgbotapi.ModeMarkdown
			_, err = b.bot.Send(msg)
			if err != nil {
				b.logger.Error(err)
			}

			time.Sleep(time.Millisecond * 300)

		}
	}
}
