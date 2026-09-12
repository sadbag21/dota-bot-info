package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sadbag21/dota-bot-info/internal/meta"
	"github.com/sadbag21/dota-bot-info/pkg/formatter"
)

type metaProvider interface {
	Get(context.Context, int) (meta.Report, error)
}

func parseMetaPosition(argument string) (int, error) {
	value := strings.TrimSpace(argument)
	if len(value) != 1 || value[0] < '1' || value[0] > '5' {
		return 0, errInvalidID
	}
	return int(value[0] - '0'), nil
}

func metaKeyboard() tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, positions := range [][]int{{1, 2, 3}, {4, 5}} {
		row := []tgbotapi.InlineKeyboardButton{}
		for _, position := range positions {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d — %s", position, meta.PositionName(position)), fmt.Sprintf("meta:%d", position)))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (b *Bot) handleMeta(message *tgbotapi.Message) {
	args := strings.TrimSpace(message.CommandArguments())
	if args == "" {
		b.sendMessage(message.Chat.ID, "🧭 <b>Мета героев</b>\nВыбери позицию. Покажу топ героев Immortal по статистике STRATZ за последние 7 полных дней в пределах актуального патча.", metaKeyboard())
		return
	}
	position, err := parseMetaPosition(args)
	if err != nil {
		b.sendMessage(message.Chat.ID, "Укажи позицию от 1 до 5: /meta 1 — или выбери кнопку.", metaKeyboard())
		return
	}
	ctx := b.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	var report meta.Report
	if b.meta == nil {
		err = meta.ErrDisabled
	} else {
		report, err = b.meta.Get(ctx, position)
	}
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		text := "⚠️ Сейчас не удалось получить свежую мету. Попробуй немного позже."
		switch {
		case errors.Is(err, meta.ErrDisabled):
			text = "⚙️ Мета ещё не настроена: владельцу бота нужно добавить STRATZ_API_TOKEN."
		case errors.Is(err, meta.ErrUnauthorized):
			text = "⚠️ STRATZ отклонил доступ. Владельцу бота нужно проверить API-токен и его права."
		case errors.Is(err, meta.ErrRateLimited):
			text = "⏳ Достигнут лимит запросов STRATZ. Попробуй позже."
		}
		slog.Warn("Meta request failed", "position", position)
		b.sendMessage(message.Chat.ID, text, metaKeyboard())
		return
	}
	b.sendMessage(message.Chat.ID, formatter.FormatMeta(report), metaKeyboard())
}
