package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	logger.Log.Infof("Telegram бот успешно инициализирован: %s", bot.Self.UserName)
	return &TelegramBot{
		Bot:        bot,
		Threads:    threads,
		APIToken:   apiToken,
		StatClient: statClient,
	}, nil
}

// Запуск прослушивания обновлений
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
				logger.Log.Errorf("Ошибка при получении статистики: %v", err)
				return
			}
			if err := t.SendMessage(chatID, thread.ThreadID, resp.Text); err != nil {
				logger.Log.Errorf("Ошибка отправки ответа: %v", err)
			}
		}()
	default:
		logger.Log.Infof("[%s] Неизвестная команда: %s", botName, command)
		return
	}
}

// Функция отправки сообщения в телегу
func (t *TelegramBot) SendMessage(chatid int64, threadid int64, msgText string) error {

	if strings.TrimSpace(msgText) == "" {
		logger.Log.Infof("SendMessageThread: пустое сообщение, отправка пропущена")
		return nil
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.APIToken)
	payload := map[string]interface{}{
		"chat_id":           chatid,
		"message_thread_id": threadid,
		"text":              msgText,
		"parse_mode":        "HTML",
	}
	data, _ := json.Marshal(payload)
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
