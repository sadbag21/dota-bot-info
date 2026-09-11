package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := `👋 Привет!

Я бот для получения информации о Dota 2 игроках.

👤 /player &lt;Dota ID&gt;
⚔️ /matches &lt;Dota ID&gt;
🦸 /heroes &lt;Dota ID&gt;
📈 /stats &lt;Dota ID&gt;
📚 /help`

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := `📚 <b>Доступные команды</b>

/start — запустить бота
/help — помощь

👤 <b>Игрок</b>
/player &lt;Dota ID&gt; — информация об игроке
/matches &lt;Dota ID&gt; — последние матчи
/heroes &lt;Dota ID&gt; — популярные герои
/stats &lt;Dota ID&gt; — расширенная статистика

🏟 <b>Матч</b>
/match &lt;Match ID&gt; — подробности матча
/impact &lt;Match ID&gt; — оценка полезности игроков

Под профилем игрока есть кнопки перехода к матчам, героям и статистике.
Под матчем — кнопки Impact и OpenDota. Повторно вводить ID не нужно.`

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

func (b *Bot) sendMessage(chatID int64, text string, keyboards ...tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)

	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	if len(keyboards) > 0 {
		msg.ReplyMarkup = keyboards[0]
	}

	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Failed to send message to chat_id=%d: %v", chatID, err)
		return
	}
}
