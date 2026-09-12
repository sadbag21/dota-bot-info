package formatter

import (
	"fmt"
	"html"
	"strings"

	"github.com/sadbag21/dota-bot-info/internal/meta"
)

func FormatMeta(report meta.Report) string {
	var text strings.Builder
	// strings.Builder writes always succeed; there is no external I/O here.
	_, _ = fmt.Fprintf(&text, "🧭 <b>Мета — %d: %s</b>\nПатч <b>%s</b> · Immortal · Ranked All Pick\n", report.Position, meta.PositionName(report.Position), html.EscapeString(report.Patch))
	if !report.From.IsZero() && !report.Through.IsZero() {
		_, _ = fmt.Fprintf(&text, "Полные дни: %s — %s (UTC)\n", report.From.Format("02.01.2006"), report.Through.Format("02.01.2006"))
	}
	if len(report.Heroes) == 0 {
		_, _ = fmt.Fprintf(&text, "\nПока недостаточно данных: нужно минимум %d матчей на героя и позицию. После выхода патча ждём полные дни статистики.\n", meta.MinMatches)
	} else {
		_, _ = text.WriteString("\n")
		for i, hero := range report.Heroes {
			if i >= 10 {
				break
			}
			winrate := float64(0)
			if hero.Matches > 0 {
				winrate = 100 * float64(hero.Wins) / float64(hero.Matches)
			}
			_, _ = fmt.Fprintf(&text, "%d. <b>%s</b> — %.1f%% · %d матчей\n", i+1, html.EscapeString(hero.Name), winrate, hero.Matches)
		}
		_, _ = fmt.Fprintf(&text, "\nМинимум %d матчей. Порядок учитывает винрейт и размер выборки; это рейтинг бота.\n", meta.MinMatches)
	}
	_, _ = fmt.Fprintf(&text, "\nИсточник: STRATZ. Патч: Valve.\nДанные получены: %s UTC.", report.FetchedAt.UTC().Format("02.01 15:04"))
	return text.String()
}
