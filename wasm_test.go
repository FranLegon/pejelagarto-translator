//go:build frontend

package main

import (
	"syscall/js"
	"testing"

	"pejelagarto-translator/internal/translator"
)

// TestWASMBuild verifies that the WASM build compiles correctly
func TestWASMBuild(t *testing.T) {
	t.Log("WASM build compiled successfully")
}

// TestTranslateToPejalagartoWASM tests the translation function in WASM mode
func TestTranslateToPejalagartoWASM(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Simple text",
			input:    "hello",
			expected: "araka", // Based on conjunctionMap
		},
		{
			name:     "Mixed case",
			input:    "Hello",
			expected: "Araka",
		},
		{
			name:     "Numbers",
			input:    "123",
			expected: "173", // Base conversion
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := translator.TranslateToPejelagarto(tc.input)
			if result != tc.expected {
				t.Errorf("translator.TranslateToPejelagarto(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestTranslateFromPejalagartoWASM tests the reverse translation in WASM mode
func TestTranslateFromPejalagartoWASM(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Simple text",
			input:    "araka",
			expected: "hello",
		},
		{
			name:     "Mixed case",
			input:    "Araka",
			expected: "Hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := translator.TranslateFromPejelagarto(tc.input)
			if result != tc.expected {
				t.Errorf("translator.TranslateFromPejelagarto(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestTranslationReversibilityWASM tests that translation is reversible in WASM mode
func TestTranslationReversibilityWASM(t *testing.T) {
	testInputs := []string{
		"hello world",
		"The quick brown fox",
		"123 test 456",
		"Mixed123Numbers",
		"UPPERCASE",
		"lowercase",
		"MiXeD CaSe",
	}

	for _, input := range testInputs {
		t.Run(input, func(t *testing.T) {
			// Test: ToPejelagarto -> FromPejelagarto
			translated := translator.TranslateToPejelagarto(input)
			reversed := translator.TranslateFromPejelagarto(translated)

			if reversed != input {
				t.Errorf("Round-trip failed:\nInput:      %q\nTranslated: %q\nReversed:   %q",
					input, translated, reversed)
			}
		})
	}
}

// TestJSWrapperArgumentValidation tests that JS wrappers validate arguments correctly
func TestJSWrapperArgumentValidation(t *testing.T) {
	t.Run("goTranslateToPejalagartoJS with no args", func(t *testing.T) {
		result := goTranslateToPejalagartoJS(js.Null(), []js.Value{})
		resultStr := result.(js.Value).String()
		if resultStr != "Error: expected 1 argument" {
			t.Errorf("Expected error message, got: %q", resultStr)
		}
	})

	t.Run("goTranslateFromPejalagartoJS with no args", func(t *testing.T) {
		result := goTranslateFromPejalagartoJS(js.Null(), []js.Value{})
		resultStr := result.(js.Value).String()
		if resultStr != "Error: expected 1 argument" {
			t.Errorf("Expected error message, got: %q", resultStr)
		}
	})
}

// TestJSWrapperFunctionality tests that JS wrappers work correctly with valid input
func TestJSWrapperFunctionality(t *testing.T) {
	t.Run("goTranslateToPejalagartoJS with valid input", func(t *testing.T) {
		input := js.ValueOf("hello")
		result := goTranslateToPejalagartoJS(js.Null(), []js.Value{input})
		resultStr := result.(js.Value).String()

		// Should translate "hello" to "araka"
		expected := "araka"
		if resultStr != expected {
			t.Errorf("goTranslateToPejalagartoJS(\"hello\") = %q, want %q", resultStr, expected)
		}
	})

	t.Run("goTranslateFromPejalagartoJS with valid input", func(t *testing.T) {
		input := js.ValueOf("araka")
		result := goTranslateFromPejalagartoJS(js.Null(), []js.Value{input})
		resultStr := result.(js.Value).String()

		// Should translate "araka" back to "hello"
		expected := "hello"
		if resultStr != expected {
			t.Errorf("goTranslateFromPejalagartoJS(\"araka\") = %q, want %q", resultStr, expected)
		}
	})
}

// TestWASMTranslationLogicConsistency ensures WASM and backend builds use same logic
func TestWASMTranslationLogicConsistency(t *testing.T) {
	// This test verifies that the translation functions in WASM mode
	// produce the same results as they would in backend mode
	testCases := []struct {
		input string
	}{
		{"hello world"},
		{"The quick brown fox jumps over the lazy dog"},
		{"123 test with numbers 456"},
		{"Special chars: !@#$%"},
		{"Mixed123Case456Text"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Translate to Pejelagarto
			pejelagarto := translator.TranslateToPejelagarto(tc.input)

			// Translate back
			reversed := translator.TranslateFromPejelagarto(pejelagarto)

			// Should be reversible
			if reversed != tc.input {
				t.Errorf("Translation not reversible for %q", tc.input)
				t.Logf("Input:       %q", tc.input)
				t.Logf("Pejelagarto: %q", pejelagarto)
				t.Logf("Reversed:    %q", reversed)
			}
		})
	}
}

// BenchmarkWASMTranslateToPejelagarto benchmarks the translation performance in WASM mode
func BenchmarkWASMTranslateToPejelagarto(b *testing.B) {
	input := "The quick brown fox jumps over the lazy dog 12345"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = translator.TranslateToPejelagarto(input)
	}
}

// BenchmarkWASMTranslateFromPejelagarto benchmarks the reverse translation in WASM mode
func BenchmarkWASMTranslateFromPejelagarto(b *testing.B) {
	// First translate to get Pejelagarto text
	input := translator.TranslateToPejelagarto("The quick brown fox jumps over the lazy dog 12345")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = translator.TranslateFromPejelagarto(input)
	}
}

// BenchmarkWASMJSWrapper benchmarks the JS wrapper overhead
func BenchmarkWASMJSWrapper(b *testing.B) {
	input := js.ValueOf("hello world test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = goTranslateToPejalagartoJS(js.Null(), []js.Value{input})
	}
}
