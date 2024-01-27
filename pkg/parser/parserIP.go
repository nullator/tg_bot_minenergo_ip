package parser

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"tg_bot_minenergo_ip/pkg/logger"

	"github.com/PuerkitoBio/goquery"
)

func Start(ctx context.Context, first_entry string, ip_code string,
	logger *logger.Logger) (string, error) {
	baseUrl := fmt.Sprintf("https://minenergo.gov.ru/node/%s", ip_code)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		logger.Errorf("Не удалось сформировать request: %s", err)
		return "ERROR", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
	req.Header.Add("Referer", baseUrl)
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		logger.Warnf("Не удалось выполнить запрос к серверу: %s", err)
		return "ERROR", err
	}
	if res.StatusCode != 200 {
		logger.Errorf("Ошибка запроса к серверу: (code %d) %s", res.StatusCode, err)
		return "ERROR", err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		logger.Warnf("Не удалось распарсить страницу: %s", err)
		return "ERROR", err
	}
	defer res.Body.Close()

	m := make(map[int]string)
	doc.Find(".file-title").Each(func(i int, s *goquery.Selection) {
		m[i] = s.Text()
	})
	gap, err := getGup(first_entry, ip_code, m)
	if err != nil {
		logger.Errorf("Ошибка получения первой записи (gap) ИП: %s", err)
		return "ERROR", err
	}

	return m[gap], nil

}

func GetIP(ctx context.Context, ip_code string, logger *logger.Logger) (string, error) {
	url := fmt.Sprintf("https://minenergo.gov.ru/api/v1/?action=organizations.getItemDetail&lang=ru&code=%s", ip_code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("Не удалось сформировать request: %s", err)
		return "ERROR", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf("Не удалось выполнить запрос к серверу: %s", err)
		return "ERROR", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("unexpected status: %s", resp.Status)
		return "ERROR", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading body: %s", err)
		return "ERROR", err
	}

	// save body to file
	file := fmt.Sprintf("%s.json", ip_code)
	err = os.WriteFile(file, body, 0644)
	if err != nil {
		slog.Error("error writing file: %s", err)
		return "ERROR", err
	}

	return "OK", nil

}

func getGup(first_entry string, ip_code string, m map[int]string) (int, error) {
	for i := 0; i < len(m); i++ {
		if m[i] == first_entry {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("не удалось распарсить код %s", ip_code)
}
