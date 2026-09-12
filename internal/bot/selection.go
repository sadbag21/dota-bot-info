package bot

import (
	"errors"
	"log/slog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// A group member's choice must not change another member's default player.
// Choices are also separate between the user's private and group chats.
type playerSelectionKey struct{ chatID, userID int64 }

type playerSelections struct {
	mu      sync.RWMutex
	players map[playerSelectionKey]int64
	path    string
}

func selectionKey(message *tgbotapi.Message) (playerSelectionKey, bool) {
	if message == nil || message.Chat == nil || message.From == nil || message.From.ID == 0 || message.SenderChat != nil {
		return playerSelectionKey{}, false
	}
	return playerSelectionKey{message.Chat.ID, message.From.ID}, true
}

func (b *Bot) rememberPlayer(message *tgbotapi.Message, accountID int64) bool {
	key, ok := selectionKey(message)
	if !ok {
		return false
	}
	if err := b.selectedPlayers.set(key, accountID); err != nil {
		slog.Error("Failed to save selected player", "error", err)
		return false
	}
	return true
}

func (b *Bot) resolveDotaID(message *tgbotapi.Message) (int64, error) {
	// Explicit input always takes precedence; invalid input must not silently
	// fall back to a different player's statistics.
	id, err := parseDotaID(message)
	if !errors.Is(err, errMissingID) {
		return id, err
	}
	key, ok := selectionKey(message)
	if !ok {
		return 0, errMissingID
	}
	b.selectedPlayers.mu.RLock()
	defer b.selectedPlayers.mu.RUnlock()
	if id, ok := b.selectedPlayers.players[key]; ok {
		return id, nil
	}
	return 0, errMissingID
}
