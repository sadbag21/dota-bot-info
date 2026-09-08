package dota

func (c *Client) GetHeroes() ([]Hero, error) {
	var heroes []Hero

	err := c.get("/heroes", &heroes)
	if err != nil {
		return nil, err
	}

	return heroes, nil
}
