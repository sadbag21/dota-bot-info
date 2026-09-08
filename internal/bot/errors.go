package bot

import (
	"errors"

	"github.com/sadbag21/dota-bot-info/internal/dota"
)

func userErrorMessage(err error) string {
	switch {
	case errors.Is(
		err,
		dota.ErrNotFound,
	):
		return "🔎 Ничего не найдено.\n\nПроверь введённый ID."

	case errors.Is(
		err,
		dota.ErrRateLimited,
	):
		return "⏳ OpenDota временно ограничил количество запросов.\n\nПопробуй немного позже."

	case errors.Is(
		err,
		dota.ErrTimeout,
	):
		return "⌛ OpenDota слишком долго отвечает.\n\nПопробуй ещё раз."

	case errors.Is(
		err,
		dota.ErrUnavailable,
	):
		return "🌐 OpenDota сейчас недоступен.\n\nПопробуй позже."

	case errors.Is(
		err,
		dota.ErrBadResponse,
	):
		return "⚠️ OpenDota вернул некорректный ответ.\n\nПопробуй позже."

	default:
		return "❌ Произошла неизвестная ошибка."
	}
}
