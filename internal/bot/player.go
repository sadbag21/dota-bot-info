package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/pkg/formatter"
	"log"
)

func (b *Bot) handlePlayer(message *tgbotapi.Message) {
	accountID, ok := b.getDotaID(
		message,
		"player",
	)

	if !ok {
		return
	}

	log.Printf(
		"Getting player information for account_id=%d",
		accountID,
	)

	playerInfo, err := b.service.GetPlayerInfo(
		accountID,
	)

	if err != nil {
		b.reportCommandError(message, accountID, err)
		return
	}

	text := formatter.FormatPlayer(
		playerInfo,
	)

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

	log.Printf(
		"Getting recent matches for account_id=%d",
		accountID,
	)

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

	log.Printf(
		"Getting hero stats for account_id=%d",
		accountID,
	)

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

	log.Printf(
		"Getting extended stats for account_id=%d",
		accountID,
	)

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
