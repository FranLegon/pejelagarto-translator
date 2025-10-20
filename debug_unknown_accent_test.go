package main

import (
	"testing"
	"unicode"
)

func TestUnknownAccent(t *testing.T) {
	input := "éȮ"
	t.Logf("Input: %q", input)

	runes := []rune(input)
	for i, r := range runes {
		t.Logf("Position %d: %c (U+%04X)", i, r, r)
		if isVowel(r) {
			vowelStr := string(unicode.ToLower(r))
			base := getBaseVowel(vowelStr)
			idx := findAccentIndex(base, vowelStr)

			wheel, ok := oneRuneAccentsWheel[base]
			if !ok || idx >= len(wheel) {
				t.Logf("  isVowel=true, lowercase=%q, accentIdx=%d, baseVowel=%c, NO WHEEL OR OUT OF RANGE",
					vowelStr, idx, base)
				continue
			}
			expectedForm := wheel[idx]
			t.Logf("  isVowel=true, lowercase=%q, accentIdx=%d, baseVowel=%c, expectedForm=%q, match=%v",
				vowelStr, idx, base, expectedForm, expectedForm == vowelStr)
		}
	}

	accented := applyAccentReplacementLogicToPejelagarto(input)
	t.Logf("Accented: %q", accented)

	reversed := applyAccentReplacementLogicFromPejelagarto(accented)
	t.Logf("Reversed: %q", reversed)

	if reversed != input {
		t.Errorf("Failed: expected %q, got %q", input, reversed)
	}
}
