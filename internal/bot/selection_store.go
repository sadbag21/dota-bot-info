package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const maxSavedPlayers = 10000
const maxSelectionFile = 2 << 20

type savedPlayer struct {
	ChatID    int64 `json:"chat_id"`
	UserID    int64 `json:"user_id"`
	AccountID int64 `json:"account_id"`
}
type selectionFile struct {
	Version int           `json:"version"`
	Players []savedPlayer `json:"players"`
}

// open is called before polling starts. A malformed file must never be silently
// replaced with an empty store on the next /player command.
func (s *playerSelections) open(directory string) error {
	if directory == "" {
		return errors.New("selection data directory is empty")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return fmt.Errorf("create selection directory: %w", err)
	}
	path := filepath.Join(directory, "player-selections.json")
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		s.path = path
		s.players = make(map[playerSelectionKey]int64)
		return nil
	}
	if err != nil {
		return fmt.Errorf("open player selections: %w", err)
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxSelectionFile+1))
	if err != nil {
		return fmt.Errorf("read player selections: %w", err)
	}
	if len(data) > maxSelectionFile {
		return errors.New("player selection file is too large")
	}
	var stored selectionFile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&stored); err != nil {
		return errors.New("invalid player selection file")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || stored.Version != 1 || stored.Players == nil || len(stored.Players) > maxSavedPlayers {
		return errors.New("invalid player selection file")
	}
	players := make(map[playerSelectionKey]int64, len(stored.Players))
	for _, entry := range stored.Players {
		key := playerSelectionKey{chatID: entry.ChatID, userID: entry.UserID}
		if !validSavedPlayer(key, entry.AccountID) {
			return errors.New("invalid player selection entry")
		}
		if _, exists := players[key]; exists {
			return errors.New("duplicate player selection entry")
		}
		players[key] = entry.AccountID
	}
	s.path, s.players = path, players
	return nil
}

func validSavedPlayer(key playerSelectionKey, accountID int64) bool {
	return key.chatID != 0 && key.userID > 0 && accountID > 0 && accountID <= 4294967295
}

func (s *playerSelections) set(key playerSelectionKey, accountID int64) error {
	if !validSavedPlayer(key, accountID) {
		return errors.New("invalid player selection")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.players[key] == accountID {
		return nil
	}
	if _, exists := s.players[key]; !exists && len(s.players) >= maxSavedPlayers {
		return errors.New("player selection capacity reached")
	}
	if s.path != "" {
		stored := selectionFile{Version: 1, Players: make([]savedPlayer, 0, len(s.players)+1)}
		for existing, id := range s.players {
			if existing != key {
				stored.Players = append(stored.Players, savedPlayer{ChatID: existing.chatID, UserID: existing.userID, AccountID: id})
			}
		}
		stored.Players = append(stored.Players, savedPlayer{ChatID: key.chatID, UserID: key.userID, AccountID: accountID})
		sort.Slice(stored.Players, func(i, j int) bool {
			a, b := stored.Players[i], stored.Players[j]
			if a.ChatID != b.ChatID {
				return a.ChatID < b.ChatID
			}
			return a.UserID < b.UserID
		})
		data, err := json.Marshal(stored)
		if err != nil {
			return fmt.Errorf("encode player selections: %w", err)
		}
		if err := replaceSelectionFile(s.path, data); err != nil {
			return err
		}
	}
	if s.players == nil {
		s.players = make(map[playerSelectionKey]int64)
	}
	s.players[key] = accountID
	return nil
}

// Write beside the destination, sync and close, then rename. Until replacement
// succeeds, the previous file and in-memory selection remain unchanged.
func replaceSelectionFile(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".player-selections-*.tmp")
	if err != nil {
		return fmt.Errorf("create selection temporary file: %w", err)
	}
	tempPath := file.Name()
	defer func() { _ = file.Close(); _ = os.Remove(tempPath) }()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write player selections: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync player selections: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close player selections: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace player selections: %w", err)
	}
	return nil
}
