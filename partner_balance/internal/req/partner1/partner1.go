package partner1

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

func GetBalance(token string) (float64, error) {
	url := "https://example.com/advertiser/balance.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0.0, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0.0, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0.0, fmt.Errorf("не удалось получить баланс, статус: %s, тело: %s", resp.Status, body)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	bodyBalance := gjson.GetBytes(resBody, "item")
	balance := bodyBalance.Float()

	return balance, nil
}
