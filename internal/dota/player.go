package dota

import "fmt"

func (c *Client) GetPlayer(accountID int64) (*Player, error) {
	return cachedGet[Player](c, fmt.Sprintf("/players/%d", accountID), playerCacheTTL, func(player *Player) error {
		if player.Profile.AccountID == 0 ||
			(player.Profile.SteamID == "" && player.Profile.Personaname == "" && player.Profile.ProfileURL == "") {
			return fmt.Errorf("%w: player %d", ErrNotFound, accountID)
		}
		return nil
	})
}

func (c *Client) GetWinLoss(accountID int64) (*WinLoss, error) {
	return cachedGet[WinLoss](c, fmt.Sprintf("/players/%d/wl", accountID), playerCacheTTL, nil)
}

func (c *Client) GetRecentMatches(accountID int64) ([]RecentMatch, error) {
	result, err := cachedGet[[]RecentMatch](c, fmt.Sprintf("/players/%d/recentMatches", accountID), recentMatchesCacheTTL, nil)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) GetPlayerHeroes(accountID int64) ([]PlayerHero, error) {
	result, err := cachedGet[[]PlayerHero](c, fmt.Sprintf("/players/%d/heroes", accountID), playerCacheTTL, nil)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func (c *Client) GetTotals(accountID int64) ([]Total, error) {
	result, err := cachedGet[[]Total](c, fmt.Sprintf("/players/%d/totals", accountID), playerCacheTTL, nil)
	if err != nil {
		return nil, err
	}
	return *result, nil
}
