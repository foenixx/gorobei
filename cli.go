package main

import (
	"github.com/alecthomas/kong"
	"strings"
)

const (
	CmdFetch   = "fetch"
	CmdSendMsg = "send-msg"
	CmdSendChatMsg = "send-chat-msg"
	CmdSendImg = "send-img"
	CmdSendChatImg = "send-chat-img"
	CmdSendAdminMsg = "send-admin-msg"
	CmdForgetImg = "forget-img"
)

type CLI struct {
	Admin string `help:"Admin user. The bot notifies admin about errors if this option is set."`
	Chat string `help:"Chat name." default:"@gorobei_posts"`
	Token string `help:"Telegram bot token." env:"GOROBEI_BOT_TOKEN"`
	Verbose bool `help:"Print verbose logs."`
	Fetch   struct {
		Limit int `help:"Stop after processing of '--limit' number of items. Default is 0 which means process all images."`
		Url string `arg:"" help:"Url to fetch data from."`
	} `cmd:"" help:"Parse the specified page content and fetch images."`
	SendMsg struct {
		Username string `help:"User to whom message is sent." required:""`
		Message string `help:"Message text." default:"test message"`
	} `cmd:"" help:"Send message to specified user."`
	SendChatMsg struct {
		Message string `help:"Message text." default:"test message"`
	} `cmd:"" help:"Send message to to default chat."`
	SendImg struct {
		Username string `help:"User to whom message is sent." required:""`
		ImagePath string `arg:"" required:"" help:"Path to image."`
		Caption string `help:"Image caption."`
	} `cmd:"" help:"Send image to specified user if --username is set and to default chat otherwise."`
	SendChatImg struct {
		ImagePath string `arg:"" required:"" help:"Path to image."`
		Caption string `help:"Image caption."`
	} `cmd:"" help:"Send image to default chat."`
	SendAdminMsg struct {
		Message string `help:"Message text."`
	} `cmd:"" help:"Send message to admin. '--admin' argument should be specified."`
	ForgetImg struct {
		Url string `arg:"" required:"" help:"Image url."`
	} `cmd:"" help:"Remove image url from internal db. Next time image's gonna be processed as new one.'"`
	command string `-`
}

func cliParse() *CLI {
	var cli CLI
	k := kong.Parse(&cli,
		kong.Name("gorobei"),
		kong.Description(""),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	switch {
	case strings.HasPrefix(k.Command(), CmdFetch):
		cli.command = CmdFetch
	case strings.HasPrefix(k.Command(), CmdSendMsg):
		cli.command = CmdSendMsg
	case strings.HasPrefix(k.Command(), CmdSendChatMsg):
		cli.command = CmdSendChatMsg
	case strings.HasPrefix(k.Command(), CmdSendImg):
		cli.command = CmdSendImg
	case strings.HasPrefix(k.Command(), CmdSendChatImg):
		cli.command = CmdSendChatImg
	case strings.HasPrefix(k.Command(), CmdSendAdminMsg):
		cli.command = CmdSendAdminMsg
		cli.SendAdminMsg.Message = "message line1\nline2\nline3\n[image](https://i.imgur.com/sMhpFyR.jpg)\n\n__error:__\n```\nerror line1\nline2```"
	case strings.HasPrefix(k.Command(), CmdForgetImg):
		cli.command = CmdForgetImg
	default:
		cli.command = "not specified"
	}
	return &cli

}
