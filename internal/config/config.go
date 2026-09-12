package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	StratzAPIToken   string
	LogLevel         slog.Level
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if token == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is not set")
	}
	var level slog.Level
	if value := strings.TrimSpace(os.Getenv("LOG_LEVEL")); value != "" {
		if err := level.UnmarshalText([]byte(value)); err != nil {
			return Config{}, fmt.Errorf("invalid LOG_LEVEL: %w", err)
		}
	}
	return Config{TelegramBotToken: token, StratzAPIToken: strings.TrimSpace(os.Getenv("STRATZ_API_TOKEN")), LogLevel: level}, nil
}
