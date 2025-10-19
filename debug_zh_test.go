package main

import (
	"fmt"
	"sort"
	"testing"
)

func TestDebugZh(t *testing.T) {
	bm := createBijectiveMap()

	fmt.Println("ToPejelagarto order:")
	indicesTo := getSortedIndices(bm, true)
	fmt.Println(indicesTo)

	fmt.Println("\nFromPejelagarto order:")
	indicesFrom := getSortedIndices(bm, false)
	fmt.Println(indicesFrom)

	fmt.Println("\nIndex 3 entries:")
	var keys3 []string
	for k := range bm[3] {
		keys3 = append(keys3, k)
	}
	sort.Strings(keys3)
	for _, k := range keys3 {
		fmt.Printf("  %q -> %q\n", k, bm[3][k])
	}

	fmt.Println("\nIndex -3 entries:")
	var keysNeg3 []string
	for k := range bm[-3] {
		keysNeg3 = append(keysNeg3, k)
	}
	sort.Strings(keysNeg3)
	for _, k := range keysNeg3 {
		fmt.Printf("  %q -> %q\n", k, bm[-3][k])
	}

	fmt.Println("\nIndex 2 entries:")
	var keys2 []string
	for k := range bm[2] {
		keys2 = append(keys2, k)
	}
	sort.Strings(keys2)
	for _, k := range keys2 {
		fmt.Printf("  %q -> %q\n", k, bm[2][k])
	}

	fmt.Println("\nIndex -2 entries:")
	var keysNeg2 []string
	for k := range bm[-2] {
		keysNeg2 = append(keysNeg2, k)
	}
	sort.Strings(keysNeg2)
	for _, k := range keysNeg2 {
		fmt.Printf("  %q -> %q\n", k, bm[-2][k])
	}

	fmt.Println("\n\nTesting 'XXH':")
	result1 := applyMapReplacementsToPejelagarto("XXH")
	fmt.Printf("ToPejelagarto('XXH') = %q\n", result1)

	result2 := applyMapReplacementsFromPejelagarto(result1)
	fmt.Printf("FromPejelagarto(%q) = %q\n", result1, result2)
}
