package service

import "sort"

type ImpactEntry struct {
	PlayerSlot int

	Name     string
	HeroName string

	Score float64
	Won   bool

	Profile       string
	SupportFactor float64

	KDAContribution           float64
	AssistsContribution       float64
	HeroDamageContribution    float64
	TowerDamageContribution   float64
	HealingContribution       float64
	StunsContribution         float64
	WardsContribution         float64
	ParticipationContribution float64
	GPMContribution           float64
	XPMContribution           float64
	NetWorthContribution      float64
	SurvivalContribution      float64
	WinContribution           float64
}

type impactCandidate struct {
	Player MatchPlayerInfo
	Won    bool

	KDA           float64
	Wards         float64
	Participation float64
}

func calculateImpactRanking(
	match *MatchDetails,
) []ImpactEntry {
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

				KDA: calculateKDA(
					player,
				),

				Wards: float64(
					player.ObsPlaced +
						player.SenPlaced,
				),

				Participation: calculateParticipation(
					player,
					match.Radiant,
				),
			},
		)
	}

	for _, player := range match.Dire {
		candidates = append(
			candidates,
			impactCandidate{
				Player: player,
				Won:    !match.RadiantWin,

				KDA: calculateKDA(
					player,
				),

				Wards: float64(
					player.ObsPlaced +
						player.SenPlaced,
				),

				Participation: calculateParticipation(
					player,
					match.Dire,
				),
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

	stunsMin, stunsMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return c.Player.Stuns
		},
	)

	wardsMin, wardsMax := minMax(
		candidates,
		func(c impactCandidate) float64 {
			return c.Wards
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

		stunsScore := normalize(
			player.Stuns,
			stunsMin,
			stunsMax,
		)

		wardsScore := normalize(
			candidate.Wards,
			wardsMin,
			wardsMax,
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

		survivalScore := 1 - normalize(
			float64(player.Deaths),
			deathsMin,
			deathsMax,
		)

		participationScore := clamp01(
			candidate.Participation,
		)

		// Не пытаемся жёстко определить позицию.
		// Вместо этого получаем число:
		//
		// 0.0 = профиль ближе к core
		// 1.0 = профиль ближе к support
		supportFactor :=
			assistsScore*0.15 +
				stunsScore*0.25 +
				wardsScore*0.25 +
				healingScore*0.15 +
				(1-gpmScore)*0.10 +
				(1-netWorthScore)*0.10

		supportFactor = clamp01(
			supportFactor,
		)

		profile := impactProfile(
			supportFactor,
		)

		winScore := 0.0

		if candidate.Won {
			winScore = 1
		}

		/*
			CORE weights:

			KDA           17%
			Assists        4%
			Hero Damage   22%
			Tower Damage  14%
			Healing        1%
			Stuns          2%
			Wards          0%
			Participation  6%
			GPM           10%
			XPM            7%
			Net Worth      8%
			Survival       4%
			Win            5%

			SUPPORT weights:

			KDA           10%
			Assists       13%
			Hero Damage    8%
			Tower Damage   3%
			Healing       10%
			Stuns         14%
			Wards         12%
			Participation 14%
			GPM            2%
			XPM            3%
			Net Worth      2%
			Survival       4%
			Win            5%
		*/

		kdaContribution :=
			kdaScore *
				blendWeight(
					0.17,
					0.10,
					supportFactor,
				) * 100

		assistsContribution :=
			assistsScore *
				blendWeight(
					0.04,
					0.13,
					supportFactor,
				) * 100

		heroDamageContribution :=
			heroDamageScore *
				blendWeight(
					0.22,
					0.08,
					supportFactor,
				) * 100

		towerDamageContribution :=
			towerDamageScore *
				blendWeight(
					0.14,
					0.03,
					supportFactor,
				) * 100

		healingContribution :=
			healingScore *
				blendWeight(
					0.01,
					0.10,
					supportFactor,
				) * 100

		stunsContribution :=
			stunsScore *
				blendWeight(
					0.02,
					0.14,
					supportFactor,
				) * 100

		wardsContribution :=
			wardsScore *
				blendWeight(
					0.00,
					0.12,
					supportFactor,
				) * 100

		participationContribution :=
			participationScore *
				blendWeight(
					0.06,
					0.14,
					supportFactor,
				) * 100

		gpmContribution :=
			gpmScore *
				blendWeight(
					0.10,
					0.02,
					supportFactor,
				) * 100

		xpmContribution :=
			xpmScore *
				blendWeight(
					0.07,
					0.03,
					supportFactor,
				) * 100

		netWorthContribution :=
			netWorthScore *
				blendWeight(
					0.08,
					0.02,
					supportFactor,
				) * 100

		survivalContribution :=
			survivalScore * 0.04 * 100

		winContribution :=
			winScore * 0.05 * 100

		score :=
			kdaContribution +
				assistsContribution +
				heroDamageContribution +
				towerDamageContribution +
				healingContribution +
				stunsContribution +
				wardsContribution +
				participationContribution +
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

				Profile:       profile,
				SupportFactor: supportFactor,

				KDAContribution: kdaContribution,

				AssistsContribution: assistsContribution,

				HeroDamageContribution: heroDamageContribution,

				TowerDamageContribution: towerDamageContribution,

				HealingContribution: healingContribution,

				StunsContribution: stunsContribution,

				WardsContribution: wardsContribution,

				ParticipationContribution: participationContribution,

				GPMContribution: gpmContribution,

				XPMContribution: xpmContribution,

				NetWorthContribution: netWorthContribution,

				SurvivalContribution: survivalContribution,

				WinContribution: winContribution,
			},
		)
	}

	sort.Slice(
		ranking,
		func(i, j int) bool {
			return ranking[i].Score >
				ranking[j].Score
		},
	)

	return ranking
}

