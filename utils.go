package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf8"
)

// https://stackoverflow.com/questions/41602230/return-first-n-chars-of-a-string
func FirstN(s string, n int) string {
	i := 0
	//iterate over runes
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

func LastN(s string, len int, n int) string {

	skipFirstN := len - n
	i := 0
	//iterate over runes
	for j := range s {
		if i == skipFirstN {
			return s[j:]
		}
		i++
	}
	return ""
}

func ShortenString(s string, first int, last int) string {
	ln := utf8.RuneCountInString(s)
	if ln <= first+last+3 {
		//nothing to shor...ten
		return s
	}
	l := LastN(s, ln, last)
	if l == "" {
		return FirstN(s, first)
	} else {
		return fmt.Sprintf("%s...%s", FirstN(s, first), l)
	}
}

func Int64ToByteArr(v int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	return b

}

func ByteArrToInt64(v []byte) (int64, error) {
	if len(v) != 8 {
		return 0, errors.New("input array has wrong length")
	}
	return int64(binary.LittleEndian.Uint64(v)), nil
}
