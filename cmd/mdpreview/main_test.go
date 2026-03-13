package main

import (
	"testing"
	"time"
)

func TestReadyMessage(t *testing.T) {
	t.Parallel()

	got := readyMessage("0.0.0.0", 8631, 42*time.Millisecond)
	want := "\n  mdpreview  ready in 42 ms\n\n  ➜  Local:   http://0.0.0.0:8631/\n\n"

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
