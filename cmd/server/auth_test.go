package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyRequired(t *testing.T) {
	h := withAPIKey("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	unauth := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, unauth)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}

	ok := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	ok.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, ok)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}

	q := httptest.NewRequest(http.MethodGet, "/api/events?api_key=secret", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, q)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("query status %d", rec.Code)
	}
}

func TestAPIKeyDisabled(t *testing.T) {
	h := withAPIKey("", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/sessions", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
}
