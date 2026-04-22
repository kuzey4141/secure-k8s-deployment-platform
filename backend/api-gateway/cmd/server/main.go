package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/config"
	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/deployments"
	"github.com/kuzey/secure-deploy-platform/backend/api-gateway/internal/httpapi"
)

// main loads configuration, opens the database, wires dependencies, and starts the HTTP server.
func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		log.Fatalf("ping database: %v", err)
	}
	cancel()

	repo := deployments.NewRepository(db)
	service := deployments.NewService(repo)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.New(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("api-gateway listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %v", err)
	}

	<-shutdownDone
}
