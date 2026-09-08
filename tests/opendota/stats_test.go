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

func TestGetPlayerTotals(t *testing.T) {
	cute.NewTestBuilder().
		Title("Get player totals").
		Description(
			"OpenDota returns aggregated player statistics",
		).
		Tags(
			"opendota",
			"stats",
			"player",
		).
		Create().
		RequestBuilder(
			cute.WithURI(
				fmt.Sprintf(
					"%s/players/%d/totals",
					baseURL,
					testAccountID,
				),
			),
			cute.WithMethod(http.MethodGet),
		).
		ExpectExecuteTimeout(30*time.Second).
		ExpectStatus(http.StatusOK).
		AssertBody(
			cutejson.Present("$[0].field"),
			cutejson.Present("$[0].n"),
			cutejson.Present("$[0].sum"),
		).
		ExecuteTest(
			context.Background(),
			t,
		)
}
