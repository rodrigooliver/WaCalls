package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

func initSentry(log *slog.Logger) func() {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return func() {}
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      os.Getenv("SENTRY_ENVIRONMENT"),
		Release:          os.Getenv("SENTRY_RELEASE"),
		AttachStacktrace: true,
	})
	if err != nil {
		log.Error("sentry init failed", "err", err)
		return func() {}
	}
	log.Info("sentry enabled")
	return func() { sentry.Flush(2 * time.Second) }
}
