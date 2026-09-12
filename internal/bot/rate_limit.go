package bot

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	commandBurst          = 5
	commandRefill         = 3 * time.Second
	commandIdleTTL        = 10 * time.Minute
	commandMaxUsers       = 4096
	commandNoticeInterval = 5 * time.Second
)

// The kind keeps Telegram users and anonymous chat senders in separate namespaces.
type commandUserKey struct {
	kind byte
	id   int64
}
type commandBucket struct {
	credit                time.Duration
	updated, seen, warned time.Time
}
type commandLimit struct {
	wait   time.Duration
	notify bool
}

type commandLimiter struct {
	mu          sync.Mutex
	now         func() time.Time
	buckets     map[commandUserKey]commandBucket
	lastCleanup time.Time
}

func newCommandLimiter() *commandLimiter {
	return &commandLimiter{now: time.Now, buckets: make(map[commandUserKey]commandBucket)}
}

func (l *commandLimiter) take(key commandUserKey) commandLimit {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= time.Minute || len(l.buckets) >= commandMaxUsers {
		for user, bucket := range l.buckets {
			if now.Sub(bucket.seen) >= commandIdleTTL {
				delete(l.buckets, user)
			}
		}
		l.lastCleanup = now
	}
	bucket, exists := l.buckets[key]
	if !exists {
		if len(l.buckets) >= commandMaxUsers {
			var oldestKey commandUserKey
			var oldest time.Time
			for user, entry := range l.buckets {
				if oldest.IsZero() || entry.seen.Before(oldest) {
					oldestKey, oldest = user, entry.seen
				}
			}
			delete(l.buckets, oldestKey)
		}
		bucket = commandBucket{credit: commandBurst * commandRefill, updated: now}
	}
	elapsed := now.Sub(bucket.updated)
	if elapsed > 0 {
		bucket.credit += min(elapsed, commandBurst*commandRefill-bucket.credit)
		bucket.updated = now
	}
	bucket.seen = now
	result := commandLimit{}
	if bucket.credit >= commandRefill {
		bucket.credit -= commandRefill
	} else {
		result.wait = commandRefill - bucket.credit
		result.notify = bucket.warned.IsZero() || now.Sub(bucket.warned) >= commandNoticeInterval
		if result.notify {
			bucket.warned = now
		}
	}
	l.buckets[key] = bucket
	return result
}

func (b *Bot) limitCommand(message *tgbotapi.Message) commandLimit {
	if b.limiter == nil || !usesDotaData(message) {
		return commandLimit{}
	}
	key := commandUserKey{kind: 'c', id: message.Chat.ID}
	if message.SenderChat != nil {
		key = commandUserKey{kind: 'c', id: message.SenderChat.ID}
	} else if message.From != nil && message.From.ID != 0 {
		key = commandUserKey{kind: 'u', id: message.From.ID}
	}
	return b.limiter.take(key)
}

func usesDotaData(message *tgbotapi.Message) bool {
	switch strings.ToLower(message.Command()) {
	case "player", "matches", "heroes", "stats", "match", "impact":
		return true
	case "meta":
		return strings.TrimSpace(message.CommandArguments()) != ""
	default:
		return false
	}
}

func commandLimitText(wait time.Duration) string {
	seconds := int((wait + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("⏳ Слишком частые запросы. Повтори через %d сек.", seconds)
}
