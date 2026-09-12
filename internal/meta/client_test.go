package meta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientCachePatchChangeAndAuthIsolation(t *testing.T) {
	now := time.Date(2026, 9, 11, 14, 0, 0, 0, time.UTC)
	var graphCalls atomic.Int32
	patch := Patch{Number: "7.41e", Timestamp: now.AddDate(0, 0, -30).Unix()}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/patches" {
			if r.Header.Get("Authorization") != "" {
				t.Error("STRATZ credential sent to patch source")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "patches": []Patch{patch}})
			return
		}
		graphCalls.Add(1)
		if r.Header.Get("Authorization") != "Bearer test-secret" || r.Header.Get("User-Agent") != "STRATZ_API" {
			t.Error("missing authorization")
		}
		var request struct{ Query string }
		_ = json.NewDecoder(r.Body).Decode(&request)
		if !strings.Contains(request.Query, "POSITION_5") || !strings.Contains(request.Query, "IMMORTAL") || !strings.Contains(request.Query, "ALL_PICK_RANKED") || strings.Contains(request.Query, "groupBy") {
			t.Error("wrong filters")
		}
		rows := []dayRow{}
		for i := 7; i >= 1; i-- {
			rows = append(rows, dayRow{Day: utcDay(now).AddDate(0, 0, -i).Unix(), HeroID: 1, Wins: 30, Matches: 50})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"constants": map[string]any{"heroes": []any{map[string]any{"id": 1, "displayName": "A"}}}, "heroStats": map[string]any{"winDay": rows}}})
	}))
	defer server.Close()
	client := NewClient("test-secret")
	defer client.Close()
	client.endpoint = server.URL + "/graphql"
	client.patchesURL = server.URL + "/patches"
	client.now = func() time.Time { return now }
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := client.Get(context.Background(), 5)
			if err != nil {
				t.Error(err)
				return
			}
			if len(r.Heroes) != 1 {
				t.Error("missing hero")
				return
			}
			r.Heroes[0].Name = "changed"
		}()
	}
	wg.Wait()
	if graphCalls.Load() != 1 {
		t.Fatalf("cache misses: %d", graphCalls.Load())
	}
	r, err := client.Get(context.Background(), 5)
	if err != nil || r.Heroes[0].Name != "A" {
		t.Fatal("mutable cache")
	}
	now = now.Add(16 * time.Minute)
	patch = Patch{Number: "7.42", Timestamp: now.Add(-time.Minute).Unix()}
	r, err = client.Get(context.Background(), 5)
	if err != nil || r.Patch != "7.42" || len(r.Heroes) != 0 {
		t.Fatalf("old patch cache reused: %+v %v", r, err)
	}
}

func TestClientErrorsAndCancellation(t *testing.T) {
	for _, tt := range []struct {
		status int
		body   string
		want   error
	}{
		{401, `test-secret`, ErrUnauthorized}, {403, `test-secret`, ErrUnauthorized}, {429, `test-secret`, ErrRateLimited},
		{500, `test-secret`, ErrUnavailable}, {200, `not json test-secret`, ErrInvalidData},
		{200, `{"errors":[{"message":"test-secret"}],"data":null}`, ErrUnavailable},
	} {
		t.Run(fmt.Sprint(tt.status, tt.body), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tt.status); fmt.Fprint(w, tt.body) }))
			defer server.Close()
			c := NewClient("test-secret")
			defer c.Close()
			c.endpoint = server.URL
			c.patch = Patch{Number: "7.41e", Timestamp: time.Now().AddDate(0, 0, -30).Unix()}
			c.patchUntil = time.Now().Add(time.Hour)
			_, err := c.Get(context.Background(), 1)
			if !errors.Is(err, tt.want) || strings.Contains(err.Error(), "test-secret") {
				t.Fatalf("unsafe or unexpected error: %v", err)
			}
		})
	}
	c := NewClient("test-secret")
	defer c.Close()
	c.gate <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Get(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	disabled := NewClient("")
	defer disabled.Close()
	if _, err := disabled.Get(context.Background(), 1); !errors.Is(err, ErrDisabled) {
		t.Fatal(err)
	}
}

func TestCancelActiveRequest(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done() }))
	defer server.Close()
	c := NewClient("test-secret")
	defer c.Close()
	c.patchesURL = server.URL
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := c.Get(ctx, 1); done <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request did not cancel")
	}
}
