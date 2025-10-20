package main

import (
	"testing"
	"unicode"
)

func TestDebugAccent(t *testing.T) {
	input := "aeiou" // Length 5, prime factors: {5:1}
	t.Logf("Input: %q, len=%d", input, len([]rune(input)))

	factors := primeFactorize(len([]rune(input)))
	t.Logf("Factors: %v", factors)

	// Should change 5th vowel (u) by power 1
	accented := applyAccentReplacementLogicToPejelagarto(input)
	t.Logf("Accented: %q", accented)

	// Check what we're finding
	accentedRunes := []rune(accented)
	for i, r := range accentedRunes {
		t.Logf("Position %d: rune=%c (U+%04X), isVowel=%v", i, r, r, isVowel(r))
		if isVowel(r) {
			vowelStr := string(unicode.ToLower(r))
			base := getBaseVowel(vowelStr)
			idx := findAccentIndex(base, vowelStr)
			t.Logf("  -> Vowel: vowelStr=%q, accentIdx=%d, baseVowel=%c", vowelStr, idx, base)
		}
	}

	reversed := applyAccentReplacementLogicFromPejelagarto(accented)
	t.Logf("Reversed: %q", reversed)

	if reversed != input {
		t.Errorf("Reversibility failed")
	}
}

func TestFindAccentIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"a", 0},
		{"à", 1},
		{"á", 2},
		{"â", 3},
		{"ã", 4},
		{"ù", 1},
	}

	for _, tt := range tests {
		baseVowel := getBaseVowel(tt.input)
		result := findAccentIndex(baseVowel, tt.input)
		t.Logf("findAccentIndex(%c, %q) = %d", baseVowel, tt.input, result)
		if result != tt.expected {
			t.Errorf("findAccentIndex(%c, %q) = %d, want %d", baseVowel, tt.input, result, tt.expected)
		}
	}
}
