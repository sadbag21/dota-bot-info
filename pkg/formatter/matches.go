package formatter

import (
	"fmt"
	"html"
	"strings"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func FormatMatches(
	accountID int64,
	matches []service.MatchInfo,
) string {
	if len(matches) == 0 {
		return fmt.Sprintf(
			"⚔️ Матчи игрока <code>%d</code> не найдены.",
			accountID,
		)
	}

	var builder strings.Builder

	builder.WriteString(
		fmt.Sprintf(
			"⚔️ <b>Последние матчи</b>\n"+
				"🆔 <code>%d</code>",
			accountID,
		),
	)

	for i, match := range matches {
		result := "🔴"

		if match.Won {
			result = "🟢"
		}

		heroName := html.EscapeString(
			match.HeroName,
		)

		matchURL := fmt.Sprintf(
			"https://www.opendota.com/matches/%d",
			match.MatchID,
		)

		builder.WriteString(
			fmt.Sprintf(
				"\n\n"+
					"%d. %s <b>%s</b>\n"+
					"   ⚔️ %d / %d / %d\n"+
					"   ⏱ %s\n"+
					"   🔗 <a href=\"%s\">Match %d</a>",
				i+1,
				result,
				heroName,
				match.Kills,
				match.Deaths,
				match.Assists,
				formatDuration(match.Duration),
				matchURL,
				match.MatchID,
			),
		)
	}

	return builder.String()
}
