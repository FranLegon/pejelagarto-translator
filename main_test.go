package main

import (
	"testing"
	"unicode/utf8"
)

// FuzzReversibility uses Go's native fuzzing to test reversibility with arbitrary inputs
func FuzzReversibility(f *testing.F) {
	// Add seed corpus with diverse test cases
	f.Add("hello world")
	f.Add("Hello World")
	f.Add("the quick brown fox")
	f.Add("this and that")
	f.Add("can you help")
	f.Add("thank you friend")
	f.Add("church and ship")
	f.Add("chat with them")
	f.Add("phone that chapter")
	f.Add("A simple test with vowels: a e i o u")
	f.Add("Hello, this is the world and you can thank your friend when you have time!")
	f.Add("Chapter one: the quick check")
	f.Add("The ship will thank the church")
	f.Add("When can you come?")
	f.Add("What is that thing?")
	f.Add("Good morning, how are you?")
	f.Add("Yes, maybe we can meet at night")
	f.Add("Please love this great day")
	f.Add("Hello friend, where are you from?")
	f.Add("Thank you so much for the evening")
	// Add some edge cases
	f.Add("")
	f.Add("a")
	f.Add("ABC")
	f.Add("123")
	f.Add("!@#$%^&*()")
	f.Add("   spaces   ")
	f.Add("newline\ntest")
	f.Add("tab\ttest")

	f.Fuzz(func(t *testing.T, original string) {
		// Skip invalid UTF-8 strings - the translator works with valid UTF-8 text only
		if !utf8.ValidString(original) {
			t.Skip("Skipping invalid UTF-8 input")
		}

		// Test: English -> Pejelagarto -> English should return original
		// This is the primary use case: translating English text to Pejelagarto and back
		translated := TranslateToPejelagarto(original)
		backToOriginal := TranslateFromPejelagarto(translated)

		if backToOriginal != original {
			t.Errorf("English→Pejelagarto→English round-trip failed!\nOriginal:   %q\nPejelagarto: %q\nBack:       %q", original, translated, backToOriginal)
		}
	})
}
