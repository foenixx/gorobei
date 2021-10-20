package main

import (
	"github.com/phuslu/log"
	"gorobei/utils"
	"os"
)

const DbPath = "./gorobei_db"

func NewGorobeiPoster(cli *CLI) (*Gorobei, error) {
	db, err := OpenDb(DbPath)
	if err != nil {
		log.Error().Err(err).Msg("cannot open db")
		return nil, err
	}
	tg, err := NewTelegram(cli.Token, db)
	if err != nil {
		err = db.Close()
		return nil, err
	}

	g := &Gorobei{d: db,
		tg: tg,
		chat:         cli.Chat,
		admin:        cli.Admin}

	err = g.Init()
	if err != nil {
		err = db.Close()
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

	log.Info().Str("token", utils.ShortenString(cli.Token, 4, 4)).Str("chat", cli.Chat).
		Int64("chat_id", g.chatId).
		Int64("admin_id", g.adminId).
		Msg("telegram Bot info")

	switch cli.command {
	case CmdFetch:
		err = g.Fetch(cli.Fetch.Url, cli.Fetch.Limit)
		if err != nil {
			log.Error().Err(err).Msg("cannot retrieve page content")
			os.Exit(1)
		}
	case CmdSendMsg:
		log.Info().Str("username", cli.SendMsg.Username).Str("message", cli.SendMsg.Message).Msg("send message command params")
		err = g.tg.SendMessage(cli.SendMsg.Username, 0, cli.SendMsg.Message, "")
		if err != nil {
			log.Error().Err(err).Msg("cannot send message")
			os.Exit(1)
		}
	case CmdSendChatMsg:
		log.Info().Str("message", cli.SendChatMsg.Message).Msg("send chat message command params")
		err = g.SendChatMessage(cli.SendChatMsg.Message)
		if err != nil {
			log.Error().Err(err).Msg("cannot send default chat message")
			os.Exit(1)
		}
	case CmdSendAdminMsg:
		log.Info().Str("message", cli.SendAdminMsg.Message).Msg("send admin message params")
		if g.admin == "" {
			log.Error().Msg("no admin specified")
		}
		err = g.SendAdminMessage(cli.SendAdminMsg.Message)
		if err != nil {
			log.Error().Err(err).Msg("cannot send admin message")
			os.Exit(1)
		}
	case CmdSendImg:
		log.Info().Str("user", cli.SendImg.Username).Str("caption", cli.SendImg.Caption).Str("path", cli.SendImg.ImagePath).Msg("send image params")
		err = g.tg.SendImage(cli.SendImg.Username, 0, cli.SendImg.ImagePath, cli.SendImg.Caption)
		if err != nil {
			log.Error().Err(err).Msg("cannot send image")
			os.Exit(1)
		}
	case CmdSendChatImg:
		log.Info().Str("chat", cli.Chat).Str("caption", cli.SendChatImg.Caption).Str("path", cli.SendChatImg.ImagePath).Msg("send chat image params")
		err = g.SendChatImage(cli.SendChatImg.ImagePath, cli.SendChatImg.Caption)
		if err != nil {
			log.Error().Err(err).Msg("cannot send image to default chat")
			os.Exit(1)
		}
	case CmdForgetImg:
		log.Info().Str("url", cli.ForgetImg.Url).Msg("forget image params")
		err = g.ForgetImg(cli.ForgetImg.Url)
		if err != nil {
			log.Error().Err(err).Msg("cannot forget image")
			os.Exit(1)
		}

	default:
		log.Error().Msg("invalid command")
		os.Exit(1)
	}
}
