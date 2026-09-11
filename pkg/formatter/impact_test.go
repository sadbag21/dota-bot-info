package formatter

import (
	"strings"
	"testing"

	"github.com/sadbag21/dota-bot-info/internal/service"
)

func TestImpactDataQualityNotice(t *testing.T) {
	for _, format := range []func(*service.MatchDetails) string{FormatImpact, FormatMatch} {
		match := &service.MatchDetails{MatchID: 123, ImpactLimited: true, ImpactRanking: []service.ImpactEntry{{PlayerSlot: 0, Name: "Example", HeroName: "Hero", Score: 50}}}
		text := format(match)
		if !strings.Contains(text, "Предварительная оценка") || !strings.Contains(text, "50.0") {
			t.Fatal("partial data must show both estimate and its limitation")
		}
		match.ImpactLimited = false
		if strings.Contains(format(match), "Предварительная оценка") {
			t.Fatal("parsed match marked as unparsed")
		}
	}
}
