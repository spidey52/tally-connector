package helper

import (
	"regexp"
	"strings"
)

func CleanString(input string) string {
	// Remove leading and trailing whitespace
	// Remove Byte Order Mark if present
	input = strings.TrimPrefix(input, "\uFEFF")

	// Ensure valid UTF-8
	input = strings.ToValidUTF8(input, "")

	// Remove illegal XML 1.0 characters: ASCII control chars except \t, \n, \r
	input = regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F]`).ReplaceAllString(input, "")

	return input
}
