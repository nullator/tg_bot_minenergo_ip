package parser

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func Start(ctx context.Context, first_entry string, ip_code string) (string, error) {
	baseUrl := fmt.Sprintf("https://minenergo.gov.ru/node/%s", ip_code)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		log.Printf("Не удалось сформировать request: %s", err)
		return "ERROR", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	req.Header.Add("Referer", baseUrl)
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		log.Printf("Не удалось выполнить запрос к серверу: %s", err)
		return "ERROR", err
	}
	if res.StatusCode != 200 {
		log.Printf("Ошибка запроса к серверу: (code %d) %s", res.StatusCode, err)
		return "ERROR", err
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
	gap, err := getGup(first_entry, ip_code, m)
	if err != nil {
		log.Printf("Ошибка получения первой записи (gap) ИП: %s", err)
	}

	return m[gap], nil

}

func getGup(first_entry string, ip_code string, m map[int]string) (int, error) {
	for i := 0; i < len(m); i++ {
		if m[i] == first_entry {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("не удалось распарсить код %s", ip_code)
}
