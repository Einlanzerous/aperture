package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aperture-dashboard/aperture/internal/api"
	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/system"
)

func main() {
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	worker := checker.NewWorker(cfg)
	worker.Start()

	sysMonitor := system.NewMonitor()
	router := api.NewRouter(worker, sysMonitor, cfg)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Aperture — \"%s\" — listening on %s", cfg.Title, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	worker.Stop()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("Stopped.")
}
