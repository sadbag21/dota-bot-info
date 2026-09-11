package bot

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func testQuery(data string) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{ID: "callback-1", Data: data, From: &tgbotapi.User{ID: 77}, Message: &tgbotapi.Message{Date: 1, Text: "Original response", Chat: &tgbotapi.Chat{ID: 42}}}
}

func TestKeyboardNavigation(t *testing.T) {
	for _, tt := range []struct {
		name     string
		keyboard tgbotapi.InlineKeyboardMarkup
		id       string
		commands []string
	}{
		{"player", playerKeyboard(4294967295), "4294967295", []string{"matches", "heroes", "stats", "player"}},
		{"match", matchKeyboard(9223372036854775807), "9223372036854775807", []string{"impact", "match"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			seen := map[string]bool{}
			for _, row := range tt.keyboard.InlineKeyboard {
				for _, button := range row {
					if button.URL != nil {
						if *button.URL != "https://www.opendota.com/matches/"+tt.id {
							t.Errorf("wrong URL: %s", *button.URL)
						}
						continue
					}
					if button.CallbackData == nil {
						t.Fatal("button has no action")
					}
					if len(*button.CallbackData) > 64 {
						t.Fatal("callback data exceeds Telegram limit")
					}
					query := testQuery(*button.CallbackData)
					message, err := callbackMessage(query)
					if err != nil {
						t.Fatal(err)
					}
					seen[message.Command()] = true
					if message.CommandArguments() != tt.id || message.Chat.ID != 42 || message.From.ID != 77 {
						t.Fatalf("incorrect routing: %+v", message)
					}
					if query.Message.Text != "Original response" {
						t.Error("original message modified")
					}
				}
			}
			for _, command := range tt.commands {
				if !seen[command] {
					t.Errorf("missing button: %s", command)
				}
			}
		})
	}
}

func TestRejectInvalidCallbacks(t *testing.T) {
	for _, data := range []string{"", "player", "player:", "player:0", "player:-1", "player:+1", "player:4294967296", "player:999999999999", "player:12 34", "player:123:456", "player: 123", "help:123", "unknown:123", "impact:9223372036854775808", "match:abc", "match:" + strings.Repeat("1", 65)} {
		t.Run(data, func(t *testing.T) {
			if _, err := callbackMessage(testQuery(data)); err == nil {
				t.Fatal("invalid callback accepted")
			}
		})
	}
	for _, query := range []*tgbotapi.CallbackQuery{nil, {}, {Message: &tgbotapi.Message{}}, {Data: "player:123", Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 42}}}} {
		if _, err := callbackMessage(query); err == nil {
			t.Fatal("inaccessible message accepted")
		}
	}
}

type telegramHTTPFunc func(*http.Request) (*http.Response, error)

func (f telegramHTTPFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestInvalidCallbackIsAnsweredWithoutService(t *testing.T) {
	var calls []string
	api, err := tgbotapi.NewBotAPIWithClient("test-token", "https://telegram.test/bot%s/%s", telegramHTTPFunc(func(req *http.Request) (*http.Response, error) {
		calls = append(calls, req.URL.Path)
		payload := `{"ok":true,"result":true}`
		if strings.HasSuffix(req.URL.Path, "/getMe") {
			payload = `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
		} else {
			if err := req.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if req.Form.Get("callback_query_id") != "callback-1" || req.Form.Get("text") == "" {
				t.Error("missing callback acknowledgement or explanation")
			}
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	b := &Bot{api: api}
	b.handleUpdate(tgbotapi.Update{CallbackQuery: testQuery("player:999999999999")})
	if len(calls) != 2 || !strings.HasSuffix(calls[1], "/answerCallbackQuery") {
		t.Fatalf("unexpected API calls: %v", calls)
	}
}

func TestSendKeyboard(t *testing.T) {
	sent := false
	api, err := tgbotapi.NewBotAPIWithClient("test-token", "https://telegram.test/bot%s/%s", telegramHTTPFunc(func(req *http.Request) (*http.Response, error) {
		payload := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
		if strings.HasSuffix(req.URL.Path, "/sendMessage") {
			sent = true
			if err := req.ParseForm(); err != nil {
				t.Fatal(err)
			}
			var markup tgbotapi.InlineKeyboardMarkup
			if err := json.Unmarshal([]byte(req.Form.Get("reply_markup")), &markup); err != nil {
				t.Fatal(err)
			}
			if len(markup.InlineKeyboard) != 2 || req.Form.Get("parse_mode") != "HTML" || req.Form.Get("chat_id") != "42" {
				t.Errorf("unexpected message: %v", req.Form)
			}
			payload = `{"ok":true,"result":{"message_id":1}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	(&Bot{api: api}).sendMessage(42, "Profile", playerKeyboard(123))
	if !sent {
		t.Fatal("message not sent")
	}
}
