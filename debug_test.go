package main

import (
	"fmt"
	"testing"
)

func TestSpecificXX00(t *testing.T) {
	input := "XX00"
	fmt.Printf("Input: %q\n", input)

	// Show bijective map for index -2
	bm := createBijectiveMap()
	fmt.Println("\nIndex -2 mappings:")
	for k, v := range bm[-2] {
		fmt.Printf("  %q -> %q\n", k, v)
	}

	translated := applyMapReplacementsToPejelagarto(input)
	fmt.Printf("\nTranslated: %q\n", translated)

	reversed := applyMapReplacementsFromPejelagarto(translated)
	fmt.Printf("Reversed: %q\n", reversed)

	if reversed != input {
		t.Errorf("Reversibility failed")
	}
}
