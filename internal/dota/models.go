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
