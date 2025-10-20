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

	// Validate accent wheels on first call
	if err := validateAccentWheels(); err != nil {
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

// Accent wheels for vowel replacement
// oneRuneAccentsWheel: single-rune accent forms (1 rune input → 1 rune output)
// twoRunesAccentsWheel: two-rune accent forms (2 runes input → 2 runes output)
// Each vowel has its own independent wheel - position 3 for 'a' can be different from position 3 for 'e'
// Only includes accents with reversible case conversion (ToUpper then ToLower returns original)

var oneRuneAccentsWheel = map[rune][]string{
	'a': {"a", "à", "á", "â", "ã", "å", "ä", "ā", "ă"}, // 9 single-rune accents for 'a'
	'e': {"e", "è", "é", "ê", "ẽ", "ė", "ë", "ē", "ĕ"}, // 9 single-rune accents for 'e'
	'i': {"i", "ì", "í", "î", "ĩ", "ï", "ī", "ĭ"},      // 8 single-rune accents for 'i' (ı excluded - case not reversible)
	'o': {"o", "ò", "ó", "ô", "õ", "ø", "ö", "ō", "ŏ"}, // 9 single-rune accents for 'o'
	'u': {"u", "ù", "ú", "û", "ũ", "ů", "ü", "ū", "ŭ"}, // 9 single-rune accents for 'u'
	'y': {"y", "ỳ", "ý", "ŷ", "ỹ", "ẏ", "ÿ", "ȳ"},      // 8 single-rune accents for 'y' (ỵ excluded if needed)
}

var twoRunesAccentsWheel = map[rune][]string{
	// Using combining diacritics (base + combining character = 2 runes)
	// U+0328 = combining ogonek, U+030C = combining caron, U+031B = combining horn
	'a': {"a\u0328", "a\u030C"},            // a+ogonek, a+caron (2 runes each)
	'e': {"e\u0328", "e\u030C"},            // e+ogonek, e+caron (2 runes each)
	'i': {"i\u0328", "i\u030C"},            // i+ogonek, i+caron (2 runes each)
	'o': {"o\u0328", "o\u030C", "o\u031B"}, // o+ogonek, o+caron, o+horn (2 runes each)
	'u': {"u\u0328", "u\u030C", "u\u031B"}, // u+ogonek, u+caron, u+horn (2 runes each)
	'y': {"y\u0328"},                       // y+ogonek (2 runes)
}

// validateAccentWheels checks that all accent forms have the expected rune count and case reversibility
func validateAccentWheels() error {
	// Validate oneRuneAccentsWheel - all values should be single runes with reversible case
	for baseVowel, accents := range oneRuneAccentsWheel {
		for idx, accentedForm := range accents {
			runeCount := utf8.RuneCountInString(accentedForm)
			if runeCount != 1 {
				return fmt.Errorf("oneRuneAccentsWheel['%c'][%d] = %q has %d runes, expected 1",
					baseVowel, idx, accentedForm, runeCount)
			}

			// Check case reversibility
			r := []rune(accentedForm)[0]
			upperForm := unicode.ToUpper(r)
			if unicode.ToLower(upperForm) != r {
				return fmt.Errorf("oneRuneAccentsWheel['%c'][%d] = %q has non-reversible case conversion",
					baseVowel, idx, accentedForm)
			}
		}
	}

	// Validate twoRunesAccentsWheel - all values should be exactly 2 runes (combining character sequences)
	// Format: base character + combining diacritic
	for baseVowel, accents := range twoRunesAccentsWheel {
		for idx, accentedForm := range accents {
			runeCount := utf8.RuneCountInString(accentedForm)
			if runeCount != 2 {
				return fmt.Errorf("twoRunesAccentsWheel['%c'][%d] = %q has %d runes, expected 2",
					baseVowel, idx, accentedForm, runeCount)
			}

			// Check case reversibility for the base character (first rune)
			runes := []rune(accentedForm)
			baseChar := runes[0]
			upperForm := unicode.ToUpper(baseChar)
			if unicode.ToLower(upperForm) != baseChar {
				return fmt.Errorf("twoRunesAccentsWheel['%c'][%d] = %q has non-reversible case conversion for base char",
					baseVowel, idx, accentedForm)
			}
		}
	}

	return nil
} // isVowel checks if a rune is a vowel (including y and accented forms)
func isVowel(r rune) bool {
	lower := unicode.ToLower(r)
	
	// Verify case conversion is reversible if character is uppercase
	// This prevents issues with characters like İ (Turkish I with dot, U+0130)
	// which lowercase to 'i' but ToUpper('i') != 'İ'
	if unicode.IsUpper(r) {
		if unicode.ToUpper(lower) != r {
			return false // Not reversible, don't treat as vowel
		}
	}
	
	lowerStr := string(lower)

	// Check oneRuneAccentsWheel
	for _, accents := range oneRuneAccentsWheel {
		for _, accentedForm := range accents {
			if accentedForm == lowerStr {
				return true
			}
		}
	}

	// Check twoRunesAccentsWheel
	for _, accents := range twoRunesAccentsWheel {
		for _, accentedForm := range accents {
			if accentedForm == lowerStr {
				return true
			}
		}
	}

	return false
} // primeFactorize returns prime factors with their powers
// Example: 245 -> map[5:1, 7:2] means 5^1 * 7^2
func primeFactorize(n int) map[int]int {
	factors := make(map[int]int)
	if n <= 1 {
		return factors
	}

	// Check for factor 2
	for n%2 == 0 {
		factors[2]++
		n = n / 2
	}

	// Check for odd factors from 3 onwards
	for i := 3; i*i <= n; i += 2 {
		for n%i == 0 {
			factors[i]++
			n = n / i
		}
	}

	// If n is still greater than 1, it's a prime factor
	if n > 1 {
		factors[n]++
	}

	return factors
}

// findAccentIndex finds the current accent index for a vowel in oneRuneAccentsWheel
func findAccentIndex(baseVowel rune, vowelStr string) int {
	accents, ok := oneRuneAccentsWheel[baseVowel]
	if !ok {
		return 0
	}
	for idx, accentedForm := range accents {
		if accentedForm == vowelStr {
			return idx
		}
	}
	return 0 // Default to no accent
}

// getBaseVowel gets the base vowel character (lowercase, no accent)
func getBaseVowel(vowelStr string) rune {
	// Try to find in oneRuneAccentsWheel
	for baseVowel, accents := range oneRuneAccentsWheel {
		for _, accentedForm := range accents {
			if accentedForm == vowelStr {
				return baseVowel
			}
		}
	}
	// Try to find in twoRunesAccentsWheel
	for baseVowel, accents := range twoRunesAccentsWheel {
		for _, accentedForm := range accents {
			if accentedForm == vowelStr {
				return baseVowel
			}
		}
	}
	// If not found, return first rune as fallback
	runes := []rune(vowelStr)
	if len(runes) > 0 {
		return unicode.ToLower(runes[0])
	}
	return 'a'
}

// applyAccentReplacementLogicToPejelagarto applies accent changes based on prime factorization
func applyAccentReplacementLogicToPejelagarto(input string) string {
	if !utf8.ValidString(input) {
		return input
	}

	runes := []rune(input)
	totalCount := len(runes)

	if totalCount == 0 {
		return input
	}

	// Get prime factors
	factors := primeFactorize(totalCount)
	if len(factors) == 0 {
		return input // No factors (totalCount is 1 or 0)
	}

	// Find all vowels and their positions
	vowelPositions := []int{}
	for i, r := range runes {
		if isVowel(r) {
			vowelPositions = append(vowelPositions, i)
		}
	}

	if len(vowelPositions) == 0 {
		return input // No vowels to modify
	}

	// Apply accent changes for each prime factor
	// Work directly with runes and ensure single-rune replacements only
	result := make([]rune, len(runes))
	copy(result, runes)

	for prime, power := range factors {
		// Find the nth vowel (1-indexed to match prime)
		vowelIndex := prime - 1 // Convert to 0-indexed

		if vowelIndex >= 0 && vowelIndex < len(vowelPositions) {
			pos := vowelPositions[vowelIndex]
			vowelRune := result[pos]
			isUpper := unicode.IsUpper(vowelRune)

			// Get current accent index and base vowel
			vowelStr := string(unicode.ToLower(vowelRune))
			baseVowel := getBaseVowel(vowelStr)
			currentAccentIndex := findAccentIndex(baseVowel, vowelStr)

			// Get the wheel for this vowel
			wheel, ok := oneRuneAccentsWheel[baseVowel]
			if !ok || len(wheel) == 0 {
				continue // Skip if no wheel for this vowel
			}

			// Move forward by power positions
			newAccentIndex := (currentAccentIndex + power) % len(wheel)

			// Get new accented form (always single rune from our wheel)
			newAccentedForm := wheel[newAccentIndex]
			newAccentRunes := []rune(newAccentedForm)

			// Only apply if:
			// 1. We got a single rune
			// 2. The new form is different from base vowel (otherwise accent info is lost)
			if len(newAccentRunes) == 1 {
				// Check if this accent actually changes the vowel
				baseVowelStr := string(baseVowel)
				if newAccentedForm != baseVowelStr || newAccentIndex == 0 {
					if isUpper {
						// Apply uppercase - but only if case conversion is reversible
						upperForm := unicode.ToUpper(newAccentRunes[0])
						if unicode.ToLower(upperForm) == newAccentRunes[0] {
							result[pos] = upperForm
						} else {
							// Case conversion not reversible, keep lowercase
							result[pos] = newAccentRunes[0]
						}
					} else {
						result[pos] = newAccentRunes[0]
					}
				}
			}
		}
	}

	return string(result)
}

// applyAccentReplacementLogicFromPejelagarto reverses accent changes based on prime factorization
func applyAccentReplacementLogicFromPejelagarto(input string) string {
	if !utf8.ValidString(input) {
		return input
	}

	runes := []rune(input)
	totalCount := len(runes)

	if totalCount == 0 {
		return input
	}

	// Get prime factors
	factors := primeFactorize(totalCount)
	if len(factors) == 0 {
		return input
	}

	// Find all vowels and their positions
	vowelPositions := []int{}
	for i, r := range runes {
		if isVowel(r) {
			vowelPositions = append(vowelPositions, i)
		}
	}

	if len(vowelPositions) == 0 {
		return input
	}

	// Apply accent changes for each prime factor (backwards)
	result := make([]rune, len(runes))
	copy(result, runes)

	for prime, power := range factors {
		// Find the nth vowel (1-indexed to match prime)
		vowelIndex := prime - 1

		if vowelIndex >= 0 && vowelIndex < len(vowelPositions) {
			pos := vowelPositions[vowelIndex]
			vowelRune := result[pos]
			isUpper := unicode.IsUpper(vowelRune)

			// Get current accent index and base vowel
			vowelStr := string(unicode.ToLower(vowelRune))
			baseVowel := getBaseVowel(vowelStr)
			currentAccentIndex := findAccentIndex(baseVowel, vowelStr)

			// Get the wheel for this vowel
			wheel, ok := oneRuneAccentsWheel[baseVowel]
			if !ok || len(wheel) == 0 {
				continue // Skip if no wheel for this vowel
			}

			// Verify the vowel is in our wheel (skip unknown accents)
			if currentAccentIndex >= len(wheel) {
				continue // Skip if accent index out of range
			}
			expectedForm := wheel[currentAccentIndex]
			if expectedForm != vowelStr {
				// Accent not in our wheel, skip transformation
				continue
			}

			// Move backward by power positions (with wrapping)
			newAccentIndex := (currentAccentIndex - power) % len(wheel)
			if newAccentIndex < 0 {
				newAccentIndex += len(wheel)
			}

			// Get new accented form (always single rune from our wheel)
			newAccentedForm := wheel[newAccentIndex]
			newAccentRunes := []rune(newAccentedForm) // Only apply if:
			// 1. We got a single rune
			// 2. The new form is different from base vowel (otherwise accent info is lost)
			if len(newAccentRunes) == 1 {
				// Check if this accent actually changes the vowel
				baseVowelStr := string(baseVowel)
				if newAccentedForm != baseVowelStr || newAccentIndex == 0 {
					if isUpper {
						// Apply uppercase - but only if case conversion is reversible
						upperForm := unicode.ToUpper(newAccentRunes[0])
						if unicode.ToLower(upperForm) == newAccentRunes[0] {
							result[pos] = upperForm
						} else {
							// Case conversion not reversible, keep lowercase
							result[pos] = newAccentRunes[0]
						}
					} else {
						result[pos] = newAccentRunes[0]
					}
				}
			}
		}
	}

	return string(result)
}

// generateFibonacci generates Fibonacci sequence up to maxIndex
func generateFibonacci(maxIndex int) []int {
	if maxIndex < 1 {
		return []int{}
	}

	fib := []int{1, 2} // Start with 1, 2 (1-indexed positions)
	for {
		next := fib[len(fib)-1] + fib[len(fib)-2]
		if next > maxIndex {
			break
		}
		fib = append(fib, next)
	}
	return fib
}

// generateTribonacci generates Tribonacci sequence up to maxIndex
func generateTribonacci(maxIndex int) []int {
	if maxIndex < 1 {
		return []int{}
	}

	trib := []int{1, 2, 4} // Start with 1, 2, 4 (1-indexed positions)
	for {
		next := trib[len(trib)-1] + trib[len(trib)-2] + trib[len(trib)-3]
		if next > maxIndex {
			break
		}
		trib = append(trib, next)
	}
	return trib
}

// applyCaseReplacementLogic inverts capitalization at Fibonacci/Tribonacci positions
// If word count is odd, use Fibonacci sequence; if even, use Tribonacci sequence
// Applying twice returns to original (reversible)
func applyCaseReplacementLogic(input string) string {
	if !utf8.ValidString(input) {
		return input
	}

	// Count words (sequences of letters/digits separated by non-letter/non-digit characters)
	runes := []rune(input)
	wordCount := 0
	inWord := false

	for _, r := range runes {
		isLetterOrDigit := unicode.IsLetter(r) || unicode.IsDigit(r)
		if isLetterOrDigit && !inWord {
			wordCount++
			inWord = true
		} else if !isLetterOrDigit {
			inWord = false
		}
	}

	if wordCount == 0 {
		return input // No words, nothing to do
	}

	// Choose sequence based on word count parity
	var sequence []int
	if wordCount%2 == 1 {
		// Odd: use Fibonacci
		sequence = generateFibonacci(len(runes))
	} else {
		// Even: use Tribonacci
		sequence = generateTribonacci(len(runes))
	}

	// Create a set of positions to invert (1-indexed in sequence, convert to 0-indexed)
	positionsToInvert := make(map[int]bool)
	for _, pos := range sequence {
		if pos > 0 && pos <= len(runes) {
			positionsToInvert[pos-1] = true // Convert to 0-indexed
		}
	}

	// Apply case inversion at specified positions
	result := make([]rune, len(runes))
	copy(result, runes)

	for i := range result {
		if positionsToInvert[i] {
			if unicode.IsUpper(result[i]) {
				lower := unicode.ToLower(result[i])
				// Check reversibility: if converting back doesn't give original, skip
				if unicode.ToUpper(lower) != result[i] {
					continue
				}
				result[i] = lower
			} else if unicode.IsLower(result[i]) {
				upper := unicode.ToUpper(result[i])
				// Check reversibility: if converting back doesn't give original, skip
				if unicode.ToLower(upper) != result[i] {
					continue
				}
				result[i] = upper
			}
		}
	}

	return string(result)
}

// TranslateToPejelagarto translates Human text to Pejelagarto
func TranslateToPejelagarto(input string) string {
	input = applyNumbersLogicToPejelagarto(input)
	input = applyMapReplacementsToPejelagarto(input)
	input = applyAccentReplacementLogicToPejelagarto(input)
	input = applyCaseReplacementLogic(input)
	return input
}

// TranslateFromPejelagarto translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
	input = applyCaseReplacementLogic(input)
	input = applyAccentReplacementLogicFromPejelagarto(input)
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
