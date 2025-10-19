package main

import (
	"fmt"
)

func debugBijectiveMap() {
	bm := createBijectiveMap()

	// Get sorted indices for To Pejelagarto
	indicesTo := getSortedIndices(bm, true)
	fmt.Println("Order for TO Pejelagarto:")
	fmt.Println(indicesTo)

	// Get sorted indices for From Pejelagarto
	indicesFrom := getSortedIndices(bm, false)
	fmt.Println("\nOrder for FROM Pejelagarto:")
	fmt.Println(indicesFrom)
}

func debugTranslation() {
	test := "hello"
	fmt.Printf("\nTranslating %q to Pejelagarto:\n", test)
	result := applyMapReplacementsToPejelagarto(test)
	fmt.Printf("Result: %q\n", result)

	fmt.Printf("\nTranslating %q from Pejelagarto:\n", result)
	reversed := applyMapReplacementsFromPejelagarto(result)
	fmt.Printf("Result: %q\n", reversed)
}
