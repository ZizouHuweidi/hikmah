package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zizouhuweidi/maktaba/internal/config"
	"github.com/zizouhuweidi/maktaba/internal/db"
	"github.com/zizouhuweidi/maktaba/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger.Info("starting Sabeel service")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	database, err := db.NewDB(cfg.Database.URL, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("connected to database")

	httpServer, err := server.New(cfg, database, logger)
	if err != nil {
		logger.Error("failed to build server", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("server starting", "port", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("server exited")
}
