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

	"server/repository"
	"server/services"

	"go.uber.org/dig"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	c := dig.New()
	c.Provide(services.NewRSSParser)
	c.Provide(repository.NewMongoDBRepository)
	c.Provide(services.NewRSSService)
	handler := NewHandler(c)
	router := http.NewServeMux()
	router.HandleFunc("/", handler.mainHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		slog.Info(fmt.Sprintf("Server listening on %s\n", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error())
		}
	}()

	sig := <-stop

	slog.Info(fmt.Sprintf("Received signal (%s), shutting down server...", sig))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Server shutdown successfully")
}
