package http_test

import (
	"bytes"
	"encoding/json"
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

	if !strings.Contains(bodyText, `id="view-toggle"`) {
		t.Fatalf("body missing view toggle button: %q", bodyText)
	}
}

func TestRenderEndpointRendersMarkdown(t *testing.T) {
	t.Parallel()

	server := apphttp.NewServer(config.Config{
		Host: "127.0.0.1",
		Port: 5536,
	})

	requestBody := bytes.NewBufferString(`{"markdown":"# Hello"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/render", requestBody)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}

	contentType := response.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}

	var payload struct {
		HTML string `json:"html"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload.HTML != "<h1>Hello</h1>" {
		t.Fatalf("html = %q, want %q", payload.HTML, "<h1>Hello</h1>")
	}
}

func TestRenderEndpointRejectsInvalidContentType(t *testing.T) {
	t.Parallel()

	server := apphttp.NewServer(config.Config{
		Host: "127.0.0.1",
		Port: 5536,
	})

	request := httptest.NewRequest(http.MethodPost, "/api/render", strings.NewReader(`{"markdown":"# Hello"}`))
	request.Header.Set("Content-Type", "text/plain")

	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("StatusCode = %d, want %d", recorder.Result().StatusCode, http.StatusUnsupportedMediaType)
	}
}
