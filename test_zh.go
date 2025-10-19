package main

import "fmt"

func testZh() {
	bm := createBijectiveMap()

	fmt.Println("Index 1 (forward letters):")
	for k, v := range bm[1] {
		if k == "z" || k == "s" {
			fmt.Printf("  %q -> %q\n", k, v)
		}
	}

	fmt.Println("\nIndex -1 (inverse letters):")
	for k, v := range bm[-1] {
		if k == "z" || k == "s" {
			fmt.Printf("  %q -> %q\n", k, v)
		}
	}

	input := "Zh00"
	fmt.Printf("\nTesting input: %q\n\n", input)

	fmt.Println("=== FromPejelagarto ===")
	result1 := applyMapReplacementsFromPejelagarto(input)
	fmt.Printf("Result: %q\n\n", result1)

	fmt.Println("=== ToPejelagarto ===")
	result2 := applyMapReplacementsToPejelagarto(result1)
	fmt.Printf("Result: %q\n\n", result2)

	fmt.Printf("Reversibility: %v\n", result2 == input)
}
