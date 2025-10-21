// Pejelagarto Translator
// Build command: go build -o pejelagarto-translator.exe main.go
// Run command: .\pejelagarto-translator.exe

package main

import (
	"fmt"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
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

// Punctuation replacement map
// NOTE: This map is independent from word/conjunction/letter maps
// Can have different lengths for keys and values
var punctuationMap = map[string]string{
	"?":  "‽",
	"!":  "¡",
	".":  "..",
	",":  "،",
	";":  "⁏",
	":":  "︰",
	"'":  "〝",
	"\"": "〞",
	"-":  "‐",
	"(":  "⦅",
	")":  "⦆",
}

// Emoji encoding maps for datetime
var dayEmojiIndex = []string{
	"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘", "🌙", "🌚",
	"🌛", "🌜", "🌝", "🌞", "⭐", "🌟", "✨", "💫", "🌠", "☀️",
	"🌤️", "⛅", "🌥️", "☁️", "🌦️", "🌧️", "⛈️", "🌩️", "🌨️", "❄️",
	"☃️",
}

var monthEmojiIndex = []string{
	"🍇", "🍈", "🍉", "🍊", "🍋", "🍌", "🍍", "🥭", "🍎", "🍏",
	"🍐", "🍑",
}

var yearEmojiIndex = []string{
	"🐀", "🐁", "🐂", "🐃", "🐄", "🐅", "🐆", "🐇", "🐈", "🐉",
	"🐊", "🐋", "🐌", "🐍", "🐎", "🐏", "🐐", "🐑", "🐒", "🐓",
	"🐔", "🐕", "🐖", "🐗", "🐘", "🐙", "🐚", "🐛", "🐜", "🐝",
	"🐞", "🐟", "🐠", "🐡", "🐢", "🐣", "🐤", "🐥", "🐦", "🐧",
	"🐨", "🐩", "🐪", "🐫", "🐬", "🐭", "🐮", "🐯", "🐰", "🐱",
	"🐲", "🐳", "🐴", "🐵", "🐶", "🐷", "🐸", "🐹", "🐺", "🐻",
	"🐼", "🐽", "🐾", "🐿️", "👀", "👁️", "👂", "👃", "👄", "👅",
	"👆", "👇", "👈", "👉", "👊", "👋", "👌", "👍", "👎", "👏",
	"👐", "👑", "👒", "👓", "👔", "👕", "👖", "👗", "👘", "👙",
	"👚", "👛", "👜", "👝", "👞", "👟", "👠", "👡", "👢",
}

var hourEmojiIndex = []string{
	"🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙",
	"🕚", "🕛", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣",
	"🕤", "🕥", "🕦", "🕧",
}

var minuteEmojiIndex = []string{
	"⏰", "⏱️", "⏲️", "⏳", "⌚", "⌛", "⏰", "🔔", "🔕", "📅",
	"📆", "📇", "📈", "📉", "📊", "📋", "📌", "📍", "📎", "📏",
	"📐", "📑", "📒", "📓", "📔", "📕", "📖", "📗", "📘", "📙",
	"📚", "📛", "📜", "📝", "📞", "📟", "📠", "📡", "📢", "📣",
	"📤", "📥", "📦", "📧", "📨", "📩", "📪", "📫", "📬", "📭",
	"📮", "📯", "📰", "📱", "📲", "📳", "📴", "📵", "📶", "📷",
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
			// remove it from the output value so it doesn't appear in Human text
			outputValue := value
			if strings.HasPrefix(key, "'") {
				// This is a Pejelagarto pattern being converted back to Human
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

// createPunctuationBijectiveMap creates a unified bijective map for punctuation replacements
func createPunctuationBijectiveMap() map[int32]map[string]string {
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
				valueLen := utf8.RuneCountInString(value)
				if valueLen > 1 {
					// Multi-rune: the actual Pejelagarto text has a quote prefix
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

			// Add to map
			if bijectiveMap[index] == nil {
				bijectiveMap[index] = make(map[string]string)
			}
			bijectiveMap[index][from] = to
		}
	}

	// Add forward mappings (positive indices)
	addEntries(punctuationMap, true)
	// Add reverse mappings (negative indices)
	addEntries(punctuationMap, false)

	return bijectiveMap
}

// getSortedPunctuationIndices returns indices sorted for the direction
func getSortedPunctuationIndices(bijectiveMap map[int32]map[string]string, toPejelagarto bool) []int32 {
	indices := make([]int32, 0, len(bijectiveMap))
	for index := range bijectiveMap {
		indices = append(indices, index)
	}

	sort.Slice(indices, func(i, j int) bool {
		iPos := indices[i] > 0
		jPos := indices[j] > 0

		if toPejelagarto {
			// ToPejelagarto: positive first, then descending by absolute value
			if iPos != jPos {
				return iPos
			}
		} else {
			// FromPejelagarto: negative first, then descending by absolute value
			if iPos != jPos {
				return !iPos
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
}

// applyPunctuationReplacementsToPejelagarto applies punctuation replacements
func applyPunctuationReplacementsToPejelagarto(input string) string {
	if !utf8.ValidString(input) {
		return input
	}

	// Use a special marker for literal quotes to avoid ambiguity
	const quoteMarker = "\uFFF3"
	input = strings.ReplaceAll(input, "'", quoteMarker)

	bijectiveMap := createPunctuationBijectiveMap()
	indices := getSortedPunctuationIndices(bijectiveMap, true)
	result := applyReplacements(input, bijectiveMap, indices)

	// Restore literal quotes as doubled quotes in the output
	result = strings.ReplaceAll(result, quoteMarker, "''")

	return result
}

// applyPunctuationReplacementsFromPejelagarto reverses punctuation replacements
func applyPunctuationReplacementsFromPejelagarto(input string) string {
	if !utf8.ValidString(input) {
		return input
	}

	// Convert doubled quotes (escaped literals) to temporary marker
	const quoteMarker = "\uFFF3"
	input = strings.ReplaceAll(input, "''", quoteMarker)

	bijectiveMap := createPunctuationBijectiveMap()
	indices := getSortedPunctuationIndices(bijectiveMap, false)
	result := applyReplacements(input, bijectiveMap, indices)

	// Restore literal quotes from marker
	result = strings.ReplaceAll(result, quoteMarker, "'")

	return result
}

// removeISO8601timestamp removes ISO 8601 timestamp from the last line if present
func removeISO8601timestamp(input string) (string, string) {
	// ISO 8601 regex pattern for timestamps like 2025-10-19T14:30:00Z or 2025-10-19T14:30:00+00:00
	iso8601Pattern := regexp.MustCompile(`(?m)^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:\d{2})$`)

	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return input, ""
	}

	// Check if last line matches ISO 8601 timestamp
	lastLine := lines[len(lines)-1]
	if iso8601Pattern.MatchString(lastLine) {
		// Remove the last line and return the timestamp
		if len(lines) == 1 {
			return "", lastLine
		}
		return strings.Join(lines[:len(lines)-1], "\n"), lastLine
	}

	return input, ""
}

// addISO8601timestamp adds timestamp as new last line to input
func addISO8601timestamp(input string, timestamp string) string {
	if timestamp == "" {
		return input
	}

	if input == "" {
		return timestamp
	}

	return input + "\n" + timestamp
}

// removeAllEmojies removes all emoji characters from the input
func removeAllEmojies(input string) string {
	// Create a set of all emojis from our maps
	emojiSet := make(map[string]bool)
	for _, emoji := range dayEmojiIndex {
		emojiSet[emoji] = true
	}
	for _, emoji := range monthEmojiIndex {
		emojiSet[emoji] = true
	}
	for _, emoji := range yearEmojiIndex {
		emojiSet[emoji] = true
	}
	for _, emoji := range hourEmojiIndex {
		emojiSet[emoji] = true
	}
	for _, emoji := range minuteEmojiIndex {
		emojiSet[emoji] = true
	}

	result := input
	for emoji := range emojiSet {
		result = strings.ReplaceAll(result, emoji, "")
	}

	return result
}

// readTimestampUsingEmojiEncoding locates emojis and returns ISO 8601 timestamp
func readTimestampUsingEmojiEncoding(input string) string {
	// Find one emoji from each category
	var day, month, year, hour, minute int = -1, -1, -1, 0, 0

	// Search for emojis in the input
	for i, emoji := range dayEmojiIndex {
		if strings.Contains(input, emoji) {
			day = i + 1 // days are 1-indexed
			break
		}
	}

	for i, emoji := range monthEmojiIndex {
		if strings.Contains(input, emoji) {
			month = i + 1 // months are 1-indexed
			break
		}
	}

	for i, emoji := range yearEmojiIndex {
		if strings.Contains(input, emoji) {
			year = 2025 + i // years start from 2025
			break
		}
	}

	for i, emoji := range hourEmojiIndex {
		if strings.Contains(input, emoji) {
			hour = i
			break
		}
	}

	for i, emoji := range minuteEmojiIndex {
		if strings.Contains(input, emoji) {
			minute = i
			break
		}
	}

	// Check if we found required components (day, month, year)
	// Hour and minute are optional and default to 0
	if day == -1 || month == -1 || year == -1 {
		return "" // Cannot determine datetime
	}

	// Create timestamp in ISO 8601 format
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:00Z", year, month, day, hour, minute)
}

// addEmojiDatetimeEncoding inserts datetime emojis at random positions
func addEmojiDatetimeEncoding(input string, timestamp string) string {
	// Use provided timestamp or current UTC datetime
	var now time.Time
	if timestamp == "" {
		now = time.Now().UTC()
	} else {
		// Parse the ISO 8601 timestamp
		parsedTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			// If parsing fails, use current time
			now = time.Now().UTC()
		} else {
			now = parsedTime.UTC()
		}
	}

	// Get emoji indices
	day := now.Day() - 1          // Convert to 0-indexed
	month := int(now.Month()) - 1 // Convert to 0-indexed
	year := now.Year() - 2025     // Years start from 2025
	hour := now.Hour()
	minute := now.Minute()

	// Validate indices
	if day < 0 || day >= len(dayEmojiIndex) {
		day = 0
	}
	if month < 0 || month >= len(monthEmojiIndex) {
		month = 0
	}
	if year < 0 || year >= len(yearEmojiIndex) {
		year = 0
	}
	if hour < 0 || hour >= len(hourEmojiIndex) {
		hour = 0
	}
	if minute < 0 || minute >= len(minuteEmojiIndex) {
		minute = 0
	}

	// Get the emojis
	emojis := []string{
		dayEmojiIndex[day],
		monthEmojiIndex[month],
		yearEmojiIndex[year],
		hourEmojiIndex[hour],
		minuteEmojiIndex[minute],
	}

	// Find all positions next to spaces or line breaks
	runes := []rune(input)
	var positions []int

	for i := 0; i < len(runes); i++ {
		if i == 0 || runes[i] == ' ' || runes[i] == '\n' {
			positions = append(positions, i)
		}
		if i == len(runes)-1 {
			positions = append(positions, i+1)
		}
	}

	// If no positions found, just append to the end
	if len(positions) == 0 {
		for _, emoji := range emojis {
			input += emoji
		}
		return input
	}

	// Shuffle positions and pick the first 5
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Insert emojis at random positions
	// Sort positions to insert from end to beginning (to maintain correct indices)
	if len(positions) > len(emojis) {
		positions = positions[:len(emojis)]
	}
	sort.Ints(positions)

	// Work with runes to avoid splitting UTF-8 sequences
	resultRunes := runes
	for i := len(positions) - 1; i >= 0 && i < len(emojis); i-- {
		pos := positions[i]
		if pos > len(resultRunes) {
			pos = len(resultRunes)
		}
		// Convert emoji to runes and insert
		emojiRunes := []rune(emojis[i])
		resultRunes = append(resultRunes[:pos], append(emojiRunes, resultRunes[pos:]...)...)
	}

	return string(resultRunes)
}

// sanitizeInvalidUTF8 replaces invalid UTF-8 bytes with soft hyphens + Private Use Area characters
// Uses a bijective mapping to maintain reversibility
// Soft hyphens (U+00AD) are invisible in most contexts, making the output cleaner
func sanitizeInvalidUTF8(input string) string {
	// Use Private Use Area characters - they won't be affected by any translation logic
	// Map each of 256 possible bytes to a unique character in range U+E000-U+E0FF
	const privateUseStart = 0xE000
	const softHyphen = '\u00AD' // Soft hyphen - invisible in most contexts

	var result strings.Builder
	result.Grow(len(input) * 2) // Reserve extra space

	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 byte - encode it invisibly using soft hyphen + private use character
			invalidByte := input[i]
			result.WriteRune(softHyphen) // Invisible marker
			result.WriteRune(rune(privateUseStart + int(invalidByte)))
			i++
		} else {
			result.WriteRune(r)
			i += size
		}
	}

	return result.String()
}

// unsanitizeInvalidUTF8 is the reverse of sanitizeInvalidUTF8
func unsanitizeInvalidUTF8(input string) string {
	const privateUseStart = 0xE000
	const softHyphen = '\u00AD'

	var result []byte
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == softHyphen && i+1 < len(runes) {
			// Check if next character is in our private use range
			nextRune := runes[i+1]
			if nextRune >= privateUseStart && nextRune < privateUseStart+256 {
				// This is an encoded invalid byte
				byteVal := byte(nextRune - privateUseStart)
				result = append(result, byteVal)
				i++ // Skip the next rune as we've processed it
				continue
			}
		}
		// Normal character
		result = append(result, []byte(string(r))...)
	}

	return string(result)
}

// TranslateToPejelagarto translates Human text to Pejelagarto
func TranslateToPejelagarto(input string) string {
	input = sanitizeInvalidUTF8(input)
	input = removeAllEmojies(input)
	input, timestamp := removeISO8601timestamp(input)
	input = applyNumbersLogicToPejelagarto(input)
	input = applyPunctuationReplacementsToPejelagarto(input)
	input = applyMapReplacementsToPejelagarto(input)
	input = applyAccentReplacementLogicToPejelagarto(input)
	input = applyCaseReplacementLogic(input)
	input = addEmojiDatetimeEncoding(input, timestamp)
	return input
}

// TranslateFromPejelagarto translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
	timestamp := readTimestampUsingEmojiEncoding(input)
	input = removeAllEmojies(input)
	input = applyCaseReplacementLogic(input)
	input = applyAccentReplacementLogicFromPejelagarto(input)
	input = applyMapReplacementsFromPejelagarto(input)
	input = applyPunctuationReplacementsFromPejelagarto(input)
	input = applyNumbersLogicFromPejelagarto(input)
	input = addISO8601timestamp(input, timestamp)
	input = unsanitizeInvalidUTF8(input)
	return input
}

