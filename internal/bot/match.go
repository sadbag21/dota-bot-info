package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/pkg/formatter"
	"log/slog"
)

func (b *Bot) handleMatch(message *tgbotapi.Message) {
	matchID, ok := b.getMatchID(
		message,
		"match",
	)

	if !ok {
		return
	}

	slog.Debug("Getting match information", "match_id", matchID)

	match, err := b.service.GetMatchDetails(
		matchID,
	)

	if err != nil {
		b.reportCommandError(message, matchID, err)
		return
	}

	text := formatter.FormatMatch(
		match,
	)

	b.sendMessage(message.Chat.ID, text, matchKeyboard(matchID))
}

func (b *Bot) handleImpact(message *tgbotapi.Message) {
	matchID, ok := b.getMatchID(
		message,
		"impact",
	)

	if !ok {
		return
	}

	slog.Debug("Getting impact analysis", "match_id", matchID)

	match, err := b.service.GetMatchDetails(
		matchID,
	)

	if err != nil {
		b.reportCommandError(message, matchID, err)
		return
	}

	text := formatter.FormatImpact(
		match,
	)

	b.sendMessage(message.Chat.ID, text, matchKeyboard(matchID))
}
