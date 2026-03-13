package main

import (
	"testing"
	"time"
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
