package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"tg_router/gateway"
	"tg_router/logger"
	"tg_router/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура бота
type TelegramBot struct {
	Bot        *tgbotapi.BotAPI
	Threads    types.Config
	APIToken   string
	StatClient gateway.StatServiceClient
}

// Инициализация бота
func NewTelegramBot(apiToken string, threads types.Config, statClient gateway.StatServiceClient) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации Telegram бота: %v", err)
	}
	logger.Log.Infof("[TelegramBot] Telegram бот успешно инициализирован: %s", bot.Self.UserName)
	return &TelegramBot{
		Bot:        bot,
		Threads:    threads,
		APIToken:   apiToken,
		StatClient: statClient,
	}, nil
}

// StartListening запускает прослушивание обновлений Telegram бота.
func (t *TelegramBot) StartListening(ctx context.Context, botName string) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := t.Bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			t.handleUpdate(ctx, update, botName)
		case <-ctx.Done():
			logger.Log.Infof("[%s] Остановка бота", botName)
			return
		}
	}
}

// Вспомогательная функция поиска Thread по chatID
func (t *TelegramBot) findThread(chatID int64) *types.Thread {
	for i, th := range t.Threads.Threads {
		if th.ChatID == chatID {
			return &t.Threads.Threads[i]
		}
	}
	return nil
}

// Главный обработчик обновлений
func (t *TelegramBot) handleUpdate(ctx context.Context, update tgbotapi.Update, botName string) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID

	thread := t.findThread(chatID)
	if thread == nil {
		logger.Log.Infof("[%s] Неизвестная группа/тема: %d", botName, chatID)
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	command := update.Message.Command()
	logger.Log.Infof("[%s] Получена команда: %s (чат: %d, network: %s)", botName, command, chatID, thread.Network)

	switch command {
	case "stat":
		go func() {
			req := &gateway.StatRequest{
				Network: thread.Network,
			}
			resp, err := t.StatClient.Stat(ctx, req)
			if err != nil {
				logger.Log.Errorf("[%s] Ошибка при получении статистики: %v", botName, err)
				return
			}
			if err := t.SendMessage(chatID, thread.ThreadID, resp.Text); err != nil {
				logger.Log.Errorf("[%s] Ошибка отправки ответа: %v", botName, err)
			}
		}()
	default:
		logger.Log.Infof("[%s] Неизвестная команда: %s", botName, command)
		return
	}
}

// SendMessage отправляет сообщение в телегу.
func (t *TelegramBot) SendMessage(chatID int64, threadID int64, msgText string) error {

	if len(bytes.TrimSpace([]byte(msgText))) == 0 {
		logger.Log.Infof("[TelegramBot] SendMessage: пустое сообщение, отправка пропущена")
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.APIToken)
	payload := map[string]interface{}{
		"chat_id":           chatID,
		"message_thread_id": threadID,
		"text":              msgText,
		"parse_mode":        "HTML",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка ответа от Telegram API: %d", resp.StatusCode)
	}
	return nil
}
