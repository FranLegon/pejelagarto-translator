//go:build !frontend

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode"

	"pejelagarto-translator/config"
)

// Global TTS configuration variables
// These are set by server_backend.go or server_frontend.go at startup
var pronunciationLanguage string
var pronunciationLanguageDropdown bool

// extractEmbeddedRequirements extracts and runs the get-requirements script to download TTS dependencies
// If singleLanguage is not empty, only that language will be downloaded
func extractEmbeddedRequirements(singleLanguage string) error {
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
	tempRequirementsDir = filepath.Join(baseDir, config.ProjectName(), "requirements")

	// Check what dependencies are missing
	piperExe := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		piperExe += ".exe"
	}

	espeakData := filepath.Join(tempRequirementsDir, "espeak-ng-data")
	piperDir := filepath.Join(tempRequirementsDir, "piper")

	// Check if all critical components exist
	piperExists := false
	espeakExists := false
	piperDirExists := false

	if _, err := os.Stat(piperExe); err == nil {
		piperExists = true
	}
	if info, err := os.Stat(espeakData); err == nil && info.IsDir() {
		espeakExists = true
	}
	if info, err := os.Stat(piperDir); err == nil && info.IsDir() {
		piperDirExists = true
	}

	// Define all available languages
	allLanguages := []string{
		"russian", "portuguese", "french", "german", "hindi", "romanian",
		"swahili", "czech", "icelandic", "kazakh", "norwegian", "swedish",
		"turkish", "vietnamese", "hungarian", "chinese",
	}

	// Determine which languages need to be checked
	var languagesToCheck []string
	if singleLanguage != "" {
		// Only check the single specified language
		languagesToCheck = []string{singleLanguage}
	} else {
		// Check all languages (dropdown enabled or default)
		languagesToCheck = allLanguages
	}

	// Check if all required language models exist
	allLanguagesExist := true
	var missingLanguages []string
	for _, lang := range languagesToCheck {
		langDir := filepath.Join(piperDir, "languages", lang)
		modelFile := filepath.Join(langDir, "model.onnx")
		modelJsonFile := filepath.Join(langDir, "model.onnx.json")

		if _, err := os.Stat(modelFile); os.IsNotExist(err) {
			allLanguagesExist = false
			missingLanguages = append(missingLanguages, lang)
			continue
		}
		if _, err := os.Stat(modelJsonFile); os.IsNotExist(err) {
			allLanguagesExist = false
			if len(missingLanguages) == 0 || missingLanguages[len(missingLanguages)-1] != lang {
				missingLanguages = append(missingLanguages, lang)
			}
		}
	}

	// If all dependencies exist, no need to download
	if piperExists && espeakExists && piperDirExists && allLanguagesExist {
		if !config.Obfuscated() {
			log.Printf("Using cached TTS requirements at: %s", tempRequirementsDir)
		}
		return nil
	}

	if !config.Obfuscated() {
		log.Printf("Downloading TTS requirements to: %s", tempRequirementsDir)
		if !piperExists {
			log.Printf("  - Missing: piper binary")
		}
		if !espeakExists {
			log.Printf("  - Missing: espeak-ng-data")
		}
		if !piperDirExists {
			log.Printf("  - Missing: piper directory (language models)")
		}
		if !allLanguagesExist && len(missingLanguages) > 0 {
			log.Printf("  - Missing language models: %v", missingLanguages)
		}
	}

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(tempRequirementsDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	var scriptContent []byte
	var scriptPath string
	var cmd *exec.Cmd
	var err error

	if runtime.GOOS == "windows" {
		// Use PowerShell script on Windows
		scriptContent, err = embeddedGetRequirements.ReadFile("scripts/requirements/get-requirements.ps1")
		if err != nil {
			return fmt.Errorf("failed to read embedded PowerShell script: %w", err)
		}

		// Create a modified version of the script that uses tempRequirementsDir
		modifiedScript := strings.Replace(string(scriptContent),
			`$RequirementsDir = Join-Path $PSScriptRoot "tts\requirements"`,
			`$RequirementsDir = "`+tempRequirementsDir+`"`,
			1)

		// Write the modified script to a temporary file with UTF-8 BOM
		scriptPath = filepath.Join(baseDir, config.ScriptSuffix()+".ps1")
		// Add UTF-8 BOM to help PowerShell parse the file correctly
		utf8Bom := []byte{0xEF, 0xBB, 0xBF}
		scriptBytes := append(utf8Bom, []byte(modifiedScript)...)
		if err := os.WriteFile(scriptPath, scriptBytes, 0755); err != nil {
			return fmt.Errorf("failed to write temporary PowerShell script: %w", err)
		}
		defer os.Remove(scriptPath) // Clean up script after execution

		// Execute the PowerShell script
		if !config.Obfuscated() {
			if singleLanguage != "" {
				log.Printf("Running PowerShell script to download dependencies for language: %s...\n", singleLanguage)
			} else {
				log.Println("Running PowerShell script to download all dependencies...")
			}
		}
		if singleLanguage != "" {
			cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", scriptPath, "-Language", singleLanguage)
		} else {
			cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
		}
	} else {
		// Use shell script on Linux/macOS
		scriptContent, err = embeddedGetRequirements.ReadFile("scripts/requirements/get-requirements.sh")
		if err != nil {
			return fmt.Errorf("failed to read embedded shell script: %w", err)
		}

		// Create a modified version of the script that uses tempRequirementsDir
		modifiedScript := strings.Replace(string(scriptContent),
			`REQUIREMENTS_DIR="$(dirname "$0")/tts/requirements"`,
			`REQUIREMENTS_DIR="`+tempRequirementsDir+`"`,
			1)

		// Write the modified script to a temporary file
		scriptPath = filepath.Join(baseDir, config.ScriptSuffix()+".sh")
		if err := os.WriteFile(scriptPath, []byte(modifiedScript), 0755); err != nil {
			return fmt.Errorf("failed to write temporary shell script: %w", err)
		}
		defer os.Remove(scriptPath) // Clean up script after execution

		// Execute the shell script
		if !config.Obfuscated() {
			if singleLanguage != "" {
				log.Printf("Running shell script to download dependencies for language: %s...\n", singleLanguage)
			} else {
				log.Println("Running shell script to download all dependencies...")
			}
		}
		if singleLanguage != "" {
			cmd = exec.Command("bash", scriptPath, singleLanguage)
		} else {
			cmd = exec.Command("bash", scriptPath)
		}
	}

	// Capture output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute requirements script: %w", err)
	}

	if !config.Obfuscated() {
		log.Println("TTS requirements downloaded successfully")
	}

	return nil
}

// Audio cache for storing normal and slow versions
var audioCache = struct {
	sync.RWMutex
	cache map[string][]byte // key: text+lang hash -> audio data
}{
	cache: make(map[string][]byte),
}

func getPiperBinaryPath() string {
	binaryPath := filepath.Join(tempRequirementsDir, "piper")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	return binaryPath
}

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
func slowDownAudio(inputPath string) (outputPath string, err error) {
	// Create a unique temporary file for the slowed output
	tempFile, err := os.CreateTemp("", "piper-tts-slow-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %w", err)
	}
	outputPath = tempFile.Name()
	tempFile.Close()

	// Use FFmpeg to slow down the audio by half (0.5x speed)
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

func handlePronunciation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	input := string(body)

	// Get language from query parameter, default to pronunciationLanguage global
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = pronunciationLanguage
	}

	// Return the preprocessed text as pronunciation
	pronunciation := preprocessTextForTTS(input, lang)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, pronunciation)
}
