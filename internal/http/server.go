package http

import (
	"net/http"
	"time"

	"github.com/Chavao/mdpreview/internal/config"
	"github.com/Chavao/mdpreview/internal/web"
)

func NewServer(cfg config.Config) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", index)

	return &http.Server{
		Addr:              cfg.Address(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(web.IndexHTML)
}
