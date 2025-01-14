package stringutils

import (
	"regexp"
	"strings"
)

var (
	ansiEscapePattern = `\x1b\[[0-9;]*[A-Za-z]`
	ansiRegex         = regexp.MustCompile(ansiEscapePattern)
)

func RemoveANSICodes(input string) string {
	if !strings.Contains(input, "\x1b") {
		return input
	}
	cleanedString := ansiRegex.ReplaceAllString(input, "")
	return cleanedString
}
