package server

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	grpcpkg "google.golang.org/grpc"
	"net"
	grpc "partner_balance/gateway"
	"partner_balance/internal/logger"
	db "partner_balance/internal/postgres"
	"partner_balance/internal/processor"
	"partner_balance/internal/utils"
	"time"
)

func Run() error {
	// Загружаем переменные окружения
	if err := utils.LoadEnv(); err != nil {
		logger.Log.Errorf("Ошибка загрузки переменных окружения: %v", err)
		return err
	}

	// Инициализируем логирование
	logger.InitLogger()

	// соединение с бд
	if err := db.InitDB(); err != nil {
		logger.Log.Errorf("Ошибка соединения postgres: %v", err)
		return err
	}

	// Загружаем конфиг с партнерскими сетками
	if err := utils.LoadConfig("config.yaml"); err != nil {
		logger.Log.Errorf("Ошибка загрузки конфигурации: %v", err)
		return err
	}

	// Вставляем партнёров из конфигурации
	if err := processor.InsertPartners(processor.PartnerList()); err != nil {
		logger.Log.Errorf("Ошибка вставки партнёров: %v", err)
		return err
	}

	// Запускаем gRPC-сервер в фоне
	go func() {
		if err := startGRPCServer(); err != nil {
			logger.Log.Errorf("Ошибка запуска gRPC-сервера: %v", err)
		}
	}()

	// Запуск планировщика
	if err := scheduler(context.Background()); err != nil {
		logger.Log.Errorf("Ошибка планировщика: %v", err)
		return err
	}
	return nil
}

// Пополнение таблиц балансов по расписанию, удаление старых записей из бд

func scheduler(ctx context.Context) error {
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc("0 4,10,16,22 * * *", func() {
		if err := processor.BalanceInsert(); err != nil {
			logger.Log.Warnf("BalanceInsert error: %v", err)
		}
	})

	c.AddFunc("0 0 0 * * *", func() {
		if err := delete(); err != nil {
			logger.Log.Warnf("delete error: %v", err)
		}
	})

	c.Start()

	// на будущее
	<-ctx.Done()
	c.Stop()
	return nil
}

func delete() error {
	err := db.DeleteOldData("admeking")
	if err != nil {
		return err
	}
	err = db.DeleteOldData("realpush")
	if err != nil {
		return err
	}
	err = db.DeleteOldData("advertrek")
	if err != nil {
		return err
	}
	return nil
}

// gRPC!!!

// statServer реализует gateway.StatServiceServer
type statServer struct {
	grpc.UnimplementedStatServiceServer
}

// Stat обрабатывает запрос StatRequest и возвращает StatReply
func (s *statServer) Stat(ctx context.Context, req *grpc.StatRequest) (*grpc.StatReply, error) {
	network := req.GetNetwork()
	text := processor.CompareBalances(network)
	return &grpc.StatReply{Text: text, Network: network}, nil
}

//  поднимаем gRPC-сервер и слушаем порт 50051
func startGRPCServer() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	grpcServer := grpcpkg.NewServer()
	grpc.RegisterStatServiceServer(grpcServer, &statServer{})
	logger.Log.Infof("gRPC-сервер запущен на :50051")
	return grpcServer.Serve(lis)
}

