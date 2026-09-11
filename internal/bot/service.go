package bot

import "github.com/sadbag21/dota-bot-info/internal/service"

type playerService interface {
	GetPlayerInfo(int64) (*service.PlayerInfo, error)
	GetRecentMatchesInfo(int64, int) ([]service.MatchInfo, error)
	GetHeroStats(int64, int) ([]service.HeroStats, error)
	GetPlayerStats(int64) (*service.PlayerStats, error)
	GetMatchDetails(int64) (*service.MatchDetails, error)
}
