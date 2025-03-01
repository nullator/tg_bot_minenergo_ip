package parser

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"tg_bot_minenergo_ip/pkg/models"
	"time"
)

// func Start(ctx context.Context, first_entry string, ip_code string) (string, error) {
// 	baseUrl := fmt.Sprintf("https://minenergo.gov.ru/node/%s", ip_code)
// 	client := &http.Client{}
// 	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
// 	if err != nil {
// 		slog.Error("Не удалось сформировать request: %s", slog.String("error", err.Error()))
// 		return "ERROR", err
// 	}
// 	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")
// 	req.Header.Add("Referer", baseUrl)
// 	req = req.WithContext(ctx)

// 	res, err := client.Do(req)
// 	if err != nil {
// 		slog.Warn("Не удалось выполнить запрос к серверу: %s", slog.String("error", err.Error()))
// 		return "ERROR", err
// 	}
// 	if res.StatusCode != 200 {
// 		slog.Error("Ошибка запроса к серверу: (code %d) %s",
// 			slog.Int("code", res.StatusCode),
// 			slog.String("status", res.Status))
// 		return "ERROR", err
// 	}
// 	doc, err := goquery.NewDocumentFromResponse(res)
// 	if err != nil {
// 		logger.Warnf("Не удалось распарсить страницу: %s", err)
// 		return "ERROR", err
// 	}
// 	defer res.Body.Close()

// 	m := make(map[int]string)
// 	doc.Find(".file-title").Each(func(i int, s *goquery.Selection) {
// 		m[i] = s.Text()
// 	})
// 	gap, err := getGup(first_entry, ip_code, m)
// 	if err != nil {
// 		logger.Errorf("Ошибка получения первой записи (gap) ИП: %s", err)
// 		return "ERROR", err
// 	}

// 	return m[gap], nil

// }

// Выполняет запрос к api Минэнерго и возвращает список записей ИП
func GetIP(ctx context.Context, ip_code string, w *models.LogCollector) ([]models.IPrecord, error) {
	baseURL := "https://minenergo.gov.ru/api/v1/"
	params := url.Values{}
	params.Add("action", "organizations.getItemDetail")
	params.Add("lang", "ru")
	params.Add("code", ip_code)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		slog.Error("Не удалось сформировать request: %s", slog.String("error", err.Error()))
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537")

	// Требуется чтобы принимался православный самоподписанный сертификат Минэнерго
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		// Проверяется не оканчивается ли текст ошибки на "context deadline exceeded"
		if err.Error()[len(err.Error())-25:] == "context deadline exceeded" {
			warn := fmt.Sprintf("Превышено время ожидания ответа от сервера Минэнерго (%s)", ip_code)
			w.Add(warn)
			return nil, nil
		}
		slog.Error("Не удалось выполнить запрос к серверу: %s",
			slog.String("error", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("unexpected status: %s", slog.String("status", resp.Status))
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if err.Error() == "context deadline exceeded" {
			warn := fmt.Sprintf("Превышено время ожидания ответа от сервера Минэнерго (%s)", ip_code)
			w.Add(warn)
			return nil, nil
		}
		slog.Error("error reading body: %s", slog.String("error", err.Error()))
		return nil, err
	}

	var IPdata models.FullData
	err = json.Unmarshal([]byte(body), &IPdata)
	if err != nil {
		slog.Error("Ошибка распаковки json в структуру ИП", slog.String("error", err.Error()))
		return nil, err
	}

	rec := IPdata.Docs[1].Recods

	return rec, nil

}
