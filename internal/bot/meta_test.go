package bot

import (
	"context"
	"errors"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/internal/meta"
)

type fakeMeta struct {
	calls []int
	err   error
}

func (f *fakeMeta) Get(ctx context.Context, position int) (meta.Report, error) {
	f.calls = append(f.calls, position)
	return meta.Report{Position: position, Patch: "7.41e"}, f.err
}

func TestMetaCommandsAndCallbacks(t *testing.T) {
	var payloads []string
	for _, row := range metaKeyboard().InlineKeyboard {
		for _, button := range row {
			if button.CallbackData == nil {
				t.Fatal("missing callback")
			}
			payloads = append(payloads, *button.CallbackData)
		}
	}
	if strings.Join(payloads, ",") != "meta:1,meta:2,meta:3,meta:4,meta:5" {
		t.Fatal(payloads)
	}
	b, _, texts := selectionTestBot(t)
	provider := &fakeMeta{}
	b.meta = provider
	for _, args := range []string{"", "0", "6", "01", "-1", "1 extra", "999999999999999"} {
		b.handleUpdate(tgbotapi.Update{Message: selectionMessage(10, 1, "meta", args)})
	}
	if len(provider.calls) != 0 {
		t.Fatal("invalid command reached source")
	}
	for position := 1; position <= 5; position++ {
		q := &tgbotapi.CallbackQuery{ID: "test", From: &tgbotapi.User{ID: 7}, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: -10}, Date: 1}, Data: "meta:" + string(rune('0'+position))}
		message, err := callbackMessage(q)
		if err != nil || message.From.ID != 7 || message.Chat.ID != -10 {
			t.Fatalf("callback: %+v %v", message, err)
		}
		b.handleUpdate(tgbotapi.Update{Message: message})
	}
	if len(provider.calls) != 5 {
		t.Fatal(provider.calls)
	}
	for i, pos := range provider.calls {
		if pos != i+1 {
			t.Fatal(provider.calls)
		}
	}
	for _, data := range []string{"meta:0", "meta:6", "meta:01", "meta:1:2"} {
		if _, err := callbackMessage(&tgbotapi.CallbackQuery{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{}, Date: 1}, Data: data}); err == nil {
			t.Fatal("accepted " + data)
		}
	}
	provider.err = meta.ErrUnauthorized
	b.handleMeta(selectionMessage(10, 1, "meta", "1"))
	if !strings.Contains((*texts)[len(*texts)-1], "отклонил доступ") {
		t.Fatal(*texts)
	}
	provider.err = errors.New("secret upstream text")
	b.handleMeta(selectionMessage(10, 1, "meta", "1"))
	if strings.Contains((*texts)[len(*texts)-1], "secret") {
		t.Fatal("leaked upstream error")
	}
	b.meta = nil
	b.handleMeta(selectionMessage(10, 1, "meta", "1"))
	if !strings.Contains((*texts)[len(*texts)-1], "STRATZ_API_TOKEN") {
		t.Fatal(*texts)
	}
}
