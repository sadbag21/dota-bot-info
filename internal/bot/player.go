package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/pkg/formatter"
	"log/slog"
)

func (b *Bot) handlePlayer(message *tgbotapi.Message) {
	accountID, ok := b.getDotaID(
		message,
		"player",
	)

	if !ok {
		return
	}

	slog.Debug("Getting player information", "account_id", accountID)

	playerInfo, err := b.service.GetPlayerInfo(
		accountID,
	)

	if err != nil {
		b.reportCommandError(message, accountID, err)
		return
	}

	b.rememberPlayer(message, accountID)
	text := formatter.FormatPlayer(
		playerInfo,
	)
	if _, ok := selectionKey(message); ok {
		text += "\n\n✅ Игрок выбран. Теперь /matches, /heroes и /stats работают без ID."
	}

	b.sendMessage(message.Chat.ID, text, playerKeyboard(accountID))
}

func (b *Bot) handleMatches(message *tgbotapi.Message) {
	accountID, ok := b.getDotaID(
		message,
		"matches",
	)

	if !ok {
		return
	}

	slog.Debug("Getting recent matches", "account_id", accountID)

	matches, err := b.service.GetRecentMatchesInfo(
		accountID,
		10,
	)

	if err != nil {
		b.reportCommandError(message, accountID, err)
		return
	}

	text := formatter.FormatMatches(
		accountID,
		matches,
	)

	b.sendMessage(message.Chat.ID, text, playerKeyboard(accountID))
}

func (b *Bot) handleHeroes(message *tgbotapi.Message) {
	accountID, ok := b.getDotaID(
		message,
		"heroes",
	)

	if !ok {
		return
	}

	slog.Debug("Getting hero stats", "account_id", accountID)

	heroes, err := b.service.GetHeroStats(
		accountID,
		10,
	)

	if err != nil {
		b.reportCommandError(message, accountID, err)
		return
	}

	text := formatter.FormatHeroes(
		accountID,
		heroes,
	)

	b.sendMessage(message.Chat.ID, text, playerKeyboard(accountID))
}

func (b *Bot) handleStats(message *tgbotapi.Message) {
	accountID, ok := b.getDotaID(
		message,
		"stats",
	)

	if !ok {
		return
	}

	slog.Debug("Getting extended stats", "account_id", accountID)

	stats, err := b.service.GetPlayerStats(
		accountID,
	)

	if err != nil {
		b.reportCommandError(message, accountID, err)
		return
	}

	text := formatter.FormatStats(
		accountID,
		stats,
	)

	b.sendMessage(message.Chat.ID, text, playerKeyboard(accountID))
}
