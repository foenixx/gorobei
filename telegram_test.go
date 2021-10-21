package main

import (
	"github.com/stretchr/testify/require"
	"gorobei/utils"
	"testing"
)

//var re1 = regexp.MustCompile(`[_*[\]()~\x60>#+\-=]`)
func Test_replaceAllChars(t *testing.T) {
	input := utils.Bt(`abcd1 _*~ []()^>#+-=\`)
	exp := utils.Bt(`abcd1 _*~ \[\]\(\)\^\>\#\+\-\=\\`)
	require.Equal(t, exp, EscapeMarkdown(input))
	//t.Log(re1.ReplaceAllString(input, `\$0`))
}

func TestEscapeMarkdown(t *testing.T) {
	//initLog(false)
	input := utils.Bt(`
~markdown~ *inline* _markup_ **test** __line__
line with [link](https://some.link/here?param1=1&param2=2)
strange [link with slashes](https://lin\k.ru)
link as text: https://some.link/here?param1=1&param2=2
line with ³back\ticks and [special]=(symbols)³
[special]=(symbols) without \ backticks
another line
³³³
	code ³line³ 1
	code\line\2
³³³`)
	exp := utils.Bt(`
~markdown~ *inline* _markup_ **test** __line__
line with [link](https://some.link/here?param1=1&param2=2)
strange [link with slashes](https://lin\\k.ru)
link as text: https://some\.link/here?param1\=1&param2\=2
line with ³back\\ticks and [special]=(symbols)³
\[special\]\=\(symbols\) without \\ backticks
another line
³³³
	code \³line\³ 1
	code\\line\\2
³³³`)
	require.Equal(t, exp, EscapeMarkdown(input))
}

