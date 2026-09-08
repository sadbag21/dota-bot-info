package formatter

import (
	"fmt"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func FormatStats(
	accountID int64,
	stats *service.PlayerStats,
) string {
	text := fmt.Sprintf(
		"📈 <b>Расширенная статистика</b>\n"+
			"🆔 <code>%d</code>\n\n"+
			"🎮 Матчей в выборке: %d\n\n"+
			"⚔️ <b>K / D / A</b>\n"+
			"├ Kills: %.2f\n"+
			"├ Deaths: %.2f\n"+
			"├ Assists: %.2f\n"+
			"└ KDA: %.2f\n\n"+
			"💰 <b>Экономика</b>\n"+
			"├ GPM: %.0f\n"+
			"├ XPM: %.0f\n"+
			"└ Last hits: %.0f\n\n"+
			"💥 <b>Урон</b>\n"+
			"├ Героям: %.0f\n"+
			"├ Башням: %.0f\n"+
			"└ Лечение: %.0f\n\n"+
			"⏱ Средняя длительность: %s",
		accountID,
		stats.MatchesAnalyzed,

		stats.AvgKills,
		stats.AvgDeaths,
		stats.AvgAssists,
		stats.KDA,

		stats.AvgGPM,
		stats.AvgXPM,
		stats.AvgLastHits,

		stats.AvgHeroDamage,
		stats.AvgTowerDamage,
		stats.AvgHeroHealing,

		formatDuration(
			int(stats.AvgDuration),
		),
	)

	return text
}
