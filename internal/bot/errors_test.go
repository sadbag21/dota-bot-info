package bot

import (
	"fmt"
	"testing"

	"github.com/sadbag21/dota-bot-info/internal/dota"
)

func TestUserErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "not found",
			err:      dota.ErrNotFound,
			expected: "🔎 Ничего не найдено.\n\nПроверь введённый ID.",
		},
		{
			name:     "rate limited",
			err:      dota.ErrRateLimited,
			expected: "⏳ OpenDota временно ограничил количество запросов.\n\nПопробуй немного позже.",
		},
		{
			name:     "timeout",
			err:      dota.ErrTimeout,
			expected: "⌛ OpenDota слишком долго отвечает.\n\nПопробуй ещё раз.",
		},
		{
			name:     "unavailable",
			err:      dota.ErrUnavailable,
			expected: "🌐 OpenDota сейчас недоступен.\n\nПопробуй позже.",
		},
		{
			name:     "bad response",
			err:      dota.ErrBadResponse,
			expected: "⚠️ OpenDota вернул некорректный ответ.\n\nПопробуй позже.",
		},
		{
			name: "wrapped timeout",
			err: fmt.Errorf(
				"service failed: %w",
				dota.ErrTimeout,
			),
			expected: "⌛ OpenDota слишком долго отвечает.\n\nПопробуй ещё раз.",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				result := userErrorMessage(
					tt.err,
				)

				if result != tt.expected {
					t.Errorf(
						"expected %q, got %q",
						tt.expected,
						result,
					)
				}
			},
		)
	}
}
