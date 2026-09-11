package dota

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func cacheTestClient(now func() time.Time, respond func(*http.Request) (int, string)) *Client {
	return &Client{baseURL: "https://opendota.test", cache: newResponseCache(512, 16*1024*1024, now), httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		status, body := respond(req)
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
}

func TestEndpointCacheExpiry(t *testing.T) {
	for _, tt := range []struct {
		name, path, body string
		ttl              time.Duration
		fetch            func(*Client) error
	}{
		{"profile", "/players/123", `{"profile":{"account_id":123,"personaname":"Player"}}`, 2 * time.Minute, func(c *Client) error { _, err := c.GetPlayer(123); return err }},
		{"win loss", "/players/123/wl", `{"win":3,"lose":2}`, 2 * time.Minute, func(c *Client) error { _, err := c.GetWinLoss(123); return err }},
		{"recent", "/players/123/recentMatches", `[{"match_id":456}]`, 45 * time.Second, func(c *Client) error { _, err := c.GetRecentMatches(123); return err }},
		{"heroes", "/players/123/heroes", `[{"hero_id":1,"games":5}]`, 2 * time.Minute, func(c *Client) error { _, err := c.GetPlayerHeroes(123); return err }},
		{"totals", "/players/123/totals", `[{"field":"kills","n":2,"sum":5}]`, 2 * time.Minute, func(c *Client) error { _, err := c.GetTotals(123); return err }},
		{"match", "/matches/456", `{"match_id":456}`, 10 * time.Minute, func(c *Client) error { _, err := c.GetMatch(456); return err }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Unix(1000, 0)
			calls := 0
			client := cacheTestClient(func() time.Time { return now }, func(req *http.Request) (int, string) {
				calls++
				if req.URL.Path != tt.path {
					t.Errorf("wrong endpoint: %s", req.URL.Path)
				}
				return 200, tt.body
			})
			if err := tt.fetch(client); err != nil {
				t.Fatal(err)
			}
			now = now.Add(tt.ttl - time.Nanosecond)
			if err := tt.fetch(client); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("cache miss before expiry: %d requests", calls)
			}
			now = now.Add(time.Nanosecond)
			if err := tt.fetch(client); err != nil {
				t.Fatal(err)
			}
			if calls != 2 {
				t.Errorf("expired entry used: %d requests", calls)
			}
		})
	}
}

func TestCacheDoesNotStoreErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
	}{
		{"not found", 404, `{}`}, {"rate limit", 429, `{}`}, {"server", 500, `{}`}, {"invalid JSON", 200, `{`}, {"empty profile", 200, `{"profile":{"account_id":123}}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			client := cacheTestClient(time.Now, func(*http.Request) (int, string) {
				calls++
				if calls == 1 {
					return tt.status, tt.body
				}
				return 200, `{"profile":{"account_id":123,"personaname":"Recovered"}}`
			})
			if _, err := client.GetPlayer(123); err == nil {
				t.Fatal("expected error")
			}
			player, err := client.GetPlayer(123)
			if err != nil || player.Profile.Personaname != "Recovered" {
				t.Fatalf("retry failed: %v", err)
			}
			if _, err := client.GetPlayer(123); err != nil {
				t.Fatal(err)
			}
			if calls != 2 {
				t.Errorf("expected error retry then cache hit, got %d requests", calls)
			}
		})
	}
}

func TestCacheIsolatesResultsAndKeys(t *testing.T) {
	client := cacheTestClient(time.Now, func(req *http.Request) (int, string) {
		switch req.URL.Path {
		case "/players/1/wl":
			return 200, `{"win":10}`
		case "/players/2/wl":
			return 200, `{"win":20}`
		case "/players/1/heroes":
			return 200, `[{"hero_id":1},{"hero_id":2}]`
		case "/matches/1":
			return 200, `{"match_id":1,"players":[{"account_id":1,"kills":5}]}`
		default:
			t.Errorf("unexpected request: %s", req.URL.Path)
			return 404, `{}`
		}
	})
	wins, err := client.GetWinLoss(1)
	if err != nil {
		t.Fatal(err)
	}
	wins.Win = 999
	wins, err = client.GetWinLoss(1)
	if err != nil || wins.Win != 10 {
		t.Fatalf("changed cached wins: %+v, %v", wins, err)
	}
	other, err := client.GetWinLoss(2)
	if err != nil || other.Win != 20 {
		t.Fatal("different players share cache")
	}
	heroes, err := client.GetPlayerHeroes(1)
	if err != nil {
		t.Fatal(err)
	}
	heroes[0], heroes[1] = heroes[1], heroes[0]
	heroes, err = client.GetPlayerHeroes(1)
	if err != nil || heroes[0].HeroID != 1 {
		t.Fatal("cached slice changed by caller")
	}
	match, err := client.GetMatch(1)
	if err != nil {
		t.Fatal(err)
	}
	match.Players[0].Kills = 999
	*match.Players[0].AccountID = 999
	match, err = client.GetMatch(1)
	if err != nil || match.Players[0].Kills != 5 || *match.Players[0].AccountID != 1 {
		t.Fatal("nested cached model changed by caller")
	}
}

func TestCacheBoundsAndCleanup(t *testing.T) {
	now := time.Unix(1000, 0)
	cache := newResponseCache(2, 6, func() time.Time { return now })
	cache.put("a", []byte("aa"), time.Second)
	cache.put("b", []byte("bb"), time.Minute)
	cache.put("c", []byte("cc"), time.Minute)
	if _, ok := cache.get("a"); ok {
		t.Fatal("earliest expiry not evicted at entry limit")
	}
	cache.put("oversized", []byte("1234567"), time.Minute)
	if _, ok := cache.get("oversized"); ok {
		t.Fatal("oversized entry cached")
	}
	cache.put("c", []byte("12345"), time.Minute)
	if cache.bytes != 5 || len(cache.entries) != 1 {
		t.Fatalf("incorrect byte accounting: %d, %d", cache.bytes, len(cache.entries))
	}
	now = now.Add(time.Minute)
	cache.put("fresh", []byte("ok"), time.Minute)
	if cache.bytes != 2 || len(cache.entries) != 1 {
		t.Fatal("expired entries not removed on write")
	}
	now = now.Add(time.Minute)
	if _, ok := cache.get("fresh"); ok || cache.bytes != 0 {
		t.Fatal("expired entry not removed on read")
	}
}

func TestCacheConcurrentReadsAndWrites(t *testing.T) {
	var calls atomic.Int64
	client := cacheTestClient(time.Now, func(*http.Request) (int, string) { calls.Add(1); return 200, `{"win":10}` })
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				result, err := client.GetWinLoss(int64(j % 4))
				if err != nil {
					t.Error(err)
					return
				}
				if result.Win != 10 {
					t.Error("cached result changed by another caller")
				}
				result.Win = 999
			}
		}()
	}
	wg.Wait()
	if calls.Load() >= 640 {
		t.Fatal("cache never reused")
	}
	if NewClient().cache == nil {
		t.Fatal("production client has no cache")
	}
}

func TestEmptyMatchIsNotCached(t *testing.T) {
	calls := 0
	client := cacheTestClient(time.Now, func(*http.Request) (int, string) { calls++; return 200, `{}` })
	for i := 0; i < 2; i++ {
		if _, err := client.GetMatch(1); err == nil {
			t.Fatal("empty match accepted")
		}
	}
	if calls != 2 {
		t.Fatal(fmt.Sprintf("empty match cached: %d calls", calls))
	}
}
