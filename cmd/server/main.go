package main

import (
	"ecommerce/config"
	"ecommerce/internal/infrastructure/database"
	"ecommerce/pkg/logger"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.New()
	log.Info("configuration loaded",
		"server_port", cfg.Server.Port,
	)
	db,err:=database.Connect(cfg.Database)
	if err != nil{
		log.Error("Failed to connect database","error",err)
		return
	}
	log.Info("database connected succefully",
	"database",cfg.Database.Name,
	)
	_=db
}
