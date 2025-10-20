package main

import (
	"testing"
)

// FuzzAccentReversibility tests the reversibility of accent replacement logic
func FuzzAccentReversibility(f *testing.F) {
	// Seed corpus with examples containing vowels
	seeds := []string{
		"hello world",
		"aeiou",
		"AEIOU",
		"The quick brown fox",
		"y is a vowel",
		"",
		"a",
		"test with accents: àéîõü",
		"Mixed Case Vowels",
		"rhythm", // has y as vowel
		"12345 numbers and vowels aeiou",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		accented := applyAccentReplacementLogicToPejelagarto(input)
		reversed := applyAccentReplacementLogicFromPejelagarto(accented)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:    %q\nAccented: %q\nReversed: %q", input, accented, reversed)
		}

		// Test: FromPejelagarto -> ToPejelagarto
		deaccented := applyAccentReplacementLogicFromPejelagarto(accented)
		reaccented := applyAccentReplacementLogicToPejelagarto(deaccented)

		if reaccented != accented {
			t.Errorf("FromPejelagarto->ToPejelagarto failed\nInput:      %q\nDeaccented: %q\nReaccented: %q", accented, deaccented, reaccented)
		}
	})
}

// TestAccentBasic tests basic accent replacement functionality
func TestAccentBasic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "Empty string", input: ""},
		{name: "Single vowel", input: "a"},
		{name: "No vowels", input: "bcdfg"},
		{name: "Multiple vowels", input: "aeiou"},
		{name: "With uppercase", input: "AEI"},
		{name: "Mixed text", input: "hello world"},
		{name: "Y as vowel", input: "yes"},
		{name: "Prime length 2", input: "ab"},
		{name: "Prime length 3", input: "abc"},
		{name: "Prime length 5", input: "abcde"},
		{name: "Prime length 7", input: "abcdefg"},
		{name: "Composite 6 (2*3)", input: "abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accented := applyAccentReplacementLogicToPejelagarto(tt.input)
			reversed := applyAccentReplacementLogicFromPejelagarto(accented)

			t.Logf("Input:    %q", tt.input)
			t.Logf("Accented: %q", accented)
			t.Logf("Reversed: %q", reversed)

			if reversed != tt.input {
				t.Errorf("Reversibility failed: input=%q, reversed=%q", tt.input, reversed)
			}
		})
	}
}

// TestPrimeFactorization tests the prime factorization helper
func TestPrimeFactorization(t *testing.T) {
	tests := []struct {
		n        int
		expected map[int]int
	}{
		{1, map[int]int{}},
		{2, map[int]int{2: 1}},
		{3, map[int]int{3: 1}},
		{4, map[int]int{2: 2}},
		{6, map[int]int{2: 1, 3: 1}},
		{8, map[int]int{2: 3}},
		{12, map[int]int{2: 2, 3: 1}},
		{245, map[int]int{5: 1, 7: 2}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := primeFactorize(tt.n)
			if len(result) != len(tt.expected) {
				t.Errorf("primeFactorize(%d) = %v, want %v", tt.n, result, tt.expected)
				return
			}
			for prime, power := range tt.expected {
				if result[prime] != power {
					t.Errorf("primeFactorize(%d)[%d] = %d, want %d", tt.n, prime, result[prime], power)
				}
			}
		})
	}
}
