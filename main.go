// Pejelagarto Translator
// Build command: go build -o pejelagarto-translator.exe main.go
// Run command: .\pejelagarto-translator.exe

package main

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Translation maps for word/syllable replacements
// NOTE: All values must have SAME length as keys (rune count)
// NOTE: Values should use ONLY letters NOT in letterMap (c,h,j,s,t,x,z) to avoid collisions
// NOTE: Values should NOT contain substrings that match conjunction patterns
var wordMap = map[string]string{
	"hello": "jhtxz",
	"world": "zcthx",
	"the":   "zjc",
}

// Translation maps for conjunction (letter pair) replacements
// NOTE: All values must have SAME length as keys (rune count)
// NOTE: Output values use ONLY letters NOT in letterMap (c,h,j,s,t,x,z) to avoid collisions
// NOTE: Avoid repeated characters to prevent ambiguity (e.g., "zz" could be confused with "z"+"z")
var conjunctionMap = map[string]string{
	"ch": "jc",
	"sh": "xs",
	"th": "zt",
}

// Translation maps for single letter replacements (must be invertible)
// NOTE: All values must have SAME length as keys (rune count)
// NOTE: Avoid letters that appear in conjunction patterns (c, h, j, s, t, x, z)
// to prevent collisions between letter outputs and conjunction inputs
var letterMap = map[string]string{
	"a": "i",
	"b": "p",
	"d": "f",
	"e": "o",
	"f": "d",
	"g": "l",
	"k": "w",
	"l": "g",
	"m": "n",
	"n": "m",
	"o": "e",
	"p": "b",
	"q": "v",
	"r": "y",
	"u": "r",
	"v": "q",
	"w": "k",
	"y": "u",
	// Letters c, h, i, j, s, t, x, z are NOT mapped to avoid conjunction collisions
}

// validateMaps checks that all mappings have equal rune lengths for keys and values
func validateMaps() error {
	maps := []struct {
		name string
		m    map[string]string
	}{
		{"wordMap", wordMap},
		{"conjunctionMap", conjunctionMap},
		{"letterMap", letterMap},
	}

	for _, mapInfo := range maps {
		for key, value := range mapInfo.m {
			keyLen := utf8.RuneCountInString(key)
			valueLen := utf8.RuneCountInString(value)
			if keyLen != valueLen {
				return fmt.Errorf("%s: key %q (len=%d) and value %q (len=%d) must have equal rune lengths",
					mapInfo.name, key, keyLen, value, valueLen)
			}
		}
	}
	return nil
}

// createBijectiveMap creates a unified bijective map from the three replacement maps
func createBijectiveMap() map[int32]map[string]string {
	// Validate that all maps have equal-length key-value pairs
	if err := validateMaps(); err != nil {
		panic(err)
	}

	bijectiveMap := make(map[int32]map[string]string)

	// Helper function to add entries to the map
	addEntries := func(sourceMap map[string]string, positive bool) {
		for key, value := range sourceMap {
			var index int32
			var from, to string

			if positive {
				// Positive index: key length, key -> value
				keyLen := utf8.RuneCountInString(key)
				index = int32(keyLen)
				from = key
				to = value
				// Prefix multi-rune values with single quote
				if utf8.RuneCountInString(value) > 1 {
					to = "'" + to
				}
			} else {
				// Negative index: value -> key (inverse)
				// For FromPejelagarto, we need to look for the value WITH the quote prefix
				valueLen := utf8.RuneCountInString(value)
				if valueLen > 1 {
					// Multi-rune: the actual Pejelagarto text has a quote prefix
					// So we need to match "'value" and return key
					index = -int32(valueLen + 1) // +1 for the quote character
					from = "'" + value
					to = key
				} else {
					// Single-rune: no quote prefix in Pejelagarto text
					index = -int32(valueLen)
					from = value
					to = key
				}
			}

			if bijectiveMap[index] == nil {
				bijectiveMap[index] = make(map[string]string)
			}
			bijectiveMap[index][from] = to
		}
	}

	// Add positive entries (key -> value)
	addEntries(wordMap, true)
	addEntries(conjunctionMap, true)
	addEntries(letterMap, true)

	// Add inverse entries (-index: value -> key)
	addEntries(wordMap, false)
	addEntries(conjunctionMap, false)
	addEntries(letterMap, false)

	return bijectiveMap
}