func calculateKDA(
	player MatchPlayerInfo,
) float64 {
	deaths := player.Deaths

	if deaths <= 0 {
		deaths = 1
	}

	return float64(
		player.Kills+player.Assists,
	) / float64(deaths)
}

func calculateParticipation(
	player MatchPlayerInfo,
	team []MatchPlayerInfo,
) float64 {
	// Если OpenDota уже посчитал этот показатель,
	// используем его.
	if player.TeamfightParticipation > 0 {
		return clamp01(
			player.TeamfightParticipation,
		)
	}

	// Fallback для матчей, где этого поля нет.
	teamKills := 0

	for _, teammate := range team {
		teamKills += teammate.Kills
	}

	if teamKills <= 0 {
		return 0
	}

	participation := float64(
		player.Kills+player.Assists,
	) / float64(teamKills)

	return clamp01(
		participation,
	)
}

func impactProfile(
	supportFactor float64,
) string {
	switch {
	case supportFactor >= 0.55:
		return "Support"

	case supportFactor <= 0.35:
		return "Core"

	default:
		return "Hybrid"
	}
}

func blendWeight(
	coreWeight float64,
	supportWeight float64,
	supportFactor float64,
) float64 {
	return coreWeight +
		(supportWeight-coreWeight)*
			supportFactor
}

func clamp01(
	value float64,
) float64 {
	if value < 0 {
		return 0
	}

	if value > 1 {
		return 1
	}

	return value
}

func normalize(
	value float64,
	min float64,
	max float64,
) float64 {
	if max == min {
		// Если у всех значение 0,
		// никто не получает баллы за показатель.
		if max == 0 {
			return 0
		}

		return 0.5
	}

	return (value - min) /
		(max - min)
}

func minMax(
	candidates []impactCandidate,
	value func(impactCandidate) float64,
) (float64, float64) {
	minValue := value(
		candidates[0],
	)

	maxValue := minValue

	for _, candidate := range candidates[1:] {
		current := value(
			candidate,
		)

		if current < minValue {
			minValue = current
		}

		if current > maxValue {
			maxValue = current
		}
	}

	return minValue, maxValue
}
