package bot

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/internal/meta"
)

func testLimiter() (*commandLimiter, *time.Time) {
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	limiter := newCommandLimiter()
	limiter.now = func() time.Time { return now }
	return limiter, &now
}

func TestCommandLimiterRefillAndNotices(t *testing.T) {
	limiter, now := testLimiter()
	key := commandUserKey{kind: 'u', id: 1}
	for range commandBurst {
		if got := limiter.take(key); got.wait != 0 {
			t.Fatal(got)
		}
	}
	if got := limiter.take(key); got.wait != 3*time.Second || !got.notify {
		t.Fatal(got)
	}
	*now = now.Add(time.Second)
	if got := limiter.take(key); got.wait != 2*time.Second || got.notify {
		t.Fatal(got)
	}
	*now = now.Add(2 * time.Second)
	if got := limiter.take(key); got.wait != 0 {
		t.Fatal("did not refill", got)
	}
	if got := limiter.take(key); got.wait != 3*time.Second || got.notify {
		t.Fatal(got)
	}
	*now = now.Add(2 * time.Second)
	if got := limiter.take(key); got.wait != time.Second || !got.notify {
		t.Fatal(got)
	}
	*now = now.Add(time.Hour)
	for range commandBurst {
		if got := limiter.take(key); got.wait != 0 {
			t.Fatal(got)
		}
	}
	if got := limiter.take(key); got.wait == 0 {
		t.Fatal("refill exceeded burst")
	}
	if commandLimitText(time.Millisecond) != "⏳ Слишком частые запросы. Повтори через 1 сек." {
		t.Fatal("wrong rounding")
	}
}

func TestCommandLimiterIdentityAndFreeCommands(t *testing.T) {
	limiter, _ := testLimiter()
	b := &Bot{limiter: limiter}
	for range commandBurst {
		b.limitCommand(selectionMessage(10, 1, "stats", "123"))
	}
	if b.limitCommand(selectionMessage(20, 1, "meta", "1")).wait == 0 {
		t.Fatal("chat switching bypassed limiter")
	}
	if b.limitCommand(selectionMessage(10, 2, "stats", "123")).wait != 0 {
		t.Fatal("another user was throttled")
	}
	for _, command := range []string{"help", "start", "unknown", "meta"} {
		if b.limitCommand(selectionMessage(10, 1, command, "")).wait != 0 {
			t.Fatal("free command throttled: " + command)
		}
	}
	// Missing/anonymous senders are bounded by chat identity, never a fake user.
	for _, kind := range []string{"no sender", "zero user", "sender chat"} {
		t.Run(kind, func(t *testing.T) {
			local, _ := testLimiter()
			bot := &Bot{limiter: local}
			message := selectionMessage(-10, 1, "stats", "123")
			switch kind {
			case "no sender":
				message.From = nil
			case "zero user":
				message.From.ID = 0
			case "sender chat":
				message.SenderChat = &tgbotapi.Chat{ID: -10}
			}
			for range commandBurst {
				bot.limitCommand(message)
			}
			if bot.limitCommand(message).wait == 0 {
				t.Fatal("anonymous sender bypassed limiter")
			}
			if bot.limitCommand(selectionMessage(-10, 1, "stats", "123")).wait != 0 {
				t.Fatal("anonymous sender consumed real user's budget")
			}
		})
	}
}

func TestCommandLimiterConcurrencyAndCapacity(t *testing.T) {
	limiter, now := testLimiter()
	var admitted atomic.Int32
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.take(commandUserKey{kind: 'u', id: 1}).wait == 0 {
				admitted.Add(1)
			}
		}()
	}
	wg.Wait()
	if admitted.Load() != commandBurst {
		t.Fatal(admitted.Load())
	}
	for user := range commandMaxUsers + 20 {
		limiter.take(commandUserKey{kind: 'u', id: int64(user)})
	}
	if len(limiter.buckets) != commandMaxUsers {
		t.Fatal("unbounded storage", len(limiter.buckets))
	}
	*now = now.Add(commandIdleTTL)
	limiter.take(commandUserKey{kind: 'u', id: 99})
	if len(limiter.buckets) != 1 {
		t.Fatal("idle entries not removed", len(limiter.buckets))
	}
}

