package main

import (
	"testing"
)

// FuzzReversibility uses fuzzing to test reversibility with random inputs
func FuzzReversibility(f *testing.F) {
	// Seed corpus with basic examples
	seeds := []string{
		"hello world",
		"The quick brown fox",
		"chapter",
		"fish and chips",
		"",
		"a",
		"HELLO",
		"HeLLo WoRLd",
		"the the the",
		"chchchch shshshsh ththth",
		"aeiou AEIOU",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		translated := applyMapReplacementsToPejelagarto(input)
		reversed := applyMapReplacementsFromPejelagarto(translated)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", input, translated, reversed)
		}

		// Test: FromPejelagarto -> ToPejelagarto
		translated2 := applyMapReplacementsFromPejelagarto(translated)
		reversed2 := applyMapReplacementsToPejelagarto(translated2)

		if reversed2 != translated {
			t.Errorf("FromPejelagarto->ToPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", translated, translated2, reversed2)
		}
	})
}
