// Pejelagarto Translator
// IMPORTANT: Before building, run: .\get-requirements.ps1
// This downloads all TTS dependencies which will be embedded into the binary
//
// Build command: go build -o pejelagarto-translator.exe main.go
// Run command (local): .\pejelagarto-translator.exe
// Run command (ngrok with random domain): .\pejelagarto-translator.exe -ngrok_token YOUR_TOKEN_HERE
// Run command (ngrok with persistent domain): .\pejelagarto-translator.exe -ngrok_token YOUR_TOKEN_HERE -ngrok_domain your-domain.ngrok-free.app
//
// Note: All TTS dependencies (Piper binary, voice models, espeak-ng-data) are embedded in the executable.
// They will be extracted to a temp directory at runtime (e.g., C:\Windows\Temp\pejelagarto-translator or /tmp/pejelagarto-translator)

package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

//go:embed tts/requirements/*
var embeddedRequirements embed.FS

// tempRequirementsDir stores the path to extracted requirements
var tempRequirementsDir string

// extractEmbeddedRequirements extracts all embedded TTS requirements to a temp directory
func extractEmbeddedRequirements() error {
	// Determine temp directory based on OS
	var baseDir string
	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("TEMP")
		if baseDir == "" {
			baseDir = os.Getenv("TMP")
		}
		if baseDir == "" {
			baseDir = "C:\\Windows\\Temp"
		}
	} else {
		baseDir = "/tmp"
	}

	// Create a unique directory for this application
	tempRequirementsDir = filepath.Join(baseDir, "pejelagarto-translator", "requirements")

	// Check if already extracted (reuse if exists)
	piperExe := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		piperExe += ".exe"
	}
	if _, err := os.Stat(piperExe); err == nil {
		// Already extracted
		log.Printf("Using cached TTS requirements at: %s", tempRequirementsDir)
		return nil
	}

	log.Printf("Extracting embedded TTS requirements to: %s", tempRequirementsDir)

	// Remove old directory if exists
	if err := os.RemoveAll(tempRequirementsDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean temp directory: %w", err)
	}

	// Create temp directory
	if err := os.MkdirAll(tempRequirementsDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Walk through embedded files and extract them
	err := fs.WalkDir(embeddedRequirements, "tts/requirements", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path (remove "tts/requirements/" prefix)
		relPath, err := filepath.Rel("tts/requirements", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Destination path
		destPath := filepath.Join(tempRequirementsDir, relPath)

		if d.IsDir() {
			// Create directory
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
		} else {
			// Read embedded file
			data, err := embeddedRequirements.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read embedded file %s: %w", path, err)
			}

			// Write to destination
			if err := os.WriteFile(destPath, data, 0755); err != nil {
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to extract embedded files: %w", err)
	}

	log.Printf("Successfully extracted TTS requirements")
	return nil
}

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

// Special character encoding maps for datetime using Unicode range U+2300 to U+23FB (avoiding emojis)
var daySpecialCharIndex = []string{
	"\u2300", "\u2301", "\u24FC", "\u2303", "\u2304", "\u2305", "\u2306", "\u2307", "\u2308", "\u2309",
	"\u230A", "\u230B", "\u230C", "\u230D", "\u230E", "\u230F", "\u2310", "\u2311", "\u2312", "\u2313",
	"\u2314", "\u2315", "\u2316", "\u2317", "\u2318", "\u2319", "\u24EA", "\u24EB", "\u231C", "\u231D",
	"\u231E", "\u231F", "\u2320", "\u2321", "\u2322", "\u2323", "\u2324", "\u2325", "\u2326", "\u2327",
	"\u24FB", "\u2329", "\u232A", "\u232B", "\u232C", "\u232D", "\u232E", "\u232F", "\u2330", "\u2331",
	"\u2332", "\u2333", "\u2334", "\u2335", "\u2336", "\u2337", "\u2338", "\u2339", "\u233A", "\u233B",
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
	"\u309A", "\u309B", "\u309C", "\uA702", "\uAAB8", "\uFBB2", "\uA950", "\uA951", "\uA926", "\uA952",
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

// escapeReservedMarkers escapes all reserved Unicode markers (U+FFF0 through U+FFFF) in the input
// using a simple encoding scheme to preserve them through translation
func escapeReservedMarkers(input string) string {
	// Reserved markers range from U+FFF0 to U+FFFF
	// We encode them as: U+FFFE followed by the last nibble (0-F)
	const escapePrefix = "\uFFFE"
	result := input

	// Escape in reverse order (highest to lowest) to avoid double-escaping
	for i := 0xFFFF; i >= 0xFFF0; i-- {
		marker := string(rune(i))
		nibble := string(rune('0' + (i & 0x0F)))
		result = strings.ReplaceAll(result, marker, escapePrefix+nibble)
	}

	return result
}

// unescapeReservedMarkers reverses the escaping done by escapeReservedMarkers
func unescapeReservedMarkers(input string) string {
	const escapePrefix = "\uFFFE"
	result := input

	// Unescape in forward order (lowest to highest)
	for i := 0xFFF0; i <= 0xFFFF; i++ {
		marker := string(rune(i))
		nibble := string(rune('0' + (i & 0x0F)))
		result = strings.ReplaceAll(result, escapePrefix+nibble, marker)
	}

	return result
}

// applyReplacements applies replacements from the bijective map in the specified order
func applyReplacements(input string, bijectiveMap map[int32]map[string]string, indices []int32) string {
	// Use special Unicode characters as markers that won't be in normal text
	const startMarker = "\uFFF0"
	const endMarker = "\uFFF1"
	const escapedStartMarker = "\uFFF3"
	const escapedEndMarker = "\uFFF4"
	const escapePrefix = "\uFFF5"

	// Escape any markers (including escaped markers) that appear in the input to preserve them
	result := strings.ReplaceAll(input, escapePrefix, escapePrefix+escapePrefix)
	result = strings.ReplaceAll(result, escapedStartMarker, escapePrefix+escapedStartMarker)
	result = strings.ReplaceAll(result, escapedEndMarker, escapePrefix+escapedEndMarker)
	result = strings.ReplaceAll(result, startMarker, escapedStartMarker)
	result = strings.ReplaceAll(result, endMarker, escapedEndMarker)

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

	// Remove all working markers
	result = strings.ReplaceAll(result, startMarker, "")
	result = strings.ReplaceAll(result, endMarker, "")

	// Restore escaped markers to original characters (reverse order of escaping)
	result = strings.ReplaceAll(result, escapedStartMarker, startMarker)
	result = strings.ReplaceAll(result, escapedEndMarker, endMarker)
	result = strings.ReplaceAll(result, escapePrefix+startMarker, escapedStartMarker)
	result = strings.ReplaceAll(result, escapePrefix+endMarker, escapedEndMarker)
	result = strings.ReplaceAll(result, escapePrefix+escapePrefix, escapePrefix)

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

	// Recursively escape all reserved markers (FFF0-FFF6) in the input
	input = escapeReservedMarkers(input)
	input = strings.ReplaceAll(input, "'", quoteMarker)

	bijectiveMap := createBijectiveMap()
	indices := getSortedIndices(bijectiveMap, true) // to Pejelagarto
	result := applyReplacements(input, bijectiveMap, indices)

	// Restore literal quotes as doubled quotes in the output
	result = strings.ReplaceAll(result, quoteMarker, "''")
	// Restore escaped markers
	result = unescapeReservedMarkers(result)
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

	// Recursively escape all reserved markers (FFF0-FFFF) in the input
	input = escapeReservedMarkers(input)
	input = strings.ReplaceAll(input, "''", quoteMarker)

	bijectiveMap := createBijectiveMap()
	indices := getSortedIndices(bijectiveMap, false) // from Pejelagarto
	result := applyReplacements(input, bijectiveMap, indices)

	// Restore literal quotes from marker
	result = strings.ReplaceAll(result, quoteMarker, "'")
	// Restore escaped markers
	result = unescapeReservedMarkers(result)
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
                <input type="checkbox" id="live-translate" onchange="toggleLiveTranslation()">
                <label for="live-translate" style="margin: 0;">Live Translation</label>
            </div>
            
            <span id="loading-indicator" class="htmx-indicator"></span>
        </div>
    </div>
    
    <script>
        let isInverted = false;
        let liveTranslateEnabled = false;
        
        // Initialize theme on page load
        (function initTheme() {
            // Check localStorage for saved preference, default to dark mode
            const savedTheme = localStorage.getItem('theme') || 'dark';
            document.documentElement.setAttribute('data-theme', savedTheme);
            updateThemeIcon(savedTheme);
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
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">Center</option><option value="hindi">South-East</option><option value="romanian">South</option><option value="icelandic">South-South-East</option><option value="arabic">South-West</option><option value="swedish">South-West-West</option><option value="vietnamese">South-South-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option><option value="kazakh">North-North-East</option></select>' : '';
            
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
                ' <select id="tts-language" onchange="watchOutputChanges()" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;"><option value="russian">North</option><option value="german">North-East</option><option value="turkish">North-East-East</option><option value="portuguese">East</option><option value="french">Center</option><option value="hindi">South-East</option><option value="romanian">South</option><option value="icelandic">South-South-East</option><option value="arabic">South-West</option><option value="swedish">South-West-West</option><option value="vietnamese">South-South-West</option><option value="czech">West</option><option value="chinese">North-West-West</option><option value="norwegian">North-West</option><option value="hungarian">North-North-West</option><option value="kazakh">North-North-East</option></select>' : '';
            
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
</body>
</html>`

// HTTP handler for the main UI
func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Conditionally inject the language dropdown
	html := htmlUI
	if pronunciationLanguageDropdown {
		dropdownHTML := ` <select id="tts-language" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;">
                    <option value="russian">North</option>
                    <option value="german">North-East</option>
                    <option value="turkish">North-East-East</option>
                    <option value="portuguese">East</option>
                    <option value="french">Center</option>
                    <option value="hindi">South-East</option>
                    <option value="romanian">South</option>
                    <option value="icelandic">South-South-East</option>
                    <option value="arabic">South-West</option>
                    <option value="swedish">South-West-West</option>
                    <option value="vietnamese">South-South-West</option>
                    <option value="czech">West</option>
                    <option value="chinese">North-West-West</option>
                    <option value="norwegian">North-West</option>
                    <option value="hungarian">North-North-West</option>
                    <option value="kazakh">North-North-East</option>
                </select>`
		html = strings.Replace(html, "{{DROPDOWN_PLACEHOLDER}}", dropdownHTML, 1)
	} else {
		html = strings.Replace(html, "{{DROPDOWN_PLACEHOLDER}}", "", 1)
	}

	fmt.Fprint(w, html)
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

// Global variable to store the pronunciation language flag
var pronunciationLanguage string
var pronunciationLanguageDropdown bool

// Audio cache for storing normal and slow versions
var audioCache = struct {
	sync.RWMutex
	cache map[string][]byte // key: text+lang hash -> audio data
}{
	cache: make(map[string][]byte),
}

// getPiperBinaryPath returns the path to the Piper binary
func getPiperBinaryPath() string {
	binaryPath := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	return binaryPath
}

// getModelPath returns the language-specific model path
func getModelPath(language string) string {
	return filepath.Join(tempRequirementsDir, "piper", "languages", language, "model.onnx")
}

// preprocessTextForTTS prepares text for TTS by:
// 1. Converting numbers from Pejelagarto format to standard format
// 2. Removing non-language-specific characters
// 3. Limiting consonant clusters to max 2 adjacent consonants
func preprocessTextForTTS(input string, pronunciationLanguage string) string {
	// Step 1: Apply number conversion from Pejelagarto format
	input = applyNumbersLogicFromPejelagarto(input)

	// Define language-specific vowels and consonants
	var vowels, consonants, allowed string

	switch pronunciationLanguage {
	case "portuguese":
		vowels = "aeiouáéíóúâêôãõàü"
		consonants = "bcdfghjklmnpqrstvwxyzç"
		allowed = vowels + consonants + "AEIOUÁÉÍÓÚÂÊÔÃÕÀÜBCDFGHJKLMNPQRSTVWXYZÇ" + "0123456789" + " .,!?;:'\"-()[]"
	case "spanish":
		vowels = "aeiouáéíóúü"
		consonants = "bcdfghjklmnñpqrstvwxyz"
		allowed = vowels + consonants + "AEIOUÁÉÍÓÚÜBCDFGHJKLMNÑPQRSTVWXYZ" + "0123456789" + " .,!?;¡¿:'\"-()[]"
	case "english":
		vowels = "aeiou"
		consonants = "bcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "AEIOUБCDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]"
	case "russian":
		vowels = "аеёиоуыэюяaeiou"
		consonants = "бвгджзйклмнпрстфхцчшщbcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "АЕЁИОУЫЭЮЯБВГДЖЗЙКЛМНПРСТФХЦЧШЩAEIOUBCDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]ъьЪЬ"
	case "czech":
		vowels = "aeiouyáéíóúůýě"
		consonants = "bcčdďfghjklmnňpqrřsštťvwxzž"
		allowed = vowels + consonants + "AEIOUYÁÉÍÓÚŮÝĚBCČDĎFGHJKLMNŇPQRŘSŠTŤVWXZŽ" + "0123456789" + " .,!?;:'\"-()[]'"
	case "romanian":
		vowels = "aeiouăâî"
		consonants = "bcdfghjklmnpqrstvwxyzșț"
		allowed = vowels + consonants + "AEIOUĂÂÎBCDFGHJKLMNPQRSTVWXYZȘȚ" + "0123456789" + " .,!?;:'\"-()[]'"
	case "french":
		vowels = "aeiouàâäéèêëïîôùûü"
		consonants = "bcçdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "AEIOUÀÂÄÉÈÊËÏÎÔÙÛÜBCÇDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]œæŒÆ"
	case "german":
		vowels = "aeiouäöü"
		consonants = "bcdfghjklmnpqrsßtvwxyz"
		allowed = vowels + consonants + "AEIOUÄÖÜBCDFGHJKLMNPQRSẞTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]ẞ"
	case "hindi":
		vowels = "अआइईउऊऋएऐओऔaeiou"
		consonants = "कखगघङचछजझञटठडढणतथदधनपफबभमयरलवशषसहaeioukcghnjṭḍnṇtdpbmyrlvśṣsh"
		allowed = vowels + consonants + "अआइईउऊऋएऐओऔकखगघङचछजझञटठडढणतथदधनपफबभमयरलवशषसहािीुूृेैोौंःँ" + "0123456789" + " .,!?;:'\"-()[]"
	case "arabic":
		vowels = "اأإآةويىاeiou"
		consonants = "بتثجحخدذرزسشصضطظعغفقكلمنهىي"
		allowed = vowels + consonants + "اأإآةويىبتثجحخدذرزسشصضطظعغفقكلمنهىيًٌٍَُِّْٰ" + "0123456789" + " .,!?;:'\"-()[]"
	case "icelandic":
		vowels = "aeiouyáéíóúýæöAEIOUYÁÉÍÓÚÝÆÖ"
		consonants = "bcdfghjklmnpqrstvwxzþðBCDFGHJKLMNPQRSTVWXZÞÐ"
		allowed = vowels + consonants + "0123456789" + " .,!?;:'\"-()[]"
	case "kazakh":
		vowels = "аәеёиоөұүыэюяaeiou"
		consonants = "бвгғджзйкқлмнңпрстфхһцчшщbcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "АӘЕЁИОӨҰҮЫЭЮЯБВГҒДЖЗЙКҚЛМНҢПРСТФХҺЦЧШЩAEIOUBCDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]ъьіЪЬІ"
	case "norwegian":
		vowels = "aeiouyæøå"
		consonants = "bcdfghjklmnpqrstvwxz"
		allowed = vowels + consonants + "AEIOUYÆØÅBCDFGHJKLMNPQRSTVWXZ" + "0123456789" + " .,!?;:'\"-()[]"
	case "swedish":
		vowels = "aeiouyäöå"
		consonants = "bcdfghjklmnpqrstvwxz"
		allowed = vowels + consonants + "AEIOUYÄÖÅBCDFGHJKLMNPQRSTVWXZ" + "0123456789" + " .,!?;:'\"-()[]"
	case "turkish":
		vowels = "aeıioöuüAEIİOÖUÜ"
		consonants = "bcçdfgğhjklmnprsştvyzBCÇDFGĞHJKLMNPRSŞTVYZ"
		allowed = vowels + consonants + "0123456789" + " .,!?;:'\"-()[]"
	case "vietnamese":
		vowels = "aăâeêioôơuưyáàảãạắằẳẵặấầẩẫậéèẻẽẹếềểễệíìỉĩịóòỏõọốồổỗộớờởỡợúùủũụứừửữựýỳỷỹỵ"
		consonants = "bcdđghklmnpqrstvx"
		allowed = vowels + consonants + "AĂÂEÊIOÔƠUƯYÁÀẢÃẠẮẰẲẴẶẤẦẨẪẬÉÈẺẼẸẾỀỂỄỆÍÌỈĨỊÓÒỎÕỌỐỒỔỖỘỚỜỞỠỢÚÙỦŨỤỨỪỬỮỰÝỲỶỸỴBCDĐGHKLMNPQRSTVX" + "0123456789" + " .,!?;:'\"-()[]"
	case "hungarian":
		vowels = "aáeéiíoóöőuúüű"
		consonants = "bcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "AÁEÉIÍOÓÖŐUÚÜŰBCDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]"
	case "chinese":
		vowels = "aeiouüāáǎàēéěèīíǐìōóǒòūúǔùǖǘǚǜ"
		consonants = "bpmfdtnlgkhzcsSrjqxyw"
		allowed = vowels + consonants + "AEIOUÜĀÁǍÀĒÉĚÈĪÍǏÌŌÓǑÒŪÚǓÙǕǗǙǛBPMFDTNLGKHZCSSRJQXYW一-龥" + "0123456789" + " .,!?;:'\"-()[]"
	default:
		// Fallback to Russian
		vowels = "аеёиоуыэюяaeiou"
		consonants = "бвгджзйклмнпрстфхцчшщbcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "АЕЁИОУЫЭЮЯБВГДЖЗЙКЛМНПРСТФХЦЧШЩAEIOUBCDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]ъьЪЬ"
	}

	// Step 1: Convert to lowercase and remove non-allowed characters
	var cleaned strings.Builder
	for _, r := range input {
		if strings.ContainsRune(allowed, r) {
			// Convert letters to lowercase, keep numbers and punctuation as-is
			if unicode.IsLetter(r) {
				cleaned.WriteRune(unicode.ToLower(r))
			} else {
				cleaned.WriteRune(r)
			}
		}
	}
	result := cleaned.String()

	// Step 2: Limit consonant clusters to max 2 adjacent consonants
	// If 3 or more consonants found, remove third onwards
	var final strings.Builder
	runes := []rune(result)

	for i := 0; i < len(runes); i++ {
		currentRune := runes[i]
		currentLower := unicode.ToLower(currentRune)

		// Always write vowels and non-letters
		if !unicode.IsLetter(currentRune) || strings.ContainsRune(vowels, currentLower) {
			final.WriteRune(currentRune)
			continue
		}

		// Current is a consonant - check how many consonants precede it
		consonantCount := 1 // Current consonant

		// Count preceding consecutive consonants
		for j := i - 1; j >= 0; j-- {
			prevRune := runes[j]
			prevLower := unicode.ToLower(prevRune)

			if !unicode.IsLetter(prevRune) {
				break
			}
			if strings.ContainsRune(vowels, prevLower) {
				break
			}
			consonantCount++
		}

		// Only write if we have 2 or fewer consecutive consonants
		if consonantCount <= 2 {
			final.WriteRune(currentRune)
		}
		// If consonantCount > 2, skip this consonant (remove third onwards)
	}

	finalResult := final.String()

	// If the result is empty or only whitespace, return a space to prevent TTS errors
	if strings.TrimSpace(finalResult) == "" {
		return " "
	}

	return finalResult
}

// textToSpeech executes the Piper Text-to-Speech binary to convert text to audio.
// It generates a unique temporary WAV file for the output audio.
func textToSpeech(input string, pronunciationLanguage string) (outputPath string, err error) {
	// Preprocess text for better TTS pronunciation
	input = preprocessTextForTTS(input, pronunciationLanguage)

	// Get language-specific model path
	modelPath := getModelPath(pronunciationLanguage)

	// Get the Piper binary path
	binaryPath := getPiperBinaryPath()

	// Check if the Piper binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("piper binary not found at %s: %w", binaryPath, err)
	}

	// Check if the voice model exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("voice model not found at %s: %w", modelPath, err)
	}

	// Create a unique temporary file for the output audio
	tempFile, err := os.CreateTemp("", "piper-tts-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %w", err)
	}
	outputPath = tempFile.Name()
	tempFile.Close()

	// Get absolute paths for the output file and model
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for output: %w", err)
	}
	absModelPath, err := filepath.Abs(modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for model: %w", err)
	}

	// Get absolute path for the binary (don't rely on PATH environment variable)
	absBinaryPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for binary: %w", err)
	}

	// Get absolute path for the requirements directory
	absRequirementsDir := tempRequirementsDir

	// Build the command to execute Piper
	// Use absolute path to binary and run from its directory to find DLLs/espeak-ng-data
	// Note: We use stdin instead of --text argument for better handling of special characters
	cmd := exec.Command(
		absBinaryPath,
		"-m", absModelPath,
		"--output_file", absOutputPath,
	)

	// Set working directory to the Piper requirements directory (absolute path)
	// This is needed so Piper can find DLLs and espeak-ng-data folder
	cmd.Dir = absRequirementsDir

	// Send text via stdin (better for special characters and Unicode)
	cmd.Stdin = strings.NewReader(input)

	// Capture both stdout and stderr for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("piper command failed: %w\nOutput: %s", err, string(output))
	}

	// Verify that the output file was created and has content
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return "", fmt.Errorf("output file not created: %w", err)
	}
	if fileInfo.Size() == 0 {
		os.Remove(outputPath)
		return "", fmt.Errorf("output file is empty (piper output: %s)", string(output))
	}

	return outputPath, nil
}

// slowDownAudio reduces the playback speed of an audio file by half using FFmpeg.
// It creates a new temporary file with the slowed-down audio.
//
// Parameters:
//   - inputPath: Path to the input WAV file
//
// Returns:
//   - outputPath: Path to the slowed-down WAV file
//   - err: Any error encountered during processing
func slowDownAudio(inputPath string) (outputPath string, err error) {
	// Create a unique temporary file for the slowed output
	tempFile, err := os.CreateTemp("", "piper-tts-slow-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %w", err)
	}
	outputPath = tempFile.Name()
	tempFile.Close()

	// Use FFmpeg to slow down the audio by half (0.5x speed)
	// The atempo filter can only adjust speed between 0.5 and 2.0
	// For 0.5x speed, we use atempo=0.5
	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-af", "atempo=0.5",
		"-y", // Overwrite output file if it exists
		outputPath,
	)

	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outputPath)
		return "", fmt.Errorf("ffmpeg command failed: %w\nOutput: %s", err, string(output))
	}

	// Verify that the output file was created and has content
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return "", fmt.Errorf("output file not created: %w", err)
	}
	if fileInfo.Size() == 0 {
		os.Remove(outputPath)
		return "", fmt.Errorf("output file is empty")
	}

	return outputPath, nil
}

// HTTP handler for text-to-speech conversion
func handleTextToSpeech(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Get language from query parameter, default to global flag value
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = pronunciationLanguage
	}

	// Get slow parameter to determine if audio should be slowed down
	slow := r.URL.Query().Get("slow") == "true"

	// Validate language
	validLanguages := map[string]bool{
		"russian": true, "portuguese": true, "french": true, "german": true,
		"hindi": true, "romanian": true, "arabic": true, "czech": true,
		"icelandic": true, "kazakh": true, "norwegian": true, "swedish": true,
		"turkish": true, "vietnamese": true, "hungarian": true, "chinese": true,
	}
	if !validLanguages[lang] {
		http.Error(w, fmt.Sprintf("Invalid language '%s'. Allowed: russian, portuguese, french, german, hindi, romanian, arabic, czech, icelandic, kazakh, norwegian, swedish, turkish, vietnamese, hungarian, chinese", lang), http.StatusBadRequest)
		return
	}

	input := string(body)

	// Generate cache key
	cacheKey := fmt.Sprintf("%s:%s:%v", input, lang, slow)

	// Check if audio is already cached
	audioCache.RLock()
	cachedAudio, exists := audioCache.cache[cacheKey]
	audioCache.RUnlock()

	if exists {
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Content-Disposition", "inline")
		w.Write(cachedAudio)
		return
	}

	// Generate normal speed audio
	wavPath, err := textToSpeech(input, lang)
	if err != nil {
		http.Error(w, fmt.Sprintf("TTS error: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(wavPath)

	// If slow mode is requested, apply audio slowdown
	if slow {
		slowWavPath, err := slowDownAudio(wavPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Audio slowdown error: %v", err), http.StatusInternalServerError)
			return
		}
		defer os.Remove(slowWavPath)
		wavPath = slowWavPath
	}

	// Read the WAV file
	wavData, err := os.ReadFile(wavPath)
	if err != nil {
		http.Error(w, "Error reading audio file", http.StatusInternalServerError)
		return
	}

	// Cache the audio
	audioCache.Lock()
	audioCache.cache[cacheKey] = wavData
	audioCache.Unlock()

	// Send the WAV file back to the client
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", "inline")
	w.Write(wavData)
}

// handleCheckSlowAudio checks if slow audio is ready and returns status
func handleCheckSlowAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = pronunciationLanguage
	}

	input := string(body)
	cacheKey := fmt.Sprintf("%s:%s:true", input, lang)

	// Check if slow audio is already cached
	audioCache.RLock()
	_, exists := audioCache.cache[cacheKey]
	audioCache.RUnlock()

	if exists {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ready":true}`)
		return
	}

	// If not cached, start generating it in the background
	go func() {
		// Check again to avoid duplicate processing
		audioCache.RLock()
		_, exists := audioCache.cache[cacheKey]
		audioCache.RUnlock()
		if exists {
			return
		}

		// Generate normal speed audio first
		wavPath, err := textToSpeech(input, lang)
		if err != nil {
			return
		}
		defer os.Remove(wavPath)

		// Generate slow version
		slowWavPath, err := slowDownAudio(wavPath)
		if err != nil {
			return
		}
		defer os.Remove(slowWavPath)

		// Read and cache the slow audio
		slowData, err := os.ReadFile(slowWavPath)
		if err != nil {
			return
		}

		audioCache.Lock()
		audioCache.cache[cacheKey] = slowData
		audioCache.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ready":false}`)
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
	// Parse command-line flags
	ngrokToken := flag.String("ngrok_token", "", "Optional ngrok auth token to expose server publicly")
	ngrokDomain := flag.String("ngrok_domain", "", "Optional ngrok persistent domain (e.g., your-domain.ngrok-free.app)")
	pronunciationLangFlag := flag.String("pronunciation_language", "russian", "TTS pronunciation language (russian, portuguese, romanian, czech)")
	pronunciationLangDropdownFlag := flag.Bool("pronunciation_language_dropdown", true, "Show language dropdown in UI for TTS")
	if !strings.HasPrefix(*ngrokDomain, "http://") && !strings.HasPrefix(*ngrokDomain, "https://") {
		*ngrokDomain = "https://" + *ngrokDomain
	}
	flag.Parse()

	// Extract embedded TTS requirements to temp directory
	log.Println("Initializing TTS requirements...")
	if err := extractEmbeddedRequirements(); err != nil {
		log.Fatalf("Failed to extract TTS requirements: %v", err)
	}

	// Validate and set pronunciation language
	validLanguages := map[string]bool{
		"russian": true, "portuguese": true, "french": true, "german": true,
		"hindi": true, "romanian": true, "arabic": true, "czech": true,
		"icelandic": true, "kazakh": true, "norwegian": true, "swedish": true,
		"turkish": true, "vietnamese": true, "hungarian": true, "chinese": true,
	}
	if !validLanguages[*pronunciationLangFlag] {
		log.Fatalf("Invalid pronunciation language '%s'. Allowed: portuguese, spanish, english, russian", *pronunciationLangFlag)
	}
	pronunciationLanguage = *pronunciationLangFlag
	pronunciationLanguageDropdown = *pronunciationLangDropdownFlag
	log.Printf("TTS pronunciation language set to: %s", pronunciationLanguage)
	log.Printf("TTS language dropdown enabled: %v", pronunciationLanguageDropdown)

	// Set up HTTP routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/to", handleTranslateTo)
	http.HandleFunc("/from", handleTranslateFrom)
	http.HandleFunc("/tts", handleTextToSpeech)
	http.HandleFunc("/tts-check-slow", handleCheckSlowAudio)

	if *ngrokToken != "" {
		// Use ngrok to expose server publicly
		log.Println("Initializing ngrok tunnel...")
		log.Printf("Using auth token: %s...\n", (*ngrokToken)[:10])

		log.Println("Connecting to ngrok service...")

		// Configure endpoint with optional domain
		var listener ngrok.Tunnel
		var err error

		if *ngrokDomain != "" {
			// Strip scheme from domain for WithDomain (it expects just hostname)
			domain := *ngrokDomain
			domain = strings.TrimPrefix(domain, "https://")
			domain = strings.TrimPrefix(domain, "http://")

			log.Printf("Using persistent domain: %s\n", domain)
			log.Println("Establishing tunnel (this may take a few seconds)...")

			// Use a channel to receive the result with timeout
			type result struct {
				listener ngrok.Tunnel
				err      error
			}
			resultChan := make(chan result)
			go func() {
				l, e := ngrok.Listen(context.Background(),
					config.HTTPEndpoint(
						config.WithDomain(domain),
					),
					ngrok.WithAuthtoken(*ngrokToken),
				)
				resultChan <- result{listener: l, err: e}
			}()

			// Wait for completion or timeout
			select {
			case res := <-resultChan:
				listener = res.listener
				err = res.err
			case <-time.After(30 * time.Second):
				log.Fatalf("Failed to start ngrok listener: connection timeout after 30 seconds")
			}
		} else {
			log.Println("Using random ngrok domain")
			log.Println("Establishing tunnel (this may take a few seconds)...")

			// Use a channel to receive the result with timeout
			type result struct {
				listener ngrok.Tunnel
				err      error
			}
			resultChan := make(chan result)
			go func() {
				l, e := ngrok.Listen(context.Background(),
					config.HTTPEndpoint(),
					ngrok.WithAuthtoken(*ngrokToken),
				)
				resultChan <- result{listener: l, err: e}
			}()

			// Wait for completion or timeout
			select {
			case res := <-resultChan:
				listener = res.listener
				err = res.err
			case <-time.After(10 * time.Second):
				log.Fatalf("Failed to start ngrok listener: connection timeout after 10 seconds")
			}
		}

		if err != nil {
			log.Fatalf("Failed to start ngrok listener: %v", err)
		}

		url := listener.URL()
		log.Printf("✓ ngrok tunnel established successfully!\n")
		log.Printf("Public URL: %s\n", url)

		// Open browser with ngrok URL
		go func() {
			time.Sleep(1 * time.Second)
			if err := openBrowser(url); err != nil {
				log.Printf("Could not open browser automatically: %v\n", err)
				log.Printf("Please open your browser and navigate to %s\n", url)
			}
		}()

		log.Println("Server is running. Press Ctrl+C to stop.")
		if err := http.Serve(listener, nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		// Use local server (default behavior)
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
}
