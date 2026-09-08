package opendota_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ozontech/cute"
	cutejson "github.com/ozontech/cute/asserts/json"
)

func TestGetHeroes(t *testing.T) {
	requireOpenDotaTests(t)
	cute.NewTestBuilder().
		Title("Get Dota heroes").
		Description(
			"OpenDota returns the hero dictionary",
		).
		Tags(
			"opendota",
			"heroes",
		).
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/heroes",
					baseURL,
				),
			),
			cute.WithMethod(http.MethodGet),
		).
		ExpectExecuteTimeout(30*time.Second).
		ExpectStatus(http.StatusOK).
		AssertBody(
			cutejson.Present("$[0].id"),
			cutejson.Present(
				"$[0].localized_name",
			),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}

func TestGetPlayerHeroes(t *testing.T) {
	requireOpenDotaTests(t)
	cute.NewTestBuilder().
		Title("Get player hero statistics").
		Description(
			"OpenDota returns hero statistics for a player",
		).
		Tags(
			"opendota",
			"heroes",
			"player",
		).
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/players/%d/heroes",
					baseURL,
					testAccountID,
				),
			),
			cute.WithMethod(http.MethodGet),
		).
		ExpectExecuteTimeout(30*time.Second).
		ExpectStatus(http.StatusOK).
		AssertBody(
			cutejson.Present("$[0].hero_id"),
			cutejson.Present("$[0].games"),
			cutejson.Present("$[0].win"),
			cutejson.Present(
				"$[0].last_played",
			),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}
