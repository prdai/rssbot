package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"go.uber.org/dig"
)

func RSSHandler(digContainer *dig.Container) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error(err.Error())
		}
		RSSFeeds := new(RSSFeeds)
		err = json.Unmarshal(bodyBytes, RSSFeeds)
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		slog.Info(fmt.Sprintf("Received request: %s", RSSFeeds))
		// TODO
		if err := h.container.Invoke(InvokeRssSync); err != nil {
			slog.Error("dig invoke failed", "err", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