// HTML UI template
const htmlUI = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pejelagarto Translator</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 900px;
            width: 100%;
        }
        
        h1 {
            text-align: center;
            color: #667eea;
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
        }
        
        .translator-box {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .text-area-container {
            display: flex;
            flex-direction: column;
        }
        
        label {
            font-weight: bold;
            margin-bottom: 8px;
            color: #333;
            font-size: 1.1em;
        }
        
        textarea {
            width: 100%;
            height: 250px;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            font-size: 14px;
            font-family: 'Courier New', monospace;
            resize: vertical;
            transition: border-color 0.3s;
        }
        
        textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        
        textarea[readonly] {
            background-color: #f5f5f5;
            cursor: not-allowed;
        }
        
        .controls {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 15px;
            flex-wrap: wrap;
        }
        
        button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
        }
        
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
        }
        
        button:active {
            transform: translateY(0);
        }
        
        .invert-btn {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            padding: 12px 20px;
            font-size: 20px;
            box-shadow: 0 4px 15px rgba(245, 87, 108, 0.4);
        }
        
        .invert-btn:hover {
            box-shadow: 0 6px 20px rgba(245, 87, 108, 0.6);
        }
        
        .checkbox-container {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 16px;
            color: #333;
        }
        
        input[type="checkbox"] {
            width: 20px;
            height: 20px;
            cursor: pointer;
        }
        
        .hidden {
            display: none !important;
        }
        
        @media (max-width: 768px) {
            .translator-box {
                grid-template-columns: 1fr;
            }
            
            h1 {
                font-size: 2em;
            }
            
            .container {
                padding: 20px;
            }
        }
        
        .htmx-indicator {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid #f3f3f3;
            border-top: 3px solid #667eea;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-left: 10px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🐊 Pejelagarto Translator 🐊</h1>
        
        <div class="translator-box">
            <div class="text-area-container">
                <label id="input-label">Human:</label>
                <textarea id="input-text" placeholder="Type your text here..."></textarea>
            </div>
            
            <div class="text-area-container">
                <label id="output-label">Pejelagarto:</label>
                <textarea id="output-text" readonly placeholder="Translation will appear here..."></textarea>
            </div>
        </div>
        
        <div class="controls">
            <button 
                id="translate-btn"
                onclick="handleTranslateClick()">
                Translate to Pejelagarto
            </button>
            
            <button class="invert-btn" onclick="invertTranslation()">⇅</button>
            
            <div class="checkbox-container">
                <input type="checkbox" id="live-translate" onchange="toggleLiveTranslation()">
                <label for="live-translate" style="margin: 0;">Live Translation</label>
            </div>
            
            <span id="loading-indicator" class="htmx-indicator"></span>
        </div>
    </div>
    
    <script>
        let isInverted = false;
        let liveTranslateEnabled = false;
        
        // Handle translate button click
        function handleTranslateClick() {
            handleLiveTranslation();
        }
        
        // Invert button functionality
        function invertTranslation() {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const inputLabel = document.getElementById('input-label');
            const outputLabel = document.getElementById('output-label');
            const translateBtn = document.getElementById('translate-btn');
            
            // Swap text content
            const temp = inputText.value;
            inputText.value = outputText.value;
            outputText.value = temp;
            
            // Swap labels
            const tempLabel = inputLabel.textContent;
            inputLabel.textContent = outputLabel.textContent;
            outputLabel.textContent = tempLabel;
            
            // Toggle button state
            isInverted = !isInverted;
            if (isInverted) {
                translateBtn.textContent = 'Translate from Pejelagarto';
            } else {
                translateBtn.textContent = 'Translate to Pejelagarto';
            }
        }
        
        // Live translation functionality
        function toggleLiveTranslation() {
            const checkbox = document.getElementById('live-translate');
            const translateBtn = document.getElementById('translate-btn');
            const inputText = document.getElementById('input-text');
            
            liveTranslateEnabled = checkbox.checked;
            
            if (liveTranslateEnabled) {
                // Hide the translate button
                translateBtn.classList.add('hidden');
                
                // Add live translation event listener
                inputText.addEventListener('input', handleLiveTranslation);
                
                // Trigger initial translation
                handleLiveTranslation();
            } else {
                // Show the translate button
                translateBtn.classList.remove('hidden');
                
                // Remove live translation event listener
                inputText.removeEventListener('input', handleLiveTranslation);
            }
        }
        
        function handleLiveTranslation() {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const endpoint = isInverted ? '/from' : '/to';
            
            // Send request to backend
            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain'
                },
                body: inputText.value
            })
            .then(response => response.text())
            .then(data => {
                outputText.value = data;
            })
            .catch(error => {
                console.error('Translation error:', error);
            });
        }
    </script>
</body>
</html>`

// HTTP handler for the main UI
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlUI)
}

// HTTP handler for translating to Pejelagarto
func handleTranslateTo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	input := string(body)
	result := TranslateToPejelagarto(input)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, result)
}

// HTTP handler for translating from Pejelagarto
func handleTranslateFrom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	input := string(body)
	result := TranslateFromPejelagarto(input)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, result)
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

func main() {
	// Set up HTTP routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/to", handleTranslateTo)
	http.HandleFunc("/from", handleTranslateFrom)

	// Server address
	addr := ":8080"
	url := "http://localhost:8080"

	// Start server in goroutine
	go func() {
		log.Printf("Starting Pejelagarto Translator server on %s\n", url)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait a moment for server to start, then open browser
	time.Sleep(500 * time.Millisecond)
	if err := openBrowser(url); err != nil {
		log.Printf("Could not open browser automatically: %v\n", err)
		log.Printf("Please open your browser and navigate to %s\n", url)
	}

	// Keep the server running
	log.Println("Server is running. Press Ctrl+C to stop.")
	select {} // Block forever
}
