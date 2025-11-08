//go:build !frontend

package main

// This file contains the main() function and server-specific code
// When building with -tags frontend, this file is excluded and wasm_main.go's main() is used instead

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"

	"pejelagarto-translator/obfuscation"
)

var pronunciationLanguage string
var pronunciationLanguageDropdown bool

// Audio cache for storing normal and slow versions
var audioCache = struct {
	sync.RWMutex
	cache map[string][]byte // key: text+lang hash -> audio data
}{
	cache: make(map[string][]byte),
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Conditionally inject the language dropdown
	html := htmlUI
	if pronunciationLanguageDropdown {
		dropdownHTML := ` <select id="tts-language" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;">
                    <option value="russian">North</option>
                    <option value="kazakh">North-North-East</option>
                    <option value="german">North-East</option>
                    <option value="turkish">North-East-East</option>
                    <option value="portuguese">East</option>
                    <option value="french">South-East-East</option>
                    <option value="hindi">South-East</option>
                    <option value="icelandic">South-South-East</option>
                    <option value="romanian">South</option>
                    <option value="vietnamese">South-South-West</option>
                    <option value="swahili">South-West</option>
                    <option value="swedish">South-West-West</option>
                    <option value="czech">West</option>
                    <option value="chinese">North-West-West</option>
                    <option value="norwegian">North-West</option>
                    <option value="hungarian">North-North-West</option>
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


// getPiperBinaryPath returns the path to the Piper binary
func getPiperBinaryPath() string {
	binaryPath := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	return binaryPath
}

// getModelPath returns the language-specific model path
// getModelPath returns the language-specific model path
// Note: language parameter must be validated by caller against validLanguages whitelist
// to prevent path traversal attacks
func getModelPath(language string) string {
	// Defensive: ensure language contains only alphanumeric characters (no path separators)
	// This is a defense-in-depth measure; validation should happen at the handler level
	for _, c := range language {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
			// Invalid character in language name - return empty to cause error
			return ""
		}
	}
	return filepath.Join(tempRequirementsDir, "piper", "languages", language, "model.onnx")
}

// getBaseVowelForTTS returns the base (unaccented) vowel for a given character
// Returns 0 if the character is not a vowel
func getBaseVowelForTTS(r rune) rune {
	// Map of accented vowels to their base forms
	accentMap := map[rune]rune{
		// a variations
		'á': 'a', 'à': 'a', 'â': 'a', 'ä': 'a', 'ã': 'a', 'å': 'a', 'ā': 'a', 'ă': 'a', 'ą': 'a', 'ǎ': 'a', 'ắ': 'a', 'ằ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a', 'ấ': 'a', 'ầ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
		// e variations
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'ĕ': 'e', 'ė': 'e', 'ę': 'e', 'ě': 'e', 'ế': 'e', 'ề': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
		// i variations
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i', 'ĩ': 'i', 'ī': 'i', 'ĭ': 'i', 'į': 'i', 'ı': 'i', 'ǐ': 'i', 'ỉ': 'i', 'ị': 'i',
		// o variations
		'ó': 'o', 'ò': 'o', 'ô': 'o', 'ö': 'o', 'õ': 'o', 'ō': 'o', 'ŏ': 'o', 'ő': 'o', 'ø': 'o', 'ǒ': 'o', 'ơ': 'o', 'ố': 'o', 'ồ': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o', 'ớ': 'o', 'ờ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
		// u variations
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ũ': 'u', 'ū': 'u', 'ŭ': 'u', 'ů': 'u', 'ű': 'u', 'ų': 'u', 'ư': 'u', 'ǔ': 'u', 'ǖ': 'u', 'ǘ': 'u', 'ǚ': 'u', 'ǜ': 'u', 'ứ': 'u', 'ừ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
		// y variations
		'ý': 'y', 'ỳ': 'y', 'ŷ': 'y', 'ÿ': 'y', 'ȳ': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
		// w variations (less common but included for completeness)
		'ẃ': 'w', 'ẁ': 'w', 'ŵ': 'w', 'ẅ': 'w',
		// Special characters
		'æ': 'e', 'œ': 'e', 'ð': 'e', 'þ': 'e',
		// Cyrillic vowels (for Russian, Kazakh, etc.)
		'а': 'a', 'е': 'e', 'ё': 'e', 'и': 'i', 'о': 'o', 'у': 'u', 'ы': 'y', 'э': 'e', 'ю': 'u', 'я': 'a',
		'ә': 'a', 'ө': 'o', 'ұ': 'u', 'ү': 'u', 'і': 'i',
	}

	// Check if it's already a base vowel
	baseVowels := "aeiouywaeiouyw"
	if strings.ContainsRune(baseVowels, r) {
		return r
	}

	// Check if we have a mapping for this accented vowel
	if base, ok := accentMap[r]; ok {
		return base
	}

	// Not a vowel
	return 0
}

// preprocessTextForTTS prepares text for TTS by:
// 1. Converting numbers from Pejelagarto format to standard format
// 2. Converting accented vowels to base vowels when accent not available in language
// 3. Removing non-language-specific characters
// 4. Limiting consonant clusters to max 2 adjacent consonants
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
	case "swahili":
		vowels = "aeiou"
		consonants = "bcdfghjklmnpqrstvwxyz"
		allowed = vowels + consonants + "AEIOUБCDFGHJKLMNPQRSTVWXYZ" + "0123456789" + " .,!?;:'\"-()[]"
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

	// Step 1: Convert to lowercase and handle characters
	var cleaned strings.Builder
	for _, r := range input {
		if strings.ContainsRune(allowed, r) {
			// Convert letters to lowercase, keep numbers and punctuation as-is
			if unicode.IsLetter(r) {
				cleaned.WriteRune(unicode.ToLower(r))
			} else {
				cleaned.WriteRune(r)
			}
		} else if unicode.IsLetter(r) {
			// Check if it's a vowel (accented or not)
			lowerR := unicode.ToLower(r)
			baseVowel := getBaseVowelForTTS(lowerR)

			// If it's a vowel and the base vowel is allowed, use the base vowel
			if baseVowel != 0 && strings.ContainsRune(vowels, baseVowel) {
				cleaned.WriteRune(baseVowel)
			}
			// Otherwise, skip this character (consonant not in allowed set)
		} else if !unicode.IsLetter(r) {
			// For non-letter characters (punctuation, numbers, spaces)
			// only include if they're in the allowed set
			if strings.ContainsRune(allowed, r) {
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
		"hindi": true, "romanian": true, "swahili": true, "czech": true,
		"icelandic": true, "kazakh": true, "norwegian": true, "swedish": true,
		"turkish": true, "vietnamese": true, "hungarian": true, "chinese": true,
	}
	if !validLanguages[lang] {
		http.Error(w, fmt.Sprintf("Invalid language '%s'. Allowed: russian, portuguese, french, german, hindi, romanian, swahili, czech, icelandic, kazakh, norwegian, swedish, turkish, vietnamese, hungarian, chinese", lang), http.StatusBadRequest)
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

// validateConstants performs comprehensive validation of all translation constants
// This function should be called before starting the server to ensure data integrity
func validateConstants() error {
	// 1. Validate conjunctionMap: len(key) == len(value) for every pair
	for key, value := range conjunctionMap {
		keyLen := utf8.RuneCountInString(key)
		valueLen := utf8.RuneCountInString(value)
		if keyLen != valueLen {
			return fmt.Errorf("conjunctionMap: key %q (len=%d) and value %q (len=%d) must have equal rune lengths",
				key, keyLen, value, valueLen)
		}
	}

	// 2. Validate letterMap: len(key) == 1 && len(value) == 1
	for key, value := range letterMap {
		keyLen := utf8.RuneCountInString(key)
		valueLen := utf8.RuneCountInString(value)
		if keyLen != 1 {
			return fmt.Errorf("letterMap: key %q must be exactly 1 rune, got %d", key, keyLen)
		}
		if valueLen != 1 {
			return fmt.Errorf("letterMap: value %q for key %q must be exactly 1 rune, got %d", value, key, valueLen)
		}
	}

	// 3. Validate escape characters are not in punctuationMap
	// Build a set of all characters in punctuationMap (keys and values)
	punctuationChars := make(map[rune]bool)
	for key, value := range punctuationMap {
		for _, r := range key {
			punctuationChars[r] = true
		}
		for _, r := range value {
			punctuationChars[r] = true
		}
	}

	if punctuationChars[internalEscapeChar] {
		return fmt.Errorf("internalEscapeChar %q found in punctuationMap (not allowed)", internalEscapeChar)
	}
	if punctuationChars[outputEscapeChar] {
		return fmt.Errorf("outputEscapeChar %q found in punctuationMap (not allowed)", outputEscapeChar)
	}

	// 4. Validate special char indices: no repeated characters
	// Collect all special characters into a single map to check for duplicates
	allSpecialChars := make(map[string]string) // map[character]source (e.g., "daySpecialCharIndex[3]")

	// Helper function to check for duplicates in an array
	checkDuplicates := func(arr []string, arrayName string) error {
		for i, char := range arr {
			if existingSource, exists := allSpecialChars[char]; exists {
				return fmt.Errorf("duplicate character %q found in %s[%d] and %s",
					char, arrayName, i, existingSource)
			}
			allSpecialChars[char] = fmt.Sprintf("%s[%d]", arrayName, i)
		}
		return nil
	}

	if err := checkDuplicates(daySpecialCharIndex, "daySpecialCharIndex"); err != nil {
		return err
	}
	if err := checkDuplicates(monthSpecialCharIndex, "monthSpecialCharIndex"); err != nil {
		return err
	}
	if err := checkDuplicates(yearSpecialCharIndex, "yearSpecialCharIndex"); err != nil {
		return err
	}
	if err := checkDuplicates(hourSpecialCharIndex, "hourSpecialCharIndex"); err != nil {
		return err
	}
	if err := checkDuplicates(minuteSpecialCharIndex, "minuteSpecialCharIndex"); err != nil {
		return err
	}

	// 5. Validate escape characters are not in special char indices
	if _, exists := allSpecialChars[string(internalEscapeChar)]; exists {
		return fmt.Errorf("internalEscapeChar %q found in special char indices (not allowed)", internalEscapeChar)
	}
	if _, exists := allSpecialChars[string(outputEscapeChar)]; exists {
		return fmt.Errorf("outputEscapeChar %q found in special char indices (not allowed)", outputEscapeChar)
	}

	// 6. Validate languages are correctly mapped to their labels in HTML dropdown
	// Extract the dropdown HTML to parse language mappings
	dropdownHTML := `<option value="russian">North</option>
                    <option value="kazakh">North-North-East</option>
                    <option value="german">North-East</option>
                    <option value="turkish">North-East-East</option>
                    <option value="portuguese">East</option>
                    <option value="french">South-East-East</option>
                    <option value="hindi">South-East</option>
                    <option value="icelandic">South-South-East</option>
                    <option value="romanian">South</option>
                    <option value="vietnamese">South-South-West</option>
                    <option value="swahili">South-West</option>
                    <option value="swedish">South-West-West</option>
                    <option value="czech">West</option>
                    <option value="chinese">North-West-West</option>
                    <option value="norwegian">North-West</option>
                    <option value="hungarian">North-North-West</option>`

	// Define expected language mappings (value -> label)
	expectedLanguageMappings := map[string]string{
		"russian":    "North",
		"kazakh":     "North-North-East",
		"german":     "North-East",
		"turkish":    "North-East-East",
		"portuguese": "East",
		"french":     "South-East-East",
		"hindi":      "South-East",
		"icelandic":  "South-South-East",
		"romanian":   "South",
		"vietnamese": "South-South-West",
		"swahili":    "South-West",
		"swedish":    "South-West-West",
		"czech":      "West",
		"chinese":    "North-West-West",
		"norwegian":  "North-West",
		"hungarian":  "North-North-West",
	}

	// Parse dropdown HTML to extract actual mappings
	optionPattern := regexp.MustCompile(`<option value="([^"]+)">([^<]+)</option>`)
	matches := optionPattern.FindAllStringSubmatch(dropdownHTML, -1)

	actualMappings := make(map[string]string)
	for _, match := range matches {
		if len(match) == 3 {
			value := match[1]
			label := match[2]
			actualMappings[value] = label
		}
	}

	// Compare expected vs actual mappings
	for lang, expectedLabel := range expectedLanguageMappings {
		actualLabel, exists := actualMappings[lang]
		if !exists {
			return fmt.Errorf("language %q missing from HTML dropdown", lang)
		}
		if actualLabel != expectedLabel {
			return fmt.Errorf("language %q has incorrect label in HTML dropdown: expected %q, got %q",
				lang, expectedLabel, actualLabel)
		}
	}

	// Check for extra languages in dropdown that shouldn't be there
	for lang := range actualMappings {
		if _, expected := expectedLanguageMappings[lang]; !expected {
			return fmt.Errorf("unexpected language %q found in HTML dropdown", lang)
		}
	}

	// 7. Validate letterMap bijectivity (no duplicate values, and reverse mapping exists)
	letterMapValues := make(map[string]string) // map[value]key
	for key, value := range letterMap {
		if existingKey, exists := letterMapValues[value]; exists {
			return fmt.Errorf("letterMap: duplicate value %q for keys %q and %q (not bijective)", value, existingKey, key)
		}
		letterMapValues[value] = key
	}
	// Check that every value has a corresponding reverse mapping
	for value, originalKey := range letterMapValues {
		reverseValue, exists := letterMap[value]
		if !exists {
			return fmt.Errorf("letterMap: value %q (from key %q) has no reverse mapping (not bijective)", value, originalKey)
		}
		if reverseValue != originalKey {
			return fmt.Errorf("letterMap: reverse mapping broken: %q -> %q -> %q, expected %q -> %q -> %q",
				originalKey, value, reverseValue, originalKey, value, originalKey)
		}
	}

	// 8. Validate special character array lengths
	if len(daySpecialCharIndex) != 31 {
		return fmt.Errorf("daySpecialCharIndex must have exactly 31 elements, got %d", len(daySpecialCharIndex))
	}
	if len(monthSpecialCharIndex) != 12 {
		return fmt.Errorf("monthSpecialCharIndex must have exactly 12 elements, got %d", len(monthSpecialCharIndex))
	}
	if len(yearSpecialCharIndex) != 100 {
		return fmt.Errorf("yearSpecialCharIndex must have exactly 100 elements, got %d", len(yearSpecialCharIndex))
	}
	if len(hourSpecialCharIndex) != 24 {
		return fmt.Errorf("hourSpecialCharIndex must have exactly 24 elements, got %d", len(hourSpecialCharIndex))
	}
	if len(minuteSpecialCharIndex) != 60 {
		return fmt.Errorf("minuteSpecialCharIndex must have exactly 60 elements, got %d", len(minuteSpecialCharIndex))
	}

	// 9. Validate punctuationMap bijectivity (no duplicate values)
	punctuationMapValues := make(map[string]string) // map[value]key
	for key, value := range punctuationMap {
		if existingKey, exists := punctuationMapValues[value]; exists {
			return fmt.Errorf("punctuationMap: duplicate value %q for keys %q and %q (not bijective)", value, existingKey, key)
		}
		punctuationMapValues[value] = key
	}

	// 10. Validate accent wheels completeness (all base vowels present)
	baseVowels := []rune{'a', 'e', 'i', 'o', 'u', 'y', 'w'}
	for _, vowel := range baseVowels {
		if _, exists := oneRuneAccentsWheel[vowel]; !exists {
			return fmt.Errorf("oneRuneAccentsWheel: missing base vowel %q", vowel)
		}
		if _, exists := twoRunesAccentsWheel[vowel]; !exists {
			return fmt.Errorf("twoRunesAccentsWheel: missing base vowel %q", vowel)
		}
	}

	// 11. Validate rune count for each value in accent wheels
	for baseVowel, accents := range oneRuneAccentsWheel {
		for idx, accentedForm := range accents {
			runeCount := utf8.RuneCountInString(accentedForm)
			if runeCount != 1 {
				return fmt.Errorf("oneRuneAccentsWheel[%q][%d] = %q has %d runes, expected 1",
					baseVowel, idx, accentedForm, runeCount)
			}
		}
	}
	for baseVowel, accents := range twoRunesAccentsWheel {
		for idx, accentedForm := range accents {
			runeCount := utf8.RuneCountInString(accentedForm)
			if runeCount != 2 {
				return fmt.Errorf("twoRunesAccentsWheel[%q][%d] = %q has %d runes, expected 2",
					baseVowel, idx, accentedForm, runeCount)
			}
		}
	}

	// 12. Validate escape characters are unique
	if internalEscapeChar == outputEscapeChar {
		return fmt.Errorf("internalEscapeChar and outputEscapeChar must be different, both are %q", internalEscapeChar)
	}

	// 13. Validate special chars don't overlap with letterMap or conjunctionMap
	// Build set of all letters used in letterMap (keys and values)
	letterMapChars := make(map[rune]bool)
	for key, value := range letterMap {
		for _, r := range key {
			letterMapChars[r] = true
		}
		for _, r := range value {
			letterMapChars[r] = true
		}
	}
	// Build set of all characters used in conjunctionMap (keys and values)
	conjunctionMapChars := make(map[rune]bool)
	for key, value := range conjunctionMap {
		for _, r := range key {
			conjunctionMapChars[r] = true
		}
		for _, r := range value {
			conjunctionMapChars[r] = true
		}
	}
	// Check special char indices don't contain letterMap or conjunctionMap characters
	for char, source := range allSpecialChars {
		for _, r := range char {
			if letterMapChars[r] {
				return fmt.Errorf("special character %q in %s conflicts with letterMap character %q", char, source, r)
			}
			if conjunctionMapChars[r] {
				return fmt.Errorf("special character %q in %s conflicts with conjunctionMap character %q", char, source, r)
			}
		}
	}

	return nil
}

