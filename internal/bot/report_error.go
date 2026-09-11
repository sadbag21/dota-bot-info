package bot

import (
	"context"
	"errors"
	"log/slog"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) reportCommandError(message *tgbotapi.Message, id int64, err error) {
	if errors.Is(err, context.Canceled) || (b.ctx != nil && b.ctx.Err() != nil) {
		return
	}
	slog.Error("Command failed", "command", message.Command(), "id", id, "chat_id", message.Chat.ID, "error", err)
	b.sendMessage(message.Chat.ID, userErrorMessage(err))
}

// Telegram's request URL contains the bot token. Log the underlying transport
// error, rather than the URL attached by net/http.
func safeTelegramError(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
}
