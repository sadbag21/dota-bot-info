package service

import (
	"fmt"

	"github.com/sadbag21/dota-bot-info/internal/dota"
)

type PlayerService struct {
	dotaClient *dota.Client
}

type PlayerInfo struct {
	Player *dota.Player

	Wins    int
	Losses  int
	Matches int
	WinRate float64
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

	matches := winLoss.Win + winLoss.Lose

	var winRate float64

	if matches > 0 {
		winRate = float64(winLoss.Win) / float64(matches) * 100
	}

	return &PlayerInfo{
		Player:  player,
		Wins:    winLoss.Win,
		Losses:  winLoss.Lose,
		Matches: matches,
		WinRate: winRate,
	}, nil
}
