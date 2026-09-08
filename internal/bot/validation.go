package bot

import (
	"errors"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const maxDotaAccountID int64 = 4294967295

var (
	errMissingID = errors.New("missing ID")
	errInvalidID = errors.New("invalid ID")
)

func parseDotaID(
	message *tgbotapi.Message,
) (int64, error) {
	args := strings.TrimSpace(
		message.CommandArguments(),
	)

	if args == "" {
		return 0, errMissingID
	}

	accountID, err := strconv.ParseInt(
		args,
		10,
		64,
	)

	if err != nil {
		return 0, errInvalidID
	}

	if accountID <= 0 ||
		accountID > maxDotaAccountID {
		return 0, errInvalidID
	}

	return accountID, nil
}

func parseMatchID(
	message *tgbotapi.Message,
) (int64, error) {
	args := strings.TrimSpace(
		message.CommandArguments(),
	)

	if args == "" {
		return 0, errMissingID
	}

	matchID, err := strconv.ParseInt(
		args,
		10,
		64,
	)

	if err != nil ||
		matchID <= 0 {
		return 0, errInvalidID
	}

	return matchID, nil
}
