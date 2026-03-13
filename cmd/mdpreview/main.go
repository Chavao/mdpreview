package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chavao/mdpreview/internal/config"
	apphttp "github.com/Chavao/mdpreview/internal/http"
)

func main() {
	startedAt := time.Now()

	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("parse config: %v", err)
	}

	server := apphttp.NewServer(cfg)
	listener, err := net.Listen("tcp", cfg.Address())
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Print(readyMessage(cfg.Host, cfg.Port, time.Since(startedAt)))

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}
}

func readyMessage(host string, port int, startup time.Duration) string {
	return fmt.Sprintf(
		"\n  mdpreview  ready in %s\n\n  ➜  Local:   http://%s/\n\n",
		formatStartupDuration(startup),
		net.JoinHostPort(host, fmt.Sprintf("%d", port)),
	)
}

func formatStartupDuration(startup time.Duration) string {
	if startup < time.Millisecond {
		return fmt.Sprintf("%d µs", startup.Microseconds())
	}

	return fmt.Sprintf("%d ms", startup.Milliseconds())
}
