package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"tg_bot_minenergo_ip/pkg/logger"
	"tg_bot_minenergo_ip/pkg/models"

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
	baseURL := "https://minenergo.gov.ru/api/v1/"
	params := url.Values{}
	params.Add("action", "organizations.getItemDetail")
	params.Add("lang", "ru")
	params.Add("code", "pao_federalnaya_setevaya_kompaniya_rosseti")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		logger.Errorf("Не удалось сформировать request: %s", err)
		return "ERROR", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537")

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
	// file := fmt.Sprintf("%s.json", ip_code)
	// err = os.WriteFile(file, body, 0644)
	// if err != nil {
	// 	slog.Error("error writing file: %s", err)
	// 	return "ERROR", err
	// }

	var IPdata models.FullData
	err = json.Unmarshal([]byte(body), &IPdata)
	if err != nil {
		logger.Errorf("Ошибка распаковки json в структуру ИП - %s", err.Error())
		return "ERROR", err
	}

	rec := IPdata.Docs[1].Recods[0].Dsc

	return rec, nil

}

func getGup(first_entry string, ip_code string, m map[int]string) (int, error) {
	for i := 0; i < len(m); i++ {
		if m[i] == first_entry {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("не удалось распарсить код %s", ip_code)
}
