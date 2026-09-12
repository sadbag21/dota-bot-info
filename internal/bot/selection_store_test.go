package bot

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/sadbag21/dota-bot-info/internal/dota"
)

func TestSelectionsSurviveReloadAndReplacement(t *testing.T) {
	dir := t.TempDir()
	b := &Bot{}
	if err := b.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []savedPlayer{{ChatID: 10, UserID: 1, AccountID: 111}, {ChatID: 10, UserID: 2, AccountID: 222}, {ChatID: -20, UserID: 1, AccountID: 333}, {ChatID: 10, UserID: 1, AccountID: 444}} {
		if !b.rememberPlayer(selectionMessage(entry.ChatID, entry.UserID, "player", ""), entry.AccountID) {
			t.Fatal("save failed")
		}
	}
	restored := &Bot{}
	if err := restored.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []savedPlayer{{ChatID: 10, UserID: 1, AccountID: 444}, {ChatID: 10, UserID: 2, AccountID: 222}, {ChatID: -20, UserID: 1, AccountID: 333}} {
		got, err := restored.resolveDotaID(selectionMessage(entry.ChatID, entry.UserID, "stats", ""))
		if err != nil || got != entry.AccountID {
			t.Fatalf("restored wrong selection: %d, %v", got, err)
		}
	}
	if got, err := restored.resolveDotaID(selectionMessage(10, 1, "stats", "999")); err != nil || got != 999 {
		t.Fatal(got, err)
	}
	if _, err := restored.resolveDotaID(selectionMessage(10, 3, "stats", "")); !errors.Is(err, errMissingID) {
		t.Fatal("selection leaked", err)
	}
	again := &Bot{}
	if err := again.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	if got, _ := again.resolveDotaID(selectionMessage(10, 1, "stats", "")); got != 444 {
		t.Fatal("explicit request altered saved player")
	}
}

func TestSelectionWriteFailureKeepsPreviousChoice(t *testing.T) {
	dir := t.TempDir()
	b, svc, texts := selectionTestBot(t)
	if err := b.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	if !b.rememberPlayer(selectionMessage(10, 1, "player", ""), 111) {
		t.Fatal("initial save failed")
	}
	originalPath := b.selectedPlayers.path
	original, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	b.selectedPlayers.path = blocked // Renaming a file over a directory fails on Windows and Linux.
	b.handlePlayer(selectionMessage(10, 1, "player", "222"))
	if !strings.Contains((*texts)[len(*texts)-1], "Не удалось сохранить") {
		t.Fatal("false success message", *texts)
	}
	if got, _ := b.resolveDotaID(selectionMessage(10, 1, "stats", "")); got != 111 {
		t.Fatal("failed write changed memory")
	}
	data, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(original) {
		t.Fatal("failed write changed saved file")
	}
	temps, err := filepath.Glob(filepath.Join(dir, ".player-selections-*.tmp"))
	if err != nil || len(temps) != 0 {
		t.Fatal("temporary files leaked", temps, err)
	}
	b.selectedPlayers.path = originalPath
	svc.profileErr = dota.ErrNotFound
	b.handlePlayer(selectionMessage(10, 1, "player", "333"))
	restored := &Bot{}
	if err := restored.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	if got, _ := restored.resolveDotaID(selectionMessage(10, 1, "stats", "")); got != 111 {
		t.Fatal("failed profile changed stored selection")
	}
	svc.profileErr = nil
	b.handlePlayer(selectionMessage(10, 1, "player", "444"))
	if got, _ := b.resolveDotaID(selectionMessage(10, 1, "stats", "")); got != 444 {
		t.Fatal("did not recover after write failure")
	}
}

func TestRejectDamagedSelectionFiles(t *testing.T) {
	for name, content := range map[string]string{
		"truncated":       `{"version":1`,
		"future version":  `{"version":2,"players":[]}`,
		"missing players": `{"version":1}`,
		"trailing object": `{"version":1,"players":[]} {}`,
		"unknown field":   `{"version":1,"players":[],"secret":1}`,
		"bad account":     `{"version":1,"players":[{"chat_id":10,"user_id":1,"account_id":4294967296}]}`,
		"bad user":        `{"version":1,"players":[{"chat_id":10,"user_id":0,"account_id":1}]}`,
		"duplicate":       `{"version":1,"players":[{"chat_id":10,"user_id":1,"account_id":1},{"chat_id":10,"user_id":1,"account_id":2}]}`,
		"oversize":        strings.Repeat(" ", maxSelectionFile+1),
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "player-selections.json")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			var selections playerSelections
			if err := selections.open(dir); err == nil {
				t.Fatal("accepted damaged file")
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != content {
				t.Fatal("damaged file overwritten", err)
			}
		})
	}
}

func TestPersistentCallbackAndConcurrentSelections(t *testing.T) {
	dir := t.TempDir()
	b, _, _ := selectionTestBot(t)
	if err := b.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	b.handleCallback(testQuery("player:222"))
	var wg sync.WaitGroup
	for user := range 20 {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			if !b.rememberPlayer(selectionMessage(10, id, "player", ""), id) {
				t.Error("save failed")
			}
		}(int64(user + 1))
	}
	wg.Wait()
	restored := &Bot{}
	if err := restored.selectedPlayers.open(dir); err != nil {
		t.Fatal(err)
	}
	if got, _ := restored.resolveDotaID(selectionMessage(42, 77, "stats", "")); got != 222 {
		t.Fatal("callback selection lost")
	}
	for user := range 20 {
		id := int64(user + 1)
		if got, _ := restored.resolveDotaID(selectionMessage(10, id, "stats", "")); got != id {
			t.Fatal("concurrent selection lost", id, got)
		}
	}
}
