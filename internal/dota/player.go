package dota

import "fmt"

func (c *Client) GetPlayer(
	accountID int64,
) (*Player, error) {
	var player Player

	err := c.get(
		fmt.Sprintf(
			"/players/%d",
			accountID,
		),
		&player,
	)

	if err != nil {
		return nil, err
	}

	if player.Profile.AccountID == 0 {
		return nil, fmt.Errorf(
			"%w: player %d",
			ErrNotFound,
			accountID,
		)
	}

	return &player, nil
}

func (c *Client) GetWinLoss(accountID int64) (*WinLoss, error) {
	var winLoss WinLoss

	err := c.get(
		fmt.Sprintf("/players/%d/wl", accountID),
		&winLoss,
	)

	if err != nil {
		return nil, err
	}

	return &winLoss, nil
}

func (c *Client) GetRecentMatches(accountID int64) ([]RecentMatch, error) {
	var matches []RecentMatch

	err := c.get(
		fmt.Sprintf("/players/%d/recentMatches", accountID),
		&matches,
	)

	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (c *Client) GetPlayerHeroes(accountID int64) ([]PlayerHero, error) {
	var heroes []PlayerHero

	err := c.get(
		fmt.Sprintf("/players/%d/heroes", accountID),
		&heroes,
	)

	if err != nil {
		return nil, err
	}

	return heroes, nil
}

func (c *Client) GetTotals(accountID int64) ([]Total, error) {
	var totals []Total

	err := c.get(
		fmt.Sprintf("/players/%d/totals", accountID),
		&totals,
	)

	if err != nil {
		return nil, err
	}

	return totals, nil
}
