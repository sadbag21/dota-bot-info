package main

import (
	"github.com/sadbag21/dota-bot-info/internal/bot"
	"github.com/sadbag21/dota-bot-info/internal/config"
	"github.com/sadbag21/dota-bot-info/internal/dota"
	"github.com/sadbag21/dota-bot-info/internal/service"
)

func main() {
	cfg := config.Load()

	dotaClient := dota.NewClient()
	playerService := service.NewPlayerService(dotaClient)

	telegramBot := bot.New(
		cfg,
		playerService,
	)

	telegramBot.Run()
}
