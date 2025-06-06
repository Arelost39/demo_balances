package partner2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ExampleResponse struct {
	BalanceCommon float64 `json:"balanceCommon"`
	BalanceReal   float64 `json:"balanceReal"`
}

func GetBalance(apiKey string) (float64, error) {
	url := "https://api.example.com/v1/public/finance/balance"

	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать запрос: %w", err)
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("X-Example-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
	}

	var data ExampleResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	return data.BalanceCommon, nil
}
