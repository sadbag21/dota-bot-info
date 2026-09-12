package formatter

import (
	"github.com/sadbag21/dota-bot-info/internal/meta"
	"strings"
	"testing"
	"time"
)

func TestFormatMeta(t *testing.T) {
	report := meta.Report{Position: 4, Patch: "7.41e", From: time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC), Through: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC), FetchedAt: time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC), Heroes: []meta.Hero{{Name: "A<&", Wins: 110, Matches: 200}}}
	text := FormatMeta(report)
	for _, part := range []string{"7.41e", "Immortal", "04.09.2026 — 10.09.2026", "A&lt;&amp;", "55.0%", "200 матчей", "STRATZ", "11.09 12:00 UTC"} {
		if !strings.Contains(text, part) {
			t.Errorf("missing %q in %s", part, text)
		}
	}
	report.Heroes = nil
	if text := FormatMeta(report); !strings.Contains(text, "недостаточно данных") || strings.Contains(text, "1. <b>") {
		t.Fatal(text)
	}
}
