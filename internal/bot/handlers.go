package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Chat == nil {
		return
	}

	message := update.Message

	if message.From != nil {
		log.Printf("Message from %s: %s", message.From.UserName, message.Text)
	}

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
