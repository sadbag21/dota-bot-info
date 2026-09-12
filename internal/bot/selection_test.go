package bot

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/internal/dota"
	"github.com/sadbag21/dota-bot-info/internal/service"
)

func selectionMessage(chatID, userID int64, command, args string) *tgbotapi.Message {
	text := "/" + command
	message := &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}, Text: text, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Length: len(text)}}}
	if args != "" {
		message.Text += " " + args
	}
	return message
}

func TestSelectionIsolationAndExplicitIDs(t *testing.T) {
	b := &Bot{}
	b.rememberPlayer(selectionMessage(10, 1, "player", "111"), 111)
	b.rememberPlayer(selectionMessage(10, 2, "player", "222"), 222)
	b.rememberPlayer(selectionMessage(20, 1, "player", "333"), 333)
	for _, tt := range []struct {
		chat, user int64
		args       string
		want       int64
		err        error
	}{
		{10, 1, "", 111, nil}, {10, 2, "", 222, nil}, {20, 1, "", 333, nil},
		{10, 3, "", 0, errMissingID}, {30, 1, "", 0, errMissingID},
		{10, 1, "444", 444, nil}, {10, 1, "abc", 0, errInvalidID}, {10, 1, "999999999999", 0, errInvalidID},
	} {
		for _, command := range []string{"player", "matches", "heroes", "stats"} {
			id, err := b.resolveDotaID(selectionMessage(tt.chat, tt.user, command, tt.args))
			if id != tt.want || !errors.Is(err, tt.err) {
				t.Errorf("%s chat=%d user=%d args=%q got %d,%v", command, tt.chat, tt.user, tt.args, id, err)
			}
		}
	}
	id, err := b.resolveDotaID(selectionMessage(10, 1, "stats", ""))
	if err != nil || id != 111 {
		t.Fatal("explicit request changed selection")
	}
	if _, err := (&Bot{}).resolveDotaID(selectionMessage(10, 1, "stats", "")); !errors.Is(err, errMissingID) {
		t.Fatal("selection leaked between bot instances")
	}
}

func TestNoSelectionForAnonymousSenders(t *testing.T) {
	b := &Bot{}
	for _, kind := range []string{"no sender", "sender chat", "zero ID"} {
		msg := selectionMessage(10, 1, "player", "111")
		switch kind {
		case "no sender":
			msg.From = nil
		case "sender chat":
			msg.SenderChat = &tgbotapi.Chat{ID: 10}
		case "zero ID":
			msg.From.ID = 0
		}
		b.rememberPlayer(msg, 111)
		msg.Text = "/player"
		if _, err := b.resolveDotaID(msg); !errors.Is(err, errMissingID) {
			t.Errorf("%s received a shared selection", kind)
		}
	}
}

type selectionService struct {
	playerService
	profileErr                              error
	profileID, statsID, matchesID, heroesID int64
}

func (s *selectionService) GetPlayerInfo(id int64) (*service.PlayerInfo, error) {
	s.profileID = id
	if s.profileErr != nil {
		return nil, s.profileErr
	}
	return &service.PlayerInfo{Player: &dota.Player{Profile: dota.Profile{AccountID: id, Personaname: "Test Player"}}}, nil
}
func (s *selectionService) GetPlayerStats(id int64) (*service.PlayerStats, error) {
	s.statsID = id
	return &service.PlayerStats{}, nil
}
func (s *selectionService) GetRecentMatchesInfo(id int64, _ int) ([]service.MatchInfo, error) {
	s.matchesID = id
	return nil, nil
}
func (s *selectionService) GetHeroStats(id int64, _ int) ([]service.HeroStats, error) {
	s.heroesID = id
	return nil, nil
}

func selectionTestBot(t *testing.T) (*Bot, *selectionService, *[]string) {
	t.Helper()
	var texts []string
	api, err := tgbotapi.NewBotAPIWithClient("test-token", "https://telegram.test/bot%s/%s", telegramHTTPFunc(func(req *http.Request) (*http.Response, error) {
		payload := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
		if strings.HasSuffix(req.URL.Path, "/sendMessage") {
			if err := req.ParseForm(); err != nil {
				t.Fatal(err)
			}
			texts = append(texts, req.Form.Get("text"))
			payload = `{"ok":true,"result":{"message_id":1}}`
		} else if strings.HasSuffix(req.URL.Path, "/answerCallbackQuery") {
			payload = `{"ok":true,"result":true}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	svc := &selectionService{}
	return &Bot{api: api, service: svc}, svc, &texts
}

func TestSelectionCommandFlow(t *testing.T) {
	b, svc, texts := selectionTestBot(t)
	run := func(command, args string) {
		b.handleUpdate(tgbotapi.Update{Message: selectionMessage(10, 1, command, args)})
	}
	run("stats", "")
	if svc.statsID != 0 || !strings.Contains((*texts)[0], "выбери игрока") {
		t.Fatal("missing selection must request an ID without calling service")
	}
	run("player", "111")
	for _, command := range []string{"matches", "heroes", "stats", "player"} {
		run(command, "")
	}
	if svc.matchesID != 111 || svc.heroesID != 111 || svc.statsID != 111 || svc.profileID != 111 {
		t.Fatal("commands did not reuse selected player")
	}
	run("stats", "222")
	run("heroes", "")
	if svc.statsID != 222 || svc.heroesID != 111 {
		t.Fatal("explicit ID should only affect that command")
	}
	svc.profileErr = dota.ErrNotFound
	run("player", "333")
	svc.profileErr = nil
	run("stats", "")
	if svc.statsID != 111 {
		t.Fatal("failed profile replaced selection")
	}
	run("player", "444")
	run("stats", "")
	if svc.statsID != 444 {
		t.Fatal("successful profile did not replace selection")
	}
	run("stats", "abc")
	if svc.statsID != 444 {
		t.Fatal("invalid input reached service")
	}
	if !strings.Contains((*texts)[len(*texts)-1], "Некорректный") {
		t.Fatal("invalid input silently fell back to selected player")
	}
	run("match", "")
	if !strings.Contains((*texts)[len(*texts)-1], "Укажи Match ID") {
		t.Fatal("match command reused a Dota ID")
	}
}

func TestCallbackSelectsPlayerForClickingUser(t *testing.T) {
	b, svc, _ := selectionTestBot(t)
	b.rememberPlayer(selectionMessage(42, 1, "player", "111"), 111)
	query := testQuery("player:222") // callback user 77, same chat 42
	b.handleUpdate(tgbotapi.Update{CallbackQuery: query})
	b.handleStats(selectionMessage(42, 77, "stats", ""))
	if svc.statsID != 222 {
		t.Fatal("profile button did not select player for clicking user")
	}
	b.handleStats(selectionMessage(42, 1, "stats", ""))
	if svc.statsID != 111 {
		t.Fatal("button changed another user's choice")
	}
}

func TestSelectionsConcurrentAccess(t *testing.T) {
	b := &Bot{}
	var wg sync.WaitGroup
	for user := int64(1); user <= 20; user++ {
		wg.Add(1)
		go func(user int64) {
			defer wg.Done()
			for range 30 {
				b.rememberPlayer(selectionMessage(10, user, "player", ""), user)
				id, err := b.resolveDotaID(selectionMessage(10, user, "stats", ""))
				if err != nil || id != user {
					t.Error("concurrent selection mismatch")
				}
			}
		}(user)
	}
	wg.Wait()
}
