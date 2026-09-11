package bot

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type lifecycleTransport func(*http.Request) (*http.Response, error)

func (f lifecycleTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestRunStopsDuringPollingAndRetry(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "polling"
		if fail {
			name = "retry"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			client := &http.Client{Transport: lifecycleTransport(func(req *http.Request) (*http.Response, error) {
				if strings.HasSuffix(req.URL.Path, "/getMe") {
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"ok":true,"result":{"id":1,"first_name":"Test"}}`)), Header: make(http.Header)}, nil
				}
				close(started)
				if fail {
					return nil, errors.New("network unavailable")
				}
				<-req.Context().Done()
				return nil, req.Context().Err()
			})}
			api, err := tgbotapi.NewBotAPIWithClient("test", "https://telegram.test/bot%s/%s", contextHTTPClient{ctx: ctx, client: client})
			if err != nil {
				t.Fatal(err)
			}
			b := &Bot{api: api, ctx: ctx, httpClient: client}
			done := make(chan struct{})
			go func() { defer close(done); b.Run() }()
			select {
			case <-started:
			case <-time.After(2 * time.Second):
				t.Fatal("polling did not start")
			}
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("shutdown waited for polling or retry timeout")
			}
		})
	}
}

func TestShutdownDoesNotSendErrorMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := &Bot{ctx: ctx}
	// A nil API would panic if shutdown attempted any outgoing messages.
	b.sendMessage(42, "test")
	b.reportCommandError(&tgbotapi.Message{}, 123, context.Canceled)
	b.Run()
}

func TestTelegramLogErrorOmitsToken(t *testing.T) {
	cause := errors.New("connection reset")
	err := &url.Error{Op: "Post", URL: "https://api.telegram.org/botSECRET/getUpdates", Err: cause}
	clean := safeTelegramError(err)
	if strings.Contains(clean.Error(), "SECRET") || !errors.Is(clean, cause) {
		t.Fatalf("unsafe or incomplete error: %v", clean)
	}
}
