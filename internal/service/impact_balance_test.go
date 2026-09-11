package service

import (
	"math"
	"testing"
)

func TestImpactRankingDominantSupportStaysWithinScale(t *testing.T) {
	match := &MatchDetails{RadiantWin: true, Radiant: []MatchPlayerInfo{
		{PlayerSlot: 0, Name: "Dominant support", Kills: 20, Deaths: 1, Assists: 30, GPM: 300, XPM: 800, NetWorth: 15000, HeroDamage: 50000, TowerDamage: 5000, HeroHealing: 20000, Stuns: 100, ObsPlaced: 20, SenPlaced: 30, TeamfightParticipation: 1},
	}, Dire: []MatchPlayerInfo{
		{PlayerSlot: 128, Name: "Farm only", Deaths: 10, GPM: 700, XPM: 600, NetWorth: 30000},
	}}
	ranking := calculateImpactRanking(match)
	for _, p := range ranking {
		if math.IsNaN(p.Score) || math.IsInf(p.Score, 0) || p.Score < 0 || p.Score > 100 {
			t.Errorf("%s has score %.6f outside 0..100", p.Name, p.Score)
		}
	}
}

func TestImpactRankingContributionSum(t *testing.T) {
	for _, p := range calculateImpactRanking(sampleImpactMatch()) {
		sum := p.KDAContribution + p.AssistsContribution + p.HeroDamageContribution + p.TowerDamageContribution + p.HealingContribution + p.StunsContribution + p.WardsContribution + p.ParticipationContribution + p.GPMContribution + p.XPMContribution + p.NetWorthContribution + p.SurvivalContribution + p.WinContribution
		if !almostEqual(sum, p.Score) {
			t.Fatalf("contributions %.6f differ from score %.6f", sum, p.Score)
		}
	}
}

func TestImpactWeightsPreserveScale(t *testing.T) {
	core, support := defaultImpactWeights()
	for _, factor := range []float64{0, .25, .5, .75, 1} {
		pairs := [][2]float64{{core.KDA, support.KDA}, {core.Assists, support.Assists}, {core.HeroDamage, support.HeroDamage}, {core.TowerDamage, support.TowerDamage}, {core.Healing, support.Healing}, {core.Stuns, support.Stuns}, {core.Wards, support.Wards}, {core.Participation, support.Participation}, {core.GPM, support.GPM}, {core.XPM, support.XPM}, {core.NetWorth, support.NetWorth}, {core.Survival, support.Survival}, {core.Win, support.Win}}
		sum := 0.0
		for _, pair := range pairs {
			weight := blendWeight(pair[0], pair[1], factor)
			if weight < 0 {
				t.Fatal("negative weight")
			}
			sum += weight
		}
		if !almostEqual(sum, 1) {
			t.Errorf("factor %.2f has total weight %.6f", factor, sum)
		}
		if !almostEqual(blendWeight(core.Win, support.Win, factor), .05) {
			t.Error("win bonus changed")
		}
	}
}
