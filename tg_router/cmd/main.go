package main

import (
	"os"
	"tg_router/logger"
	"tg_router/server"
)

func main() {
	if err := server.Run(); err != nil {
		logger.Log.Errorf("Ошибка: %v", err)
		os.Exit(1)
	}
}
