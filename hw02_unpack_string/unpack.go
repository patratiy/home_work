// Package hw02unpackstring implements primitive string unpacking.
package hw02unpackstring

import (
	"errors"
	"strings"
)

// ErrInvalidString is returned when the input string has errors.
var ErrInvalidString = errors.New("invalid string")

// isDigit reports whether the rune is an ASCII digit.
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// Unpack performs primitive string unpacking.
func Unpack(input string) (string, error) {
	runes := []rune(input)
	var builder strings.Builder

	for i := 0; i < len(runes); {
		r := runes[i]

		// Determine the literal rune to unpack.
		var char rune
		switch {
		case r == '\\':
			// Escape sequence: only a digit or a backslash may be escaped.
			if i+1 >= len(runes) {
				return "", ErrInvalidString
			}
			next := runes[i+1]
			if next != '\\' && !isDigit(next) {
				return "", ErrInvalidString
			}
			char = next
			i += 2
		case isDigit(r):
			// A digit cannot start a sequence.
			return "", ErrInvalidString
		default:
			char = r
			i++
		}

		// Determine the repetition count: 1 by default.
		count := 1
		if i < len(runes) && isDigit(runes[i]) {
			// Numbers are not allowed: a digit must not be followed by another digit.
			if i+1 < len(runes) && isDigit(runes[i+1]) {
				return "", ErrInvalidString
			}
			count = int(runes[i] - '0')
			i++
		}

		if count > 0 {
			builder.WriteString(strings.Repeat(string(char), count))
		}
	}

	return builder.String(), nil
}
