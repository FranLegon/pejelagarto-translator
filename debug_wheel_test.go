package main

import (
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestWhichWheel(t *testing.T) {
	tests := []struct {
		char          string
		expectedWheel string // "one" or "two"
	}{
		{"à", "one"},
		{"Ǫ", "two"},
		{"ǫ", "two"},
		{"ò", "one"},
		{"Ò", "one"},
	}

	for _, tt := range tests {
		t.Run(tt.char, func(t *testing.T) {
			lowerStr := string(unicode.ToLower([]rune(tt.char)[0]))
			baseVowel := getBaseVowel(lowerStr)

			t.Logf("Char: %q, lowercase: %q, base: %c, runeCount: %d",
				tt.char, lowerStr, baseVowel, utf8.RuneCountInString(tt.char))

			// Check oneRuneAccentsWheel
			if wheel, ok := oneRuneAccentsWheel[baseVowel]; ok {
				for idx, form := range wheel {
					if form == lowerStr {
						t.Logf("  Found in oneRuneAccentsWheel[%c][%d] = %q", baseVowel, idx, form)
					}
				}
			}

			// Check twoRunesAccentsWheel
			if wheel, ok := twoRunesAccentsWheel[baseVowel]; ok {
				for idx, form := range wheel {
					if form == lowerStr {
						t.Logf("  Found in twoRunesAccentsWheel[%c][%d] = %q", baseVowel, idx, form)
					}
				}
			}
		})
	}
}
