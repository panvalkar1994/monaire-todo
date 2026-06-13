package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo/internal/config"
	"todo/internal/database"
	"todo/internal/handler"
	gormrepo "todo/internal/repository/gorm"
	"todo/internal/service"

	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 15 * time.Second

func notifyShutdown(c chan<- os.Signal) {
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
}

func Run(cfg *config.Config) error {
	gin.SetMode(cfg.Server.GinMode)

	slog.Info("starting server", "addr", cfg.Server.Addr, "gin_mode", cfg.Server.GinMode)

	db, err := database.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	repo := gormrepo.NewTodoRepository(db)
	svc := service.NewTodoService(repo)
	router := handler.NewRouter(svc)

	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: router,
	}

	slog.Info("listening", "addr", cfg.Server.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	notifyShutdown(quit)

	select {
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		slog.Error("server failed", "error", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	slog.Info("shutting down server", "timeout", shutdownTimeout.String())
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		return fmt.Errorf("shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}
