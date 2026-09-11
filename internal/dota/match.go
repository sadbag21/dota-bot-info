package dota

import "fmt"

func (c *Client) GetMatch(matchID int64) (*Match, error) {
	return cachedGet[Match](c, fmt.Sprintf("/matches/%d", matchID), matchCacheTTL, func(match *Match) error {
		if match.MatchID == 0 {
			return fmt.Errorf("%w: match %d", ErrNotFound, matchID)
		}
		return nil
	})
}
