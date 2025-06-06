package partner3

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type FeedDetailStatisticsResponse struct {
	Data struct {
		Balance string `json:"balance"`
	} `json:"data"`
}

func GetBalance(token string) (float64, error) {
	feedID := "11111"
	url := fmt.Sprintf("https://example.com/api/v1/?api_token=%s&start_date=2025-03-31&end_date=2025-03-31&group_by=feed&feed_ids=%s", token, feedID)

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
	}

	var apiResp FeedDetailStatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	balance, err := strconv.ParseFloat(apiResp.Data.Balance, 64)
	if err != nil {
		return 0, fmt.Errorf("ошибка преобразования баланса: %w", err)
	}

	return balance, nil
}
