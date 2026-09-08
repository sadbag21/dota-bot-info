package service

import (
	"math"
	"testing"
)

func TestImpactRanking_StrongCoreBeatsWeakCore(t *testing.T) {
	match := &MatchDetails{
		RadiantWin: true,

		Radiant: []MatchPlayerInfo{
			{
				PlayerSlot: 0,
				Name:       "Strong Core",
				HeroName:   "Phantom Assassin",

				Kills:   20,
				Deaths:  3,
				Assists: 10,

				GPM:      850,
				XPM:      1000,
				NetWorth: 40000,

				HeroDamage:  70000,
				TowerDamage: 12000,
				HeroHealing: 0,

				Stuns:     0,
				ObsPlaced: 0,
				SenPlaced: 0,

				TeamfightParticipation: 0.80,
			},

			{
				PlayerSlot: 1,
				Name:       "Weak Core",
				HeroName:   "Juggernaut",

				Kills:   3,
				Deaths:  12,
				Assists: 4,

				GPM:      350,
				XPM:      450,
				NetWorth: 15000,

				HeroDamage:  12000,
				TowerDamage: 500,
				HeroHealing: 0,

				Stuns:     0,
				ObsPlaced: 0,
				SenPlaced: 0,

				TeamfightParticipation: 0.25,
			},
		},
	}

	ranking := calculateImpactRanking(match)

	if len(ranking) != 2 {
		t.Fatalf(
			"expected 2 players in ranking, got %d",
			len(ranking),
		)
	}

	if ranking[0].Name != "Strong Core" {
		t.Errorf(
			"expected Strong Core to be first, got %s",
			ranking[0].Name,
		)
	}

	if ranking[0].Score <= ranking[1].Score {
		t.Errorf(
			"expected Strong Core score %.2f to be greater than Weak Core score %.2f",
			ranking[0].Score,
			ranking[1].Score,
		)
	}
}

func TestImpactRanking_SupportContributionIsRewarded(t *testing.T) {
	match := &MatchDetails{
		RadiantWin: true,

		Radiant: []MatchPlayerInfo{
			{
				PlayerSlot: 0,
				Name:       "Good Support",
				HeroName:   "Disruptor",

				Kills:   2,
				Deaths:  5,
				Assists: 28,

				GPM:      320,
				XPM:      500,
				NetWorth: 13000,

				HeroDamage:  16000,
				TowerDamage: 200,
				HeroHealing: 1500,

				Stuns: 75,

				ObsPlaced: 12,
				SenPlaced: 22,

				TeamfightParticipation: 0.88,
			},

			{
				PlayerSlot: 1,
				Name:       "Low Impact",
				HeroName:   "Sniper",

				Kills:   1,
				Deaths:  14,
				Assists: 2,

				GPM:      360,
				XPM:      420,
				NetWorth: 14000,

				HeroDamage:  8000,
				TowerDamage: 100,
				HeroHealing: 0,

				Stuns: 0,

				ObsPlaced: 0,
				SenPlaced: 0,

				TeamfightParticipation: 0.10,
			},
		},
	}

	ranking := calculateImpactRanking(match)

	if len(ranking) != 2 {
		t.Fatalf(
			"expected 2 players, got %d",
			len(ranking),
		)
	}

	if ranking[0].Name != "Good Support" {
		t.Errorf(
			"expected support contribution to be rewarded, first place is %s",
			ranking[0].Name,
		)
	}

	var support ImpactEntry

	for _, player := range ranking {
		if player.Name == "Good Support" {
			support = player
			break
		}
	}

	if support.StunsContribution <= 0 {
		t.Error("expected support to receive stun contribution")
	}

	if support.WardsContribution <= 0 {
		t.Error("expected support to receive ward contribution")
	}

	if support.ParticipationContribution <= 0 {
		t.Error("expected support to receive participation contribution")
	}
}

