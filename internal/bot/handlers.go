package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"strings"
	"time"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
		return
	}
	if update.Message == nil || update.Message.Chat == nil {
		return
	}

	message := update.Message

	if !message.IsCommand() {
		return
	}

	started := time.Now()
	defer func() {
		userID := int64(0)
		if message.From != nil {
			userID = message.From.ID
		}
		slog.Info("Command processed", "command", message.Command(), "telegram_user", userID, "chat_id", message.Chat.ID, "duration", time.Since(started))
	}()
	switch strings.ToLower(message.Command()) {
	case "start":
		b.handleStart(message)

	case "help":
		b.handleHelp(message)
	case "meta":
		b.handleMeta(message)

	case "player":
		b.handlePlayer(message)

	case "matches":
		b.handleMatches(message)

	case "heroes":
		b.handleHeroes(message)

	case "stats":
		b.handleStats(message)

	case "match":
		b.handleMatch(message)

	case "impact":
		b.handleImpact(message)

	default:
		b.sendMessage(
			message.Chat.ID,
			"❌ Неизвестная команда. Используй /help",
		)
	}
}
