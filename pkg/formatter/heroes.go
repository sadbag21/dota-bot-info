package formatter

import (
	"fmt"
	"html"
	"strings"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func FormatHeroes(
	accountID int64,
	heroes []service.HeroStats,
) string {
	if len(heroes) == 0 {
		return fmt.Sprintf(
			"🦸 Статистика героев игрока <code>%d</code> не найдена.",
			accountID,
		)
	}

	var builder strings.Builder

	builder.WriteString(
		fmt.Sprintf(
			"🦸 <b>Самые популярные герои</b>\n"+
				"🆔 <code>%d</code>",
			accountID,
		),
	)

	for i, hero := range heroes {
		heroName := html.EscapeString(
			hero.HeroName,
		)

		builder.WriteString(
			fmt.Sprintf(
				"\n\n"+
					"%d. <b>%s</b>\n"+
					"   🎮 Матчи: %d\n"+
					"   🟢 Победы: %d\n"+
					"   🔴 Поражения: %d\n"+
					"   📈 Winrate: %.2f%%",
				i+1,
				heroName,
				hero.Games,
				hero.Wins,
				hero.Losses,
				hero.WinRate,
			),
		)
	}

	return builder.String()
}
