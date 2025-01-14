package stringutils

import (
	"testing"
)

func TestRemoveANSICodes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello \x1b[31mWorld\x1b[0m", "Hello World"},
		{"\x1b[1mBold\x1b[0m", "Bold"},
		{"No ANSI codes here", "No ANSI codes here"},
		{"\x1b[32mGreen\x1b[0m and \x1b[31mRed\x1b[0m", "Green and Red"},
		{"\x1b[0m\x1b[1mMultiple\x1b[0m sequences", "Multiple sequences"},
	}

	for _, test := range tests {
		result := RemoveANSICodes(test.input)
		if result != test.expected {
			t.Errorf("RemoveANSICodes(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

var tests = map[string]string{
	"nocodes": "No ANSI codes here",
	"short":   "Hello \x1b[31mWorld\x1b[0m",
	"medium":  "Hello \x1b[31mWorld\x1b[0m and \x1b[32mUniverse\x1b[0m with \x1b[1mBold\x1b[0m text",
	"long":    "Hello \x1b[31mWorld\x1b[0m and \x1b[32mUniverse\x1b[0m with \x1b[1mBold\x1b[0m text and \x1b[4mUnderline\x1b[0m text and \x1b[7mInverted\x1b[0m text and \x1b[9mStrikethrough\x1b[0m text and \x1b[3mItalic\x1b[0m text and \x1b[5mBlink\x1b[0m text",
}

func BenchmarkRemoveANSICodesNoCodes(b *testing.B) {
	s := tests["nocodes"]
	for i := 0; i < b.N; i++ {
		RemoveANSICodes(s)
	}
}

func BenchmarkRemoveANSICodesLengthShort(b *testing.B) {
	s := tests["short"]
	for i := 0; i < b.N; i++ {
		RemoveANSICodes(s)
	}
}

func BenchmarkRemoveANSICodesLengthMedium(b *testing.B) {
	s := tests["medium"]
	for i := 0; i < b.N; i++ {
		RemoveANSICodes(s)
	}
}

func BenchmarkRemoveANSICodesLengthLong(b *testing.B) {
	s := tests["long"]
	for i := 0; i < b.N; i++ {
		RemoveANSICodes(s)
	}
}
