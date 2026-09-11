package bot

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func playerKeyboard(accountID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(accountID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️ Матчи", "matches:"+id),
			tgbotapi.NewInlineKeyboardButtonData("🦸 Герои", "heroes:"+id),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 Статистика", "stats:"+id),
			tgbotapi.NewInlineKeyboardButtonData("👤 Профиль", "player:"+id),
		),
	)
}

func matchKeyboard(matchID int64) tgbotapi.InlineKeyboardMarkup {
	id := strconv.FormatInt(matchID, 10)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Impact", "impact:"+id),
			tgbotapi.NewInlineKeyboardButtonData("🏟 Матч", "match:"+id),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 OpenDota", "https://www.opendota.com/matches/"+id),
		),
	)
}
