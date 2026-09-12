package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

func (b *Bot) handleStart(message *tgbotapi.Message) {
	text := `👋 Привет!

Я бот для получения информации о Dota 2 игроках.

👤 /player &lt;Dota ID&gt;
⚔️ /matches &lt;Dota ID&gt;
🦸 /heroes &lt;Dota ID&gt;
📈 /stats &lt;Dota ID&gt;
🧭 /meta — мета героев по позициям
📚 /help

Начни с /player и Dota ID. После выбора игрока команды /matches, /heroes и /stats работают без ID.`

	b.sendMessage(
		message.Chat.ID,
		text,
	)
}

func (b *Bot) handleHelp(message *tgbotapi.Message) {
	text := `📚 <b>Доступные команды</b>

/start — запустить бота
/help — помощь

Запросы статистики: до 5 подряд, затем один новый запрос каждые 3 секунды.
Команды и кнопки используют общий лимит. Справка и меню /meta доступны всегда.

🧭 <b>Мета</b>
/meta — выбрать позицию
/meta [1–5] — топ героев Immortal за последние 7 полных дней актуального патча

👤 <b>Игрок</b>
/player [Dota ID] — выбрать игрока или показать выбранного
/matches [Dota ID] — последние матчи
/heroes [Dota ID] — популярные герои
/stats [Dota ID] — расширенная статистика

После успешного /player ID можно не повторять. Выбор личный для каждого участника чата и сохраняется между перезапусками бота.

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
	if b.ctx != nil && b.ctx.Err() != nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)

	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	if len(keyboards) > 0 {
		msg.ReplyMarkup = keyboards[0]
	}

	_, err := b.api.Send(msg)
	if err != nil {
		slog.Error("Failed to send message", "chat_id", chatID, "error", safeTelegramError(err))
		return
	}
}
