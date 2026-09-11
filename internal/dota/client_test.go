package dota

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type timeoutError struct{}

func (timeoutError) Error() string   { return "network timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestGetResponse(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr error
	}{
		{"success", 200, `{"win":12,"lose":3}`, nil},
		{"not found", 404, `not JSON`, ErrNotFound},
		{"rate limited", 429, `not JSON`, ErrRateLimited},
		{"server error", 500, `not JSON`, ErrUnavailable},
		{"bad gateway", 502, `not JSON`, ErrUnavailable},
		{"origin timeout", 522, `not JSON`, ErrUnavailable},
		{"bad request", 400, `{}`, ErrBadResponse},
		{"no content", 204, ``, ErrBadResponse},
		{"broken JSON", 200, `{"win":`, ErrBadResponse},
		{"HTML", 200, `<html>error</html>`, ErrBadResponse},
		{"empty body", 200, ``, ErrBadResponse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader(tt.body)}
			client := &Client{baseURL: "https://opendota.test/api", httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet || req.URL.Path != "/api/players/123/wl" {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL)
				}
				if req.Header.Get("Accept") != "application/json" {
					t.Error("missing JSON Accept header")
				}
				return &http.Response{StatusCode: tt.status, Body: body, Header: make(http.Header)}, nil
			})}}
			var result WinLoss
			err := client.get("/players/123/wl", &result)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
			if !body.closed {
				t.Error("response body was not closed")
			}
			if tt.wantErr == nil && (result.Win != 12 || result.Lose != 3) {
				t.Errorf("unexpected result: %+v", result)
			}
		})
	}
}

func TestGetIOErrors(t *testing.T) {
	connectionErr := errors.New("connection refused")
	for _, tt := range []struct {
		name        string
		cause, want error
	}{
		{"deadline", context.DeadlineExceeded, ErrTimeout},
		{"wrapped deadline", fmt.Errorf("read: %w", context.DeadlineExceeded), ErrTimeout},
		{"network timeout", timeoutError{}, ErrTimeout},
		{"connection failure", connectionErr, ErrUnavailable},
	} {
		for _, phase := range []string{"request", "body"} {
			t.Run(tt.name+"/"+phase, func(t *testing.T) {
				body := &trackedBody{Reader: failingReader{tt.cause}}
				client := &Client{baseURL: "https://opendota.test", httpClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					if phase == "request" {
						return nil, tt.cause
					}
					return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
				})}}
				var result WinLoss
				err := client.get("/test", &result)
				want := tt.want
				if phase == "body" && want == ErrUnavailable {
					want = ErrBadResponse
				}
				if !errors.Is(err, want) || !errors.Is(err, tt.cause) {
					t.Fatalf("got %v, want category %v and cause %v", err, want, tt.cause)
				}
				if phase == "body" && !body.closed {
					t.Error("body was not closed after read failure")
				}
			})
		}
	}
}

func TestGetPlayerProfile(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		wantErr    error
	}{
		{"null profile", `{"profile":null}`, ErrNotFound},
		{"missing profile", `{}`, ErrNotFound},
		{"ID only", `{"profile":{"account_id":123}}`, ErrNotFound},
		{"missing ID", `{"profile":{"personaname":"Player"}}`, ErrNotFound},
		{"name", `{"profile":{"account_id":123,"personaname":"Player"}}`, nil},
		{"Steam ID only", `{"profile":{"account_id":123,"steamid":"76561197960265851"}}`, nil},
		{"profile URL only", `{"profile":{"account_id":123,"profileurl":"https://steamcommunity.com/profiles/76561197960265851"}}`, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{baseURL: "https://opendota.test", httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/players/123" {
					t.Errorf("unexpected path: %s", req.URL.Path)
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(tt.body)), Header: make(http.Header)}, nil
			})}}
			player, err := client.GetPlayer(123)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && player != nil {
				t.Error("invalid profile returned as a player")
			}
			if tt.wantErr == nil && (player == nil || player.Profile.AccountID != 123) {
				t.Errorf("unexpected player: %+v", player)
			}
		})
	}
}
