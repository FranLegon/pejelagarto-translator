// Pejelagarto Translator - A reversible fictional language translator
// Build command: go build -o pejelagarto-translator main.go
// Run command: go run main.go
// The server will start on http://localhost:8080

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Accent wheel for vowel modifications
// The wheel loops around when moving forward or backward
// Each level maps base vowel (lowercase) to its accented form
var accentWheel = []map[rune]string{
	// No accent (base vowels)
	{'a': "a", 'e': "e", 'i': "i", 'o': "o", 'u': "u"},
	// Grave accent
	{'a': "à", 'e': "è", 'i': "ì", 'o': "ò", 'u': "ù"},
	// Acute accent
	{'a': "á", 'e': "é", 'i': "í", 'o': "ó", 'u': "ú"},
	// Circumflex accent
	{'a': "â", 'e': "ê", 'i': "î", 'o': "ô", 'u': "û"},
	// Tilde accent
	{'a': "ã", 'e': "ẽ", 'i': "ĩ", 'o': "õ", 'u': "ũ"},
	// Ring above
	{'a': "å", 'e': "e̊", 'i': "i̊", 'o': "o̊", 'u': "ů"},
	// Diaeresis
	{'a': "ä", 'e': "ë", 'i': "ï", 'o': "ö", 'u': "ü"},
	// Macron
	{'a': "ā", 'e': "ē", 'i': "ī", 'o': "ō", 'u': "ū"},
	// Breve
	{'a': "ă", 'e': "ĕ", 'i': "ĭ", 'o': "ŏ", 'u': "ŭ"},
}

// Translation maps for the Pejelagarto language
var (
	// wordMap replaces whole words or syllables
	wordMap = map[string]string{
		"hello":   "jetzo",
		"world":   "vorlag",
		"the":     "ze",
		"is":      "ez",
		"and":     "ung",
		"you":     "yux",
		"are":     "irr",
		"this":    "zez",
		"that":    "zit",
		"will":    "vell",
		"can":     "kin",
		"have":    "jiv",
		"with":    "veez",
		"from":    "frux",
		"they":    "zey",
		"what":    "vit",
		"when":    "ven",
		"where":   "verr",
		"how":     "jov",
		"why":     "vey",
		"good":    "gux",
		"great":   "grit",
		"thank":   "zink",
		"please":  "plix",
		"sorry":   "surry",
		"yes":     "yiz",
		"no":      "nux",
		"maybe":   "mibby",
		"friend":  "frund",
		"love":    "luv",
		"time":    "tym",
		"day":     "diy",
		"night":   "nyt",
		"morning": "murneng",
		"evening": "ivneng",
	}

	// conjunctionMap replaces letter pairs
	conjunctionMap = map[string]string{
		"ch": "jj",
		"sh": "xx",
		"th": "zz",
		"ph": "ff",
		"ck": "kk",
		"ng": "gg",
		"qu": "kv",
	}

	// letterMap replaces single letters (must be invertible)
	letterMap = map[string]string{
		"a": "i",
		"e": "o",
		"i": "a",
		"o": "e",
		"u": "y",
	}
)

// invertMap creates a reverse mapping from the original map
func invertMap(m map[string]string) map[string]string {
	inverted := make(map[string]string)
	for k, v := range m {
		inverted[v] = k
	}
	return inverted
}

// makeBidirectionalMap creates a bidirectional map by adding inverse mappings
// This prevents Pejelagarto sequences in the input from being mis-translated
func makeBidirectionalMap(m map[string]string) map[string]string {
	bidirectional := make(map[string]string)
	// Add original mappings
	for k, v := range m {
		bidirectional[k] = v
	}
	// Add inverse mappings (value maps to itself to prevent re-translation)
	for _, v := range m {
		if _, exists := bidirectional[v]; !exists {
			bidirectional[v] = v // Identity mapping
		}
	}
	return bidirectional
}

