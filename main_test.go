package main

import (
	"testing"
)

// FuzzApplyMapReplacements uses fuzzing to test map replacement reversibility with random inputs
func FuzzApplyMapReplacements(f *testing.F) {
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

// FuzzApplyNumbersLogic uses fuzzing to test number base conversion reversibility
func FuzzApplyNumbersLogic(f *testing.F) {
	// Minimal seed corpus - fuzzer will generate the rest
	seeds := []string{
		"0",
		"7",
		"-1",
		"text 42 more",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		pejelagarto := applyNumbersLogicToPejelagarto(input)
		reversed := applyNumbersLogicFromPejelagarto(pejelagarto)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:       %q\nPejelagarto: %q\nReversed:    %q", input, pejelagarto, reversed)
		}
	})
}

// FuzzApplyAccentReplacementLogic uses fuzzing to test accent replacement reversibility
func FuzzApplyAccentReplacementLogic(f *testing.F) {
	// Seed corpus with vowel-containing examples
	seeds := []string{
		"aeiou",
		"AEIOU",
		"hello",
		"world",
		"",
		"a",
		"test",
		"múltiple áccents",
		"MiXeD CaSe",
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
	})
}

// FuzzTranslatePejelagarto uses fuzzing to test full translation pipeline reversibility
func FuzzTranslatePejelagarto(f *testing.F) {
	// Seed corpus with diverse examples
	seeds := []string{
		"hello world",
		"The quick brown fox jumps over the lazy dog",
		"test 123 numbers",
		"aeiou vowels",
		"",
		"a",
		"CHAPTER",
		"mixed 42 content",
		"the fish and chips",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test: TranslateToPejelagarto -> TranslateFromPejelagarto
		pejelagarto := TranslateToPejelagarto(input)
		reversed := TranslateFromPejelagarto(pejelagarto)

		if reversed != input {
			t.Errorf("TranslateToPejelagarto->TranslateFromPejelagarto failed\nInput:       %q\nPejelagarto: %q\nReversed:    %q", input, pejelagarto, reversed)
		}
	})
}

// TestNumberConversionBasic is removed - use FuzzApplyNumbersLogic instead

// TestNumberConversionReversibility is removed - use FuzzApplyNumbersLogic instead

// TestConvertToBase7 and TestConvertFromBase7 are removed - math/big handles conversions internally
