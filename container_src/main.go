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

	"github.com/prdai/rssbot/repository"
	"github.com/prdai/rssbot/services"

	"go.uber.org/dig"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	c := dig.New()
	must(c.Provide(services.NewRSSParser))
	must(c.Provide(repository.NewMongoDBRepository, dig.As(new(repository.Repository))))
	must(c.Provide(services.NewRSSService, dig.As(new(services.RSSService))))
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
