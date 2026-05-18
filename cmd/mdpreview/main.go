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
	listener, err := listenWithPortDiscovery(&cfg)
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

func listenWithPortDiscovery(cfg *config.Config) (net.Listener, error) {
	var lastErr error
	attempts := cfg.PortDiscoveryAttempts
	if attempts < 1 {
		attempts = config.DefaultPortDiscoveryAttempts
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		listener, err := net.Listen("tcp", cfg.Address())

		if err == nil {
			return listener, nil
		}

		lastErr = err

		cfg.Port++
		log.Printf("retrying with port discovery, current address http://%s:%d", cfg.Host, cfg.Port)
	}

	return nil, lastErr
}

func readyMessage(host string, port int, startup time.Duration) string {
	const (
		reset     = "\x1b[0m"
		blue      = "\x1b[38;2;10;170;255m"
		boldBlue  = "\x1b[1;38;2;10;170;255m"
		lightGray = "\x1b[38;2;163;163;163m"

		styledAppName = boldBlue + "mdpreview" + reset
		styledArrow   = blue + "➜" + reset
		styledURL     = lightGray + "http://%s/" + reset
	)

	return fmt.Sprintf(
		"\n  %s λ ready in %s\n\n  %s  Local:   "+styledURL+"\n\n",
		styledAppName,
		formatStartupDuration(startup),
		styledArrow,
		net.JoinHostPort(host, fmt.Sprintf("%d", port)),
	)
}

func formatStartupDuration(startup time.Duration) string {
	if startup < time.Millisecond {
		return fmt.Sprintf("%d µs", startup.Microseconds())
	}

	return fmt.Sprintf("%d ms", startup.Milliseconds())
}
