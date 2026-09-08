package opendota_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ozontech/cute"
	"github.com/ozontech/cute/asserts/headers"
	cutejson "github.com/ozontech/cute/asserts/json"
)

func TestGetRecentMatches(t *testing.T) {
	cute.NewTestBuilder().
		Title("Get recent player matches").
		Description(
			"OpenDota returns recent matches for a player",
		).
		Tags(
			"opendota",
			"matches",
			"player",
		).
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/players/%d/recentMatches",
					baseURL,
					testAccountID,
				),
			),
			cute.WithMethod(http.MethodGet),
		).
		ExpectExecuteTimeout(30*time.Second).
		ExpectStatus(http.StatusOK).
		AssertHeaders(
			headers.Present("Content-Type"),
		).
		AssertBody(
			cutejson.Present("$[0].match_id"),
			cutejson.Present("$[0].hero_id"),
			cutejson.Present("$[0].player_slot"),
			cutejson.Present("$[0].radiant_win"),
			cutejson.Present("$[0].duration"),
			cutejson.Present("$[0].kills"),
			cutejson.Present("$[0].deaths"),
			cutejson.Present("$[0].assists"),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}

func TestGetMatch(t *testing.T) {
	cute.NewTestBuilder().
		Title("Get Dota match").
		Description(
			"OpenDota returns parsed match information by match ID",
		).
		Tags(
			"opendota",
			"match",
		).
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/matches/%d",
					baseURL,
					testMatchID,
				),
			),
			cute.WithMethod(http.MethodGet),
		).
		ExpectExecuteTimeout(30*time.Second).
		ExpectStatus(http.StatusOK).
		AssertHeaders(
			headers.Present("Content-Type"),
		).
		AssertBody(
			cutejson.Equal(
				"$.match_id",
				testMatchID,
			),
			cutejson.Present("$.duration"),
			cutejson.Present("$.radiant_win"),

			cutejson.Length(
				"$.players",
				10,
			),

			cutejson.Present(
				"$.players[0].hero_id",
			),

			cutejson.Present(
				"$.players[0].kills",
			),

			cutejson.Present(
				"$.players[0].deaths",
			),

			cutejson.Present(
				"$.players[0].assists",
			),

			cutejson.Present(
				"$.players[0].gold_per_min",
			),

			cutejson.Present(
				"$.players[0].xp_per_min",
			),

			cutejson.Present(
				"$.players[0].hero_damage",
			),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}