func TestImpactRanking_WinBonusIsFivePoints(t *testing.T) {
	match := &MatchDetails{
		RadiantWin: true,

		Radiant: []MatchPlayerInfo{
			{
				PlayerSlot: 0,
				Name:       "Radiant Player",
				HeroName:   "Axe",

				Kills:   10,
				Deaths:  5,
				Assists: 10,

				GPM:      500,
				XPM:      600,
				NetWorth: 20000,

				HeroDamage:  20000,
				TowerDamage: 1000,

				Stuns: 10,

				TeamfightParticipation: 0.5,
			},
		},

		Dire: []MatchPlayerInfo{
			{
				PlayerSlot: 128,
				Name:       "Dire Player",
				HeroName:   "Axe",

				Kills:   10,
				Deaths:  5,
				Assists: 10,

				GPM:      500,
				XPM:      600,
				NetWorth: 20000,

				HeroDamage:  20000,
				TowerDamage: 1000,

				Stuns: 10,

				TeamfightParticipation: 0.5,
			},
		},
	}

	ranking := calculateImpactRanking(match)

	if len(ranking) != 2 {
		t.Fatalf(
			"expected 2 players, got %d",
			len(ranking),
		)
	}

	var radiant ImpactEntry
	var dire ImpactEntry

	for _, player := range ranking {
		switch player.Name {
		case "Radiant Player":
			radiant = player

		case "Dire Player":
			dire = player
		}
	}

	if !almostEqual(radiant.WinContribution, 5) {
		t.Errorf(
			"expected winner contribution 5, got %.2f",
			radiant.WinContribution,
		)
	}

	if !almostEqual(dire.WinContribution, 0) {
		t.Errorf(
			"expected loser contribution 0, got %.2f",
			dire.WinContribution,
		)
	}

	diff := radiant.Score - dire.Score

	if !almostEqual(diff, 5) {
		t.Errorf(
			"expected score difference of 5 points, got %.2f",
			diff,
		)
	}
}

func TestImpactRanking_IsSortedDescending(t *testing.T) {
	match := sampleImpactMatch()

	ranking := calculateImpactRanking(match)

	if len(ranking) != 10 {
		t.Fatalf(
			"expected 10 players, got %d",
			len(ranking),
		)
	}

	for i := 1; i < len(ranking); i++ {
		if ranking[i-1].Score < ranking[i].Score {
			t.Errorf(
				"ranking is not sorted: position %d has %.2f, position %d has %.2f",
				i,
				ranking[i-1].Score,
				i+1,
				ranking[i].Score,
			)
		}
	}
}

func TestImpactRanking_ScoreIsBetweenZeroAndHundred(t *testing.T) {
	match := sampleImpactMatch()

	ranking := calculateImpactRanking(match)

	for _, player := range ranking {
		if player.Score < 0 || player.Score > 100 {
			t.Errorf(
				"%s has invalid score %.2f",
				player.Name,
				player.Score,
			)
		}
	}
}

func TestCalculateParticipation_UsesOpenDotaValue(t *testing.T) {
	player := MatchPlayerInfo{
		TeamfightParticipation: 0.73,
	}

	participation := calculateParticipation(
		player,
		nil,
	)

	if !almostEqual(participation, 0.73) {
		t.Errorf(
			"expected 0.73, got %.4f",
			participation,
		)
	}
}

func TestCalculateParticipation_Fallback(t *testing.T) {
	player := MatchPlayerInfo{
		Kills:   2,
		Assists: 8,
	}

	team := []MatchPlayerInfo{
		{
			Kills: 8,
		},
		{
			Kills: 5,
		},
		{
			Kills: 4,
		},
		{
			Kills: 2,
		},
		{
			Kills: 1,
		},
	}

	participation := calculateParticipation(
		player,
		team,
	)

	// (2 kills + 8 assists) / 20 team kills = 0.5
	if !almostEqual(participation, 0.5) {
		t.Errorf(
			"expected participation 0.5, got %.4f",
			participation,
		)
	}
}

func TestImpactProfile(t *testing.T) {
	tests := []struct {
		name          string
		supportFactor float64
		expected      string
	}{
		{
			name:          "core",
			supportFactor: 0.20,
			expected:      "Core",
		},
		{
			name:          "hybrid",
			supportFactor: 0.45,
			expected:      "Hybrid",
		},
		{
			name:          "support",
			supportFactor: 0.75,
			expected:      "Support",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				result := impactProfile(
					tt.supportFactor,
				)

				if result != tt.expected {
					t.Errorf(
						"expected %s, got %s",
						tt.expected,
						result,
					)
				}
			},
		)
	}
}

func almostEqual(a, b float64) bool {
	const epsilon = 0.0001

	return math.Abs(a-b) < epsilon
}

