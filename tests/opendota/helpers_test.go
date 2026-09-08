package opendota_test

import (
	"os"
	"testing"
)

func requireOpenDotaTests(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_OPENDOTA_TESTS") != "1" {
		t.Skip(
			"OpenDota integration tests are disabled; " +
				"set RUN_OPENDOTA_TESTS=1 to enable",
		)
	}
}
