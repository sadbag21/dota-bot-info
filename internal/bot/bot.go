package bot

import (
	"log"

	"github.com/sadbag21/dota-bot-info/internal/config"
	"github.com/sadbag21/dota-bot-info/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	service *service.PlayerService
}

func New(
	cfg config.Config,
	playerService *service.PlayerService,
) *Bot {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"Authorized on account %s",
		api.Self.UserName,
	)

	return &Bot{
		api:     api,
		service: playerService,
	}
}

func (b *Bot) Run() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := b.api.GetUpdatesChan(updateConfig)

	for update := range updates {
		b.handleUpdate(update)
	}
}
