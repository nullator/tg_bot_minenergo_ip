package parser_ip

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func Parse(ip_code string) (string, error) {
	baseUrl := fmt.Sprintf("https://minenergo.gov.ru/node/%s", ip_code)
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		log.Printf("Не удалось сформировать request: %s", err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	req.Header.Add("Referer", baseUrl)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Не удалось выполнить запрос к серверу: %s", err)
		return "ERROR", err
	}
	if res.StatusCode != 200 {
		log.Printf("Ошибка запроса к серверу: (code %d) %s", res.StatusCode, err)
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Printf("Не удалось распарсить страницу: %s", err)
		return "ERROR", err
	}
	defer res.Body.Close()

	m := make(map[int]string)
	doc.Find(".file-title").Each(func(i int, s *goquery.Selection) {
		m[i] = s.Text()
	})
	gap, err := getGup(ip_code, m)
	if err != nil {
		log.Printf("Ошибка получения первой записи (gap) ИП: %s", err)
	}
	return m[gap], nil

}

// rosseti_volga   = "4190"
// fsk_ees         = "4174"
// rosseti_cip     = "4178"
// rosseti_yug     = "4191"
// rosseti_centr   = "4192"
// rosseti_sibir   = "4185"
// rosseti_ural    = "4189"
// rosseti_sev_zap = "4193"
// rosseti_sev_kav = "4177"
// rusgydro        = "4195"
// drsk            = "4217"
// krea            = "4224"
// setevaya_komp   = "4361"

func getGup(ip_code string, m map[int]string) (int, error) {
	switch ip_code {
	case "4190":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность ПАО МРСК Волги за 2016 год (Дата публикации: 30.03.2017)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4190")
	case "4174":
		for i := 0; i < len(m); i++ {
			if m[i] == "ПАО ФСК ЕЭС от 31.03.2016 (Дата публикации: 07.04.2016)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4174")
	case "4178":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность ПАО «МРСК Центра и Приволжья» за 2017 год (Дата публикации: 28.03.2018)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4178")

	case "4191":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность ПАО \"МРСК Юга\" за 2018 год (Дата публикации: 29.03.2019)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4191")

	case "4192":
		for i := 0; i < len(m); i++ {
			if m[i] == "ПАО МРСК Центра от 13.04.2016 (Дата публикации: 13.04.2016)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4192")

	case "4185":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность ПАО \"МРСК Сибири\" за 2017 год (Дата публикации: 30.03.2018)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4185")

	case "4189":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность ОАО \"МРСК Урала\" за 2017 год (Дата публикации: 31.03.2018)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4189")

	case "4193":
		for i := 0; i < len(m); i++ {
			if m[i] == "Годовая финансовая отчетность и аудиторское заключение за 2015 год (Дата публикации: 14.03.2018)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4193")

	case "4177":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность ПАО \"МРСК Северного Кавказа\" за 2017 год (Дата публикации: 29.03.2018)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4177")

	case "4195":
		for i := 0; i < len(m); i++ {
			if m[i] == "ПАО «РусГидро» от 31.03.2016" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4195")

	case "4217":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность АО ДРСК за 2016 год от 23.03.2017 (Дата публикации: 23.03.2017)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4217")

	case "4224":
		for i := 0; i < len(m); i++ {
			if m[i] == "Бухгалтерская отчетность за 2016 год с аудиторским заключением" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4224")

	case "4361":
		for i := 0; i < len(m); i++ {
			if m[i] == "ОАО Сетевая компания от 12.04.2016 (Дата публикации: 12.04.2016)" {
				return i + 1, nil
			}
		}
		return 0, errors.New("не удалось распарсить 4361")

	}
	return 0, errors.New("не удалось распарсить 4361")
}
