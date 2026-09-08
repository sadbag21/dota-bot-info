package dota

import "fmt"

func (c *Client) GetMatch(matchID int64) (*Match, error) {
	var match Match

	err := c.get(
		fmt.Sprintf("/matches/%d", matchID),
		&match,
	)

	if err != nil {
		return nil, err
	}

	return &match, nil
}
