package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := `Привет! 👋

Я бот для получения информации о Dota 2 игроках.

Используй:

/player <Dota ID> — информация об игроке
/help — список команд`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := `📚 Доступные команды:

/start — запустить бота
/help — помощь
/player <Dota ID> — информация об игроке`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true

	_, err := b.api.Send(msg)
	if err != nil {
		return
	}
}
