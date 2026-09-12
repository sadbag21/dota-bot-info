package bot

import (
	"errors"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	if query == nil {
		return
	}
	message, err := callbackMessage(query)
	answer := tgbotapi.NewCallback(query.ID, "")
	var limit commandLimit
	if err != nil {
		answer.Text = "Кнопка недоступна. Выполни команду заново через /help."
	} else {
		limit = b.limitCommand(message)
		if limit.wait > 0 {
			answer.Text = commandLimitText(limit.wait)
		}
	}
	// Always acknowledge callbacks, including throttled ones, before API requests.
	if _, ackErr := b.api.Request(answer); ackErr != nil {
		slog.Error("Failed to answer callback", "error", safeTelegramError(ackErr))
	}
	if err != nil || limit.wait > 0 {
		return
	}
	b.handleCommand(message)
}

// Reuse command validation and handlers without modifying the original message.
func callbackMessage(query *tgbotapi.CallbackQuery) (*tgbotapi.Message, error) {
	if query == nil || query.Message == nil || query.Message.Chat == nil || query.Message.Date == 0 {
		return nil, errors.New("callback message unavailable")
	}
	command, id, ok := strings.Cut(query.Data, ":")
	if !ok || id == "" || len(query.Data) > 64 {
		return nil, errInvalidID
	}
	for _, digit := range id {
		if digit < '0' || digit > '9' {
			return nil, errInvalidID
		}
	}
	message := &tgbotapi.Message{
		Chat:     query.Message.Chat,
		From:     query.From,
		Text:     "/" + command + " " + id,
		Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command) + 1}},
	}
	var err error
	switch command {
	case "meta":
		_, err = parseMetaPosition(id)
	case "player", "matches", "heroes", "stats":
		_, err = parseDotaID(message)
	case "match", "impact":
		_, err = parseMatchID(message)
	default:
		return nil, errors.New("unknown callback command")
	}
	if err != nil {
		return nil, err
	}
	return message, nil
}
