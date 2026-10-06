package hw02unpackstring

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "a4bc2d5e", expected: "aaaabccddddde"},
		{input: "abccd", expected: "abccd"},
		{input: "", expected: ""},
		{input: "aaa0b", expected: "aab"},
		{input: "🙃0", expected: ""},
		{input: "aaф0b", expected: "aab"},
		{input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
		{input: "🙃3", expected: "🙃🙃🙃"},
		{input: "при5вет", expected: "прииииивет"},
		{input: `qwe\4\5`, expected: `qwe45`},
		{input: `qwe\45`, expected: `qwe44444`},
		{input: `qwe\\5`, expected: `qwe\\\\\`},
		{input: `qwe\\\3`, expected: `qwe\3`},
		{input: `qwe\5\4`, expected: `qwe54`},
		{input: `\5`, expected: `5`},
		{input: "a1b2", expected: "abb"},
		{input: "a9", expected: "aaaaaaaaa"},
		{input: "a 2b", expected: "a  b"},
		{input: "a2\n2b", expected: "aa\n\nb"},
		{input: `\5\5`, expected: "55"},
		{input: `a\34`, expected: "a3333"},
		{input: `qwe\40`, expected: "qwe"},
		{input: `\\`, expected: `\`},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			result, err := Unpack(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalidStrings := []string{
		"3abc",
		"45",
		"aaa10b",
		"a45",
		"10a",
		`qw\ne`,
		`qw\q`,
		`\`,
		`qwe\`,
		`qwe\456`,
		`qwe\\45`,
		"a03",
		"1",
		"00",
		`qw\🙃`,
	}
	for _, tc := range invalidStrings {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			_, err := Unpack(tc)
			require.Truef(t, errors.Is(err, ErrInvalidString), "actual error %q", err)
		})
	}
}
