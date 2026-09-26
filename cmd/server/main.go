package main

import (
	"ecommerce/config"
	"ecommerce/internal/bootstrap"
	"ecommerce/internal/infrastructure/database"
	"ecommerce/pkg/logger"
)

func main() {
	cfg := config.LoadConfig()
	log, logCloser, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer logCloser.Close()
	log.Info("configuration loaded",
		"server_port", cfg.Server.Port,
	)
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Error("Failed to connect database", "error", err)
		return
	}
	log.Info("database connected successfully",
		"database", cfg.Database.Name,
	)
	app := bootstrap.NewApplication(db, log, cfg.Redis)
	log.Info("Server starting", "port", cfg.Server.Port)
	if app.Router.Run(":" + cfg.Server.Port); err != nil {
		log.Error("server stopped unexpectedly", "error", err)
	}
}
