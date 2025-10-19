package main

import (
	"testing"
)

func TestBugZJC(t *testing.T) {
	input := "'ZJC"
	t.Logf("Input: %q", input)

	pej := applyMapReplacementsToPejelagarto(input)
	t.Logf("ToPejelagarto: %q", pej)

	back := applyMapReplacementsFromPejelagarto(pej)
	t.Logf("FromPejelagarto: %q", back)

	if back != input {
		t.Errorf("MISMATCH! Expected %q, got %q", input, back)
	}
}
