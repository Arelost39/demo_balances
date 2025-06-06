package server

import (
	"context"
	"os"
	"tg_router/gateway"
	"tg_router/logger"
	"tg_router/tg"
	"tg_router/types"
	"time"

	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)


func Run() error {
	ctx := context.Background()

	// задаем свой токен телеги
	apiToken := os.Getenv("TG_BOT_TOKEN")

	// конфиг grpc
	grpcAddress := os.Getenv("GRPC_ADDRESS")
	if grpcAddress == "" {
		grpcAddress = "partner_balance:50051"
	}

	// Создаем gRPC клиент
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Errorf("Ошибка подключения к gRPC серверу: %v", err)
		return err
	}
	defer conn.Close()

	statClient := gateway.NewStatServiceClient(conn)

	threads, err := LoadThreadsFromFile("threads.yaml")
	if err != nil {
		logger.Log.Errorf("Ошибка загрузки конфига чатов: %v", err)
		return err
	}

	bot, err := tg.NewTelegramBot(apiToken, threads, statClient)
	if err != nil {
		logger.Log.Errorf("Ошибка запуска бота %v", err)
		return err
	}
	go bot.StartListening(ctx, "Multibot запущен")

	// Запуск планировщика
	scheduler(bot, ctx)
	return nil
}

// LoadThreadsFromFile загружает конфигурацию тредов из YAML файла
func LoadThreadsFromFile(filename string) (types.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return types.Config{}, err
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return types.Config{}, err
	}

	return config, nil
}

func scheduler(bot *tg.TelegramBot, ctx context.Context) {
	c := cron.New(cron.WithLocation(time.Local))
	c.AddFunc("15 10,17 * * *", func() {
		for _, thread := range bot.Threads.Threads {
			req := &gateway.StatRequest{
				Network: thread.Network,
			}
			resp, err := bot.StatClient.Stat(ctx, req)
			if err != nil {
				logger.Log.Errorf("Ошибка при получении статистики: %v", err)
				return
			}
			bot.SendMessage(thread.ChatID, thread.ThreadID, resp.Text)
		}
	})
	c.Start()
	select {}
}
