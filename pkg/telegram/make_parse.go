package telegram

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"tg_bot_minenergo_ip/pkg/parser"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) LoadIP() {

	for {
		var wg sync.WaitGroup
		wg.Add(len(b.config.IP))
		start_time := time.Now().UnixMilli()
		for ip := range b.config.IP {
			go func(ip string) {
				new_report, err := parser.Parse(b.config.IP[ip].First_entry, ip)
				if err != nil {
					log.Printf("Ошибка парсинга: %s", err)
				}
				old_report, err := b.base.Get(ip, ip)
				if err != nil {
					log.Printf("Ошибка чтения из БД при парсинге новости: %s", err)
				}
				if (new_report != old_report) && (new_report != "ERROR") {
					log.Printf("Обнаружена новая запись ИП %s: %s", b.config.IP[ip].Name, new_report)
					b.make_notify(ip, new_report)
					err = b.base.Save(ip, new_report, ip)
					if err != nil {
						log.Printf("Ошибка сохранения в БД новой новости по ИП: %s", err)
					}
				}
				wg.Done()
			}(ip)
		}
		wg.Wait()
		end_time := time.Now().UnixMilli()
		delta := end_time - start_time
		log.Printf("Выполнен парсинг сайта МЭ за время %v милисекунд", delta)
		time.Sleep(time.Minute * 10)
	}
}

func (b *Bot) make_notify(ip string, news string) {
	users, err := b.base.GetAll(ip)
	if err != nil {
		log.Printf("Ошибка чтения из БД данных о подписчиках")
	}

	for id, key := range users {
		if key == "subscride" {
			msg_text := fmt.Sprintf("*%s*\nРазмещена новая запись:\n%s\n[https://minenergo.gov.ru/node/%s]", b.config.IP[ip].Name, news, ip)
			id_int64, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Println(err)
			}
			msg := tgbotapi.NewMessage(id_int64, msg_text)
			msg.ParseMode = "Markdown"
			_, err = b.bot.Send(msg)
			if err != nil {
				log.Println(err)
			}

		}
	}
}
