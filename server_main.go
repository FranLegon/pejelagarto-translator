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
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"

	"pejelagarto-translator/obfuscation"
)

// Global TTS configuration variables are now declared in tts.go

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

// handleIsDownloadable returns JSON indicating if this build supports downloads
func handleIsDownloadable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if isDownloadable {
		fmt.Fprint(w, `{"downloadable": true}`)
	} else {
		fmt.Fprint(w, `{"downloadable": false}`)
	}
}

// handleDownloadWindows serves the embedded Windows binary
func handleDownloadWindows(w http.ResponseWriter, r *http.Request) {
	if !isDownloadable {
		http.Error(w, "Downloads not available in this build", http.StatusNotFound)
		return
	}

	data, err := embeddedBinaries.ReadFile("bin/pejelagarto-translator.exe")
	if err != nil {
		http.Error(w, "Windows binary not found", http.StatusNotFound)
		log.Printf("Error reading Windows binary: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=pejelagarto-translator.exe")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

// handleDownloadLinux serves the embedded Linux/Mac binary
func handleDownloadLinux(w http.ResponseWriter, r *http.Request) {
	if !isDownloadable {
		http.Error(w, "Downloads not available in this build", http.StatusNotFound)
		return
	}

	data, err := embeddedBinaries.ReadFile("bin/pejelagarto-translator")
	if err != nil {
		http.Error(w, "Linux/Mac binary not found", http.StatusNotFound)
		log.Printf("Error reading Linux/Mac binary: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=pejelagarto-translator")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
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
	var ngrokToken *string
	var ngrokDomain *string

	if useNgrokDefault {
		// Use hardcoded values for ngrok_default builds
		token := defaultNgrokToken
		domain := defaultNgrokDomain
		ngrokToken = &token
		ngrokDomain = &domain
		if !obfuscation.Obfuscated() {
			log.Println("Using hardcoded ngrok configuration (ngrok_default build)")
		}
	} else {
		// Use command-line flags for regular builds
		ngrokToken = flag.String("ngrok_token", "", getFlagUsage("Optional ngrok auth token to expose server publicly"))
		ngrokDomain = flag.String("ngrok_domain", "", getFlagUsage("Optional ngrok persistent domain (e.g., your-domain.ngrok-free.app)"))
	}

	pronunciationLangFlag := flag.String("pronunciation_language", "russian", getFlagUsage("TTS pronunciation language (russian, portuguese, romanian, czech)"))
	pronunciationLangDropdownFlag := flag.Bool("pronunciation_language_dropdown", true, getFlagUsage("Show language dropdown in UI for TTS"))

	flag.Parse()

	if !strings.HasPrefix(*ngrokDomain, "http://") && !strings.HasPrefix(*ngrokDomain, "https://") {
		*ngrokDomain = "https://" + *ngrokDomain
	}

	// Extract embedded TTS requirements to temp directory
	log.Println("Initializing TTS requirements...")
	var languageToDownload string
	if !*pronunciationLangDropdownFlag {
		// Dropdown is disabled, download only the selected language
		languageToDownload = *pronunciationLangFlag
	}
	// If dropdown is enabled, languageToDownload remains empty and all languages are downloaded
	if err := extractEmbeddedRequirements(languageToDownload); err != nil {
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
	http.HandleFunc("/api/is-downloadable", handleIsDownloadable)
	http.HandleFunc("/download/windows", handleDownloadWindows)
	http.HandleFunc("/download/linux", handleDownloadLinux)

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
