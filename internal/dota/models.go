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
