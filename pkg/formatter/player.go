package formatter

import (
	"fmt"
	"html"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func FormatPlayer(info *service.PlayerInfo) string {
	player := info.Player

	name := player.Profile.Personaname

	if name == "" {
		name = player.Profile.Name
	}

	if name == "" {
		name = "Unknown player"
	}

	name = html.EscapeString(name)

	rank := formatRank(player.RankTier)

	text := fmt.Sprintf(
		"🎮 <b>%s</b>\n\n"+
			"🆔 Dota ID: <code>%d</code>\n"+
			"🏆 Rank: %s",
		name,
		player.Profile.AccountID,
		rank,
	)

	if player.LeaderboardRank != nil {
		text += fmt.Sprintf(
			"\n🏅 Leaderboard: #%d",
			*player.LeaderboardRank,
		)
	}

	text += fmt.Sprintf(
		"\n\n"+
			"📊 <b>Статистика</b>\n"+
			"├ Матчи: %d\n"+
			"├ Победы: %d\n"+
			"├ Поражения: %d\n"+
			"└ Winrate: %.2f%%",
		info.Matches,
		info.Wins,
		info.Losses,
		info.WinRate,
	)

	if len(info.RecentMatches) > 0 {
		text += "\n\n⚔️ <b>Последние матчи</b>"

		for i, match := range info.RecentMatches {
			result := "🔴"

			if match.Won {
				result = "🟢"
			}

			heroName := html.EscapeString(
				match.HeroName,
			)

			text += fmt.Sprintf(
				"\n\n%d. %s <b>%s</b>\n"+
					"   %d / %d / %d • %s",
				i+1,
				result,
				heroName,
				match.Kills,
				match.Deaths,
				match.Assists,
				formatDuration(match.Duration),
			)
		}
	}

	if player.Profile.ProfileURL != "" {
		text += fmt.Sprintf(
			"\n\n🔗 <a href=\"%s\">Steam profile</a>",
			html.EscapeString(
				player.Profile.ProfileURL,
			),
		)
	}

	return text
}

func formatRank(rankTier int) string {
	if rankTier == 0 {
		return "Unknown"
	}

	medal := rankTier / 10
	star := rankTier % 10

	if medal == 8 {
		return "Immortal"
	}

	medals := map[int]string{
		1: "Herald",
		2: "Guardian",
		3: "Crusader",
		4: "Archon",
		5: "Legend",
		6: "Ancient",
		7: "Divine",
	}

	medalName, ok := medals[medal]
	if !ok {
		return "Unknown"
	}

	if star == 0 {
		return medalName
	}

	return fmt.Sprintf(
		"%s %d",
		medalName,
		star,
	)
}

func formatDuration(seconds int) string {
	minutes := seconds / 60
	remainingSeconds := seconds % 60

	return fmt.Sprintf(
		"%d:%02d",
		minutes,
		remainingSeconds,
	)
}
