package main

import (
jsoniter "github.com/json-iterator/go"
"github.com/mymmrac/telego"
"github.com/phuslu/log"

"os"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func NewTelegramBot(token string) (*telego.Bot, error) {
	bot, err := telego.NewBot(token)


	if err != nil {
		return nil, err
	}
	// Call method getMe (https://core.telegram.org/bots/api#getme)
	user, err := bot.GetMe()
	if err != nil {
		return nil, err
	}
	// Print Bot information
	log.Info().Msgf("Bot user: %#v\n", user)
	return bot, nil
}


// Helper function to open file or panic
func mustOpen(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return file
}