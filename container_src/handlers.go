package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/prdai/rssbot/services"

	"go.uber.org/dig"
)

func NewHandler(container *dig.Container) handler {
	return handler{container}
}

type handler struct {
	container *dig.Container
}

func (h *handler) mainHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(err.Error())
	}
	RSSFeeds := RSSFeeds{}
	err = json.Unmarshal(bodyBytes, &RSSFeeds)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	slog.Info(fmt.Sprintf("Received request: %s", RSSFeeds))
	if err := h.container.Invoke(func(s services.RSSService) error {
		slog.Info("Invoking Sync RSS Feeds")
		s.SyncRSSFeeds(RSSFeeds.Feeds, r.Context())
		return nil
	}); err != nil {
		slog.Error("dig invoke failed", "err", err)
	}
	w.WriteHeader(http.StatusAccepted)
}
