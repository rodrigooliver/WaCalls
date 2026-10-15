package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
)

type sessionRow struct {
	ID            string
	Name          string
	JID           string
	WebhookURL    string
	WebhookSecret string
}

type sessionStore struct{ db *sql.DB }

func newSessionStore(ctx context.Context, db *sql.DB) (*sessionStore, error) {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS sessions (
		id   TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		jid  TEXT
	)`)
	if err != nil {
		return nil, err
	}
	for _, stmt := range []string{
		`ALTER TABLE sessions ADD COLUMN webhook_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE sessions ADD COLUMN webhook_secret TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := db.ExecContext(ctx, stmt); err != nil && !strings.Contains(err.Error(), "duplicate column") {
			return nil, err
		}
	}
	return &sessionStore{db: db}, nil
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *sessionStore) list(ctx context.Context) ([]sessionRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, COALESCE(jid, ''), COALESCE(webhook_url, ''), COALESCE(webhook_secret, '') FROM sessions ORDER BY rowid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []sessionRow
	for rows.Next() {
		var r sessionRow
		if err := rows.Scan(&r.ID, &r.Name, &r.JID, &r.WebhookURL, &r.WebhookSecret); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *sessionStore) insert(ctx context.Context, id, name, webhookURL, webhookSecret string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sessions (id, name, jid, webhook_url, webhook_secret) VALUES (?, ?, NULL, ?, ?)`, id, name, webhookURL, webhookSecret)
	return err
}

func (s *sessionStore) setJID(ctx context.Context, id, jid string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET jid = ? WHERE id = ?`, jid, id)
	return err
}

func (s *sessionStore) delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id)
	return err
}
