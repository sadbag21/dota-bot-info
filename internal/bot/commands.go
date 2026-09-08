package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := `👋 Привет!

Я бот для получения информации о Dota 2 игроках.

Основные команды:

👤 /player &lt;Dota ID&gt;
⚔️ /matches &lt;Dota ID&gt;
📚 /help`

	b.sendMessage(message.Chat.ID, text)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := `📚 <b>Доступные команды</b>

/start — запустить бота
/help — помощь

/player &lt;Dota ID&gt; — информация об игроке
/matches &lt;Dota ID&gt; — последние матчи
/heroes &lt;Dota ID&gt; — популярные герои`

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
