package bot

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/internal/config"
	"github.com/sadbag21/dota-bot-info/internal/meta"
	"github.com/sadbag21/dota-bot-info/internal/service"
)

type Bot struct {
	api             *tgbotapi.BotAPI
	service         playerService
	meta            metaProvider
	limiter         *commandLimiter
	selectedPlayers playerSelections
	ctx             context.Context
	httpClient      *http.Client
}

func New(ctx context.Context, cfg config.Config, playerService *service.PlayerService) (*Bot, error) {
	b := &Bot{service: playerService, limiter: newCommandLimiter(), ctx: ctx}
	if err := b.selectedPlayers.open(cfg.DataDir); err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 40 * time.Second}
	api, err := tgbotapi.NewBotAPIWithClient(cfg.TelegramBotToken, tgbotapi.APIEndpoint, contextHTTPClient{ctx: ctx, client: client})
	if err != nil {
		client.CloseIdleConnections()
		return nil, fmt.Errorf("authorize Telegram bot: %w", safeTelegramError(err))
	}
	slog.Info("Telegram bot authorized", "username", api.Self.UserName)
	b.api, b.httpClient, b.meta = api, client, meta.NewClient(cfg.StratzAPIToken)
	return b, nil
}

func (b *Bot) Run() {
	if closer, ok := b.meta.(interface{ Close() }); ok {
		defer closer.Close()
	}
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if b.httpClient != nil {
		defer b.httpClient.CloseIdleConnections()
	}
	updates := tgbotapi.NewUpdate(0)
	updates.Timeout = 30
	slog.Info("Bot started")
	defer slog.Info("Bot stopped")
	for ctx.Err() == nil {
		batch, err := b.api.GetUpdates(updates)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("Telegram polling failed", "error", safeTelegramError(err))
			timer := time.NewTimer(3 * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			continue
		}
		for _, update := range batch {
			if ctx.Err() != nil {
				return
			}
			if update.UpdateID < updates.Offset {
				continue
			}
			b.handleUpdate(update)
			updates.Offset = update.UpdateID + 1
		}
	}
}

// Telegram creates requests with a background context. Bind them to the
// application's lifetime before http.Client applies its request timeout.
type contextHTTPClient struct {
	ctx    context.Context
	client *http.Client
}

func (c contextHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req.Clone(c.ctx))
}
