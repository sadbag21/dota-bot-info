package service

import "sort"

type ImpactEntry struct {
	PlayerSlot int
	Name       string
	HeroName   string
	Score      float64
	Won        bool
}

type impactCandidate struct {
	Player MatchPlayerInfo
	Won    bool
	KDA    float64
}

func calculateImpactRanking(match *MatchDetails) []ImpactEntry {
	candidates := make(
		[]impactCandidate,
		0,
		len(match.Radiant)+len(match.Dire),
	)

	for _, player := range match.Radiant {
		candidates = append(
			candidates,
			impactCandidate{
				Player: player,
				Won:    match.RadiantWin,
				KDA:    calculateKDA(player),
			},
		)
	}

	for _, player := range match.Dire {
		candidates = append(
			candidates,
			impactCandidate{
				Player: player,
				Won:    !match.RadiantWin,
				KDA:    calculateKDA(player),
			},
		)
	}

	if len(candidates) == 0 {
		return nil
	}

	kdaMin, kdaMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return c.KDA
		},
	)

	assistsMin, assistsMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.Assists)
		},
	)

	heroDamageMin, heroDamageMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.HeroDamage)
		},
	)

	towerDamageMin, towerDamageMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.TowerDamage)
		},
	)

	healingMin, healingMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.HeroHealing)
		},
	)

	gpmMin, gpmMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.GPM)
		},
	)

	xpmMin, xpmMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.XPM)
		},
	)

	netWorthMin, netWorthMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.NetWorth)
		},
	)

	deathsMin, deathsMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return float64(c.Player.Deaths)
		},
	)

	ranking := make(
		[]ImpactEntry,
		0,
		len(candidates),
	)

	for _, candidate := range candidates {
		player := candidate.Player

		kdaScore := normalize(
			candidate.KDA,
			kdaMin,
			kdaMax,
		)

		assistsScore := normalize(
			float64(player.Assists),
			assistsMin,
			assistsMax,
		)

		heroDamageScore := normalize(
			float64(player.HeroDamage),
			heroDamageMin,
			heroDamageMax,
		)

		towerDamageScore := normalize(
			float64(player.TowerDamage),
			towerDamageMin,
			towerDamageMax,
		)

		healingScore := normalize(
			float64(player.HeroHealing),
			healingMin,
			healingMax,
		)

		gpmScore := normalize(
			float64(player.GPM),
			gpmMin,
			gpmMax,
		)

		xpmScore := normalize(
			float64(player.XPM),
			xpmMin,
			xpmMax,
		)

		netWorthScore := normalize(
			float64(player.NetWorth),
			netWorthMin,
			netWorthMax,
		)

		// Чем меньше смертей, тем выше значение.
		survivalScore := 1 - normalize(
			float64(player.Deaths),
			deathsMin,
			deathsMax,
		)

		winScore := 0.0

		if candidate.Won {
			winScore = 1.0
		}

		score := kdaScore*0.20 +
			assistsScore*0.10 +
			heroDamageScore*0.20 +
			towerDamageScore*0.12 +
			healingScore*0.10 +
			gpmScore*0.06 +
			xpmScore*0.05 +
			netWorthScore*0.05 +
			survivalScore*0.04 +
			winScore*0.08

		score *= 100

		ranking = append(
			ranking,
			ImpactEntry{
				PlayerSlot: player.PlayerSlot,
				Name:       player.Name,
				HeroName:   player.HeroName,
				Score:      score,
				Won:        candidate.Won,
			},
		)
	}

	sort.Slice(
		ranking,
		func(i, j int) bool {
			return ranking[i].Score > ranking[j].Score
		},
	)

	return ranking
}

func calculateKDA(player MatchPlayerInfo) float64 {
	deaths := player.Deaths

	if deaths <= 0 {
		deaths = 1
	}

	return float64(player.Kills+player.Assists) / float64(deaths)
}

func normalize(
	value float64,
	min float64,
	max float64,
) float64 {
	if max == min {
		return 0.5
	}

	return (value - min) / (max - min)
}

func minMax(
	candidates []impactCandidate,
	value func(impactCandidate) float64,
) (float64, float64) {
	minValue := value(candidates[0])
	maxValue := minValue

	for _, candidate := range candidates[1:] {
		current := value(candidate)

		if current < minValue {
			minValue = current
		}

		if current > maxValue {
			maxValue = current
		}
	}

	return minValue, maxValue
}
