// Package meta builds a position-specific ranking from STRATZ daily counters.
package meta

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

var (
	ErrDisabled     = errors.New("STRATZ token not configured")
	ErrUnavailable  = errors.New("meta source unavailable")
	ErrUnauthorized = errors.New("STRATZ access denied")
	ErrRateLimited  = errors.New("STRATZ rate limit")
	ErrInvalidData  = errors.New("invalid meta source data")
)

const MinMatches = 200

type Patch struct {
	Number    string `json:"patch_number"`
	Timestamp int64  `json:"patch_timestamp"`
}

type Hero struct {
	ID      int
	Name    string
	Wins    int64
	Matches int64
	Score   float64
}

type Report struct {
	Position                 int
	Patch                    string
	From, Through, FetchedAt time.Time
	Heroes                   []Hero
}

type dayRow struct {
	Day     int64 `json:"day"`
	HeroID  int   `json:"heroId"`
	Wins    int64 `json:"winCount"`
	Matches int64 `json:"matchCount"`
}

func PositionName(position int) string {
	switch position {
	case 1:
		return "Керри"
	case 2:
		return "Мид"
	case 3:
		return "Оффлейн"
	case 4:
		return "Поддержка"
	case 5:
		return "Полная поддержка"
	default:
		return "Неизвестная позиция"
	}
}

func utcDay(t time.Time) time.Time { return t.UTC().Truncate(24 * time.Hour) }

// Exclude the entire release day: a daily bucket can contain two patches.
func windowStart(now time.Time, patch Patch) time.Time {
	start := utcDay(now).AddDate(0, 0, -7)
	afterRelease := utcDay(time.Unix(patch.Timestamp, 0)).AddDate(0, 0, 1)
	if afterRelease.After(start) {
		return afterRelease
	}
	return start
}

func rank(rows []dayRow, names map[int]string, patch Patch, position int, now time.Time) (Report, error) {
	report := Report{Position: position, Patch: patch.Number, FetchedAt: now.UTC()}
	start, end := windowStart(now, patch), utcDay(now)
	if !start.Before(end) {
		return report, nil
	}
	totals := map[int]Hero{}
	days := map[int64]bool{}
	seen := map[[2]int64]bool{}
	for _, row := range rows {
		day := time.Unix(row.Day, 0).UTC()
		if day.Before(start) || !day.Before(end) {
			continue
		}
		key := [2]int64{row.Day, int64(row.HeroID)}
		if !day.Equal(utcDay(day)) || row.HeroID <= 0 || row.Matches < 0 || row.Wins < 0 || row.Wins > row.Matches || seen[key] {
			return Report{}, ErrInvalidData
		}
		seen[key] = true
		if row.Matches == 0 {
			continue
		}
		days[row.Day] = true
		if report.From.IsZero() || day.Before(report.From) {
			report.From = day
		}
		if report.Through.IsZero() || day.After(report.Through) {
			report.Through = day
		}
		hero := totals[row.HeroID]
		hero.ID = row.HeroID
		hero.Name = names[row.HeroID]
		if hero.Name == "" {
			hero.Name = fmt.Sprintf("Герой #%d", row.HeroID)
		}
		hero.Wins += row.Wins
		hero.Matches += row.Matches
		totals[row.HeroID] = hero
	}
	if report.Through.IsZero() {
		// A newly released patch can legitimately have no complete daily bucket yet.
		if end.Sub(start) <= 48*time.Hour {
			return report, nil
		}
		return Report{}, ErrUnavailable
	}
	if end.Sub(report.Through) > 48*time.Hour {
		return Report{}, ErrUnavailable
	}
	// Do not silently rank a partial response with missing days.
	for day := start; !day.After(report.Through); day = day.AddDate(0, 0, 1) {
		if !days[day.Unix()] {
			return Report{}, ErrInvalidData
		}
	}
	for _, hero := range totals {
		if hero.Matches < MinMatches {
			continue
		}
		hero.Score = wilson(hero.Wins, hero.Matches)
		report.Heroes = append(report.Heroes, hero)
	}
	sort.Slice(report.Heroes, func(i, j int) bool {
		a, b := report.Heroes[i], report.Heroes[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Matches != b.Matches {
			return a.Matches > b.Matches
		}
		return a.ID < b.ID
	})
	if len(report.Heroes) > 10 {
		report.Heroes = report.Heroes[:10]
	}
	return report, nil
}

func wilson(wins, matches int64) float64 {
	if matches <= 0 {
		return 0
	}
	n, p, z := float64(matches), float64(wins)/float64(matches), 1.96
	return (p + z*z/(2*n) - z*math.Sqrt((p*(1-p)+z*z/(4*n))/n)) / (1 + z*z/n)
}
