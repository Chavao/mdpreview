package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/Chavao/mdpreview/internal/markdown"
)

const maxRenderBodyBytes = 1 << 20

type renderRequest struct {
	Markdown string `json:"markdown"`
}

type renderResponse struct {
	HTML string `json:"html"`
}

func render(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "content type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRenderBodyBytes)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request renderRequest
	if err := decoder.Decode(&request); err != nil {
		if errors.Is(err, io.EOF) {
			http.Error(w, "request body is required", http.StatusBadRequest)
			return
		}

		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "request body must contain a single JSON object", http.StatusBadRequest)
		return
	}

	if !utf8.ValidString(request.Markdown) {
		http.Error(w, "markdown must be valid UTF-8", http.StatusBadRequest)
		return
	}

	response := renderResponse{HTML: markdown.Render(request.Markdown)}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
