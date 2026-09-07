package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/sadbag21/dota-bot-info/pkg/formatter"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	message := update.Message

	log.Printf(
		"Message from %s: %s",
		message.From.UserName,
		message.Text,
	)

	if !message.IsCommand() {
		return
	}

	switch strings.ToLower(message.Command()) {
	case "start":
		b.handleStart(message)

	case "help":
		b.handleHelp(message)

	case "player":
		b.handlePlayer(message)

	default:
		b.sendMessage(
			message.Chat.ID,
			"❌ Неизвестная команда. Используй /help",
		)
	}
}

func (b *Bot) handlePlayer(message *tgbotapi.Message) {
	// Получаем аргументы команды.
	args := message.CommandArguments()

	if args == "" {
		b.sendMessage(
			message.Chat.ID,
			"❌ Укажи Dota ID.\n\nПример:\n<code>/player 123456789</code>",
		)
		return
	}

	// Убираем пробелы.
	args = strings.TrimSpace(args)

	// Проверяем, что ID состоит из числа.
	accountID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		b.sendMessage(
			message.Chat.ID,
			"❌ Dota ID должен быть числом.\n\nПример:\n<code>/player 123456789</code>",
		)
		return
	}

	// Dota ID не может быть отрицательным или нулём.
	if accountID <= 0 {
		b.sendMessage(
			message.Chat.ID,
			"❌ Dota ID должен быть положительным числом.",
		)
		return
	}

	log.Printf(
		"Getting player information for account_id=%d",
		accountID,
	)

	// Запрашиваем игрока через Service.
	playerInfo, err := b.service.GetPlayerInfo(accountID)
	if err != nil {
		log.Printf(
			"Failed to get player %d: %v",
			accountID,
			err,
		)

		b.sendMessage(
			message.Chat.ID,
			fmt.Sprintf(
				"❌ Не удалось получить информацию об игроке.\n\nID: <code>%d</code>\n\nПопробуй ещё раз позже.",
				accountID,
			),
		)

		return
	}

	// Формируем красивый текст.
	text := formatter.FormatPlayer(playerInfo)

	// Отправляем его пользователю.
	b.sendMessage(message.Chat.ID, text)
}
