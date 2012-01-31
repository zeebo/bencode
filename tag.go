package bencode

import (
	"strings"
	"unicode"
)

func isValidTag(key string) bool {
	if key == "" {
		return false
	}
	for _, c := range key {
		if c != ' ' && c != '$' && c != '-' && c != '_' && !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func matchName(key string) func(string) bool {
	return func(s string) bool {
		return strings.ToLower(key) == strings.ToLower(s)
	}
}
