package main

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func TestUnicodeNormalization(t *testing.T) {
	// Check all characters in both wheels
	allChars := make(map[string]string)

	// Collect from oneRuneAccentsWheel
	for base, accents := range oneRuneAccentsWheel {
		for _, accent := range accents {
			allChars[accent] = fmt.Sprintf("oneRune[%c]", base)
		}
	}

	// Collect from twoRunesAccentsWheel
	for base, accents := range twoRunesAccentsWheel {
		for _, accent := range accents {
			if prev, exists := allChars[accent]; exists {
				t.Logf("DUPLICATE: %q in both %s and twoRune[%c]", accent, prev, base)
			}
			allChars[accent] = fmt.Sprintf("twoRune[%c]", base)
		}
	}

	t.Logf("\nChecking Unicode properties of all accent characters:")
	t.Logf("%-5s %-10s %-8s %-8s %-8s %-20s", "Char", "Wheel", "Runes", "NFD", "Bytes", "Codepoint")

	for char, wheel := range allChars {
		if char == "a" || char == "e" || char == "i" || char == "o" || char == "u" || char == "y" {
			continue // Skip base vowels
		}

		runeCount := utf8.RuneCountInString(char)
		nfd := norm.NFD.String(char)
		nfdRuneCount := utf8.RuneCountInString(nfd)
		byteCount := len(char)

		codepoint := ""
		for _, r := range char {
			codepoint += fmt.Sprintf("U+%04X ", r)
		}

		t.Logf("%-5s %-10s %-8d %-8d %-8d %-20s", char, wheel, runeCount, nfdRuneCount, byteCount, codepoint)
	}
}
