package main

import (
	"embed"
	"fmt"
	"math/big"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

//go:embed get-requirements.ps1 get-requirements.sh
var embeddedGetRequirements embed.FS

// tempRequirementsDir stores the path to extracted requirements
var tempRequirementsDir string

// Translation maps for conjunction (letter pair) replacements
// NOTE: All values must have SAME length as keys (rune count)
// NOTE: Output values use ONLY letters NOT in letterMap (c,h,j,s,t,x,z) to avoid collisions
// NOTE: Avoid repeated characters to prevent ambiguity (e.g., "zz" could be confused with "z"+"z")
var conjunctionMap = map[string]string{
	"hello": "araka",
	"hola":  "arak",
	"fran":  "filo",
	"the":   "ele",
	"el":    "le",
	"la":    "al",
	"leg":   "ady",
	"ch":    "jc",
	"sh":    "xs",
	"th":    "zt",
}

// Translation maps for single letter replacements (must be invertible)
// NOTE: All values must have SAME length as keys (rune count)
// NOTE: Avoid letters that appear in conjunction patterns (c, h, j, s, t, x, z)
// to prevent collisions between letter outputs and conjunction inputs
// NOTE: Consonants map to consonants, vowels map to vowels (y and w are vowels)
// NOTE: Each letter must map to another letter that maps back to it (true bijective pairs)
var letterMap = map[string]string{
	"a": "u",
	"b": "p",
	"d": "f",
	"e": "w",
	"f": "d",
	"g": "l",
	"i": "o",
	"k": "r",
	"l": "g",
	"m": "n",
	"n": "m",
	"o": "i",
	"p": "b",
	"q": "v",
	"r": "k",
	"u": "a",
	"v": "q",
	"w": "e",
	"y": "y",
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
	"\"": "〞",
	"-":  "‐",
	"(":  "⦅",
	")":  "⦆",
}

// Escape characters for internal and output escaping
const (
	internalEscapeChar = '\\'     // Backslash - used internally, removed before output
	outputEscapeChar   = '\u00AD' // Soft hyphen - used in output, visible in Pejelagarto text
)

// Special character encoding maps for datetime using Unicode range U+2300 to U+23FB (avoiding emojis)
var daySpecialCharIndex = []string{
	"\u2300", "\u2301", "\u24FC", "\u2303", "\u2304", "\u2305", "\u2306", "\u2307", "\u2308", "\u2309",
	"\u230A", "\u230B", "\u230C", "\u230D", "\u230E", "\u230F", "\u2310", "\u2311", "\u2312", "\u2313",
	"\u2314", "\u2315", "\u2316", "\u2317", "\u2318", "\u2319", "\u24EA", "\u24EB", "\u231C", "\u231D",
	"\u231E",
}

var monthSpecialCharIndex = []string{
	"\u233C", "\u233D", "\u233E", "\u233F", "\u2340", //"\u2341", "\u2342", "\u2343", "\u2344", "\u00A0",
	"\uA4F8", "\uA4F9", "\uA4FA", "\uA4FB", "\uA4FC",
	"\u2B4E", "\u2B4F",
}

var yearSpecialCharIndex = []string{
	/*
		"\u00AD", "\u034F", "\u061C", "\u070F", "\u115F", "\u1160", "\u17B4", "\u17B5", "\u180B", "\u180C",
		"\u180D", "\u180E", "\u200B", "\u200C", "\u200D", "\u200E", "\u200F", "\u202A", "\u202B", "\u202C",
		"\u202D", "\u202E", "\u2060", "\u2061", "\u2062", "\u2063", "\u2064", "\u2065", "\u2066", "\u2067",
		"\u2068", "\u2069", "\u206A", "\u206B", "\u206C", "\u206D", "\u206E", "\u206F", "\u3164", "\uFE00",
		"\uFE01", "\uFE02", "\uFE03", "\uFE04", "\uFE05", "\uFE06", "\uFE07", "\uFE08", "\uFE09", "\uFE0A",
		"\uFE0B", "\uFE0C", "\uFE0D", "\uFE0E", "\uFE0F", "\uFEFF", "\uFFF9", "\uFFFA", "\uFFFB", "\u2383",
		"\u2384", "\u2385", "\u2386", "\u2387", "\u2388", "\u2389", "\u238A", "\u238B", "\u238C", "\u238D",
		"\u238E", "\u238F", "\u2390", "\u2391", "\u2392", "\u2393", "\u2394", "\u2395", "\u2396", "\u2397",
		"\u2398", "\u2399", "\u239A", "\u239B", "\u239C", "\u239D", "\u239E", "\u239F", "\u23A0", "\u23A1",
	*/
	/*
		"\uE0020", "\uE0021", "\uE0022", "\uE0023", "\uE0024", "\uE0025", "\uE0026", "\uE0027", "\uE0028", "\uE0029",
		"\uE002A", "\uE002B", "\uE002C", "\uE002D", "\uE002E", "\uE002F", "\uE0030", "\uE0031", "\uE0032", "\uE0033",
		"\uE0034", "\uE0035", "\uE0036", "\uE0037", "\uE0038", "\uE0039", "\uE003A", "\uE003B", "\uE003C", "\uE003D",
		"\uE003E", "\uE003F", "\uE0040", "\uE0041", "\uE0042", "\uE0043", "\uE0044", "\uE0045", "\uE0046", "\uE0047",
		"\uE0048", "\uE0049", "\uE004A", "\uE004B", "\uE004C", "\uE004D", "\uE004E", "\uE004F", "\uE0050", "\uE0051",
		"\uE0052", "\uE0053", "\uE0054", "\uE0055", "\uE0056", "\uE0057", "\uE0058", "\uE0059", "\uE005A", "\uE005B",
		"\uE005C", "\uE005D", "\uE005E", "\uE005F", "\uE0060", "\uE0061", "\uE0062", "\uE0063", "\uE0064", "\uE0065",
		"\uE0066", "\uE0067", "\uE0068", "\uE0069", "\uE006A", "\uE006B", "\uE006C", "\uE006D", "\uE006E", "\uE006F",
		"\uE0070", "\uE0071", "\uE0072", "\uE0073", "\uE0074", "\uE0075", "\uE0076", "\uE0077", "\uE0078", "\uE0079",
	*/
	"\uFE70", "\uFE71", "\uFE72", "\uFE73", "\uFE74", "\uFE75", "\uFE76", "\uFE77", "\uFE78", "\uFE79",
	"\uFE7A", "\uFE7B", "\uFE7C", "\uFE7D", "\uFE7E", "\uFC5E", "\uFC5F", "\uFC60", "\uFC61", "\uFC62",
	"\uFC63", "\uFBB2", "\uFBB3", "\uFBB4", "\uFBB5", "\uFBB6", "\uFBB7", "\uFBB8", "\uFBB9", "\uFBBA",
	"\uFBBB", "\uFBBC", "\uFBBD", "\uFBBE", "\uFBBF", "\uFBC0", "\uFBC1", "\uFBC2", "\uFBC3", "\uFBC4",
	"\uFBC5", "\uFBC6", "\uFBC7", "\uFBC8", "\uFBC9", "\uFBCA", "\uFBCB", "\uFBCC", "\uFBCD", "\uFBCE",
	"\uFBCF", "\uFBD0", "\uFBD1", "\uFBD2", "\uA674", "\uA675", "\uA676", "\uA677", "\uA678", "\uA679",
	"\uA67A", "\uA67B", "\uA67C", "\uA67D", "\uA67E", "\uA67F", "\u3192", "\u3193", "\u3194", "\u3195",
	"\u3196", "\u3197", "\u3198", "\u3199", "\u319A", "\u319B", "\u319C", "\u319D", "\u319E", "\u319F",
	"\u2E2F", "\u2E30", "\u2E31", "\u2E32", "\u2E33", "\u2E34", "\u2E35", "\u2E44", "\u2E49", "\u2E4E",
	"\u23A2", "\u23A3", "\u23A4", "\u23A5", "\u23A6", "\u23A7", "\u23A8", "\u23A9", "\u2DE0", "\u23A0",
}

var hourSpecialCharIndex = []string{
	"\u23AA", "\u23AB", "\u23AC", "\u23AD", "\u23AE", "\u23AF", "\u23B0", "\u23B1", "\u23B2", "\u23B3",
	"\u23B4", "\u23B5", "\u23B6", "\u23B7", "\u23B8", "\u23B9", "\u23BA", "\u23BB", "\u23BC", "\u23BD",
	"\u23BE", "\u0F0B", "\u23C0", "\u02B9",
}

var minuteSpecialCharIndex = []string{
	/*
		"\u23C1", "\u23C2", "\u23C3", "\u23C4", "\u23C5", "\u23C6", "\u23C7", "\u23C8", "\u23C9", "\u23CA",
		"\u23CB", "\u23CC", "\u23CD", "\u23CE", "\u24FD", "\u23D0", "\u23D1", "\u23D2", "\u23D3", "\u23D4",
		"\u23D5", "\u23D6", "\u23D7", "\u23D8", "\u23D9", "\u23DA", "\u23DB", "\u23DC", "\u23DD", "\u23DE",
		"\u23DF", "\u23E0", "\u23E1", "\u23E2", "\u23E3", "\u23E4", "\u1085", "\u1058", "\u1071", "\u141D",
	*/
	"\u2DE1", "\u2DE2", "\u2DE3", "\u2DE4", "\u2DE5", "\u2DE6", "\u2DE7", "\u2DE8", "\u2DE9", "\u2DEA",
	"\u2DEB", "\u2DEC", "\u2DED", "\u2DEE", "\u2DEF", "\u2DF0", "\u2DF1", "\u2DF2", "\u2DF3", "\u2DF4",
	"\u2DF5", "\u2DF6", "\u2DF7", "\u2DF8", "\u2DF9", "\u2DFA", "\u2DFB", "\u2DFC", "\u2DFD", "\u2DFE",
	"\u2DFF", "\u2E00", "\u2E01", "\u2E02", "\u2E03", "\u2E04", "\u2E05", "\u2E06", "\u2E07", "\u2E08",
	/*
		"\u24EC", "\u24ED", "\u24EE", "\u24EF", "\u24F0", "\u24F1", "\u24F2", "\u24F3", "\u24F4", "\u24F5",
		"\u24F6", "\u23F4", "\u23F5", "\u23F6", "\u23F7", "\u24F7", "\u24F8", "\u24F9", "\u24FA",
	*/
	"\u2427", "\u2428", "\u2429", "\u302A", "\u302B", "\u2FFC", "\u2FFD", "\u2FFE", "\u2FFF", "\u3099",
	"\u309A", "\u309B", "\u309C", "\uA702", "\uAAB8", "\u061C", "\uA950", "\uA951", "\uA926", "\uA952",
}

// validateMaps checks that all mappings have equal rune lengths for keys and values
func validateMaps() error {
	maps := []struct {
		name string
		m    map[string]string
	}{
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
	addEntries(conjunctionMap, true)
	addEntries(letterMap, true)

	// Add inverse entries (-index: value -> key)
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

// internalEscape escapes characters using backslash prefix (internal use only, removed before output)
// This is used to protect characters during translation processing
// Escapes the escape character itself first to avoid conflicts
func internalEscape(input string, charsToEscape string) string {
	var result strings.Builder
	result.Grow(len(input) * 2)

	// Build a map for quick lookup
	escapeMap := make(map[rune]bool)
	for _, r := range charsToEscape {
		escapeMap[r] = true
	}
	// Always escape the escape character itself
	escapeMap[internalEscapeChar] = true

	for _, r := range input {
		if escapeMap[r] {
			result.WriteRune(internalEscapeChar)
		}
		result.WriteRune(r)
	}

	return result.String()
}

// internalUnescape removes backslash escaping (reverses internalEscape)
func internalUnescape(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		if runes[i] == internalEscapeChar && i+1 < len(runes) {
			// Skip the escape character, add the next character
			i++
			result.WriteRune(runes[i])
		} else {
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

// outputEscape escapes characters using soft hyphen prefix (present in Pejelagarto output)
// This escaping is visible in the translated text
func outputEscape(input string, charsToEscape string) string {
	var result strings.Builder
	result.Grow(len(input) * 2)

	// Build a map for quick lookup
	escapeMap := make(map[rune]bool)
	for _, r := range charsToEscape {
		escapeMap[r] = true
	}
	// Always escape the escape character itself
	escapeMap[outputEscapeChar] = true

	for _, r := range input {
		if escapeMap[r] {
			result.WriteRune(outputEscapeChar)
		}
		result.WriteRune(r)
	}

	return result.String()
}

// outputUnescape removes soft hyphen escaping (reverses outputEscape)
func outputUnescape(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		if runes[i] == outputEscapeChar && i+1 < len(runes) {
			// Skip the escape character, add the next character
			i++
			result.WriteRune(runes[i])
		} else {
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

// applyReplacements applies replacements from the bijective map in the specified order
func applyReplacements(input string, bijectiveMap map[int32]map[string]string, indices []int32) string {
	// Use special Unicode characters as markers that won't be in normal text
	const startMarker = "\uFFF0"
	const endMarker = "\uFFF1"

	// Escape any markers that appear in the input to preserve them
	// Convert markers to runes for proper escaping
	startMarkerRune := []rune(startMarker)[0]
	endMarkerRune := []rune(endMarker)[0]
	result := internalEscape(input, string([]rune{startMarkerRune, endMarkerRune}))

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

			// Pre-calculate marker positions and escape positions for O(n) performance
			markerMap := make(map[int]int)   // pos -> marker depth at that position
			escapedMap := make(map[int]bool) // pos -> is this position escaped
			depth := 0
			startMarkerRune := []rune(startMarker)[0]
			endMarkerRune := []rune(endMarker)[0]

			// First pass: identify escaped characters
			// We need to check for BOTH backslash escapes (internal) and soft hyphen escapes (output)
			for i := 0; i < len(resultRunes); i++ {
				if (resultRunes[i] == internalEscapeChar || resultRunes[i] == outputEscapeChar) && i+1 < len(resultRunes) {
					escapedMap[i+1] = true
					escapedMap[i] = true // Mark the escape character itself as escaped too
				}
			}

			// Second pass: calculate marker depth, skipping escaped characters
			for i := 0; i < len(resultRunes); i++ {
				markerMap[i] = depth

				// Only count as marker if not escaped
				if !escapedMap[i] {
					if resultRunes[i] == startMarkerRune {
						depth++
					} else if resultRunes[i] == endMarkerRune {
						depth--
					}
				}
			}

			for pos < len(resultRunes) {
				// Check if current character is escaped
				if escapedMap[pos] {
					// This character is escaped, just copy it
					newResult.WriteRune(resultRunes[pos])
					pos++
					continue
				}

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

	// Remove all working markers, but NOT escaped ones
	// We need to manually iterate to skip escaped markers
	var cleanResult strings.Builder
	resultRunes := []rune(result)
	for i := 0; i < len(resultRunes); i++ {
		// Check if this is an escaped character (preceded by backslash)
		isEscaped := i > 0 && resultRunes[i-1] == internalEscapeChar

		// Skip unescaped markers
		if !isEscaped && (resultRunes[i] == startMarkerRune || resultRunes[i] == endMarkerRune) {
			continue
		}

		cleanResult.WriteRune(resultRunes[i])
	}
	result = cleanResult.String()

	// THEN restore escaped characters (this restores original markers that were in the input)
	result = internalUnescape(result)

	return result
}

// applyMapReplacementsToPejelagarto translates text to Pejelagarto using map replacements
func applyMapReplacementsToPejelagarto(input string) string {
	// If input is not valid UTF-8, return it unchanged
	if !utf8.ValidString(input) {
		return input
	}

	// Escape quotes in input using output escaping (soft hyphen prefix)
	// This will be visible in the Pejelagarto output
	input = outputEscape(input, "'")

	bijectiveMap := createBijectiveMap()
	indices := getSortedIndices(bijectiveMap, true)
	result := applyReplacements(input, bijectiveMap, indices)

	return result
}

// applyMapReplacementsFromPejelagarto translates text from Pejelagarto using map replacements
func applyMapReplacementsFromPejelagarto(input string) string {
	// If input is not valid UTF-8, return it unchanged
	if !utf8.ValidString(input) {
		return input
	}

	bijectiveMap := createBijectiveMap()
	indices := getSortedIndices(bijectiveMap, false) // from Pejelagarto
	result := applyReplacements(input, bijectiveMap, indices)

	// Unescape output-escaped quotes (soft hyphen prefix)
	result = outputUnescape(result)

	return result
}

// applyNumbersLogicToPejelagarto applies number transformation for Pejelagarto encoding
// Positive numbers: converts base-10 to base-8
// Negative numbers: converts base-10 to base-7
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

				// Convert to base-8 for positive, base-7 for negative
				var convertedStr string
				if isNegative {
					// Negative: convert to base-7
					convertedStr = absValue.Text(7)
				} else {
					// Positive: convert to base-8
					convertedStr = absValue.Text(8)
				}

				// Write sign if negative
				if isNegative {
					result.WriteRune('-')
				}
				// Preserve leading zeros
				for j := 0; j < leadingZeros; j++ {
					result.WriteRune('0')
				}
				result.WriteString(convertedStr)
			}
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}

	return result.String()
}

// applyNumbersLogicFromPejelagarto applies number transformation from Pejelagarto encoding
// Positive numbers: converts base-8 to base-10
// Negative numbers: converts base-7 to base-10
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
		// Check if we're at the start of a number (base 7 or base 8, including negative)
		if isBase8Digit(runes[i]) || (runes[i] == '-' && i+1 < len(runes) && isBase7Digit(runes[i+1])) {
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
			// For positive: base-8 (0-7), for negative: base-7 (0-6)
			digitStart := i
			if isNegative {
				// Negative: expect base-7 digits (0-6)
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
			} else {
				// Positive: expect base-8 digits (0-7)
				for i < len(runes) && isBase8Digit(runes[i]) {
					i++
				}

				// Check if followed by digits 8-9 (which means it's actually a base-10 number, not base-8)
				// If so, we need to consume those digits and skip transformation
				hasHighDigits := false
				highDigitEnd := i
				if i < len(runes) && runes[i] >= '8' && runes[i] <= '9' {
					hasHighDigits = true
					// Consume remaining base-10 digits
					for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
						i++
					}
					highDigitEnd = i
				}

				// If we found digits 8-9, this is a base-10 number, not base-8 - preserve as-is
				if hasHighDigits {
					numberStr := string(runes[digitStart:highDigitEnd])
					for j := 0; j < leadingZeros; j++ {
						result.WriteRune('0')
					}
					result.WriteString(numberStr)
					continue
				}
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
				// Parse as big.Int from appropriate base
				var base10Str string
				if isNegative {
					// Negative: parse from base-7
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
					base10Str = base7Value.Text(10)
				} else {
					// Positive: parse from base-8
					base8Value := new(big.Int)
					_, ok := base8Value.SetString(numberStr, 8)
					if !ok {
						// Parse failed, preserve as-is
						for j := 0; j < leadingZeros; j++ {
							result.WriteRune('0')
						}
						result.WriteString(numberStr)
						continue
					}
					base10Str = base8Value.Text(10)
				}

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

// isBase8Digit checks if a rune is a valid base 8 digit (0-7)
func isBase8Digit(r rune) bool {
	return r >= '0' && r <= '7'
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
	'w': {"w", "ẁ", "ẃ", "ŵ", "ẅ"},                     // 5 single-rune accents for 'w'
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
	'w': {"w\u0328", "w\u030C"},            // w+ogonek, w+caron (2 runes each)
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
			// Try converting to lowercase first
			lower := unicode.ToLower(result[i])
			if lower != result[i] && unicode.ToUpper(lower) == result[i] {
				// Character can be lowercased and conversion is reversible
				result[i] = lower
				continue
			}

			// Try converting to uppercase
			upper := unicode.ToUpper(result[i])
			if upper != result[i] && unicode.ToLower(upper) == result[i] {
				// Character can be uppercased and conversion is reversible
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

	// Escape quotes using output escaping (soft hyphen prefix)
	input = outputEscape(input, "'")

	bijectiveMap := createPunctuationBijectiveMap()
	indices := getSortedPunctuationIndices(bijectiveMap, true)
	result := applyReplacements(input, bijectiveMap, indices)

	return result
}

// applyPunctuationReplacementsFromPejelagarto reverses punctuation replacements
func applyPunctuationReplacementsFromPejelagarto(input string) string {
	if !utf8.ValidString(input) {
		return input
	}

	bijectiveMap := createPunctuationBijectiveMap()
	indices := getSortedPunctuationIndices(bijectiveMap, false)
	result := applyReplacements(input, bijectiveMap, indices)

	// Unescape output-escaped quotes (soft hyphen prefix)
	result = outputUnescape(result)

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

// removeTimestampSpecialCharacters removes all special characters used for timestamp encoding
func removeTimestampSpecialCharacters(input string) string {
	// Build a map of all special characters used for timestamp encoding
	specialCharsMap := make(map[string]bool)
	for _, char := range daySpecialCharIndex {
		specialCharsMap[char] = true
	}
	for _, char := range monthSpecialCharIndex {
		specialCharsMap[char] = true
	}
	for _, char := range yearSpecialCharIndex {
		specialCharsMap[char] = true
	}
	for _, char := range hourSpecialCharIndex {
		specialCharsMap[char] = true
	}
	for _, char := range minuteSpecialCharIndex {
		specialCharsMap[char] = true
	}

	var result strings.Builder
	runes := []rune(input)
	for i := range runes {
		char := string(runes[i])
		// Keep character only if it's NOT in our special characters map
		if !specialCharsMap[char] {
			result.WriteString(char)
		}
	}
	return result.String()
}

// readTimestampUsingSpecialCharEncoding locates special characters and returns ISO 8601 timestamp
func readTimestampUsingSpecialCharEncoding(input string) string {
	// Find one special character from each category
	var day, month, year, hour, minute int = -1, -1, -1, 0, 0

	// Search for special characters in the input
	for i, specialChar := range daySpecialCharIndex {
		if strings.Contains(input, specialChar) {
			day = i + 1 // days are 1-indexed
			break
		}
	}

	for i, specialChar := range monthSpecialCharIndex {
		if strings.Contains(input, specialChar) {
			month = i + 1 // months are 1-indexed
			break
		}
	}

	for i, specialChar := range yearSpecialCharIndex {
		if strings.Contains(input, specialChar) {
			year = 2025 + i // years start from 2025
			break
		}
	}

	for i, specialChar := range hourSpecialCharIndex {
		if strings.Contains(input, specialChar) {
			hour = i
			break
		}
	}

	for i, specialChar := range minuteSpecialCharIndex {
		if strings.Contains(input, specialChar) {
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

// addSpecialCharDatetimeEncoding inserts datetime special characters at random positions
func addSpecialCharDatetimeEncoding(input string, timestamp string) string {
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

	// Get special character indices
	day := now.Day() - 1          // Convert to 0-indexed
	month := int(now.Month()) - 1 // Convert to 0-indexed
	year := now.Year() - 2025     // Years start from 2025
	hour := now.Hour()
	minute := now.Minute()

	// Validate indices
	if day < 0 || day >= len(daySpecialCharIndex) {
		day = 0
	}
	if month < 0 || month >= len(monthSpecialCharIndex) {
		month = 0
	}
	if year < 0 || year >= len(yearSpecialCharIndex) {
		year = 0
	}
	if hour < 0 || hour >= len(hourSpecialCharIndex) {
		hour = 0
	}
	if minute < 0 || minute >= len(minuteSpecialCharIndex) {
		minute = 0
	}

	// Get the special characters
	specialChars := []string{
		daySpecialCharIndex[day],
		monthSpecialCharIndex[month],
		yearSpecialCharIndex[year],
		hourSpecialCharIndex[hour],
		minuteSpecialCharIndex[minute],
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
		for _, specialChar := range specialChars {
			input += specialChar
		}
		return input
	}

	// Shuffle positions
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Work with runes to avoid splitting UTF-8 sequences
	resultRunes := runes

	// Insert as many special characters as we have positions
	numToInsert := len(specialChars)
	if len(positions) < numToInsert {
		numToInsert = len(positions)
	}

	// Sort the positions we'll use (insert from end to beginning)
	selectedPositions := positions[:numToInsert]
	sort.Ints(selectedPositions)

	for i := len(selectedPositions) - 1; i >= 0; i-- {
		pos := selectedPositions[i]
		if pos > len(resultRunes) {
			pos = len(resultRunes)
		}
		// Convert special character to runes and insert
		specialCharRunes := []rune(specialChars[i])
		resultRunes = append(resultRunes[:pos], append(specialCharRunes, resultRunes[pos:]...)...)
	}

	// If we couldn't insert all special characters, append the rest at the end
	for i := numToInsert; i < len(specialChars); i++ {
		specialCharRunes := []rune(specialChars[i])
		resultRunes = append(resultRunes, specialCharRunes...)
	}

	return string(resultRunes)
}

// sanitizeInvalidUTF8 replaces invalid UTF-8 bytes with Hangul Filler + Private Use Area characters
// Uses a bijective mapping to maintain reversibility
// Hangul Filler (U+3164) is invisible in most contexts, making the output cleaner
func sanitizeInvalidUTF8(input string) string {
	// Use Private Use Area characters - they won't be affected by any translation logic
	// Map each of 256 possible bytes to a unique character in range U+E000-U+E0FF
	const privateUseStart = 0xE000
	const hangulFiller = '\u3164' // Hangul Filler - invisible in most contexts

	var result strings.Builder
	result.Grow(len(input) * 2) // Reserve extra space

	for i := 0; i < len(input); {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8 byte - encode it invisibly using Hangul Filler + private use character
			invalidByte := input[i]
			result.WriteRune(hangulFiller) // Invisible marker
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
	const hangulFiller = '\u3164'

	var result []byte
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == hangulFiller && i+1 < len(runes) {
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
	input = removeTimestampSpecialCharacters(input)
	input, timestamp := removeISO8601timestamp(input)
	input = applyNumbersLogicToPejelagarto(input)
	input = applyPunctuationReplacementsToPejelagarto(input)
	input = applyMapReplacementsToPejelagarto(input)
	input = applyAccentReplacementLogicToPejelagarto(input)
	input = applyCaseReplacementLogic(input)
	input = addSpecialCharDatetimeEncoding(input, timestamp)
	return input
}

// TranslateFromPejelagarto translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
	timestamp := readTimestampUsingSpecialCharEncoding(input)
	input = removeTimestampSpecialCharacters(input)
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
        :root {
            --bg-gradient-start: #1a1a2e;
            --bg-gradient-end: #16213e;
            --container-bg: #0f3460;
            --text-primary: #e1e1e1;
            --text-secondary: #b0b0b0;
            --heading-color: #53a8e2;
            --button-gradient-start: #53a8e2;
            --button-gradient-end: #3d7ea6;
            --button-shadow: rgba(83, 168, 226, 0.4);
            --button-hover-shadow: rgba(83, 168, 226, 0.6);
            --invert-btn-gradient-start: #e94560;
            --invert-btn-gradient-end: #d62839;
            --invert-btn-shadow: rgba(233, 69, 96, 0.4);
            --invert-btn-hover-shadow: rgba(233, 69, 96, 0.6);
            --border-color: #2a2a40;
            --textarea-bg: #1a1a2e;
            --textarea-readonly-bg: #16213e;
            --textarea-focus-border: #53a8e2;
            --theme-btn-bg: #53a8e2;
            --theme-btn-hover: #3d7ea6;
        }

        [data-theme="light"] {
            --bg-gradient-start: #667eea;
            --bg-gradient-end: #764ba2;
            --container-bg: white;
            --text-primary: #333;
            --text-secondary: #666;
            --heading-color: #667eea;
            --button-gradient-start: #667eea;
            --button-gradient-end: #764ba2;
            --button-shadow: rgba(102, 126, 234, 0.4);
            --button-hover-shadow: rgba(102, 126, 234, 0.6);
            --invert-btn-gradient-start: #f093fb;
            --invert-btn-gradient-end: #f5576c;
            --invert-btn-shadow: rgba(245, 87, 108, 0.4);
            --invert-btn-hover-shadow: rgba(245, 87, 108, 0.6);
            --border-color: #e0e0e0;
            --textarea-bg: white;
            --textarea-readonly-bg: #f5f5f5;
            --textarea-focus-border: #667eea;
            --theme-btn-bg: #ffd700;
            --theme-btn-hover: #ffed4e;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, var(--bg-gradient-start) 0%, var(--bg-gradient-end) 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
            transition: background 0.3s ease;
        }
        
        .container {
            background: var(--container-bg);
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 900px;
            width: 100%;
            position: relative;
            transition: background 0.3s ease;
        }
        
        .theme-toggle {
            position: absolute;
            top: 20px;
            right: 20px;
            background: var(--theme-btn-bg);
            border: none;
            border-radius: 50%;
            width: 45px;
            height: 45px;
            cursor: pointer;
            display: flex;
            justify-content: center;
            align-items: center;
            font-size: 24px;
            transition: all 0.3s ease;
            box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
            z-index: 10;
        }
        
        .theme-toggle:hover {
            background: var(--theme-btn-hover);
            transform: scale(1.1) rotate(15deg);
            box-shadow: 0 6px 15px rgba(0, 0, 0, 0.3);
        }
        
        h1 {
            text-align: center;
            color: var(--heading-color);
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
            transition: color 0.3s ease;
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
            color: var(--text-primary);
            font-size: 1.1em;
            transition: color 0.3s ease;
            display: flex;
            align-items: center;
            gap: 8px;
            min-height: 40px;
        }
        
        textarea {
            width: 100%;
            height: 250px;
            padding: 15px;
            border: 2px solid var(--border-color);
            border-radius: 10px;
            font-size: 14px;
            font-family: 'Courier New', monospace;
            resize: none;
            transition: all 0.3s ease;
            background-color: var(--textarea-bg);
            color: var(--text-primary);
        }
        
        textarea:focus {
            outline: none;
            border-color: var(--textarea-focus-border);
        }
        
        textarea[readonly] {
            background-color: var(--textarea-readonly-bg);
            cursor: not-allowed;
            resize: none;
        }
        
        .controls {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 15px;
            flex-wrap: wrap;
        }
        
        button {
            background: linear-gradient(135deg, var(--button-gradient-start) 0%, var(--button-gradient-end) 100%);
            color: white;
            border: none;
            padding: 12px 30px;
            border-radius: 25px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: background 0.3s ease;
            box-shadow: 0 4px 15px var(--button-shadow);
        }
        
        button:active {
            transform: translateY(0);
        }
        
        .invert-btn {
            background: linear-gradient(135deg, var(--invert-btn-gradient-start) 0%, var(--invert-btn-gradient-end) 100%);
            padding: 12px 20px;
            font-size: 20px;
            box-shadow: 0 4px 15px var(--invert-btn-shadow);
        }
        
        .checkbox-container {
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 16px;
            color: var(--text-primary);
            transition: color 0.3s ease;
        }
        
        input[type="checkbox"] {
            width: 20px;
            height: 20px;
            cursor: pointer;
        }
        
        .play-btn {
            background: linear-gradient(135deg, #56ab2f 0%, #a8e063 100%);
            padding: 8px 16px;
            font-size: 18px;
            box-shadow: 0 4px 15px rgba(86, 171, 47, 0.4);
            min-width: auto;
        }
        
        .play-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        
        .hidden {
            display: none !important;
        }
        
        @media (max-width: 768px) {
            body {
                padding: 10px;
                align-items: flex-start;
            }
            
            .container {
                padding: 15px;
                border-radius: 15px;
                margin-top: 10px;
            }
            
            .translator-box {
                grid-template-columns: 1fr;
                gap: 15px;
                margin-bottom: 15px;
            }
            
            h1 {
                font-size: 1.5em;
                margin-bottom: 20px;
                padding-right: 50px;
            }
            
            .theme-toggle {
                width: 40px;
                height: 40px;
                top: 15px;
                right: 15px;
                font-size: 20px;
            }
            
            label {
                font-size: 0.95em;
                margin-bottom: 6px;
            }
            
            textarea {
                height: 120px;
                padding: 10px;
                font-size: 13px;
            }
            
            button {
                padding: 10px 20px;
                font-size: 14px;
            }
            
            .play-btn {
                padding: 6px 12px;
                font-size: 16px;
            }
            
            .invert-btn {
                padding: 10px 16px;
                font-size: 18px;
            }
            
            .controls {
                gap: 10px;
            }
            
            .checkbox-container {
                font-size: 14px;
            }
            
            input[type="checkbox"] {
                width: 18px;
                height: 18px;
            }
        }
        
        .htmx-indicator {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid var(--border-color);
            border-top: 3px solid var(--button-gradient-start);
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-left: 10px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .download-section {
            margin-top: 30px;
            padding: 15px;
            background: var(--textarea-bg);
            border-radius: 8px;
            border: 1px solid var(--border-color);
        }
        
        .download-buttons {
            display: flex;
            gap: 10px;
            justify-content: center;
            flex-wrap: wrap;
        }
        
        .download-btn {
            display: inline-block;
            padding: 8px 16px;
            background: linear-gradient(135deg, var(--button-gradient-start), var(--button-gradient-end));
            color: var(--text-primary);
            text-decoration: none;
            border-radius: 6px;
            font-size: 13px;
            font-weight: 500;
            transition: all 0.3s ease;
            border: 1px solid var(--border-color);
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
        }
        
        .download-btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.25);
            filter: brightness(1.1);
        }
        
        .download-btn:active {
            transform: translateY(0);
            box-shadow: 0 1px 2px rgba(0, 0, 0, 0.15);
        }
        
        /* Desktop: align download section to bottom left */
        @media (min-width: 769px) {
            .download-section {
                position: fixed;
                bottom: 20px;
                left: 20px;
                margin-top: 0;
                max-width: 280px;
                z-index: 100;
            }
            
            .download-buttons {
                flex-direction: column;
                gap: 8px;
            }
            
            .download-btn {
                width: 100%;
                text-align: center;
            }
        }
        
        .version-display {
            position: fixed;
            bottom: 10px;
            right: 10px;
            font-size: 12px;
            color: var(--text-secondary);
            opacity: 0.7;
            font-family: 'Courier New', monospace;
            z-index: 1000;
        }
        
        .version-display a {
            color: var(--text-secondary);
            text-decoration: none;
            transition: opacity 0.2s ease;
        }
        
        .version-display a:hover {
            opacity: 1;
            text-decoration: underline;
        }
        
        @media (max-width: 768px) {
            .version-display {
                font-size: 10px;
                bottom: 5px;
                right: 5px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <button class="theme-toggle" onclick="toggleTheme()" aria-label="Toggle theme">
            <span id="theme-icon">🌙</span>
        </button>
        <h1>🐊 Pejelagarto Translator 🐊</h1>
        
        <div class="translator-box">
            <div class="text-area-container">
                <label id="input-label">Human:</label>
                <textarea id="input-text" placeholder="Type your text here..."></textarea>
            </div>
            
            <div class="text-area-container">
                <label id="output-label">Pejelagarto: <button class="play-btn" id="play-output" onclick="playAudio('output', false)">🔊 Play</button>{{DROPDOWN_PLACEHOLDER}}</label>
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
                <input type="checkbox" id="live-translate" onchange="toggleLiveTranslation()" checked>
                <label for="live-translate" style="margin: 0;">Live Translation</label>
            </div>
            
            <span id="loading-indicator" class="htmx-indicator"></span>
        </div>
    </div>
    
    <div id="download-section" class="download-section" style="display: none;">
        <h3 style="color: var(--text-primary); margin-bottom: 10px; font-size: 16px;">Download Translator</h3>
        <div class="download-buttons">
            <a href="/download/windows" download="pejelagarto-translator.exe" class="download-btn">
                💻 Windows
            </a>
            <a href="/download/linux" download="pejelagarto-translator" class="download-btn">
                🐧 Linux/Mac
            </a>
        </div>
    </div>
    
    <script>
        let isInverted = false;
        let liveTranslateEnabled = true;
        
        // Initialize theme on page load
        (function initTheme() {
            // Check localStorage for saved preference, default to dark mode
            const savedTheme = localStorage.getItem('theme') || 'dark';
            document.documentElement.setAttribute('data-theme', savedTheme);
            updateThemeIcon(savedTheme);
        })();
        
        // Initialize live translation on page load
        (function initLiveTranslation() {
            const translateBtn = document.getElementById('translate-btn');
            const inputText = document.getElementById('input-text');
            
            // Hide translate button since live translation is on
            translateBtn.classList.add('hidden');
            
            // Add event listener for live translation
            inputText.addEventListener('input', handleLiveTranslation);
        })();
        
        // Check if downloadable version and show download section
        (function initDownloadSection() {
            fetch('/api/is-downloadable')
                .then(response => response.json())
                .then(data => {
                    if (data.downloadable) {
                        document.getElementById('download-section').style.display = 'block';
                    }
                })
                .catch(err => console.log('Download check failed:', err));
        })();
        
        // Toggle theme function
        function toggleTheme() {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            
            document.documentElement.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            updateThemeIcon(newTheme);
        }
        
        // Update theme icon
        function updateThemeIcon(theme) {
            const icon = document.getElementById('theme-icon');
            icon.textContent = theme === 'dark' ? '🌙' : '☀️';
        }
        
        // Handle translate button click
        function handleTranslateClick() {
            handleLiveTranslation();
        }
        
        // Invert button functionality
        function invertTranslation() {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const translateBtn = document.getElementById('translate-btn');
            
            // Swap text content
            const temp = inputText.value;
            inputText.value = outputText.value;
            outputText.value = temp;
            
            // Toggle inverted state
            isInverted = !isInverted;
            
            // Update translate button text
            if (isInverted) {
                translateBtn.textContent = 'Translate from Pejelagarto';
            } else {
                translateBtn.textContent = 'Translate to Pejelagarto';
            }
            
            // Force reset to single button state with proper labels
            resetToSingleButton();
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
        
        // Track current output text, language, and slow audio availability
        let currentOutputText = '';
        let currentLanguage = '';
        let currentInvertedState = false;
        let slowAudioReady = {}; // Tracks which text+language combinations have slow audio ready
        
        // Watch for output text changes or language changes to reset buttons
        function watchOutputChanges() {
            const outputText = document.getElementById('output-text');
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            // Check if state has changed (text, language, or invert state)
            if (outputText.value !== currentOutputText || selectedLanguage !== currentLanguage || isInverted !== currentInvertedState) {
                currentOutputText = outputText.value;
                currentLanguage = selectedLanguage;
                currentInvertedState = isInverted;
                resetToSingleButton();
                
                // Check if slow audio is already ready for this text+language combination
                const cacheKey = currentOutputText + ':' + selectedLanguage;
                if (slowAudioReady[cacheKey]) {
                    // Already have slow audio, split button immediately
                    const source = isInverted ? 'input' : 'output';
                    const container = isInverted ? document.getElementById('input-label') : document.getElementById('output-label');
                    splitButton(source, container);
                }
            }
        }
        
        // Reset to single button state
        function resetToSingleButton() {
            const outputLabel = document.getElementById('output-label');
            const inputLabel = document.getElementById('input-label');
            
            // Get current language selection before recreating dropdown
            const oldDropdown = document.getElementById('tts-language');
            const selectedLang = oldDropdown ? oldDropdown.value : 'russian';
            
            const dropdownHTML = document.getElementById('tts-language') ? 
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="kazakh">North-North-East</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">South-East-East</option><option value="hindi">South-East</option><option value="icelandic">South-South-East</option><option value="romanian">South</option><option value="vietnamese">South-South-West</option><option value="swahili">South-West</option><option value="swedish">South-West-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option></select>' : '';
            
            // Always reset both labels to ensure clean state
            if (isInverted) {
                inputLabel.innerHTML = 'Pejelagarto: <button class="play-btn" id="play-input" onclick="playAudio(&quot;input&quot;, false)" style="width: 104px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🔊 Play</button>' + dropdownHTML;
                outputLabel.textContent = 'Human:';
            } else {
                outputLabel.innerHTML = 'Pejelagarto: <button class="play-btn" id="play-output" onclick="playAudio(&quot;output&quot;, false)" style="width: 104px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🔊 Play</button>' + dropdownHTML;
                inputLabel.textContent = 'Human:';
            }
            
            // Restore language selection
            const newDropdown = document.getElementById('tts-language');
            if (newDropdown) {
                newDropdown.value = selectedLang;
            }
        }
        
        // Start watching for changes
        setInterval(watchOutputChanges, 500);
        
        // Play audio function - only called when play button is clicked
        function playAudio(source, slow) {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            const playInputBtn = document.getElementById('play-input');
            const playOutputBtn = document.getElementById('play-output');
            const playInputSlowBtn = document.getElementById('play-input-slow');
            const playOutputSlowBtn = document.getElementById('play-output-slow');
            
            // Determine which text to convert to speech
            let textToSpeak = '';
            let button = null;
            let container = null;
            
            if (source === 'input') {
                textToSpeak = inputText.value;
                button = slow ? playInputSlowBtn : playInputBtn;
                container = document.getElementById('input-label');
            } else {
                textToSpeak = outputText.value;
                button = slow ? playOutputSlowBtn : playOutputBtn;
                container = document.getElementById('output-label');
            }
            
            // Check if there's text to speak
            if (!textToSpeak || textToSpeak.trim() === '') {
                alert('No text to convert to speech!');
                return;
            }
            
            // Disable button and show loading state
            button.disabled = true;
            const originalText = button.textContent;
            button.textContent = '⏳';
            
            // Get selected language from dropdown if available
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            // Build URL with language parameter and slow parameter
            let url = selectedLanguage ? '/tts?lang=' + selectedLanguage : '/tts';
            if (slow) {
                url += (selectedLanguage ? '&' : '?') + 'slow=true';
            }
            
            // Send request to TTS endpoint
            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain'
                },
                body: textToSpeak
            })
            .then(async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(errorText || 'TTS request failed: ' + response.statusText);
                }
                return response.blob();
            })
            .then(blob => {
                // Create an audio element and play it
                const audioUrl = URL.createObjectURL(blob);
                const audio = new Audio(audioUrl);
                
                audio.onended = function() {
                    URL.revokeObjectURL(audioUrl);
                    button.disabled = false;
                    button.textContent = originalText;
                };
                
                audio.onerror = function() {
                    URL.revokeObjectURL(audioUrl);
                    button.disabled = false;
                    button.textContent = originalText;
                    alert('Error playing audio');
                };
                
                audio.play();
                button.textContent = '▶️';
                
                // If this is normal speed, start checking for slow version
                if (!slow) {
                    checkForSlowAudio(textToSpeak, selectedLanguage, source, container);
                }
            })
            .catch(error => {
                console.error('TTS error:', error);
                button.disabled = false;
                button.textContent = originalText;
                
                let errorMsg = error.message;
                if (errorMsg.includes('voice model not found')) {
                    const lang = selectedLanguage || 'portuguese';
                    errorMsg = 'Language model not installed for: ' + lang + '\\n\\nTo install the model, run:\\ncd tts/requirements/piper/languages\\n.\\\\download_models.ps1\\n\\nOr download manually from:\\ntts/requirements/piper/languages/README.md';
                }
                
                alert('Text-to-speech error:\\n\\n' + errorMsg);
            });
        }
        
        // Check periodically if slow audio is ready
        function checkForSlowAudio(text, language, source, container) {
            const url = language ? '/tts-check-slow?lang=' + language : '/tts-check-slow';
            const cacheKey = text + ':' + language;
            
            const checkInterval = setInterval(() => {
                fetch(url, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'text/plain'
                    },
                    body: text
                })
                .then(response => response.json())
                .then(data => {
                    if (data.ready) {
                        clearInterval(checkInterval);
                        // Mark this text+language combination as ready
                        slowAudioReady[cacheKey] = true;
                        splitButton(source, container);
                    }
                })
                .catch(error => {
                    console.error('Error checking slow audio:', error);
                    clearInterval(checkInterval);
                });
            }, 1000); // Check every second
            
            // Stop checking after 30 seconds
            setTimeout(() => clearInterval(checkInterval), 30000);
        }
        
        // Split button into fast and slow versions
        function splitButton(source, container) {
            // Validate that we're modifying the correct container based on current state
            const expectedSource = isInverted ? 'input' : 'output';
            if (source !== expectedSource) {
                // State has changed since this was called, don't modify
                return;
            }
            
            // Double-check the container is the correct label
            const expectedContainer = isInverted ? document.getElementById('input-label') : document.getElementById('output-label');
            if (container !== expectedContainer) {
                // Container mismatch, state has changed
                return;
            }
            
            // Get current language selection before recreating dropdown
            const oldDropdown = document.getElementById('tts-language');
            const selectedLang = oldDropdown ? oldDropdown.value : 'russian';
            
            const dropdownHTML = document.getElementById('tts-language') ? 
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="kazakh">North-North-East</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">South-East-East</option><option value="hindi">South-East</option><option value="icelandic">South-South-East</option><option value="romanian">South</option><option value="vietnamese">South-South-West</option><option value="swahili">South-West</option><option value="swedish">South-West-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option></select>' : '';
            
            const label = 'Pejelagarto:';
            const buttonId = source === 'input' ? 'play-input' : 'play-output';
            const slowButtonId = source === 'input' ? 'play-input-slow' : 'play-output-slow';
            
            container.innerHTML = label + 
                ' <button class="play-btn" id="' + buttonId + '" onclick="playAudio(&quot;' + source + '&quot;, false)" style="width: 50px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🐇🔊</button>' +
                ' <button class="play-btn" id="' + slowButtonId + '" onclick="playAudio(&quot;' + source + '&quot;, true)" style="width: 50px; height: 38px; padding: 4px 2px; font-size: 16px; overflow: hidden; white-space: nowrap;">🐌🔊</button>' +
                dropdownHTML;
            
            // Restore language selection
            const newDropdown = document.getElementById('tts-language');
            if (newDropdown) {
                newDropdown.value = selectedLang;
            }
        }
    </script>
    
    <div class="version-display"><a href="https://github.com/FranLegon/pejelagarto-translator" target="_blank">{{VERSION}}</a></div>
</body>
</html>`

// HTTP handler for the main UI
