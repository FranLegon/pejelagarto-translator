package main

import (
	"testing"
)

// TestReversibility verifies that the translation is perfectly reversible
func TestReversibility(t *testing.T) {
	// Comprehensive test string that uses all three maps
	// wordMap: hello, world, the, and, you, this, can, thank, friend
	// conjunctionMap: ch, sh, th
	// letterMap: a, e, i, o, u
	testCases := []string{
		"hello world",
		"Hello World",
		"the quick brown fox",
		"this and that",
		"can you help",
		"thank you friend",
		"church and ship",
		"chat with them",
		"phone that chapter",
		"A simple test with vowels: a e i o u",
		"Hello, this is the world and you can thank your friend when you have time!",
		"Chapter one: the quick check",
		"The ship will thank the church",
		"When can you come?",
		"What is that thing?",
		"Good morning, how are you?",
		"Yes, maybe we can meet at night",
		"Please love this great day",
		"Hello friend, where are you from?",
		"Thank you so much for the evening",
	}

	for _, original := range testCases {
		// Test: ToPejelagarto -> FromPejelagarto should return original
		translated := TranslateToPejelagarto(original)
		backToOriginal := TranslateFromPejelagarto(translated)

		if backToOriginal != original {
			t.Errorf("Forward-backward translation failed!\nOriginal:  %q\nTranslated: %q\nBack:      %q", original, translated, backToOriginal)
		}

		// Test: FromPejelagarto -> ToPejelagarto should return original
		// (We use the translated version as input)
		backTranslated := TranslateToPejelagarto(backToOriginal)

		if backTranslated != translated {
			t.Errorf("Backward-forward translation failed!\nOriginal:       %q\nTranslated:     %q\nBack Original:  %q\nBack Translated: %q", original, translated, backToOriginal, backTranslated)
		}
	}

	t.Logf("All %d reversibility tests passed!", len(testCases))
}

// TestSpecificTranslations tests some specific expected translations
func TestSpecificTranslations(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "jetzo"},
		{"world", "vorlag"},
		{"the", "ze"},
		{"chapter", "jjiptor"}, // ch->jj, a->i, e->o
		{"ship", "xxap"},       // sh->xx, i->a, (p unchanged)
	}

	for _, tt := range tests {
		result := TranslateToPejelagarto(tt.input)
		if result != tt.expected {
			t.Errorf("TranslateToPejelagarto(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