func sampleImpactMatch() *MatchDetails {
	return &MatchDetails{
		RadiantWin: true,

		Radiant: []MatchPlayerInfo{
			{
				PlayerSlot:             0,
				Name:                   "Radiant Carry",
				HeroName:               "Phantom Assassin",
				Kills:                  18,
				Deaths:                 4,
				Assists:                8,
				GPM:                    800,
				XPM:                    950,
				NetWorth:               38000,
				HeroDamage:             65000,
				TowerDamage:            9000,
				TeamfightParticipation: 0.70,
			},
			{
				PlayerSlot:             1,
				Name:                   "Radiant Mid",
				HeroName:               "Puck",
				Kills:                  12,
				Deaths:                 5,
				Assists:                16,
				GPM:                    650,
				XPM:                    850,
				NetWorth:               30000,
				HeroDamage:             50000,
				TowerDamage:            3000,
				Stuns:                  15,
				TeamfightParticipation: 0.80,
			},
			{
				PlayerSlot:             2,
				Name:                   "Radiant Offlane",
				HeroName:               "Axe",
				Kills:                  8,
				Deaths:                 7,
				Assists:                20,
				GPM:                    500,
				XPM:                    700,
				NetWorth:               24000,
				HeroDamage:             35000,
				TowerDamage:            1500,
				Stuns:                  40,
				TeamfightParticipation: 0.82,
			},
			{
				PlayerSlot:             3,
				Name:                   "Radiant Support",
				HeroName:               "Rubick",
				Kills:                  4,
				Deaths:                 8,
				Assists:                25,
				GPM:                    370,
				XPM:                    570,
				NetWorth:               16000,
				HeroDamage:             22000,
				TowerDamage:            300,
				Stuns:                  55,
				ObsPlaced:              8,
				SenPlaced:              12,
				TeamfightParticipation: 0.86,
			},
			{
				PlayerSlot:             4,
				Name:                   "Radiant Hard Support",
				HeroName:               "Dazzle",
				Kills:                  2,
				Deaths:                 6,
				Assists:                30,
				GPM:                    310,
				XPM:                    500,
				NetWorth:               13000,
				HeroDamage:             15000,
				HeroHealing:            18000,
				Stuns:                  5,
				ObsPlaced:              14,
				SenPlaced:              20,
				TeamfightParticipation: 0.90,
			},
		},

		Dire: []MatchPlayerInfo{
			{
				PlayerSlot:             128,
				Name:                   "Dire Carry",
				HeroName:               "Juggernaut",
				Kills:                  11,
				Deaths:                 8,
				Assists:                5,
				GPM:                    600,
				XPM:                    700,
				NetWorth:               26000,
				HeroDamage:             40000,
				TowerDamage:            4000,
				TeamfightParticipation: 0.50,
			},
			{
				PlayerSlot:             129,
				Name:                   "Dire Mid",
				HeroName:               "Invoker",
				Kills:                  9,
				Deaths:                 9,
				Assists:                12,
				GPM:                    550,
				XPM:                    680,
				NetWorth:               24000,
				HeroDamage:             38000,
				TowerDamage:            1200,
				Stuns:                  20,
				TeamfightParticipation: 0.65,
			},
			{
				PlayerSlot:             130,
				Name:                   "Dire Offlane",
				HeroName:               "Centaur Warrunner",
				Kills:                  5,
				Deaths:                 10,
				Assists:                15,
				GPM:                    430,
				XPM:                    560,
				NetWorth:               19000,
				HeroDamage:             28000,
				TowerDamage:            600,
				Stuns:                  35,
				TeamfightParticipation: 0.62,
			},
			{
				PlayerSlot:             131,
				Name:                   "Dire Support",
				HeroName:               "Lion",
				Kills:                  3,
				Deaths:                 12,
				Assists:                17,
				GPM:                    300,
				XPM:                    450,
				NetWorth:               12000,
				HeroDamage:             17000,
				Stuns:                  50,
				ObsPlaced:              6,
				SenPlaced:              10,
				TeamfightParticipation: 0.60,
			},
			{
				PlayerSlot:             132,
				Name:                   "Dire Hard Support",
				HeroName:               "Crystal Maiden",
				Kills:                  1,
				Deaths:                 14,
				Assists:                10,
				GPM:                    250,
				XPM:                    350,
				NetWorth:               9000,
				HeroDamage:             9000,
				Stuns:                  15,
				ObsPlaced:              10,
				SenPlaced:              15,
				TeamfightParticipation: 0.40,
			},
		},
	}
}
