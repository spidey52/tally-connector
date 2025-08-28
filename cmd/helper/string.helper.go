package helper

import (
	"regexp"
	"strings"
)

// Remove leading and trailing whitespace
// Remove Byte Order Mark if present
// input = strings.TrimPrefix(input, "\uFEFF")

// // Ensure valid UTF-8
// input = strings.ToValidUTF8(input, "")

// // Remove illegal XML 1.0 characters: ASCII control chars except \t, \n, \r
// input = regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F]`).ReplaceAllString(input, "")

// return input

var illegalEntity = regexp.MustCompile(`&#[0-8];|&#1[0-9];|&#2[0-9];|&#3[0-1];`)

func CleanString(input string) string {
	// Remove leading and trailing whitespace
	// Remove Byte Order Mark if present
	input = strings.TrimPrefix(input, "\uFEFF")

	// Ensure valid UTF-8
	input = strings.ToValidUTF8(input, "")

	// Remove illegal XML 1.0 characters: ASCII control chars except \t, \n, \r
	input = regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F]`).ReplaceAllString(input, "")

	// Remove illegal XML entities
	input = illegalEntity.ReplaceAllString(input, "")

	return input
}
