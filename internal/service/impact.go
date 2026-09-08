package service

import "sort"

type ImpactEntry struct {
	PlayerSlot int

	Name     string
	HeroName string

	Score float64
	Won   bool

	KDAContribution         float64
	AssistsContribution     float64
	HeroDamageContribution  float64
	TowerDamageContribution float64
	HealingContribution     float64
	GPMContribution         float64
	XPMContribution         float64
	NetWorthContribution    float64
	SurvivalContribution    float64
	WinContribution         float64
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

		kdaContribution := kdaScore * 20
		assistsContribution := assistsScore * 10
		heroDamageContribution := heroDamageScore * 20
		towerDamageContribution := towerDamageScore * 12
		healingContribution := healingScore * 10
		gpmContribution := gpmScore * 6
		xpmContribution := xpmScore * 5
		netWorthContribution := netWorthScore * 5
		survivalContribution := survivalScore * 4
		winContribution := winScore * 8

		score := kdaContribution +
			assistsContribution +
			heroDamageContribution +
			towerDamageContribution +
			healingContribution +
			gpmContribution +
			xpmContribution +
			netWorthContribution +
			survivalContribution +
			winContribution

		ranking = append(
			ranking,
			ImpactEntry{
				PlayerSlot: player.PlayerSlot,

				Name:     player.Name,
				HeroName: player.HeroName,

				Score: score,
				Won:   candidate.Won,

				KDAContribution:         kdaContribution,
				AssistsContribution:     assistsContribution,
				HeroDamageContribution:  heroDamageContribution,
				TowerDamageContribution: towerDamageContribution,
				HealingContribution:     healingContribution,
				GPMContribution:         gpmContribution,
				XPMContribution:         xpmContribution,
				NetWorthContribution:    netWorthContribution,
				SurvivalContribution:    survivalContribution,
				WinContribution:         winContribution,
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
