package service

import (
	"fmt"
	"sort"
)

type MatchDetails struct {
	MatchID    int64
	Duration   int
	RadiantWin bool

	Radiant []MatchPlayerInfo
	Dire    []MatchPlayerInfo

	ImpactRanking []ImpactEntry
}

type MatchPlayerInfo struct {
	AccountID *int64

	PlayerSlot int

	Name     string
	HeroName string

	Kills   int
	Deaths  int
	Assists int

	GPM int
	XPM int

	NetWorth int

	HeroDamage  int
	TowerDamage int
	HeroHealing int

	LastHits int
	Denies   int
}

func (s *PlayerService) GetMatchDetails(
	matchID int64,
) (*MatchDetails, error) {
	match, err := s.dotaClient.GetMatch(matchID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get match: %w",
			err,
		)
	}

	heroNames, err := s.getHeroNames()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get heroes: %w",
			err,
		)
	}

	radiant := make([]MatchPlayerInfo, 0, 5)
	dire := make([]MatchPlayerInfo, 0, 5)

	for _, player := range match.Players {
		heroName := heroNames[player.HeroID]

		if heroName == "" {
			heroName = fmt.Sprintf(
				"Hero %d",
				player.HeroID,
			)
		}

		name := player.Personaname

		if name == "" {
			name = "Anonymous"
		}

		info := MatchPlayerInfo{
			AccountID: player.AccountID,

			PlayerSlot: player.PlayerSlot,

			Name:     name,
			HeroName: heroName,

			Kills:   player.Kills,
			Deaths:  player.Deaths,
			Assists: player.Assists,

			GPM: player.GoldPerMin,
			XPM: player.XPPerMin,

			NetWorth: player.NetWorth,

			HeroDamage:  player.HeroDamage,
			TowerDamage: player.TowerDamage,
			HeroHealing: player.HeroHealing,

			LastHits: player.LastHits,
			Denies:   player.Denies,
		}

		if player.PlayerSlot < 128 {
			radiant = append(
				radiant,
				info,
			)
		} else {
			dire = append(
				dire,
				info,
			)
		}
	}

	sort.Slice(
		radiant,
		func(i, j int) bool {
			return radiant[i].PlayerSlot <
				radiant[j].PlayerSlot
		},
	)

	sort.Slice(
		dire,
		func(i, j int) bool {
			return dire[i].PlayerSlot <
				dire[j].PlayerSlot
		},
	)

	details := &MatchDetails{
		MatchID:    match.MatchID,
		Duration:   match.Duration,
		RadiantWin: match.RadiantWin,
		Radiant:    radiant,
		Dire:       dire,
	}

	details.ImpactRanking = calculateImpactRanking(details)

	return details, nil
}
