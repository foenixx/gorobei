package main

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func bt(s string) string {
	const ph3 = "^^^"
	const ph1 = "^"
	return strings.ReplaceAll(strings.ReplaceAll(s, ph1, "`"), ph3, "```")
}

func Test_replaceOthers(t *testing.T) {
	input := bt(`_*[]()~^>#+-=|{}.!`)
	exp := bt(`\_\*\[\]\(\)\~\^\>\#\+\-\=\|\{\}\.\!`)
	require.Equal(t, exp, string(replOthers([]byte(input))))
}

func TestEscapeMarkdown(t *testing.T) {
	//initLog(false)
	input := bt(`
markdown \ test line1
line with ^back\ticks^
another line
^^^
  code ^line^ 1
  code\line\2
^^^`)
	exp := bt(`
markdown \ test line1
line with ^back\\ticks^
another line
^^^
  code \^line\^ 1
  code\\line\\2
^^^`)
	require.Equal(t, exp, EscapeMarkdown(input))
}
