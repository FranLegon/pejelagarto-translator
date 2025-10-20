package main

import (
	"testing"
	"unicode"
	"unicode/utf8"
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

// FuzzApplyPunctuationReplacements tests punctuation replacement reversibility
func FuzzApplyPunctuationReplacements(f *testing.F) {
	// Seed corpus with punctuation examples
	seeds := []string{
		"Hello, world!",
		"What? Really!",
		"Test... more dots..",
		"Quote: 'hello'",
		`Quote: "hello"`,
		"Dash-separated-words",
		"(parentheses) test",
		"Mix: of; punctuation, marks!",
		"?.!,;:'-\"",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

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
	// Seed corpus with diverse examples
	seeds := []string{
		"hello world",               // 2 words (even) -> Tribonacci
		"Go programming language",   // 3 words (odd) -> Fibonacci
		"one two three four five",   // 5 words (odd) -> Fibonacci
		"A B C D",                   // 4 words (even) -> Tribonacci
		"test",                      // 1 word (odd) -> Fibonacci
		"Testing reversibility",     // 2 words (even) -> Tribonacci
		"",                          // 0 words -> no-op
		"a",                         // 1 word (odd) -> Fibonacci
		"UPPERCASE lowercase MiXeD", // 3 words (odd) -> Fibonacci
		"123 456",                   // 2 words (even) -> Tribonacci
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

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

// FuzzEmojiDateTimeEncoding tests emoji datetime encoding with special non-reversibility handling
func FuzzEmojiDateTimeEncoding(f *testing.F) {
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
		// Skip invalid UTF-8 as Go's string handling will convert invalid bytes to replacement characters
		if !utf8.ValidString(input) {
			return
		}
		
		// Emoji datetime logic isn't fully reversible because:
		// 1. Emojis encode current time, which changes
		// 2. Random placement of emojis
		// But we can verify correct behavior by comparing after removing emojis and timestamps
		
		translated := TranslateToPejelagarto(input)
		restored := TranslateFromPejelagarto(translated)
		
		// Clean both for comparison (remove emojis and timestamps)
		inputCleaned := removeAllEmojies(removeISO8601timestamp(input))
		restoredCleaned := removeAllEmojies(removeISO8601timestamp(restored))

		if inputCleaned != restoredCleaned {
			t.Errorf("Reversibility failed after cleaning.\nInput (cleaned):    %q\nRestored (cleaned): %q", inputCleaned, restoredCleaned)
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
		// Skip invalid UTF-8 as Go's string handling will convert invalid bytes to replacement characters
		if !utf8.ValidString(input) {
			return
		}
		
		// Test: TranslateToPejelagarto -> TranslateFromPejelagarto
		pejelagarto := TranslateToPejelagarto(input)
		reversed := TranslateFromPejelagarto(pejelagarto)

		// Since emoji/timestamp logic is now integrated, we need to clean for comparison
		inputCleaned := removeAllEmojies(removeISO8601timestamp(input))
		reversedCleaned := removeAllEmojies(removeISO8601timestamp(reversed))

		if reversedCleaned != inputCleaned {
			t.Errorf("TranslateToPejelagarto->TranslateFromPejelagarto failed\nInput (cleaned):       %q\nPejelagarto: %q\nReversed (cleaned):    %q", inputCleaned, pejelagarto, reversedCleaned)
		}
	})
}

// TestTranslateToPejelagarto tests the full translation to Pejelagarto
func TestTranslateToPejelagarto(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "Simple word", input: "hello"},
		{name: "Word with number", input: "hello 42"},
		{name: "The quick brown fox", input: "The quick brown fox"},
		{name: "Text with conjunctions", input: "fish and chips"},
		{name: "Number only", input: "123"},
		{name: "Negative number", input: "-5"},
		{name: "Text with quotes", input: "I'm happy"},
		{name: "Empty string", input: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateToPejelagarto(tt.input)
			// Verify reversibility (with emoji/timestamp cleaning)
			reversed := TranslateFromPejelagarto(result)
			inputCleaned := removeAllEmojies(removeISO8601timestamp(tt.input))
			reversedCleaned := removeAllEmojies(removeISO8601timestamp(reversed))
			if reversedCleaned != inputCleaned {
				t.Errorf("Reversibility failed: TranslateToPejelagarto(%q) = %q, but TranslateFromPejelagarto(%q) = %q",
					tt.input, result, result, reversed)
			}
		})
	}
}

// TestTranslateFromPejelagarto tests the full translation from Pejelagarto
func TestTranslateFromPejelagarto(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "Simple word", input: "'jhtxz"},
		{name: "Word with number", input: "'jhtxz 3333333333333333423"},
		{name: "The quick brown fox", input: "'Zjc vracw pyekm dex"},
		{name: "Text with conjunctions", input: "da'xs ĩmf 'jcabs"},
		{name: "Number only", input: "3333333333333333600"},
		{name: "Negative number", input: "-3333333333333333341"},
		{name: "Text with escaped quotes", input: "At''s jibbu"},
		{name: "Empty string", input: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateFromPejelagarto(tt.input)
			// Verify reversibility (with emoji/timestamp cleaning)
			reversed := TranslateToPejelagarto(result)
			inputCleaned := removeAllEmojies(removeISO8601timestamp(tt.input))
			reversedCleaned := removeAllEmojies(removeISO8601timestamp(reversed))
			if reversedCleaned != inputCleaned {
				t.Errorf("Reversibility failed: TranslateFromPejelagarto(%q) = %q, but TranslateToPejelagarto(%q) = %q",
					tt.input, result, result, reversed)
			}
		})
	}
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

			if reversed != tt.input {
				t.Errorf("Reversibility failed: input=%q, accented=%q, reversed=%q", tt.input, accented, reversed)
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
