package service

import (
	"fmt"
	"sort"
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

type HeroStats struct {
	HeroID   int
	HeroName string

	Games   int
	Wins    int
	Losses  int
	WinRate float64
}

type PlayerStats struct {
	MatchesAnalyzed int

	AvgKills   float64
	AvgDeaths  float64
	AvgAssists float64
	KDA        float64

	AvgGPM      float64
	AvgXPM      float64
	AvgLastHits float64

	AvgHeroDamage  float64
	AvgTowerDamage float64
	AvgHeroHealing float64

	AvgDuration float64
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
		5,
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
	limit int,
) []MatchInfo {
	if limit <= 0 {
		return []MatchInfo{}
	}

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

func (s *PlayerService) GetRecentMatchesInfo(
	accountID int64,
	limit int,
) ([]MatchInfo, error) {
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

	return buildMatchInfos(
		recentMatches,
		heroNames,
		limit,
	), nil
}

func (s *PlayerService) GetHeroStats(
	accountID int64,
	limit int,
) ([]HeroStats, error) {
	playerHeroes, err := s.dotaClient.GetPlayerHeroes(accountID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get player heroes: %w",
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

	sort.Slice(
		playerHeroes,
		func(i, j int) bool {
			return playerHeroes[i].Games > playerHeroes[j].Games
		},
	)

	if limit <= 0 {
		return []HeroStats{}, nil
	}

	count := len(playerHeroes)

	if count > limit {
		count = limit
	}

	result := make([]HeroStats, 0, count)

	for i := 0; i < count; i++ {
		hero := playerHeroes[i]

		name := heroNames[hero.HeroID]

		if name == "" {
			name = fmt.Sprintf(
				"Hero %d",
				hero.HeroID,
			)
		}

		losses := hero.Games - hero.Win

		var winRate float64

		if hero.Games > 0 {
			winRate =
				float64(hero.Win) /
					float64(hero.Games) *
					100
		}

		result = append(
			result,
			HeroStats{
				HeroID:   hero.HeroID,
				HeroName: name,
				Games:    hero.Games,
				Wins:     hero.Win,
				Losses:   losses,
				WinRate:  winRate,
			},
		)
	}

	return result, nil
}

func averageTotal(total dota.Total) float64 {
	if total.N <= 0 {
		return 0
	}

	return total.Sum / float64(total.N)
}

func (s *PlayerService) GetPlayerStats(
	accountID int64,
) (*PlayerStats, error) {
	totals, err := s.dotaClient.GetTotals(accountID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get player totals: %w",
			err,
		)
	}

	totalMap := make(map[string]dota.Total)

	for _, total := range totals {
		totalMap[total.Field] = total
	}

	avgKills := averageTotal(
		totalMap["kills"],
	)

	avgDeaths := averageTotal(
		totalMap["deaths"],
	)

	avgAssists := averageTotal(
		totalMap["assists"],
	)

	var kda float64

	if avgDeaths > 0 {
		kda = (avgKills + avgAssists) / avgDeaths
	}

	matchesAnalyzed := totalMap["kills"].N

	return &PlayerStats{
		MatchesAnalyzed: matchesAnalyzed,

		AvgKills:   avgKills,
		AvgDeaths:  avgDeaths,
		AvgAssists: avgAssists,
		KDA:        kda,

		AvgGPM: averageTotal(
			totalMap["gold_per_min"],
		),

		AvgXPM: averageTotal(
			totalMap["xp_per_min"],
		),

		AvgLastHits: averageTotal(
			totalMap["last_hits"],
		),

		AvgHeroDamage: averageTotal(
			totalMap["hero_damage"],
		),

		AvgTowerDamage: averageTotal(
			totalMap["tower_damage"],
		),

		AvgHeroHealing: averageTotal(
			totalMap["hero_healing"],
		),

		AvgDuration: averageTotal(
			totalMap["duration"],
		),
	}, nil
}
