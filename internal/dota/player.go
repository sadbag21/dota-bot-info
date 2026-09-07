package dota

import "fmt"

func (c *Client) GetPlayer(accountID int64) (*Player, error) {
	var player Player

	err := c.get(
		fmt.Sprintf("/players/%d", accountID),
		&player,
	)

	if err != nil {
		return nil, err
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
