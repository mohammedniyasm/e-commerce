package main

import (
	"ecommerce/config"
	"ecommerce/internal/bootstrap"
	"ecommerce/internal/infrastructure/database"
	"ecommerce/pkg/logger"
	"log"
	"os"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config", "error", err)
		return
	}
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
	app, err := bootstrap.NewApplication(db, log, cfg)
	if err != nil {
		log.Error("failed to start application", "error", err)
		os.Exit(1)
	}
	log.Info("Server starting", "port", cfg.Server.Port)
	if app.Router.Run(":" + cfg.Server.Port); err != nil {
		log.Error("server stopped unexpectedly", "error", err)
	}
}
