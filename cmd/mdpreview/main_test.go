package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Chavao/mdpreview/internal/config"
)

func TestReadyMessage(t *testing.T) {
	t.Parallel()

	got := readyMessage("0.0.0.0", 8631, 42*time.Millisecond)
	want := "\n  \x1b[1;38;2;10;170;255mmdpreview\x1b[0m λ ready in 42 ms\n\n  \x1b[38;2;10;170;255m➜\x1b[0m  Local:   \x1b[38;2;163;163;163mhttp://0.0.0.0:8631/\x1b[0m\n\n"

	if got != want {
		t.Fatalf("readyMessage() = %q, want %q", got, want)
	}
}

func TestFormatStartupDurationUsesMicroseconds(t *testing.T) {
	t.Parallel()

	got := formatStartupDuration(250 * time.Microsecond)

	if got != "250 µs" {
		t.Fatalf("formatStartupDuration() = %q, want %q", got, "250 µs")
	}
}

func TestFormatStartupDurationUsesMilliseconds(t *testing.T) {
	t.Parallel()

	got := formatStartupDuration(42 * time.Millisecond)

	if got != "42 ms" {
		t.Fatalf("formatStartupDuration() = %q, want %q", got, "42 ms")
	}
}

func TestListenWithPortDiscoveryUsesConfiguredPortWhenAvailable(t *testing.T) {
	t.Parallel()

	probe, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		t.Fatalf("reserve probe port: %v", err)
	}

	port := probe.Addr().(*net.TCPAddr).Port
	if err := probe.Close(); err != nil {
		t.Fatalf("close probe listener: %v", err)
	}

	cfg := config.Config{
		Host: "127.0.0.1",
		Port: port,
	}

	listener, err := listenWithPortDiscovery(&cfg)
	if err != nil {
		t.Fatalf("listenWithPortDiscovery() error = %v", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("listener address type = %T, want *net.TCPAddr", listener.Addr())
	}

	if addr.Port != cfg.Port {
		t.Fatalf("listener port = %d, want %d", addr.Port, cfg.Port)
	}
}

func TestMainFallsBackToNextPortWhenConfiguredPortIsBusy(t *testing.T) {
	host := "127.0.0.1"
	basePort, listeners := reserveConsecutivePorts(t, host, 2)
	defer closeListeners(listeners)

	if err := listeners[1].Close(); err != nil {
		t.Fatalf("close free port listener: %v", err)
	}

	cmd, stdout, stderr := newMainHelperCommand(t, host, basePort)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper process: %v", err)
	}

	expectedURL := fmt.Sprintf("http://%s:%d/", host, basePort+1)
	if err := waitForServerReady(expectedURL, 5*time.Second); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("wait for server ready: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("signal helper process: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for helper process: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), expectedURL) {
		t.Fatalf("stdout = %q, want substring %q", stdout.String(), expectedURL)
	}
}

func TestMainFailsAfterTenBusyPorts(t *testing.T) {
	host := "127.0.0.1"
	basePort, listeners := reserveConsecutivePorts(t, host, 10)
	defer closeListeners(listeners)

	cmd, stdout, stderr := newMainHelperCommand(t, host, basePort)

	err := cmd.Run()
	if err == nil {
		t.Fatalf("main() error = nil, want non-nil\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("main() error = %T, want *exec.ExitError", err)
	}

	if !strings.Contains(stderr.String(), "listen failed:") {
		t.Fatalf("stderr = %q, want listen failure", stderr.String())
	}

	if strings.Contains(stdout.String(), "λ ready") {
		t.Fatalf("stdout = %q, want no ready message", stdout.String())
	}
}

func TestMainHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	host := os.Getenv("MDPREVIEW_HELPER_HOST")
	port := os.Getenv("MDPREVIEW_HELPER_PORT")
	if host == "" || port == "" {
		t.Fatalf("helper process missing host or port")
	}

	originalArgs := os.Args
	os.Args = []string{os.Args[0], "--host", host, "--port", port}
	defer func() {
		os.Args = originalArgs
	}()

	main()
}

func newMainHelperCommand(t *testing.T, host string, port int) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainHelperProcess$")
	cmd.Env = append(
		os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		fmt.Sprintf("MDPREVIEW_HELPER_HOST=%s", host),
		fmt.Sprintf("MDPREVIEW_HELPER_PORT=%d", port),
	)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd, stdout, stderr
}

func reserveConsecutivePorts(t *testing.T, host string, count int) (int, []net.Listener) {
	t.Helper()

	for attempt := 0; attempt < 200; attempt++ {
		probe, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			t.Fatalf("reserve probe port: %v", err)
		}

		basePort := probe.Addr().(*net.TCPAddr).Port
		if err := probe.Close(); err != nil {
			t.Fatalf("close probe port: %v", err)
		}

		listeners := make([]net.Listener, 0, count)
		ok := true

		for offset := 0; offset < count; offset++ {
			listener, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(basePort+offset)))
			if err != nil {
				ok = false
				closeListeners(listeners)
				break
			}

			listeners = append(listeners, listener)
		}

		if ok {
			return basePort, listeners
		}
	}

	t.Fatalf("reserve %d consecutive ports: no available range found", count)
	return 0, nil
}

func closeListeners(listeners []net.Listener) {
	for _, listener := range listeners {
		if listener == nil {
			continue
		}

		_ = listener.Close()
	}
}

func waitForServerReady(url string, timeout time.Duration) error {
	client := http.Client{
		Timeout: 100 * time.Millisecond,
	}

	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}

			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		time.Sleep(25 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready at %s: %w", url, lastErr)
}
