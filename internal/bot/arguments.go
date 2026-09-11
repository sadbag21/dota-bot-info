package bot

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) getDotaID(
	message *tgbotapi.Message,
	command string,
) (int64, bool) {
	accountID, err := b.resolveDotaID(
		message,
	)

	if err == nil {
		return accountID, true
	}

	if errors.Is(err, errMissingID) {
		b.sendMessage(
			message.Chat.ID,
			"❌ Сначала выбери игрока через /player или укажи Dota ID в команде.\n\n"+
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
