package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sadbag21/dota-bot-info/internal/bot"
	"github.com/sadbag21/dota-bot-info/internal/config"
	"github.com/sadbag21/dota-bot-info/internal/dota"
	"github.com/sadbag21/dota-bot-info/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Bot failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel})))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dotaClient := dota.NewClientWithContext(ctx)
	defer dotaClient.Close()
	telegramBot, err := bot.New(ctx, cfg, service.NewPlayerService(dotaClient))
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	telegramBot.Run()
	return nil
}
