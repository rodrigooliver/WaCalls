package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

var webhookHTTP = &http.Client{Timeout: 8 * time.Second}

func signWebhook(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (m *SessionManager) deliverWebhook(sessionID string, ev any) {
	s, ok := m.Get(sessionID)
	if !ok || s.webhookURL == "" {
		return
	}
	body, err := json.Marshal(ev)
	if err != nil {
		return
	}
	url, secret := s.webhookURL, s.webhookSecret
	go m.postWebhook(sessionID, url, secret, body)
}

func (m *SessionManager) postWebhook(sessionID, url, secret string, body []byte) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		m.log.Warn("webhook request", "session", sessionID, "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "wacalls-webhook")
	if secret != "" {
		req.Header.Set("X-Wacalls-Signature", signWebhook(secret, body))
	}
	resp, err := webhookHTTP.Do(req)
	if err != nil {
		m.log.Warn("webhook delivery", "session", sessionID, "err", err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		m.log.Warn("webhook status", "session", sessionID, "status", resp.StatusCode)
	}
}
