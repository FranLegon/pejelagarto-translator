package main

import (
	"strings"
	"testing"
	"time"
)

// TestProblematicSeed tests the specific seed that causes fuzzing to hang
func TestProblematicSeed(t *testing.T) {
	// Recreate the problematic input from the seed file bda087e7ff257266
	problematicInput := "0" + strings.Repeat("\x05", 2300)

	t.Logf("Testing input: \"0\" + %d repetitions of \\x05", 2300)
	t.Logf("Total length: %d characters\n", len(problematicInput))

	// Test ToPejelagarto
	t.Log("Testing ToPejelagarto...")
	start := time.Now()
	translated := applyMapReplacementsToPejelagarto(problematicInput)
	elapsed := time.Since(start)
	t.Logf("ToPejelagarto Time: %v", elapsed)
	t.Logf("Output length: %d", len(translated))
	if len(translated) <= 50 {
		t.Logf("Output: %q", translated)
	} else {
		t.Logf("Output (first 50 chars): %q...", translated[:50])
	}

	// Test FromPejelagarto
	t.Log("\nTesting FromPejelagarto...")
	start = time.Now()
	reversed := applyMapReplacementsFromPejelagarto(translated)
	elapsed = time.Since(start)
	t.Logf("FromPejelagarto Time: %v", elapsed)
	t.Logf("Output length: %d", len(reversed))

	// Check reversibility
	if reversed != problematicInput {
		t.Errorf("Reversibility FAILED")
		t.Errorf("Expected length: %d, Got length: %d", len(problematicInput), len(reversed))
		if len(reversed) <= 100 {
			t.Errorf("Expected: %q", problematicInput)
			t.Errorf("Got: %q", reversed)
		} else {
			t.Errorf("Expected (first 100 chars): %q...", problematicInput[:100])
			t.Errorf("Got (first 100 chars): %q...", reversed[:100])
		}
	} else {
		t.Log("✓ Reversibility: PASSED")
	}
}
