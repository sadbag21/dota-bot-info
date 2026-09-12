package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type cachedReport struct {
	report Report
	until  time.Time
}

type Client struct {
	token                string
	http                 *http.Client
	endpoint, patchesURL string
	gate                 chan struct{}
	now                  func() time.Time
	patch                Patch
	patchUntil           time.Time
	cache                map[int]cachedReport
}

func NewClient(token string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp4", addr)
	}
	return &Client{
		token:      strings.TrimSpace(token),
		http:       &http.Client{Transport: transport, Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		endpoint:   "https://api.stratz.com/graphql",
		patchesURL: "https://www.dota2.com/datafeed/patchnoteslist?language=english",
		gate:       make(chan struct{}, 1), now: time.Now, cache: map[int]cachedReport{},
	}
}

func (c *Client) Close() { c.http.CloseIdleConnections() }

// One gate coalesces concurrent cache misses and bounds source traffic to one
// active request. Waiters can cancel without waiting for the network request.
func (c *Client) Get(ctx context.Context, position int) (Report, error) {
	if position < 1 || position > 5 {
		return Report{}, ErrInvalidData
	}
	if c.token == "" {
		return Report{}, ErrDisabled
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	select {
	case c.gate <- struct{}{}:
	case <-ctx.Done():
		return Report{}, ctx.Err()
	}
	defer func() { <-c.gate }()
	now := c.now().UTC()
	if !now.Before(c.patchUntil) {
		patch, err := c.fetchPatch(ctx, now)
		if err != nil {
			return Report{}, err
		}
		if patch != c.patch {
			c.cache = map[int]cachedReport{}
		}
		c.patch, c.patchUntil = patch, now.Add(15*time.Minute)
	}
	if cached, ok := c.cache[position]; ok && now.Before(cached.until) && utcDay(cached.report.FetchedAt).Equal(utcDay(now)) {
		return clone(cached.report), nil
	}
	if !windowStart(now, c.patch).Before(utcDay(now)) {
		return Report{Position: position, Patch: c.patch.Number, FetchedAt: now}, nil
	}
	// No groupBy: preserve hero/day buckets so release-day data can be excluded.
	query := fmt.Sprintf(`{ constants { heroes { id displayName } } heroStats { winDay(take: 7, skip: 0, bracketIds: [IMMORTAL], positionIds: [POSITION_%d], gameModeIds: [ALL_PICK_RANKED]) { day heroId winCount matchCount } } }`, position)
	body, _ := json.Marshal(map[string]string{"query": query})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return Report{}, ErrUnavailable
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "STRATZ_API")
	req.Header.Set("Authorization", "Bearer "+c.token)
	var response struct {
		Errors []json.RawMessage `json:"errors"`
		Data   *struct {
			Constants struct {
				Heroes []struct {
					ID   int    `json:"id"`
					Name string `json:"displayName"`
				} `json:"heroes"`
			} `json:"constants"`
			Stats struct {
				Rows []dayRow `json:"winDay"`
			} `json:"heroStats"`
		} `json:"data"`
	}
	if err := c.request(req, &response); err != nil {
		return Report{}, err
	}
	// Do not expose upstream messages: they may echo request input or credentials.
	if len(response.Errors) > 0 || response.Data == nil {
		return Report{}, ErrUnavailable
	}
	if len(response.Data.Constants.Heroes) == 0 {
		return Report{}, ErrInvalidData
	}
	names := map[int]string{}
	for _, hero := range response.Data.Constants.Heroes {
		names[hero.ID] = hero.Name
	}
	report, err := rank(response.Data.Stats.Rows, names, c.patch, position, now)
	if err != nil {
		return Report{}, err
	}
	c.cache[position] = cachedReport{report: clone(report), until: now.Add(time.Hour)}
	return report, nil
}

func clone(r Report) Report { r.Heroes = append([]Hero(nil), r.Heroes...); return r }

var patchNumber = regexp.MustCompile(`^\d+\.\d+[a-z]?$`)

func (c *Client) fetchPatch(ctx context.Context, now time.Time) (Patch, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.patchesURL, nil)
	if err != nil {
		return Patch{}, ErrUnavailable
	}
	var feed struct {
		Success bool    `json:"success"`
		Patches []Patch `json:"patches"`
	}
	if err := c.request(req, &feed); err != nil {
		return Patch{}, err
	}
	if !feed.Success {
		return Patch{}, ErrInvalidData
	}
	var latest Patch
	for _, patch := range feed.Patches {
		if patch.Timestamp > 0 && patch.Timestamp <= now.Unix() && patch.Timestamp > latest.Timestamp {
			if !patchNumber.MatchString(patch.Number) {
				return Patch{}, ErrInvalidData
			}
			latest = patch
		}
	}
	if latest.Number == "" {
		return Patch{}, ErrInvalidData
	}
	return latest, nil
}

func (c *Client) request(req *http.Request, target any) error {
	response, err := c.http.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return req.Context().Err()
		}
		return ErrUnavailable
	}
	defer func() {
		// Closing cannot change the decoded result or the request/read error.
		_ = response.Body.Close()
	}()
	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return ErrUnavailable
	}
	const maxBody = 4 << 20
	data, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil {
		if req.Context().Err() != nil {
			return req.Context().Err()
		}
		return ErrUnavailable
	}
	if len(data) > maxBody || json.Unmarshal(data, target) != nil {
		return ErrInvalidData
	}
	return nil
}
