package service

import (
	"fmt"
	"sync"

	"github.com/sadbag21/dota-bot-info/internal/dota"
)

type PlayerService struct {
	dotaClient *dota.Client

	heroesMu sync.RWMutex
	heroes   map[int]string
}

type PlayerInfo struct {
	Player *dota.Player

	Wins    int
	Losses  int
	Matches int
	WinRate float64

	RecentMatches []MatchInfo
}

type MatchInfo struct {
	MatchID int64

	HeroName string

	Kills   int
	Deaths  int
	Assists int

	Duration int
	Won      bool
}

func NewPlayerService(dotaClient *dota.Client) *PlayerService {
	return &PlayerService{
		dotaClient: dotaClient,
	}
}

func (s *PlayerService) GetPlayerInfo(accountID int64) (*PlayerInfo, error) {
	player, err := s.dotaClient.GetPlayer(accountID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get player: %w",
			err,
		)
	}

	winLoss, err := s.dotaClient.GetWinLoss(accountID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get win/loss: %w",
			err,
		)
	}

	recentMatches, err := s.dotaClient.GetRecentMatches(accountID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get recent matches: %w",
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

	matches := winLoss.Win + winLoss.Lose

	var winRate float64

	if matches > 0 {
		winRate = float64(winLoss.Win) / float64(matches) * 100
	}

	matchInfos := buildMatchInfos(
		recentMatches,
		heroNames,
	)

	return &PlayerInfo{
		Player:        player,
		Wins:          winLoss.Win,
		Losses:        winLoss.Lose,
		Matches:       matches,
		WinRate:       winRate,
		RecentMatches: matchInfos,
	}, nil
}

func (s *PlayerService) getHeroNames() (map[int]string, error) {
	s.heroesMu.RLock()

	if s.heroes != nil {
		heroes := s.heroes

		s.heroesMu.RUnlock()

		return heroes, nil
	}

	s.heroesMu.RUnlock()

	s.heroesMu.Lock()
	defer s.heroesMu.Unlock()

	// Пока мы ждали Lock, другой запрос
	// мог уже загрузить героев.
	if s.heroes != nil {
		return s.heroes, nil
	}

	heroes, err := s.dotaClient.GetHeroes()
	if err != nil {
		return nil, err
	}

	heroNames := make(map[int]string, len(heroes))

	for _, hero := range heroes {
		heroNames[hero.ID] = hero.LocalizedName
	}

	s.heroes = heroNames

	return s.heroes, nil
}

func buildMatchInfos(
	matches []dota.RecentMatch,
	heroNames map[int]string,
) []MatchInfo {
	const limit = 5

	count := len(matches)

	if count > limit {
		count = limit
	}

	result := make([]MatchInfo, 0, count)

	for i := 0; i < count; i++ {
		match := matches[i]

		heroName := heroNames[match.HeroID]

		if heroName == "" {
			heroName = fmt.Sprintf(
				"Hero %d",
				match.HeroID,
			)
		}

		isRadiant := match.PlayerSlot < 128

		won := isRadiant == match.RadiantWin

		result = append(
			result,
			MatchInfo{
				MatchID:  match.MatchID,
				HeroName: heroName,
				Kills:    match.Kills,
				Deaths:   match.Deaths,
				Assists:  match.Assists,
				Duration: match.Duration,
				Won:      won,
			},
		)
	}

	return result
}
