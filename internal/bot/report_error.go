package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) reportCommandError(message *tgbotapi.Message, id int64, err error) {
	log.Printf("Command %s failed for id=%d: %v", message.Command(), id, err)
	b.sendMessage(message.Chat.ID, userErrorMessage(err))
}
