package config

import (
	"log/slog"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	for _, tt := range []struct {
		value string
		level slog.Level
	}{{"", slog.LevelInfo}, {"DEBUG", slog.LevelDebug}, {"warn", slog.LevelWarn}, {"ERROR", slog.LevelError}} {
		t.Setenv("LOG_LEVEL", tt.value)
		cfg, err := Load()
		if err != nil || cfg.LogLevel != tt.level {
			t.Fatalf("LOG_LEVEL=%q: %+v, %v", tt.value, cfg, err)
		}
	}
	t.Setenv("LOG_LEVEL", "invalid")
	if _, err := Load(); err == nil {
		t.Fatal("invalid log level accepted")
	}
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	if _, err := Load(); err == nil {
		t.Fatal("empty token accepted")
	}
}
