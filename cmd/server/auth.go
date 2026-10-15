package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"
)

func withAPIKey(apiKey string, next http.Handler) http.Handler {
	if apiKey == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !apiKeyMatches(presentedAPIKey(r), apiKey) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="wacalls"`)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func presentedAPIKey(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(h, prefix) {
			return strings.TrimSpace(h[len(prefix):])
		}
	}
	if k := r.Header.Get("X-API-Key"); k != "" {
		return k
	}
	return r.URL.Query().Get("api_key")
}

func apiKeyMatches(got, want string) bool {
	sum := func(s string) []byte {
		h := sha256.Sum256([]byte(s))
		return h[:]
	}
	return subtle.ConstantTimeCompare(sum(got), sum(want)) == 1
}
