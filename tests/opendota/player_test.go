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

func TestGetPlayer(t *testing.T) {
	requireOpenDotaTests(t)
	cute.NewTestBuilder().
		Title("Get Dota player").
		Description("OpenDota returns player information by account ID").
		Tag("opendota").
		Tag("player").
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/players/%d",
					baseURL,
					testAccountID,
				),
			),
			cute.WithMethod(http.MethodGet),
		).
		ExpectExecuteTimeout(30*time.Second).
		ExpectStatus(http.StatusOK).
		AssertBody(
			cutejson.Equal(
				"$.profile.account_id",
				testAccountID,
			),
			cutejson.Present(
				"$.profile.personaname",
			),
			cutejson.Present(
				"$.profile.profileurl",
			),
			cutejson.Present(
				"$.rank_tier",
			),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}

func TestGetPlayerWinLoss(t *testing.T) {
	requireOpenDotaTests(t)
	cute.NewTestBuilder().
		Title("Get player win loss").
		Description(
			"OpenDota returns wins and losses for player",
		).
		Tags(
			"opendota",
			"player",
			"wl",
		).
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/players/%d/wl",
					baseURL,
					testAccountID,
				),
			),
			cute.WithMethod(
				http.MethodGet,
			),
		).
		ExpectExecuteTimeout(
			30*time.Second,
		).
		ExpectStatus(
			http.StatusOK,
		).
		AssertBody(
			cutejson.Present("$.win"),
			cutejson.Present("$.lose"),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}
