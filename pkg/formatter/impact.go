package formatter

import (
	"fmt"
	"html"
	"strings"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func FormatImpact(match *service.MatchDetails) string {
	if len(match.ImpactRanking) == 0 {
		return fmt.Sprintf(
			"❌ Impact Score для матча <code>%d</code> недоступен.",
			match.MatchID,
		)
	}

	var builder strings.Builder

	winner := "🔴 Dire"

	if match.RadiantWin {
		winner = "🟢 Radiant"
	}

	builder.WriteString(
		fmt.Sprintf(
			"📊 <b>Impact analysis</b>\n"+
				"🏟 Матч: <code>%d</code>\n"+
				"🏆 Победитель: %s\n\n"+
				"<i>Оценка 0–100 является внутренней метрикой бота.</i>",
			match.MatchID,
			winner,
		),
	)

	mvp := match.ImpactRanking[0]
	lowest := match.ImpactRanking[len(match.ImpactRanking)-1]

	builder.WriteString(
		formatImpactDetails(
			"⭐ MVP",
			mvp,
		),
	)

	builder.WriteString(
		formatImpactDetails(
			"💀 Lowest impact",
			lowest,
		),
	)

	builder.WriteString(
		"\n\n🏁 <b>Полный рейтинг</b>",
	)

	for i, player := range match.ImpactRanking {
		icon := "▫️"

		switch i {
		case 0:
			icon = "🥇"
		case 1:
			icon = "🥈"
		case 2:
			icon = "🥉"
		}

		result := "❌"

		if player.Won {
			result = "✅"
		}

		builder.WriteString(
			fmt.Sprintf(
				"\n%s %d. %s — %s — <b>%.1f</b> [%s] %s",
				icon,
				i+1,
				html.EscapeString(
					player.HeroName,
				),
				html.EscapeString(
					player.Name,
				),
				player.Score,
				player.Profile,
				result,
			),
		)
	}

	return builder.String()
}

func formatImpactDetails(
	title string,
	player service.ImpactEntry,
) string {
	return fmt.Sprintf(
		"\n\n%s: <b>%s</b> — %s\n"+
			"🧩 Профиль: <b>%s</b>\n"+
			"🎯 Итог: <b>%.1f / 100</b>\n\n"+
			"├ ⚔️ KDA: +%.1f\n"+
			"├ 🤝 Assists: +%.1f\n"+
			"├ 💥 Hero Damage: +%.1f\n"+
			"├ 🏰 Tower Damage: +%.1f\n"+
			"├ ❤️ Healing: +%.1f\n"+
			"├ 💫 Stuns: +%.1f\n"+
			"├ 👁 Wards: +%.1f\n"+
			"├ 👥 Participation: +%.1f\n"+
			"├ 💰 GPM: +%.1f\n"+
			"├ ✨ XPM: +%.1f\n"+
			"├ 💵 Net Worth: +%.1f\n"+
			"├ 🛡 Survival: +%.1f\n"+
			"└ 🏆 Win: +%.1f",
		title,

		html.EscapeString(
			player.HeroName,
		),

		html.EscapeString(
			player.Name,
		),

		player.Profile,
		player.Score,

		player.KDAContribution,
		player.AssistsContribution,
		player.HeroDamageContribution,
		player.TowerDamageContribution,
		player.HealingContribution,
		player.StunsContribution,
		player.WardsContribution,
		player.ParticipationContribution,
		player.GPMContribution,
		player.XPMContribution,
		player.NetWorthContribution,
		player.SurvivalContribution,
		player.WinContribution,
	)
}
