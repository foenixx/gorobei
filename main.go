package main

import (
	"fmt"
	"github.com/phuslu/log"
	"gorobei/clock"
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
		fetcher: &httpFetcherImpl{},
		chat:         cli.Chat,
		admin:        cli.Admin,
		clock: &clock.RealClock{}}

	err = g.Init()
	if err != nil {
		err = db.Close()
		return nil, err
	}
	return g, nil
}

func must(err error, msg string) {
	if err == nil {
		return
	}
	log.Error().Err(err).Msg(msg)
	os.Exit(1)
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
			// try to notify an admin
			_ = g.SendAdminMessage(fmt.Sprintf("Cannot get page content!\n[link](%s)\n\n__error__:\n```\n%s\n```", cli.Fetch.Url, err.Error()))
			os.Exit(1)
		}
	case CmdSendMsg:
		log.Info().Str("username", cli.SendMsg.Username).Str("message", cli.SendMsg.Message).Msg("send message command params")
		err = g.tg.SendMessageText(cli.SendMsg.Username, 0, cli.SendMsg.Message)
		must(err, "cannot send message")
	case CmdSendChatMsg:
		log.Info().Str("message", cli.SendChatMsg.Message).Msg("send chat message command params")
		err = g.SendChatMessage(cli.SendChatMsg.Message)
		must(err, "cannot send default chat message")
	case CmdSendAdminMsg:
		log.Info().Str("message", cli.SendAdminMsg.Message).Msg("send admin message params")
		if g.admin == "" {
			log.Error().Msg("no admin specified")
		}
		err = g.SendAdminMessage(cli.SendAdminMsg.Message)
		must(err, "cannot send admin message")
	case CmdSendImg:
		log.Info().Str("user", cli.SendImg.Username).Str("caption", cli.SendImg.Caption).Str("path", cli.SendImg.ImagePath).Msg("send image params")
		err = g.tg.SendImage(cli.SendImg.Username, 0, cli.SendImg.ImagePath, cli.SendImg.Caption)
		must(err, "cannot send image")
	case CmdSendChatImg:
		log.Info().Str("chat", cli.Chat).Str("caption", cli.SendChatImg.Caption).Str("path", cli.SendChatImg.ImagePath).Msg("send chat image params")
		err = g.SendChatImage(cli.SendChatImg.ImagePath, cli.SendChatImg.Caption)
		must(err, "cannot send image to default chat")
	case CmdForgetImg:
		log.Info().Str("url", cli.ForgetImg.Url).Msg("forget image params")
		err = g.ForgetImg(cli.ForgetImg.Url)
		must(err, "cannot forget image")
	case CmdReport:
		log.Info().Msg("send report to admin")
		var r *DailyReport
		r, err = g.ReadOrCreateDailyReport()
		must(err, "cannot read daily report")
		if g.admin == "" {
			log.Error().Msg("no admin specified")
		}
		err = g.SendAdminMessage(g.FormatDailyReport(r))
		must(err, "cannot read daily report")
	default:
		log.Error().Msg("invalid command")
		os.Exit(1)
	}
}