type rateMetaSource struct {
	calls  int
	events *[]string
}

func (s *rateMetaSource) Get(_ context.Context, position int) (meta.Report, error) {
	s.calls++
	*s.events = append(*s.events, "fetch")
	return meta.Report{Position: position, Patch: "7.41e"}, nil
}

func TestRateLimitCommandsAndCallbacksShareBudget(t *testing.T) {
	var texts, acks, events []string
	api, err := tgbotapi.NewBotAPIWithClient("test-token", "https://telegram.test/bot%s/%s", telegramHTTPFunc(func(req *http.Request) (*http.Response, error) {
		if err := req.ParseForm(); err != nil {
			t.Fatal(err)
		}
		payload := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test"}}`
		switch {
		case strings.HasSuffix(req.URL.Path, "/answerCallbackQuery"):
			acks = append(acks, req.Form.Get("text"))
			events = append(events, "ack")
			payload = `{"ok":true,"result":true}`
		case strings.HasSuffix(req.URL.Path, "/sendMessage"):
			texts = append(texts, req.Form.Get("text"))
			payload = `{"ok":true,"result":{"message_id":1}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	limiter, now := testLimiter()
	source := &rateMetaSource{events: &events}
	svc := &selectionService{}
	b := &Bot{api: api, limiter: limiter, meta: source, service: svc}
	b.handleUpdate(tgbotapi.Update{Message: selectionMessage(42, 77, "meta", "1")})
	for range 4 {
		b.handleUpdate(tgbotapi.Update{CallbackQuery: testQuery("meta:2")})
	}
	if source.calls != 5 {
		t.Fatalf("callbacks consumed two tokens: %d", source.calls)
	}
	if strings.Join(events, ",") != "fetch,ack,fetch,ack,fetch,ack,fetch,ack,fetch" {
		t.Fatal("ack must precede fetch", events)
	}
	b.handleUpdate(tgbotapi.Update{CallbackQuery: testQuery("meta:3")})
	if source.calls != 5 || len(acks) != 5 || !strings.Contains(acks[4], "3 сек.") {
		t.Fatal("throttled callback not acknowledged", acks)
	}
	// Data commands are blocked before OpenDota as well, across chats.
	b.handleUpdate(tgbotapi.Update{Message: selectionMessage(99, 77, "stats", "123")})
	if svc.statsID != 0 {
		t.Fatal("throttled request reached OpenDota")
	}
	before := len(texts)
	for range 10 {
		b.handleUpdate(tgbotapi.Update{Message: selectionMessage(42, 77, "meta", "1")})
	}
	if len(texts) != before || source.calls != 5 {
		t.Fatal("spam produced repeated messages or requests")
	}
	b.handleUpdate(tgbotapi.Update{Message: selectionMessage(42, 77, "help", "")})
	b.handleUpdate(tgbotapi.Update{Message: selectionMessage(42, 77, "meta", "")})
	if len(texts) != before+2 {
		t.Fatal("help/menu unavailable while throttled")
	}
	b.handleUpdate(tgbotapi.Update{Message: selectionMessage(42, 78, "stats", "456")})
	if svc.statsID != 456 {
		t.Fatal("another user was blocked")
	}
	*now = now.Add(commandRefill)
	b.handleUpdate(tgbotapi.Update{CallbackQuery: testQuery("meta:4")})
	if source.calls != 6 || acks[len(acks)-1] != "" {
		t.Fatal("did not resume after refill")
	}
	b.handleCallback(nil)
}

func TestRateLimitCommandNotice(t *testing.T) {
	b, _, texts := selectionTestBot(t)
	limiter, _ := testLimiter()
	b.limiter = limiter
	source := &fakeMeta{}
	b.meta = source
	for range commandBurst + 3 {
		b.handleUpdate(tgbotapi.Update{Message: selectionMessage(10, 1, "meta", "1")})
	}
	if len(source.calls) != commandBurst || len(*texts) != commandBurst+1 || !strings.Contains((*texts)[commandBurst], "3 сек.") {
		t.Fatal("wrong command throttle", *texts)
	}
}
