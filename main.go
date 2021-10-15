package main

import (
	"github.com/mymmrac/telego"
	"github.com/phuslu/log"
	"os"
)

const DbPath = "./gorobei_db"

func NewGorobeiPoster(cli *CLI) (*Gorobei, error) {
	db, err := OpenDb(DbPath)
	if err != nil {
		log.Error().Err(err).Msg("cannot open db")
		return nil, err
	}
	tel, err := telego.NewBot(cli.Token)
	if err != nil {
		db.Close()
		return nil, err
	}

	g := &Gorobei{d: db, bot: tel, chat: cli.Chat}
	err = g.Init()
	if err != nil {
		db.Close()
		return nil, err
	}
	return g, nil
}

func main() {
	cli := cliParse()
	initLog(cli.Verbose)

	log.Info().Str("cmd", cli.command).Str("url", cli.Fetch.Url).Msg("command details")
	g, err := NewGorobeiPoster(cli)
	if err != nil {
		log.Error().Err(err).Msg("initialization error")
		os.Exit(1)
	}

	log.Info().Str("token", ShortenString(cli.Token, 4, 4)).Str("chat", cli.Chat).
		Int64("chat_id", g.chatID).
		Msg("telegram bot info")

	switch cli.command {
	case CmdFetch:
		err = g.Fetch(cli.Fetch.Url)
		if err != nil {
			log.Error().Err(err).Msg("cannot retrieve page content")
			os.Exit(1)
		}
	case CmdTestMsg:
		//err = g.TestMessage(cli.Chat, cli.TestMsg.Message)
		err = g.TestImage()
		if err != nil {
			log.Error().Err(err).Msg("cannot send test message")
			os.Exit(1)
		}
	default:
		log.Error().Msg("invalid command")
		os.Exit(1)
	}
}
