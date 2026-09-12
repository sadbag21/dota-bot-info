package meta

import (
	"errors"
	"testing"
	"time"
)

func TestRankingWindowAndWeights(t *testing.T) {
	now := time.Date(2026, 9, 11, 14, 0, 0, 0, time.UTC)
	patch := Patch{Number: "7.41e", Timestamp: time.Date(2026, 9, 8, 15, 0, 0, 0, time.UTC).Unix()}
	rows := []dayRow{
		{Day: time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC).Unix(), HeroID: 1, Wins: 10000, Matches: 10000},
		{Day: time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC).Unix(), HeroID: 1, Wins: 90, Matches: 100},
		{Day: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC).Unix(), HeroID: 1, Wins: 110, Matches: 300},
		{Day: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC).Unix(), HeroID: 2, Wins: 180, Matches: 300},
		{Day: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC).Unix(), HeroID: 3, Wins: 199, Matches: 199},
		{Day: utcDay(now).Unix(), HeroID: 4, Wins: 900, Matches: 900},
	}
	r, err := rank(rows, map[int]string{1: "A", 2: "B"}, patch, 5, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Heroes) != 2 || r.Heroes[0].ID != 2 || r.Heroes[1].Wins != 200 || r.Heroes[1].Matches != 400 {
		t.Fatalf("wrong aggregation: %+v", r)
	}
	if r.From.Day() != 9 || r.Through.Day() != 10 || r.Position != 5 {
		t.Fatalf("wrong period: %+v", r)
	}
	if wilson(55, 100) >= wilson(5500, 10000) {
		t.Fatal("small sample ranked too highly")
	}
}

func TestRankingRejectsBadOrStaleData(t *testing.T) {
	now := time.Date(2026, 9, 11, 14, 0, 0, 0, time.UTC)
	patch := Patch{Number: "7.41e", Timestamp: now.AddDate(0, 0, -30).Unix()}
	valid := []dayRow{}
	for i := 7; i >= 1; i-- {
		valid = append(valid, dayRow{Day: utcDay(now).AddDate(0, 0, -i).Unix(), HeroID: 1, Wins: 30, Matches: 50})
	}
	tests := map[string][]dayRow{
		"missing day": append(append([]dayRow{}, valid[:2]...), valid[3:]...),
		"duplicate":   append(append([]dayRow{}, valid...), valid[0]),
		"stale":       valid[:3],
		"empty":       nil,
	}
	bad := append([]dayRow{}, valid...)
	bad[0].Wins = 51
	tests["wins exceed matches"] = bad
	for name, rows := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := rank(rows, nil, patch, 1, now); err == nil {
				t.Fatal("accepted bad data")
			}
		})
	}
	if _, err := rank(valid, nil, patch, 1, now); err != nil {
		t.Fatal(err)
	}
	fresh := Patch{Number: "7.42", Timestamp: now.Add(-time.Hour).Unix()}
	if r, err := rank(valid, nil, fresh, 1, now); err != nil || len(r.Heroes) != 0 {
		t.Fatalf("fresh patch reused old data: %+v %v", r, err)
	}
	if _, err := rank(nil, nil, patch, 1, now); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
