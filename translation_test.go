package main

import (
	"testing"
	"unicode"
	"unicode/utf8"
)

// FuzzApplyMapReplacements uses fuzzing to test map replacement reversibility with random inputs
func FuzzApplyMapReplacements(f *testing.F) {
	// Seed corpus with basic cases
	f.Add("")
	f.Add("hello")
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
	// Seed corpus with basic cases
	f.Add("")
	f.Add("123")
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
	// Seed corpus with basic cases
	f.Add("")
	f.Add("café")
	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		accented := applyAccentReplacementLogicToPejelagarto(input)
		reversed := applyAccentReplacementLogicFromPejelagarto(accented)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:    %q\nAccented: %q\nReversed: %q", input, accented, reversed)
		}
	})
}

// FuzzApplyPunctuationReplacements tests punctuation replacement reversibility
func FuzzApplyPunctuationReplacements(f *testing.F) {
	// Seed corpus with basic cases
	f.Add("")
	f.Add("Hello, world!")
	f.Fuzz(func(t *testing.T, input string) {
		if !utf8.ValidString(input) {
			t.Skip("invalid utf8")
		}

		// Test: ToPejelagarto -> FromPejelagarto
		translated := applyPunctuationReplacementsToPejelagarto(input)
		reversed := applyPunctuationReplacementsFromPejelagarto(translated)

		if reversed != input {
			t.Errorf("applyPunctuationReplacementsToPejelagarto->FromPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", input, translated, reversed)
		}

		// Test: FromPejelagarto -> ToPejelagarto
		translated2 := applyPunctuationReplacementsFromPejelagarto(translated)
		reversed2 := applyPunctuationReplacementsToPejelagarto(translated2)

		if reversed2 != translated {
			t.Errorf("applyPunctuationReplacementsFromPejelagarto->ToPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", translated, translated2, reversed2)
		}
	})
}

// FuzzApplyCaseReplacementLogic tests case replacement logic reversibility
func FuzzApplyCaseReplacementLogic(f *testing.F) {
	// Seed corpus with basic cases
	f.Add("")
	f.Add("Hello World")
	f.Fuzz(func(t *testing.T, input string) {
		if !utf8.ValidString(input) {
			t.Skip("invalid utf8")
		}

		// Apply case replacement twice - should return to original (self-reversing)
		once := applyCaseReplacementLogic(input)
		twice := applyCaseReplacementLogic(once)

		if input != twice {
			t.Errorf("applyCaseReplacementLogic not reversible:\nInput: %q\nOnce:  %q\nTwice: %q", input, once, twice)
		}

		// Word count should not change
		originalWords := countWordsInString(input)
		onceWords := countWordsInString(once)

		if originalWords != onceWords {
			t.Errorf("Word count changed: %d -> %d\nInput: %q\nOnce:  %q", originalWords, onceWords, input, once)
		}
	})
}

// countWordsInString counts words for testing (matches applyCaseReplacementLogic logic)
func countWordsInString(input string) int {
	runes := []rune(input)
	wordCount := 0
	inWord := false

	for _, r := range runes {
		isLetterOrDigit := unicode.IsLetter(r) || unicode.IsDigit(r)
		if isLetterOrDigit && !inWord {
			wordCount++
			inWord = true
		} else if !isLetterOrDigit {
			inWord = false
		}
	}

	return wordCount
}

// FuzzSpecialCharDateTimeEncoding tests special character datetime encoding with special non-reversibility handling
func FuzzSpecialCharDateTimeEncoding(f *testing.F) {
	// Seed corpus with basic cases
	f.Add("")
	f.Add("test!@#")
	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8 as Go's string handling will convert invalid bytes to replacement characters
		if !utf8.ValidString(input) {
			return
		}

		// Special character datetime logic isn't fully reversible because:
		// 1. Special characters encode current time, which changes
		// 2. Random placement of special characters
		// But we can verify correct behavior by comparing after removing special characters and timestamps

		translated := TranslateToPejelagarto(input)
		restored := TranslateFromPejelagarto(translated)

		// Clean both for comparison (remove special characters and timestamps)
		inputCleanedTemp, _ := removeISO8601timestamp(input)
		inputCleaned := removeTimestampSpecialCharacters(inputCleanedTemp)
		restoredCleanedTemp, _ := removeISO8601timestamp(restored)
		restoredCleaned := removeTimestampSpecialCharacters(restoredCleanedTemp)

		if inputCleaned != restoredCleaned {
			t.Errorf("Reversibility failed after cleaning.\nInput (cleaned):    %q\nRestored (cleaned): %q", inputCleaned, restoredCleaned)
		}
	})
}

// FuzzTranslatePejelagarto uses fuzzing to test full translation pipeline reversibility
func FuzzTranslatePejelagarto(f *testing.F) {
	// Seed corpus with basic cases
	f.Add("")
	f.Add("Hello, World! 123")
	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8 as Go's string handling will convert invalid bytes to replacement characters
		if !utf8.ValidString(input) {
			return
		}

		// Test: TranslateToPejelagarto -> TranslateFromPejelagarto
		pejelagarto := TranslateToPejelagarto(input)
		reversed := TranslateFromPejelagarto(pejelagarto)

		// Since special character/timestamp logic is now integrated, we need to clean for comparison
		inputCleanedTemp, _ := removeISO8601timestamp(input)
		inputCleaned := removeTimestampSpecialCharacters(inputCleanedTemp)
		reversedCleanedTemp, _ := removeISO8601timestamp(reversed)
		reversedCleaned := removeTimestampSpecialCharacters(reversedCleanedTemp)

		if reversedCleaned != inputCleaned {
			t.Errorf("TranslateToPejelagarto->TranslateFromPejelagarto failed\nInput (cleaned):       %q\nPejelagarto: %q\nReversed (cleaned):    %q", inputCleaned, pejelagarto, reversedCleaned)
		}

		/*
			// Test: TranslateFromPejelagarto -> TranslateToPejelagarto

			// dont validate if input contains invalid pejelagarto numbers (not negative base 7 or positive base 8)
			if !containsInvalidPejelagartoNumbers(input) {
				human := TranslateFromPejelagarto(input)
				reversed = TranslateToPejelagarto(human)

				reversedCleanedTemp, _ = removeISO8601timestamp(reversed)
				reversedCleaned = removeTimestampSpecialCharacters(reversedCleanedTemp)

				if reversedCleaned != inputCleaned {
					t.Errorf("TranslateFromPejelagarto->TranslateToPejelagarto failed\nInput (cleaned):       %q\nHuman:      %q\nReversed (cleaned):    %q", inputCleaned, human, reversedCleaned)
				}
			}
		*/

	})
}