// applyMapReplacementsToPejelagarto applies the translation maps in order: word, conjunction, letter
// Uses a greedy approach: at each position, tries to match the longest pattern first
func applyMapReplacementsToPejelagarto(input string) string {
	// Make maps bidirectional to prevent Pejelagarto sequences in input from being mis-translated
	bidirectionalWordMap := makeBidirectionalMap(wordMap)
	bidirectionalConjunctionMap := makeBidirectionalMap(conjunctionMap)

	runes := []rune(input)
	result := []rune{}

	for i := 0; i < len(runes); {
		// Try wordMap first (case-insensitive) - find longest match
		longestWordValue := ""
		longestWordLen := 0
		for key, value := range bidirectionalWordMap {
			keyRunes := []rune(strings.ToLower(key))
			if len(keyRunes) > longestWordLen && i+len(keyRunes) <= len(runes) {
				// Check if we have a match (case-insensitive)
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if unicode.ToLower(runes[i+j]) != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestWordValue = value
					longestWordLen = len(keyRunes)
				}
			}
		}
		if longestWordLen > 0 {
			// Preserve capitalization of first letter
			valueRunes := []rune(longestWordValue)
			if unicode.IsUpper(runes[i]) && len(valueRunes) > 0 {
				valueRunes[0] = unicode.ToUpper(valueRunes[0])
			}
			result = append(result, valueRunes...)
			i += longestWordLen
			continue
		}

		// Try conjunctionMap - find longest match
		longestConjValue := ""
		longestConjLen := 0
		for key, value := range bidirectionalConjunctionMap {
			keyRunes := []rune(key)
			if len(keyRunes) > longestConjLen && i+len(keyRunes) <= len(runes) {
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if runes[i+j] != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestConjValue = value
					longestConjLen = len(keyRunes)
				}
			}
		}
		if longestConjLen > 0 {
			result = append(result, []rune(longestConjValue)...)
			i += longestConjLen
			continue
		}

		// Try letterMap
		charMatched := false
		for key, value := range letterMap {
			if string(runes[i]) == key {
				result = append(result, []rune(value)...)
				charMatched = true
				break
			}
		}
		if charMatched {
			i++
		} else {
			// No match, keep original character
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// applyMapReplacementsFromPejelagarto reverses the translation in reverse order: word, conjunction, letter
// Uses a greedy approach: at each position, tries to match the longest pattern first
func applyMapReplacementsFromPejelagarto(input string) string {
	// Invert the maps
	invertedLetterMap := invertMap(letterMap)
	invertedConjunctionMap := invertMap(conjunctionMap)
	invertedWordMap := invertMap(wordMap)

	runes := []rune(input)
	result := []rune{}

	for i := 0; i < len(runes); {
		// Try inverted wordMap first (case-insensitive) - find longest match
		longestWordValue := ""
		longestWordLen := 0
		for key, value := range invertedWordMap {
			keyRunes := []rune(strings.ToLower(key))
			if len(keyRunes) > longestWordLen && i+len(keyRunes) <= len(runes) {
				// Check if we have a match (case-insensitive)
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if unicode.ToLower(runes[i+j]) != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestWordValue = value
					longestWordLen = len(keyRunes)
				}
			}
		}
		if longestWordLen > 0 {
			// Preserve capitalization of first letter
			valueRunes := []rune(longestWordValue)
			if unicode.IsUpper(runes[i]) && len(valueRunes) > 0 {
				valueRunes[0] = unicode.ToUpper(valueRunes[0])
			}
			result = append(result, valueRunes...)
			i += longestWordLen
			continue
		}

		// Try inverted conjunctionMap - find longest match
		longestConjValue := ""
		longestConjLen := 0
		for key, value := range invertedConjunctionMap {
			keyRunes := []rune(key)
			if len(keyRunes) > longestConjLen && i+len(keyRunes) <= len(runes) {
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if runes[i+j] != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestConjValue = value
					longestConjLen = len(keyRunes)
				}
			}
		}
		if longestConjLen > 0 {
			result = append(result, []rune(longestConjValue)...)
			i += longestConjLen
			continue
		}

		// Try inverted letterMap
		charMatched := false
		for key, value := range invertedLetterMap {
			if string(runes[i]) == key {
				result = append(result, []rune(value)...)
				charMatched = true
				break
			}
		}
		if charMatched {
			i++
		} else {
			// No match, keep original character
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// applyNumbersFromBase10ToBase7 converts all base-10 numbers in the input to base-7
func applyNumbersFromBase10ToBase7(input string) string {
	runes := []rune(input)
	result := []rune{}

	for i := 0; i < len(runes); {
		// Check if current character is a digit
		if unicode.IsDigit(runes[i]) {
			// Collect all consecutive digits
			numStr := ""
			for i < len(runes) && unicode.IsDigit(runes[i]) {
				numStr += string(runes[i])
				i++
			}

			// Convert base-10 to base-7
			var base10Num int
			if _, err := fmt.Sscanf(numStr, "%d", &base10Num); err == nil {
				base7Str := convertToBase7(base10Num)
				result = append(result, []rune(base7Str)...)
			} else {
				// If parsing fails, keep original
				result = append(result, []rune(numStr)...)
			}
		} else {
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// applyNumbersFromBase7ToBase10 converts all base-7 numbers in the input to base-10
func applyNumbersFromBase7ToBase10(input string) string {
	runes := []rune(input)
	result := []rune{}

	for i := 0; i < len(runes); {
		// Check if current character is a base-7 digit (0-6)
		if unicode.IsDigit(runes[i]) {
			// Collect all consecutive base-7 digits
			numStart := i
			numStr := ""
			isValidBase7 := true
			for i < len(runes) && unicode.IsDigit(runes[i]) {
				digit := runes[i]
				if digit >= '0' && digit <= '6' {
					numStr += string(digit)
				} else {
					// Not a valid base-7 number
					isValidBase7 = false
					break
				}
				i++
			}

			// Convert base-7 to base-10
			if isValidBase7 && numStr != "" {
				base10Num := convertFromBase7(numStr)
				result = append(result, []rune(fmt.Sprintf("%d", base10Num))...)
			} else {
				// If not valid base-7, keep original characters
				for j := numStart; j < i; j++ {
					result = append(result, runes[j])
				}
				if i < len(runes) {
					result = append(result, runes[i])
					i++
				}
			}
		} else {
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// convertToBase7 converts a base-10 integer to base-7 string
func convertToBase7(num int) string {
	if num == 0 {
		return "0"
	}

	isNegative := num < 0
	if isNegative {
		num = -num
	}

	result := ""
	for num > 0 {
		digit := num % 7
		result = string(rune('0'+digit)) + result
		num = num / 7
	}

	if isNegative {
		result = "-" + result
	}

	return result
}

// convertFromBase7 converts a base-7 string to base-10 integer
func convertFromBase7(base7Str string) int {
	isNegative := false
	if len(base7Str) > 0 && base7Str[0] == '-' {
		isNegative = true
		base7Str = base7Str[1:]
	}

	result := 0
	multiplier := 1

	// Process from right to left
	for i := len(base7Str) - 1; i >= 0; i-- {
		digit := int(base7Str[i] - '0')
		result += digit * multiplier
		multiplier *= 7
	}

	if isNegative {
		result = -result
	}

	return result
}

// primeFactorize returns the prime factorization of n as a map of prime -> power
func primeFactorize(n int) map[int]int {
	if n <= 1 {
		return map[int]int{}
	}

	factors := make(map[int]int)

	// Handle factor of 2
	for n%2 == 0 {
		factors[2]++
		n = n / 2
	}

	// Handle odd factors from 3 onwards
	for i := 3; i*i <= n; i += 2 {
		for n%i == 0 {
			factors[i]++
			n = n / i
		}
	}

	// If n is still greater than 1, then it's a prime factor
	if n > 1 {
		factors[n]++
	}

	return factors
}

// isVowel checks if a string (potentially multi-byte char) is a vowel (accented or not)
func isVowel(s string) bool {
	// Check base vowels (case-insensitive)
	baseVowels := "aeiou"
	lowerS := strings.ToLower(s)
	if len(lowerS) == 1 && strings.ContainsRune(baseVowels, rune(lowerS[0])) {
		return true
	}

	// Check if the string appears in any accent wheel level (checking both cases)
	for _, accentMap := range accentWheel {
		for _, accentedVowel := range accentMap {
			if s == accentedVowel || lowerS == accentedVowel {
				return true
			}
			// Also check if uppercase version matches
			if s == strings.ToUpper(accentedVowel) {
				return true
			}
		}
	}

	return false
}

// findVowelInWheel returns the wheel index and base vowel for a given string
// Returns (-1, 0) if not found
func findVowelInWheel(s string) (wheelIndex int, baseVowel rune) {
	// Normalize to lowercase for comparison
	lowerS := strings.ToLower(s)

	for idx, accentMap := range accentWheel {
		for base, accented := range accentMap {
			// Compare lowercase versions
			if lowerS == accented || lowerS == strings.ToLower(accented) {
				return idx, base
			}
			// Also check uppercase version of accented vowel
			upperAccented := strings.ToUpper(accented)
			if s == upperAccented {
				return idx, base
			}
		}
	}
	return -1, 0
}

// applyAccentReplacementLogicToPejelagarto applies accent modifications based on prime factorization
// Uses the character count of the input (post-translation text)
func applyAccentReplacementLogicToPejelagarto(input string) string {
	// Count total characters (runes) - this is the POST-TRANSLATION count
	runeCount := utf8.RuneCountInString(input)

	if runeCount <= 1 {
		return input
	}

	// Get prime factorization of the total count
	factors := primeFactorize(runeCount)

	if len(factors) == 0 {
		return input
	}

	// Convert to slice of strings (each character as a string to handle multi-byte)
	chars := []string{}
	for _, r := range input {
		chars = append(chars, string(r))
	}

	// Find all vowel positions
	vowelPositions := []int{}
	for i, char := range chars {
		if isVowel(char) {
			vowelPositions = append(vowelPositions, i)
		}
	}

	if len(vowelPositions) == 0 {
		return input
	}

	// Apply accent changes for each prime factor
	for prime, power := range factors {
		// Calculate which vowel to modify (prime-th vowel, 1-indexed)
		vowelIndex := (prime - 1) % len(vowelPositions)
		charIndex := vowelPositions[vowelIndex]

		currentChar := chars[charIndex]
		wheelIndex, baseVowel := findVowelInWheel(currentChar)

		if wheelIndex == -1 {
			continue // Not a vowel we can modify
		}

		// Check if original was uppercase
		wasUpper := unicode.IsUpper([]rune(currentChar)[0])

		// Move forward in the wheel by 'power' positions
		newWheelIndex := (wheelIndex + power) % len(accentWheel)
		newChar := accentWheel[newWheelIndex][baseVowel]

		// Preserve case - convert to uppercase if needed
		if wasUpper {
			// Convert first rune to uppercase
			runes := []rune(newChar)
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
				newChar = string(runes)
			}
		}

		chars[charIndex] = newChar
	}

	return strings.Join(chars, "")
}

// applyAccentReplacementLogicFromPejelagarto reverses accent modifications
// Uses the character count of the input (pre-reverse-translation text, which is the Pejelagarto text)
func applyAccentReplacementLogicFromPejelagarto(input string) string {
	// Count total characters (runes) - this is the PRE-REVERSE-TRANSLATION count (same as POST-TRANSLATION from encoding)
	runeCount := utf8.RuneCountInString(input)

	if runeCount <= 1 {
		return input
	}

	// Get prime factorization of the total count
	factors := primeFactorize(runeCount)

	if len(factors) == 0 {
		return input
	}

	// Convert to slice of strings (each character as a string to handle multi-byte)
	chars := []string{}
	for _, r := range input {
		chars = append(chars, string(r))
	}

	// Find all vowel positions
	vowelPositions := []int{}
	for i, char := range chars {
		if isVowel(char) {
			vowelPositions = append(vowelPositions, i)
		}
	}

	if len(vowelPositions) == 0 {
		return input
	}

	// Apply accent changes for each prime factor (moving backward)
	for prime, power := range factors {
		// Calculate which vowel to modify (prime-th vowel, 1-indexed)
		vowelIndex := (prime - 1) % len(vowelPositions)
		charIndex := vowelPositions[vowelIndex]

		currentChar := chars[charIndex]
		wheelIndex, baseVowel := findVowelInWheel(currentChar)

		if wheelIndex == -1 {
			continue // Not a vowel we can modify
		}

		// Check if original was uppercase
		wasUpper := unicode.IsUpper([]rune(currentChar)[0])

		// Move backward in the wheel by 'power' positions
		newWheelIndex := (wheelIndex - power) % len(accentWheel)
		if newWheelIndex < 0 {
			newWheelIndex += len(accentWheel)
		}

		// Look up the new character
		newChar, exists := accentWheel[newWheelIndex][baseVowel]
		if !exists {
			// If baseVowel not found, continue (shouldn't happen)
			continue
		}

		// Preserve case - convert to uppercase if needed
		if wasUpper {
			// Convert first rune to uppercase
			runes := []rune(newChar)
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
				newChar = string(runes)
			}
		}

		chars[charIndex] = newChar
	}

	return strings.Join(chars, "")
}

// TranslateToPejelagarto translates Human text to Pejelagarto
func TranslateToPejelagarto(input string) string {
	input = applyNumbersFromBase10ToBase7(input)
	input = applyMapReplacementsToPejelagarto(input)
	return applyAccentReplacementLogicToPejelagarto(input)
}

// TranslateFromPejelagarto translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
	input = applyAccentReplacementLogicFromPejelagarto(input)
	input = applyMapReplacementsFromPejelagarto(input)
	return applyNumbersFromBase7ToBase10(input)
}

// HTML template for the web interface
const htmlTemplate = `<!DOCTYPE html>
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
        }
        .translator-box {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }
        .text-group {
            display: flex;
            flex-direction: column;
        }
        label {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            font-size: 1.1em;
        }
        textarea {
            width: 100%;
            min-height: 200px;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            font-size: 16px;
            font-family: inherit;
            resize: vertical;
            transition: border-color 0.3s;
        }
        textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        textarea[readonly] {
            background-color: #f5f5f5;
            cursor: default;
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
            font-size: 16px;
            font-weight: 600;
            border-radius: 25px;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
        .invert-btn {
            background: white;
            color: #667eea;
            border: 2px solid #667eea;
            padding: 10px 20px;
            font-size: 20px;
        }
        .invert-btn:hover {
            background: #667eea;
            color: white;
        }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        input[type="checkbox"] {
            width: 20px;
            height: 20px;
            cursor: pointer;
        }
        .checkbox-label {
            font-weight: 500;
            color: #333;
            cursor: pointer;
            margin: 0;
        }
        @media (max-width: 768px) {
            .translator-box {
                grid-template-columns: 1fr;
            }
            h1 {
                font-size: 2em;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Pejelagarto Translator</h1>
        <div class="translator-box">
            <div class="text-group">
                <label id="input-label">Human</label>
                <textarea id="input-text" placeholder="Type here..."></textarea>
            </div>
            <div class="text-group">
                <label id="output-label">Pejelagarto</label>
                <textarea id="output-text" readonly placeholder="Translation will appear here..."></textarea>
            </div>
        </div>
        <div class="controls">
            <button 
                id="translate-btn"
                onclick="translateText(); return false;"
            >
                Translate to Pejelagarto
            </button>
            <button class="invert-btn" id="invert-btn">⇅</button>
            <div class="checkbox-group">
                <input type="checkbox" id="live-translate">
                <label for="live-translate" class="checkbox-label">Live Translation</label>
            </div>
        </div>
    </div>

    <script>
        let isInverted = false;
        let liveTranslateEnabled = false;

        const inputText = document.getElementById('input-text');
        const outputText = document.getElementById('output-text');
        const inputLabel = document.getElementById('input-label');
        const outputLabel = document.getElementById('output-label');
        const translateBtn = document.getElementById('translate-btn');
        const invertBtn = document.getElementById('invert-btn');
        const liveTranslateCheckbox = document.getElementById('live-translate');

        // Invert button handler
        invertBtn.addEventListener('click', () => {
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
                translateBtn.setAttribute('hx-post', '/from');
            } else {
                translateBtn.textContent = 'Translate to Pejelagarto';
                translateBtn.setAttribute('hx-post', '/to');
            }

            // Re-initialize HTMX for the button
            htmx.process(translateBtn);
        });

        // Live translation handler
        liveTranslateCheckbox.addEventListener('change', (e) => {
            liveTranslateEnabled = e.target.checked;
            
            if (liveTranslateEnabled) {
                translateBtn.style.display = 'none';
                inputText.addEventListener('input', handleLiveTranslate);
                // Trigger initial translation
                handleLiveTranslate();
            } else {
                translateBtn.style.display = 'block';
                inputText.removeEventListener('input', handleLiveTranslate);
            }
        });

        // Live translate function
        function handleLiveTranslate() {
            const endpoint = isInverted ? '/from' : '/to';
            
            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
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

        // Manual translate function for button
        function translateText() {
            const endpoint = isInverted ? '/from' : '/to';
            
            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
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

// Handler for the root path - serves the HTML UI
func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlTemplate)
}

// Handler for translating TO Pejelagarto
func handleTranslateTo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	input := string(body)
	translated := TranslateToPejelagarto(input)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, translated)
}

// Handler for translating FROM Pejelagarto
func handleTranslateFrom(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	input := string(body)
	translated := TranslateFromPejelagarto(input)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, translated)
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser: %v\n", err)
	}
}

func main() {
	// Set up routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/to", handleTranslateTo)
	http.HandleFunc("/from", handleTranslateFrom)

	// Start server
	addr := ":8080"
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Printf("Starting Pejelagarto Translator server on %s\n", url)
	fmt.Println("Press Ctrl+C to stop the server")

	// Open browser after a short delay to ensure server is ready
	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("Opening %s in your browser...\n", url)
		openBrowser(url)
	}()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
