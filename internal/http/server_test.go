package http_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Chavao/mdpreview/internal/config"
	apphttp "github.com/Chavao/mdpreview/internal/http"
)

func TestNewServerServesIndex(t *testing.T) {
	t.Parallel()

	server := apphttp.NewServer(config.Config{
		Host: "127.0.0.1",
		Port: 5536,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	server.Handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}

	contentType := response.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	bodyText := string(body)
	if !strings.Contains(bodyText, "markdown-app") {
		t.Fatalf("body missing app root: %q", bodyText)
	}

	if !strings.Contains(bodyText, "<textarea") {
		t.Fatalf("body missing textarea: %q", bodyText)
	}
}
