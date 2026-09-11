package service

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/sadbag21/dota-bot-info/internal/dota"
)

func TestImpactRealMatchFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/impact/matches.json")
	if err != nil {
		t.Fatal(err)
	}
	var matches []dota.Match
	if err := json.Unmarshal(data, &matches); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("testdata/impact/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline []struct {
		MatchID int64
		Ranking []ImpactEntry
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if len(matches) != 15 || len(baseline) != 15 {
		t.Fatal("expected 15 real match fixtures")
	}
	for index, match := range matches {
		details := buildMatchDetails(&match, nil)
		if match.MatchID != baseline[index].MatchID {
			t.Fatal("fixture IDs do not match baseline")
		}
		if len(details.ImpactRanking) != 10 {
			t.Fatalf("match %d: missing players", match.MatchID)
		}
		if details.ImpactLimited != (match.Version == nil) {
			t.Fatalf("match %d: parse status lost", match.MatchID)
		}
		for i, p := range details.ImpactRanking {
			expected := baseline[index].Ranking[i]
			if p.PlayerSlot != expected.PlayerSlot || p.Profile != expected.Profile || !almostEqual(p.Score, expected.Score) {
				t.Errorf("match %d rank %d changed: got slot %d score %.6f profile %s", match.MatchID, i, p.PlayerSlot, p.Score, p.Profile)
			}
			if math.IsNaN(p.Score) || math.IsInf(p.Score, 0) || p.Score < 0 || p.Score > 100 {
				t.Errorf("match %d: invalid score", match.MatchID)
			}
			sum := p.KDAContribution + p.AssistsContribution + p.HeroDamageContribution + p.TowerDamageContribution + p.HealingContribution + p.StunsContribution + p.WardsContribution + p.ParticipationContribution + p.GPMContribution + p.XPMContribution + p.NetWorthContribution + p.SurvivalContribution + p.WinContribution
			if !almostEqual(sum, p.Score) {
				t.Errorf("match %d: contribution total differs from score", match.MatchID)
			}
		}
	}
}

func TestImpactParsedZeroVersionAndZeroSupportStats(t *testing.T) {
	version := 0
	match := &dota.Match{MatchID: 1, Version: &version, Players: []dota.MatchPlayer{{PlayerSlot: 0}}}
	if buildMatchDetails(match, nil).ImpactLimited {
		t.Fatal("present zero version and zero stats mistaken for missing data")
	}
}
