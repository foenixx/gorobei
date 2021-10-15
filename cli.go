package main

import (
	"github.com/alecthomas/kong"
	"strings"
)

const (
	CmdFetch = "fetch"
	CmdTestMsg = "test-msg"
)

type CLI struct {
	Chat string `help:"Chat name." default:"@gorobei_posts"`
	Token string `help:"Telegram bot token." env:"GOROBEI_BOT_TOKEN"`
	Verbose bool `help:"Print verbose logs."`
	Fetch   struct {
		Url string `arg:"" help:"Url to fetch data from."`
	} `cmd:"" help:"Parse the specified page content and fetch images."`
	TestMsg struct {
		Message string `help:"Message text." default:"test message"`
	} `cmd:"" help:"Send test message."`
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
	case strings.HasPrefix(k.Command(), CmdTestMsg):
		cli.command = CmdTestMsg
	default:
		cli.command = "not specified"
	}
	return &cli

}
