package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"HobitsService/internal/app"
	"HobitsService/internal/config"
	"HobitsService/internal/infrastructure/database"
	httpserver "HobitsService/internal/infrastructure/http"
	"HobitsService/internal/logger"
	"HobitsService/internal/metrics"
)

func main() {
	cfg := config.MustLoad()

	if err := logger.Init(cfg.Env); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Logger.Sync()

	if err := metrics.Init(); err != nil {
		log.Fatalf("failed to init metrics: %v", err)
	}

	ctx := context.Background()

	db, err := database.New(ctx, &cfg.Postgres)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal("failed to get working directory", zap.Error(err))
	}

	migrationsPath := filepath.Join(wd, "migrations")
	if err := db.RunMigrations(migrationsPath); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	logger.Info("Successfully connected to database and ran migrations")

	application := app.NewApp(db)
	defer application.Close()

	logger.Info("Application initialized successfully")

	httpSrv := httpserver.NewServer(8080)
	if err := httpSrv.Start(); err != nil {
		logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
	logger.Info("HTTP server started on port 8080 for /metrics and /health")

	if err := application.GRPCServer.Start(); err != nil {
		logger.Fatal("failed to start gRPC server", zap.Error(err))
	}

	application.Scheduler.Start()
	logger.Info("Scheduler started for streak checks and reminders")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info("received signal", zap.Any("signal", sig))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Stop(ctx); err != nil {
		logger.Error("failed to stop HTTP server", zap.Error(err))
	}

	if err := application.GRPCServer.Stop(); err != nil {
		logger.Error("failed to stop gRPC server", zap.Error(err))
	}

	application.Close()
	logger.Info("Application shutdown completed")
}
