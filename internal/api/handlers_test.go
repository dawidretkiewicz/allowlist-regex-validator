package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListAllowlistsEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler().Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/allowlists", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var items []PresetAllowlist
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("expected 4 presets, got %d", len(items))
	}
}

func TestValidateEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler().Register(mux)

	payload := ProcessRequest{
		JSONRegexObject:  `{"g": ["^https:\\/\\/example\\.com\\/.*$", "("]}`,
		MultilineRegexes: "",
		DecodeJSON:       true,
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/validate", bytes.NewReader(b))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ProcessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if resp.Summary.TotalCount != 2 || resp.Summary.InvalidCount != 1 {
		t.Fatalf("unexpected summary: %#v", resp.Summary)
	}
}

func TestMatchEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler().Register(mux)

	payload := ProcessRequest{
		MultilineRegexes: "^https://a\\.com/.*$",
		MultilineURLs:    "https://a.com/x\nhttps://b.com",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", bytes.NewReader(b))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ProcessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if resp.Match == nil || len(resp.Match.Rows) != 2 {
		t.Fatalf("missing match result")
	}
}
