package utils

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

var reTest = regexp.MustCompile(`def(.*?)jkl`)

func replMatched(i [][]byte) [][]byte {
	return i
}

func replaceNonMatched(s []byte) []byte {
	return []byte("*")
}

func TestReplaceAllSubmatchFunc2(t *testing.T) {
	in := "abc def ghi jkl mno def pqr jkl stu"
	exp := "*def ghi jkl*def pqr jkl*"
	out := ReplaceAllSubmatchFunc2(reTest, []byte(in), replMatched, replaceNonMatched, -1)
	require.Equal(t, exp, string(out))
}
