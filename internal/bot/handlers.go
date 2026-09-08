package bot

import (
	"errors"
	"log"
	"strings"

	"github.com/sadbag21/dota-bot-info/pkg/formatter"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	message := update.Message

	log.Printf(
		"Message from %s: %s",
		message.From.UserName,
		message.Text,
	)

	if !message.IsCommand() {
		return
	}

	switch strings.ToLower(message.Command()) {
	case "start":
		b.handleStart(message)

	case "help":
		b.handleHelp(message)

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
		log.Printf(
			"Failed to get player %d: %v",
			accountID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			userErrorMessage(err),
		)

		return
	}

	text := formatter.FormatPlayer(
		playerInfo,
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
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
		log.Printf(
			"Failed to get recent matches for %d: %v",
			accountID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			userErrorMessage(err),
		)

		return
	}

	text := formatter.FormatMatches(
		accountID,
		matches,
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
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
		log.Printf(
			"Failed to get hero stats for %d: %v",
			accountID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			userErrorMessage(err),
		)

		return
	}

	text := formatter.FormatHeroes(
		accountID,
		heroes,
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
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
		log.Printf(
			"Failed to get stats for %d: %v",
			accountID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			userErrorMessage(err),
		)

		return
	}

	text := formatter.FormatStats(
		accountID,
		stats,
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

func (b *Bot) handleMatch(message *tgbotapi.Message) {
	matchID, ok := b.getMatchID(
		message,
		"match",
	)

	if !ok {
		return
	}

	log.Printf(
		"Getting match information for match_id=%d",
		matchID,
	)

	match, err := b.service.GetMatchDetails(
		matchID,
	)

	if err != nil {
		log.Printf(
			"Failed to get match %d: %v",
			matchID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			userErrorMessage(err),
		)

		return
	}

	text := formatter.FormatMatch(
		match,
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

func (b *Bot) handleImpact(message *tgbotapi.Message) {
	matchID, ok := b.getMatchID(
		message,
		"impact",
	)

	if !ok {
		return
	}

	log.Printf(
		"Getting impact analysis for match_id=%d",
		matchID,
	)

	match, err := b.service.GetMatchDetails(
		matchID,
	)

	if err != nil {
		log.Printf(
			"Failed to get impact for match %d: %v",
			matchID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			userErrorMessage(err),
		)

		return
	}

	text := formatter.FormatImpact(
		match,
	)

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

// getDotaID получает Dota ID из команды.
// При ошибке сам отправляет сообщение пользователю.
func (b *Bot) getDotaID(
	message *tgbotapi.Message,
	command string,
) (int64, bool) {
	accountID, err := parseDotaID(
		message,
	)

	if err == nil {
		return accountID, true
	}

	if errors.Is(err, errMissingID) {
		b.sendMessage(
			message.Chat.ID,
			"❌ Укажи Dota ID.\n\n"+
				"Пример:\n"+
				"<code>/"+command+" 1677175114</code>",
		)

		return 0, false
	}

	b.sendMessage(
		message.Chat.ID,
		"❌ Некорректный Dota ID.\n\n"+
			"Пример:\n"+
			"<code>/"+command+" 1677175114</code>",
	)

	return 0, false
}

// getMatchID получает Match ID из команды.
// При ошибке сам отправляет сообщение пользователю.
func (b *Bot) getMatchID(
	message *tgbotapi.Message,
	command string,
) (int64, bool) {
	matchID, err := parseMatchID(
		message,
	)

	if err == nil {
		return matchID, true
	}

	if errors.Is(err, errMissingID) {
		b.sendMessage(
			message.Chat.ID,
			"❌ Укажи Match ID.\n\n"+
				"Пример:\n"+
				"<code>/"+command+" 8983647546</code>",
		)

		return 0, false
	}

	b.sendMessage(
		message.Chat.ID,
		"❌ Некорректный Match ID.\n\n"+
			"Пример:\n"+
			"<code>/"+command+" 8983647546</code>",
	)

	return 0, false
}
