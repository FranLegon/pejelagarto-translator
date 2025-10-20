package main

import (
	"testing"
)

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
			// Just verify it doesn't panic and produces output
			t.Logf("TranslateToPejelagarto(%q) = %q", tt.input, result)

			// Verify reversibility
			reversed := TranslateFromPejelagarto(result)
			if reversed != tt.input {
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
		{name: "Text with conjunctions", input: "da'xs imf 'jcabs"},
		{name: "Number only", input: "3333333333333333600"},
		{name: "Negative number", input: "-3333333333333333341"},
		{name: "Text with escaped quotes", input: "At''s jibbu"},
		{name: "Empty string", input: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateFromPejelagarto(tt.input)
			// Just verify it doesn't panic and produces output
			t.Logf("TranslateFromPejelagarto(%q) = %q", tt.input, result)

			// Verify reversibility
			reversed := TranslateToPejelagarto(result)
			if reversed != tt.input {
				t.Errorf("Reversibility failed: TranslateFromPejelagarto(%q) = %q, but TranslateToPejelagarto(%q) = %q",
					tt.input, result, result, reversed)
			}
		})
	}
}

// TestFullReversibility tests that translating to and from Pejelagarto is reversible
func TestFullReversibility(t *testing.T) {
	tests := []string{
		"hello world",
		"The quick brown fox jumps over the lazy dog",
		"I have 42 apples and 7 oranges",
		"fish and chips",
		"chapter 123",
		"It's a beautiful day!",
		"Numbers: 0, -5, 999, -1000",
		"Mixed: hello 42 world 7 test",
		"'quoted text'",
		"a b c d e f g",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Translate to Pejelagarto
			pejelagarto := TranslateToPejelagarto(input)
			t.Logf("Original: %q", input)
			t.Logf("Pejelagarto: %q", pejelagarto)

			// Translate back to Human
			reversed := TranslateFromPejelagarto(pejelagarto)
			t.Logf("Reversed: %q", reversed)

			// Check reversibility
			if reversed != input {
				t.Errorf("Reversibility failed:\n  Input:       %q\n  Pejelagarto: %q\n  Reversed:    %q", input, pejelagarto, reversed)
			}
		})
	}
}

// FuzzFullTranslation uses fuzzing to test full translation reversibility
func FuzzFullTranslation(f *testing.F) {
	// Seed corpus
	seeds := []string{
		"hello world",
		"The quick brown fox",
		"42",
		"-5",
		"fish and chips",
		"I'm happy",
		"",
		"a",
		"chapter 1",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		pejelagarto := TranslateToPejelagarto(input)
		reversed := TranslateFromPejelagarto(pejelagarto)

		if reversed != input {
			t.Errorf("Full translation reversibility failed\nInput:       %q\nPejelagarto: %q\nReversed:    %q", input, pejelagarto, reversed)
		}
	})
}
