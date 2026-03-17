package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aperture-dashboard/aperture/internal/api"
	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/store"
	"github.com/aperture-dashboard/aperture/internal/store/postgres"
	"github.com/aperture-dashboard/aperture/internal/store/sqlite"
	"github.com/aperture-dashboard/aperture/internal/system"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	// Open persistent store (nil if storage is not configured).
	historyStore, err := openStore(context.Background(), cfg.Storage)
	if err != nil {
		slog.Error("open store", "error", err)
		os.Exit(1)
	}
	if historyStore != nil {
		slog.Info("storage enabled", "driver", cfg.Storage.Driver)
	}

	worker := checker.NewWorker(cfg)
	if historyStore != nil {
		worker.SetStore(historyStore)
	}
	worker.Start()

	sysMonitor := system.NewMonitor()
	router := api.NewRouter(worker, sysMonitor, cfg, historyStore)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "title", cfg.Title, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	worker.Stop(ctx)
	if historyStore != nil {
		if err := historyStore.Close(); err != nil {
			slog.Error("store close", "error", err)
		}
	}
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("stopped")
}

func openStore(ctx context.Context, cfg config.StorageConfig) (store.Store, error) {
	switch cfg.Driver {
	case "":
		return nil, nil
	case "sqlite":
		return sqlite.Open(ctx, cfg.DSN)
	case "postgres":
		return postgres.Open(ctx, cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %q", cfg.Driver)
	}
}
