package main

import (
	"context"
	"ecommerce/config"
	"ecommerce/internal/bootstrap"
	"ecommerce/internal/infrastructure/database"
	"ecommerce/pkg/logger"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           app.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() {
		log.Info("HTTP server starting", "port", cfg.Server.Port)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server stopped unexpectedly", "error", err)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful HTTP shutdown failed", "error", err)
		if closeErr := server.Close(); closeErr != nil {
			log.Error("forced HTTP shutdown failed", "error", closeErr)
		}
	}
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		app.WorkerManager.Wait()
	}()
	select {
	case <-workerDone:
		log.Info("background workers stopped successfully")
	case <-shutdownCtx.Done():
		log.Info("timed out waiting for background workers to stop")
	}
	log.Info("server shutdown complete")
}
