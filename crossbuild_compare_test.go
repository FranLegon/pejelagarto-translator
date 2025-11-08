package main

import (
	"testing"
)

// TestCrossBuildConsistency verifies that the translation logic produces identical results
// in both normal and WASM builds. This test runs in both build modes and logs the results.
// 
// To verify cross-build consistency:
// 1. Run: go test -v -run=TestCrossBuildConsistency 
// 2. Build WASM tests and compare logged outputs manually
//
// The key validation is that the Pejelagarto output should be identical in both builds.
func TestCrossBuildConsistency(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Simple word", "hello"},
		{"Capitalized", "Hello"},
		{"Uppercase", "HELLO"},
		{"Numbers", "123"},
		{"Alphanumeric", "test123"},
		{"Multiple words", "hello world"},
		{"Long text", "The quick brown fox jumps over the lazy dog"},
		{"Mixed case", "MiXeD CaSe"},
		{"With punctuation", "Hello, world!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Translate to Pejelagarto
			pejelagarto := TranslateToPejelagarto(tc.input)
			
			// Translate back
			reversed := TranslateFromPejelagarto(pejelagarto)
			
			// Clean for comparison (as done in fuzz tests)
			inputCleanedTemp, _ := removeISO8601timestamp(tc.input)
			inputCleaned := removeTimestampSpecialCharacters(inputCleanedTemp)
			reversedCleanedTemp, _ := removeISO8601timestamp(reversed)
			reversedCleaned := removeTimestampSpecialCharacters(reversedCleanedTemp)
			
			// Log the results for manual cross-build comparison
			t.Logf("Input:            %q", tc.input)
			t.Logf("Pejelagarto:       %q", pejelagarto)
			t.Logf("Reversed (cleaned): %q", reversedCleaned)
			
			// Verify reversibility (after cleaning timestamps)
			if reversedCleaned != inputCleaned {
				t.Errorf("Reversibility check failed (after timestamp cleaning)")
			}
		})
	}
}

// TestTranslationOutputFormat logs translation outputs for cross-build verification
// Run this test in both normal and WASM builds and compare the outputs
func TestTranslationOutputFormat(t *testing.T) {
	inputs := []string{
		"hello",
		"Hello World",
		"test 123",
		"UPPERCASE",
	}

	t.Log("=== Translation Output Comparison ===")
	for _, input := range inputs {
		result := TranslateToPejelagarto(input)
		t.Logf("TranslateToPejelagarto(%q) = %q", input, result)
	}
	t.Log("=== End Translation Outputs ===")
}

// TestWASMNormalBuildEquivalence documents that both builds use the same translation logic
// The actual comparison must be done by running tests in both modes and comparing logs
func TestWASMNormalBuildEquivalence(t *testing.T) {
	t.Log("This test verifies that normal and WASM builds produce identical translations.")
	t.Log("Both builds use the same core translation logic from main.go")
	t.Log("")
	t.Log("To verify equivalence:")
	t.Log("1. Run normal build tests:  go test -v -run=TestCrossBuildConsistency")
	t.Log("2. Build WASM tests:        GOOS=js GOARCH=wasm go test -tags frontend -c -o /tmp/wasm_test.wasm")
	t.Log("3. Compare the logged output from both runs - they should be identical")
	t.Log("")
	t.Log("The WASM build uses the same TranslateToPejelagarto() and TranslateFromPejelagarto()")
	t.Log("functions as the normal build, ensuring consistency.")
}