// getSortedIndices returns indices sorted appropriately for the direction
func getSortedIndices(bijectiveMap map[int32]map[string]string, toPejelagarto bool) []int32 {
	indices := make([]int32, 0, len(bijectiveMap))
	for index := range bijectiveMap {
		indices = append(indices, index)
	}

	sort.Slice(indices, func(i, j int) bool {
		iPos := indices[i] > 0
		jPos := indices[j] > 0

		if toPejelagarto {
			// ToPejelagarto: sign(index) desc, abs(index) desc
			// Positive first, then negative; within each group: descending by absolute value
			if iPos != jPos {
				return iPos // positive before negative
			}
		} else {
			// FromPejelagarto: sign(index) asc, abs(index) desc
			// Negative first, then positive; within each group: descending by absolute value
			if iPos != jPos {
				return !iPos // negative before positive
			}
		}

		// Within same sign group: descending by absolute value
		absI := indices[i]
		if absI < 0 {
			absI = -absI
		}
		absJ := indices[j]
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	return indices
} // matchCase applies the casing pattern from original to replacement
func matchCase(original, replacement string) string {
	origRunes := []rune(original)
	replRunes := []rune(replacement)

	result := make([]rune, len(replRunes))
	copy(result, replRunes)

	// Handle quote prefix specially - skip it in case matching
	origOffset := 0
	replOffset := 0
	if len(origRunes) > 0 && origRunes[0] == '\'' {
		origOffset = 1
	}
	if len(replRunes) > 0 && replRunes[0] == '\'' {
		replOffset = 1
	}

	for i := replOffset; i < len(result) && (i-replOffset+origOffset) < len(origRunes); i++ {
		origIdx := i - replOffset + origOffset
		origChar := origRunes[origIdx]
		replChar := result[i]

		if unicode.IsUpper(origChar) {
			upperReplChar := unicode.ToUpper(replChar)
			// Only apply case conversion if it's reversible
			// Check: upper -> lower -> upper gives back the same character
			if unicode.ToUpper(unicode.ToLower(upperReplChar)) == upperReplChar {
				result[i] = upperReplChar
			}
		} else if unicode.IsLower(origChar) {
			lowerReplChar := unicode.ToLower(replChar)
			// Only apply case conversion if it's reversible
			// Check: lower -> upper -> lower gives back the same character
			if unicode.ToLower(unicode.ToUpper(lowerReplChar)) == lowerReplChar {
				result[i] = lowerReplChar
			}
		}
	}

	return string(result)
}

// applyReplacements applies replacements from the bijective map in the specified order
func applyReplacements(input string, bijectiveMap map[int32]map[string]string, indices []int32) string {
	// Use special Unicode characters as markers that won't be in normal text
	const startMarker = "\uFFF0"
	const endMarker = "\uFFF1"

	result := input

	for _, index := range indices {
		replacements := bijectiveMap[index]

		// Sort keys by length descending, then alphabetically
		keys := make([]string, 0, len(replacements))
		for key := range replacements {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			if len(keys[i]) != len(keys[j]) {
				return len(keys[i]) > len(keys[j])
			}
			return keys[i] < keys[j]
		})

		for _, key := range keys {
			value := replacements[key]

			// For FromPejelagarto: if the key has a quote prefix (Pejelagarto multi-rune pattern),
			// remove it from the output value so it doesn't appear in English text
			outputValue := value
			if strings.HasPrefix(key, "'") {
				// This is a Pejelagarto pattern being converted back to English
				// Remove any quote prefix from the output
				outputValue = strings.TrimPrefix(value, "'")
			}

			// Find and replace all occurrences (case-insensitive)
			newResult := strings.Builder{}
			pos := 0
			resultRunes := []rune(result)
			keyRunes := []rune(key)

			// Pre-calculate marker positions for O(n) performance instead of O(n²)
			markerMap := make(map[int]int) // pos -> marker depth at that position
			depth := 0
			startMarkerRune := []rune(startMarker)[0]
			endMarkerRune := []rune(endMarker)[0]
			for i := 0; i < len(resultRunes); i++ {
				markerMap[i] = depth
				if resultRunes[i] == startMarkerRune {
					depth++
				} else if resultRunes[i] == endMarkerRune {
					depth--
				}
			}

			for pos < len(resultRunes) {
				// Check if we're inside markers using pre-calculated map
				if markerMap[pos] > 0 {
					// Skip characters inside markers
					newResult.WriteRune(resultRunes[pos])
					pos++
					continue
				}

				// Check if we're inside a quoted multi-rune pattern (for non-quoted keys only)
				// A quote at the start of a word protects the entire word from being re-matched
				// BUT only if the quote hasn't been processed yet (i.e., not inside markers)
				if len(keyRunes) > 0 && keyRunes[0] != '\'' {
					inQuotedWord := false
					// Look backwards in the current word for an unprocessed quote
					// Optimized: limit backward scan to word boundary
					wordStart := pos
					for i := pos - 1; i >= 0 && i >= pos-50; i-- { // limit backward scan
						if !unicode.IsLetter(resultRunes[i]) && resultRunes[i] != '\'' {
							wordStart = i + 1
							break
						}
						if i == 0 {
							wordStart = 0
						}
					}

					for i := pos - 1; i >= wordStart; i-- {
						if resultRunes[i] == '\'' {
							// Check if this quote is inside markers using pre-calculated map
							if markerMap[i] == 0 {
								inQuotedWord = true
							}
							break
						}
						if !unicode.IsLetter(resultRunes[i]) {
							break
						}
					}
					if inQuotedWord {
						// Skip this character - it's part of a quoted pattern
						newResult.WriteRune(resultRunes[pos])
						pos++
						continue
					}
				}

				// Check if current position matches the key (case-insensitive)
				if pos+len(keyRunes) <= len(resultRunes) {
					matched := true
					matchEndPos := pos + len(keyRunes)

					// Check if the match would span across a quote character (but not start with one)
					// Quotes act as boundaries - patterns can't span across them
					if len(keyRunes) > 0 && keyRunes[0] != '\'' {
						for i := pos; i < matchEndPos; i++ {
							if resultRunes[i] == '\'' {
								matched = false
								break
							}
						}
					}

					// Check if any part of the potential match is inside markers
					// Use pre-calculated marker map
					if matched {
						for i := pos; i < matchEndPos; i++ {
							if markerMap[i] > 0 {
								matched = false
								break
							}
						}
					}

					// Also check for case-insensitive character match
					if matched {
						for i := 0; i < len(keyRunes); i++ {
							resultChar := resultRunes[pos+i]
							keyChar := keyRunes[i]

							// Check if characters match (case-insensitive)
							if unicode.ToLower(resultChar) != unicode.ToLower(keyChar) {
								matched = false
								break
							}

							// Additional check: ensure case conversion is reversible for this character
							// Skip match if either character has non-reversible case conversion
							if unicode.IsLetter(resultChar) {
								// Check if upper->lower->upper is reversible
								if unicode.ToUpper(unicode.ToLower(resultChar)) != unicode.ToUpper(resultChar) {
									matched = false
									break
								}
							}
							if unicode.IsLetter(keyChar) {
								// Check if upper->lower->upper is reversible
								if unicode.ToUpper(unicode.ToLower(keyChar)) != unicode.ToUpper(keyChar) {
									matched = false
									break
								}
							}
						}
					}

					if matched {
						// Extract matched text with original casing
						matchedText := string(resultRunes[pos : pos+len(keyRunes)])
						// Apply case matching
						casedValue := matchCase(matchedText, outputValue)
						// Wrap in markers and add
						newResult.WriteString(startMarker)
						newResult.WriteString(casedValue)
						newResult.WriteString(endMarker)
						pos += len(keyRunes)
						continue
					}
				}

				// No match, just copy the character
				newResult.WriteRune(resultRunes[pos])
				pos++
			}

			result = newResult.String()
		}
	}

	// Remove all markers
	result = strings.ReplaceAll(result, startMarker, "")
	result = strings.ReplaceAll(result, endMarker, "")

	return result
}

// applyMapReplacementsToPejelagarto translates text to Pejelagarto using map replacements
func applyMapReplacementsToPejelagarto(input string) string {
	// If input is not valid UTF-8, return it unchanged
	if !utf8.ValidString(input) {
		return input
	}
	// Use a special marker for literal quotes in input to avoid ambiguity
	// with quote prefixes used in Pejelagarto output
	const quoteMarker = "\uFFF2"
	input = strings.ReplaceAll(input, "'", quoteMarker)

	bijectiveMap := createBijectiveMap()
	indices := getSortedIndices(bijectiveMap, true) // to Pejelagarto
	result := applyReplacements(input, bijectiveMap, indices)

	// Restore literal quotes as doubled quotes in the output
	result = strings.ReplaceAll(result, quoteMarker, "''")
	return result
}

// applyMapReplacementsFromPejelagarto translates text from Pejelagarto using map replacements
func applyMapReplacementsFromPejelagarto(input string) string {
	// If input is not valid UTF-8, return it unchanged
	if !utf8.ValidString(input) {
		return input
	}
	// Convert doubled quotes (escaped literals) to temporary marker
	const quoteMarker = "\uFFF2"
	input = strings.ReplaceAll(input, "''", quoteMarker)

	bijectiveMap := createBijectiveMap()
	indices := getSortedIndices(bijectiveMap, false) // from Pejelagarto
	result := applyReplacements(input, bijectiveMap, indices)

	// Restore literal quotes from marker
	result = strings.ReplaceAll(result, quoteMarker, "'")
	return result
}

// applyNumbersLogicToPejelagarto applies number transformation for Pejelagarto encoding
// Adds offset (5699447592686571) to all base-10 numbers and converts to base-7
// Preserves leading zeros and handles signs separately
// Uses arbitrary-precision arithmetic to handle any size number
func applyNumbersLogicToPejelagarto(input string) string {
	// If input is not valid UTF-8, return it unchanged
	if !utf8.ValidString(input) {
		return input
	}
	var result strings.Builder
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		// Check if we're at the start of a number (including negative)
		// Only process ASCII digits 0-9, not Unicode digits
		if (runes[i] >= '0' && runes[i] <= '9') || (runes[i] == '-' && i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '9') {
			// Extract sign
			isNegative := false
			if runes[i] == '-' {
				isNegative = true
				i++
			}

			// Count leading zeros
			leadingZeros := 0
			for i < len(runes) && runes[i] == '0' {
				leadingZeros++
				i++
			}

			// Get the rest of the digits
			digitStart := i
			for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
				i++
			}

			numberStr := string(runes[digitStart:i])

			// If only zeros, handle specially
			if numberStr == "" {
				// Only zeros (e.g., "00", "000", "-0")
				if isNegative {
					result.WriteRune('-')
				}
				for j := 0; j < leadingZeros; j++ {
					result.WriteRune('0')
				}
			} else {
				// Parse as big.Int
				absValue := new(big.Int)
				_, ok := absValue.SetString(numberStr, 10)
				if !ok {
					// Parse failed, preserve as-is
					if isNegative {
						result.WriteRune('-')
					}
					for j := 0; j < leadingZeros; j++ {
						result.WriteRune('0')
					}
					result.WriteString(numberStr)
					continue
				}

				// Add offset and convert to base-7
				offset := big.NewInt(5699447592686571)
				offsetValue := new(big.Int).Add(absValue, offset)
				base7 := offsetValue.Text(7)

				// Write sign if negative
				if isNegative {
					result.WriteRune('-')
				}
				// Preserve leading zeros
				for j := 0; j < leadingZeros; j++ {
					result.WriteRune('0')
				}
				result.WriteString(base7)
			}
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}

	return result.String()
}

