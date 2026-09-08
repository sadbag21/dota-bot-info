package dota

type Player struct {
	Profile Profile `json:"profile"`

	RankTier        int  `json:"rank_tier"`
	LeaderboardRank *int `json:"leaderboard_rank"`
}

type Profile struct {
	AccountID    int64  `json:"account_id"`
	Personaname  string `json:"personaname"`
	Name         string `json:"name"`
	Avatar       string `json:"avatar"`
	AvatarFull   string `json:"avatarfull"`
	AvatarMedium string `json:"avatarmedium"`
	ProfileURL   string `json:"profileurl"`
}

type WinLoss struct {
	Win  int `json:"win"`
	Lose int `json:"lose"`
}

type RecentMatch struct {
	MatchID    int64 `json:"match_id"`
	PlayerSlot int   `json:"player_slot"`
	RadiantWin bool  `json:"radiant_win"`
	Duration   int   `json:"duration"`
	HeroID     int   `json:"hero_id"`
	Kills      int   `json:"kills"`
	Deaths     int   `json:"deaths"`
	Assists    int   `json:"assists"`
	StartTime  int64 `json:"start_time"`
}

type Hero struct {
	ID            int    `json:"id"`
	LocalizedName string `json:"localized_name"`
}

type PlayerHero struct {
	HeroID     int   `json:"hero_id"`
	LastPlayed int64 `json:"last_played"`
	Games      int   `json:"games"`
	Win        int   `json:"win"`
}

type Total struct {
	Field string  `json:"field"`
	N     int     `json:"n"`
	Sum   float64 `json:"sum"`
}

type Match struct {
	MatchID    int64 `json:"match_id"`
	Duration   int   `json:"duration"`
	StartTime  int64 `json:"start_time"`
	RadiantWin bool  `json:"radiant_win"`

	Players []MatchPlayer `json:"players"`
}

type MatchPlayer struct {
	AccountID  *int64 `json:"account_id"`
	PlayerSlot int    `json:"player_slot"`
	HeroID     int    `json:"hero_id"`

	Personaname string `json:"personaname"`

	Kills   int `json:"kills"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`

	GoldPerMin int `json:"gold_per_min"`
	XPPerMin   int `json:"xp_per_min"`
	NetWorth   int `json:"net_worth"`

	HeroDamage  int `json:"hero_damage"`
	TowerDamage int `json:"tower_damage"`
	HeroHealing int `json:"hero_healing"`
	LastHits    int `json:"last_hits"`
	Denies      int `json:"denies"`

	Stuns                  float64 `json:"stuns"`
	ObsPlaced              int     `json:"obs_placed"`
	SenPlaced              int     `json:"sen_placed"`
	TeamfightParticipation float64 `json:"teamfight_participation"`
}
