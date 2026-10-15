package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignWebhook(t *testing.T) {
	body := []byte(`{"type":"session-qr"}`)
	got := signWebhook("secret", body)
	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write(body)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