// applyNumbersLogicFromPejelagarto applies number transformation from Pejelagarto encoding
// Converts all base-7 numbers to base-10 and subtracts offset (5699447592686571)
// Preserves leading zeros and handles signs separately
// Uses arbitrary-precision arithmetic to handle any size number
func applyNumbersLogicFromPejelagarto(input string) string {
	// If input is not valid UTF-8, return it unchanged
	if !utf8.ValidString(input) {
		return input
	}
	var result strings.Builder
	runes := []rune(input)
	i := 0

	for i < len(runes) {
		// Check if we're at the start of a base 7 number (including negative)
		if isBase7Digit(runes[i]) || (runes[i] == '-' && i+1 < len(runes) && isBase7Digit(runes[i+1])) {
			// Extract sign
			isNegative := false
			if runes[i] == '-' {
				isNegative = true
				i++
			}

			// Count leading zeros
			leadingZeros := 0
			for i < len(runes) && runes[i] == '0' {
				leadingZeros++
				i++
			}

			// Get the rest of the digits (must be base 7)
			digitStart := i
			for i < len(runes) && isBase7Digit(runes[i]) {
				i++
			}

			// Check if followed by digits 7-9 (which means it's actually a base-10 number, not base-7)
			// If so, we need to consume those digits and skip transformation
			hasHighDigits := false
			highDigitEnd := i
			if i < len(runes) && runes[i] >= '7' && runes[i] <= '9' {
				hasHighDigits = true
				// Consume remaining base-10 digits
				for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
					i++
				}
				highDigitEnd = i
			}

			// If we found digits 7-9, this is a base-10 number, not base-7 - preserve as-is
			if hasHighDigits {
				numberStr := string(runes[digitStart:highDigitEnd])
				if isNegative {
					result.WriteRune('-')
				}
				for j := 0; j < leadingZeros; j++ {
					result.WriteRune('0')
				}
				result.WriteString(numberStr)
				continue
			}

			numberStr := string(runes[digitStart:i])

			// If only zeros, handle specially
			if numberStr == "" {
				// Only zeros (e.g., "00", "000", "-0")
				if isNegative {
					result.WriteRune('-')
				}
				for j := 0; j < leadingZeros; j++ {
					result.WriteRune('0')
				}
			} else {
				// Parse as big.Int from base-7
				base7Value := new(big.Int)
				_, ok := base7Value.SetString(numberStr, 7)
				if !ok {
					// Parse failed, preserve as-is
					if isNegative {
						result.WriteRune('-')
					}
					for j := 0; j < leadingZeros; j++ {
						result.WriteRune('0')
					}
					result.WriteString(numberStr)
					continue
				}

				// Convert from base-7 and subtract offset
				offset := big.NewInt(5699447592686571)
				resultValue := new(big.Int).Sub(base7Value, offset)
				base10Str := resultValue.Text(10)

				// Write sign if negative
				if isNegative {
					result.WriteRune('-')
				}
				// Preserve leading zeros
				for j := 0; j < leadingZeros; j++ {
					result.WriteRune('0')
				}
				result.WriteString(base10Str)
			}
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}

	return result.String()
}

// isBase7Digit checks if a rune is a valid base 7 digit (0-6)
func isBase7Digit(r rune) bool {
	return r >= '0' && r <= '6'
}

// TranslateToPejelagarto translates Human text to Pejelagarto
func TranslateToPejelagarto(input string) string {
	input = applyNumbersLogicToPejelagarto(input)
	input = applyMapReplacementsToPejelagarto(input)
	return input
}

// TranslateFromPejelagarto translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
	input = applyMapReplacementsFromPejelagarto(input)
	input = applyNumbersLogicFromPejelagarto(input)
	return input
}

func main() {
	// Placeholder for future web server implementation
	fmt.Println("Pejelagarto Translator - Implementation in progress")

	// Quick test
	input := "hello"
	pej := applyMapReplacementsToPejelagarto(input)
	back := applyMapReplacementsFromPejelagarto(pej)
	fmt.Println("Input:", input)
	fmt.Println("ToPejelagarto:", pej)
	fmt.Println("FromPejelagarto:", back)
}