// getFlagUsage returns the usage string for flags based on build mode
// Returns the actual usage for normal builds, empty string for obfuscated builds
func getFlagUsage(usage string) string {
	if !obfuscation.Obfuscated() {
		return usage
	}
	return ""
}


// HTTP handler for the main UI
func main() {
	// Disable -help flag for obfuscated builds
	if obfuscation.Obfuscated() {
		flag.Usage = func() {}
	}

	// Validate all constants before starting the server
	if err := validateConstants(); err != nil {
		log.Fatalf("Constants validation failed: %v", err)
	}
	log.Println("Constants validation passed ✓")

	// Parse command-line flags
	ngrokToken := flag.String("ngrok_token", "", getFlagUsage("Optional ngrok auth token to expose server publicly"))
	ngrokDomain := flag.String("ngrok_domain", "", getFlagUsage("Optional ngrok persistent domain (e.g., your-domain.ngrok-free.app)"))
	pronunciationLangFlag := flag.String("pronunciation_language", "russian", getFlagUsage("TTS pronunciation language (russian, portuguese, romanian, czech)"))
	pronunciationLangDropdownFlag := flag.Bool("pronunciation_language_dropdown", true, getFlagUsage("Show language dropdown in UI for TTS"))
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
		"hindi": true, "romanian": true, "swahili": true, "czech": true,
		"icelandic": true, "kazakh": true, "norwegian": true, "swedish": true,
		"turkish": true, "vietnamese": true, "hungarian": true, "chinese": true,
	}
	if !validLanguages[*pronunciationLangFlag] {
		log.Fatalf("Invalid pronunciation language '%s'. Allowed: portuguese, spanish, english, russian", *pronunciationLangFlag)
	}
	pronunciationLanguage = *pronunciationLangFlag
	pronunciationLanguageDropdown = *pronunciationLangDropdownFlag
	if !obfuscation.Obfuscated() {
		log.Printf("TTS pronunciation language set to: %s", pronunciationLanguage)
		log.Printf("TTS language dropdown enabled: %v", pronunciationLanguageDropdown)
	}

	// Set up HTTP routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/to", handleTranslateTo)
	http.HandleFunc("/from", handleTranslateFrom)
	http.HandleFunc("/tts", handleTextToSpeech)
	http.HandleFunc("/tts-check-slow", handleCheckSlowAudio)

	if *ngrokToken != "" {
		// Use ngrok to expose server publicly
		if !obfuscation.Obfuscated() {
			log.Println("Initializing ngrok tunnel...")
			log.Printf("Using auth token: %s...\n", (*ngrokToken)[:10])
			log.Println("Connecting to ngrok service...")
		}

		// Configure endpoint with optional domain
		var listener ngrok.Tunnel
		var err error

		if *ngrokDomain != "" {
			// Strip scheme from domain for WithDomain (it expects just hostname)
			domain := *ngrokDomain
			domain = strings.TrimPrefix(domain, "https://")
			domain = strings.TrimPrefix(domain, "http://")

			if !obfuscation.Obfuscated() {
				log.Printf("Using persistent domain: %s\n", domain)
				log.Println("Establishing tunnel (this may take a few seconds)...")
			}

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
			if !obfuscation.Obfuscated() {
				log.Println("Using random ngrok domain")
				log.Println("Establishing tunnel (this may take a few seconds)...")
			}

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
		if !obfuscation.Obfuscated() {
			log.Printf("ngrok tunnel established successfully! ✓\n")
			log.Printf("Public URL: %s\n", url)
		}

		// Open browser with ngrok URL (only if configured to do so)
		if obfuscation.ShouldOpenBrowser() {
			go func() {
				time.Sleep(1 * time.Second)
				if err := openBrowser(url); err != nil {
					if !obfuscation.Obfuscated() {
						log.Printf("Could not open browser automatically: %v\n", err)
						log.Printf("Please open your browser and navigate to %s\n", url)
					}
				}
			}()
		}

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
			if !obfuscation.Obfuscated() {
				log.Printf("Starting Pejelagarto Translator server on %s\n", url)
			}
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Fatalf("Server failed to start: %v", err)
			}
		}()

		// Wait a moment for server to start, then open browser (only if configured to do so)
		time.Sleep(500 * time.Millisecond)
		if obfuscation.ShouldOpenBrowser() {
			if err := openBrowser(url); err != nil {
				if !obfuscation.Obfuscated() {
					log.Printf("Could not open browser automatically: %v\n", err)
					log.Printf("Please open your browser and navigate to %s\n", url)
				}
			}
		}

		// Keep the server running
		log.Println("Server is running. Press Ctrl+C to stop.")
		select {} // Block forever
	}
}
