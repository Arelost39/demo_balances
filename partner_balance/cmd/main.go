package main

import (	
	"os"
	"partner_balance/internal/logger"
	server "partner_balance/internal/server"
)

func main() {
    if err := server.Run(); err != nil {
        logger.Log.Errorf("Ошибка: %v", err)
        os.Exit(1)
    }
}