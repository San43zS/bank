package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"banking-platform/internal/app"
)

func main() {
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- application.Run()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Printf("Failed to shutdown: %v", err)
		}
	case err := <-runErrCh:
		if err != nil {
			log.Fatalf("Failed to run app: %v", err)
		}
	}
}
