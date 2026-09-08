package formatter

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func FormatMatch(
	match *service.MatchDetails,
) string {
	var builder strings.Builder

	winner := "🔴 Dire"

	if match.RadiantWin {
		winner = "🟢 Radiant"
	}

	builder.WriteString(
		fmt.Sprintf(
			"🏟 <b>Матч %d</b>\n"+
				"⏱ Длительность: %s\n"+
				"🏆 Победитель: %s",
			match.MatchID,
			formatDuration(match.Duration),
			winner,
		),
	)

	builder.WriteString(
		"\n\n🟢 <b>RADIANT</b>",
	)

	for i, player := range match.Radiant {
		builder.WriteString(
			formatMatchPlayer(
				i+1,
				player,
			),
		)
	}

	builder.WriteString(
		"\n\n🔴 <b>DIRE</b>",
	)

	for i, player := range match.Dire {
		builder.WriteString(
			formatMatchPlayer(
				i+6,
				player,
			),
		)
	}

	matchURL := fmt.Sprintf(
		"https://www.opendota.com/matches/%d",
		match.MatchID,
	)

	builder.WriteString(
		fmt.Sprintf(
			"\n\n🔗 <a href=\"%s\">Открыть матч в OpenDota</a>",
			matchURL,
		),
	)

	return builder.String()
}

func formatMatchPlayer(
	number int,
	player service.MatchPlayerInfo,
) string {
	name := html.EscapeString(
		player.Name,
	)

	hero := html.EscapeString(
		player.HeroName,
	)

	accountID := ""

	if player.AccountID != nil {
		accountID = " • <code>" +
			strconv.FormatInt(
				*player.AccountID,
				10,
			) +
			"</code>"
	}

	return fmt.Sprintf(
		"\n\n"+
			"%d. <b>%s</b> — %s%s\n"+
			"   ⚔️ %d / %d / %d\n"+
			"   💰 GPM/XPM: %d / %d\n"+
			"   💵 NW: %s\n"+
			"   💥 Damage: %s",
		number,
		hero,
		name,
		accountID,

		player.Kills,
		player.Deaths,
		player.Assists,

		player.GPM,
		player.XPM,

		formatNumber(player.NetWorth),
		formatNumber(player.HeroDamage),
	)
}

func formatNumber(number int) string {
	s := strconv.Itoa(number)

	n := len(s)

	if n <= 3 {
		return s
	}

	var builder strings.Builder

	first := n % 3

	if first > 0 {
		builder.WriteString(
			s[:first],
		)

		if first < n {
			builder.WriteString(" ")
		}
	}

	for i := first; i < n; i += 3 {
		builder.WriteString(
			s[i : i+3],
		)

		if i+3 < n {
			builder.WriteString(" ")
		}
	}

	return builder.String()
}
