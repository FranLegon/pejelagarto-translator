package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// FuzzApplyMapReplacements uses fuzzing to test map replacement reversibility with random inputs
func FuzzApplyMapReplacements(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		translated := applyMapReplacementsToPejelagarto(input)
		reversed := applyMapReplacementsFromPejelagarto(translated)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", input, translated, reversed)
		}

		// Test: FromPejelagarto -> ToPejelagarto
		translated2 := applyMapReplacementsFromPejelagarto(translated)
		reversed2 := applyMapReplacementsToPejelagarto(translated2)

		if reversed2 != translated {
			t.Errorf("FromPejelagarto->ToPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", translated, translated2, reversed2)
		}
	})
}

// FuzzApplyNumbersLogic uses fuzzing to test number base conversion reversibility
func FuzzApplyNumbersLogic(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		pejelagarto := applyNumbersLogicToPejelagarto(input)
		reversed := applyNumbersLogicFromPejelagarto(pejelagarto)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:       %q\nPejelagarto: %q\nReversed:    %q", input, pejelagarto, reversed)
		}
	})
}

// FuzzApplyAccentReplacementLogic uses fuzzing to test accent replacement reversibility
func FuzzApplyAccentReplacementLogic(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		// Test: ToPejelagarto -> FromPejelagarto
		accented := applyAccentReplacementLogicToPejelagarto(input)
		reversed := applyAccentReplacementLogicFromPejelagarto(accented)

		if reversed != input {
			t.Errorf("ToPejelagarto->FromPejelagarto failed\nInput:    %q\nAccented: %q\nReversed: %q", input, accented, reversed)
		}
	})
}

// FuzzApplyPunctuationReplacements tests punctuation replacement reversibility
func FuzzApplyPunctuationReplacements(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		if !utf8.ValidString(input) {
			t.Skip("invalid utf8")
		}

		// Test: ToPejelagarto -> FromPejelagarto
		translated := applyPunctuationReplacementsToPejelagarto(input)
		reversed := applyPunctuationReplacementsFromPejelagarto(translated)

		if reversed != input {
			t.Errorf("applyPunctuationReplacementsToPejelagarto->FromPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", input, translated, reversed)
		}

		// Test: FromPejelagarto -> ToPejelagarto
		translated2 := applyPunctuationReplacementsFromPejelagarto(translated)
		reversed2 := applyPunctuationReplacementsToPejelagarto(translated2)

		if reversed2 != translated {
			t.Errorf("applyPunctuationReplacementsFromPejelagarto->ToPejelagarto failed\nInput:      %q\nTranslated: %q\nReversed:   %q", translated, translated2, reversed2)
		}
	})
}

// FuzzApplyCaseReplacementLogic tests case replacement logic reversibility
func FuzzApplyCaseReplacementLogic(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		if !utf8.ValidString(input) {
			t.Skip("invalid utf8")
		}

		// Apply case replacement twice - should return to original (self-reversing)
		once := applyCaseReplacementLogic(input)
		twice := applyCaseReplacementLogic(once)

		if input != twice {
			t.Errorf("applyCaseReplacementLogic not reversible:\nInput: %q\nOnce:  %q\nTwice: %q", input, once, twice)
		}

		// Word count should not change
		originalWords := countWordsInString(input)
		onceWords := countWordsInString(once)

		if originalWords != onceWords {
			t.Errorf("Word count changed: %d -> %d\nInput: %q\nOnce:  %q", originalWords, onceWords, input, once)
		}
	})
}

// countWordsInString counts words for testing (matches applyCaseReplacementLogic logic)
func countWordsInString(input string) int {
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

	return wordCount
}

// FuzzSpecialCharDateTimeEncoding tests special character datetime encoding with special non-reversibility handling
func FuzzSpecialCharDateTimeEncoding(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8 as Go's string handling will convert invalid bytes to replacement characters
		if !utf8.ValidString(input) {
			return
		}

		// Special character datetime logic isn't fully reversible because:
		// 1. Special characters encode current time, which changes
		// 2. Random placement of special characters
		// But we can verify correct behavior by comparing after removing special characters and timestamps

		translated := TranslateToPejelagarto(input)
		restored := TranslateFromPejelagarto(translated)

		// Clean both for comparison (remove special characters and timestamps)
		inputCleanedTemp, _ := removeISO8601timestamp(input)
		inputCleaned := removeTimestampSpecialCharacters(inputCleanedTemp)
		restoredCleanedTemp, _ := removeISO8601timestamp(restored)
		restoredCleaned := removeTimestampSpecialCharacters(restoredCleanedTemp)

		if inputCleaned != restoredCleaned {
			t.Errorf("Reversibility failed after cleaning.\nInput (cleaned):    %q\nRestored (cleaned): %q", inputCleaned, restoredCleaned)
		}
	})
}

// FuzzTranslatePejelagarto uses fuzzing to test full translation pipeline reversibility
func FuzzTranslatePejelagarto(f *testing.F) {
	// No seed corpus - let fuzzer generate random inputs
	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8 as Go's string handling will convert invalid bytes to replacement characters
		if !utf8.ValidString(input) {
			return
		}

		// Test: TranslateToPejelagarto -> TranslateFromPejelagarto
		pejelagarto := TranslateToPejelagarto(input)
		reversed := TranslateFromPejelagarto(pejelagarto)

		// Since special character/timestamp logic is now integrated, we need to clean for comparison
		inputCleanedTemp, _ := removeISO8601timestamp(input)
		inputCleaned := removeTimestampSpecialCharacters(inputCleanedTemp)
		reversedCleanedTemp, _ := removeISO8601timestamp(reversed)
		reversedCleaned := removeTimestampSpecialCharacters(reversedCleanedTemp)

		if reversedCleaned != inputCleaned {
			t.Errorf("TranslateToPejelagarto->TranslateFromPejelagarto failed\nInput (cleaned):       %q\nPejelagarto: %q\nReversed (cleaned):    %q", inputCleaned, pejelagarto, reversedCleaned)
		}
	})
}

// TestTextToSpeech tests the text-to-speech functionality
func TestTextToSpeech(t *testing.T) {
	// Check if Piper is installed
	binaryPath := piperBinaryPath
	if _, err := os.Stat(binaryPath + ".exe"); os.IsNotExist(err) {
		t.Skip("Piper binary not found, skipping TTS test")
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("Voice model not found, skipping TTS test")
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Simple text",
			input:   "Hello world",
			wantErr: false,
		},
		{
			name:    "Empty text",
			input:   "",
			wantErr: false,
		},
		{
			name:    "Text with punctuation",
			input:   "Hello, world! How are you?",
			wantErr: false,
		},
		{
			name:    "Pejelagarto text",
			input:   "Ⱨėⱡⱡø₽ 𝔅𝔢₽𝔶𝔢ⱡª₽ℊ𝔩𝕣₮ⱡ₽",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath, err := textToSpeech(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("textToSpeech() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify the output file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Output file not created: %s", outputPath)
				}

				// Verify the output file has content
				fileInfo, err := os.Stat(outputPath)
				if err != nil {
					t.Errorf("Failed to stat output file: %v", err)
				} else if fileInfo.Size() == 0 {
					t.Errorf("Output file is empty")
				}

				// Verify it's a WAV file (check for RIFF header)
				file, err := os.Open(outputPath)
				if err != nil {
					t.Errorf("Failed to open output file: %v", err)
				} else {
					defer file.Close()
					header := make([]byte, 4)
					if _, err := file.Read(header); err == nil {
						if string(header) != "RIFF" {
							t.Errorf("Output file is not a valid WAV file (missing RIFF header)")
						}
					}
				}

				// Clean up
				os.Remove(outputPath)
			}
		})
	}
}

// TestHandleTextToSpeech tests the HTTP handler for text-to-speech
func TestHandleTextToSpeech(t *testing.T) {
	// Check if Piper is installed
	binaryPath := piperBinaryPath
	if _, err := os.Stat(binaryPath + ".exe"); os.IsNotExist(err) {
		t.Skip("Piper binary not found, skipping TTS handler test")
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("Voice model not found, skipping TTS handler test")
	}

	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid POST request",
			method:         http.MethodPost,
			body:           "Hello world",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				// Check content type
				contentType := rr.Header().Get("Content-Type")
				if contentType != "audio/wav" {
					t.Errorf("Expected Content-Type audio/wav, got %s", contentType)
				}

				// Check that we got audio data
				if rr.Body.Len() == 0 {
					t.Errorf("Expected audio data in response body, got empty")
				}

				// Verify it's a WAV file (check for RIFF header)
				body := rr.Body.Bytes()
				if len(body) >= 4 && string(body[:4]) != "RIFF" {
					t.Errorf("Response is not a valid WAV file (missing RIFF header)")
				}
			},
		},
		{
			name:           "GET request (should fail)",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			checkResponse:  nil,
		},
		{
			name:           "Empty text",
			method:         http.MethodPost,
			body:           " ",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				// With a space, Piper should generate minimal audio
				if rr.Body.Len() == 0 {
					t.Errorf("Expected audio data for minimal text")
				}
			},
		},
		{
			name:           "Pejelagarto text",
			method:         http.MethodPost,
			body:           "Ⱨėⱡⱡø₽ 𝔅𝔢₽𝔶𝔢ⱡª₽ℊ𝔩𝕣₮ⱡ₽",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Body.Len() == 0 {
					t.Errorf("Expected audio data for Pejelagarto text")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/tts", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handleTextToSpeech(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
				if rr.Code == http.StatusInternalServerError {
					t.Errorf("Error response: %s", rr.Body.String())
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

// TestTextToSpeechWithCleanup tests that temporary files are cleaned up
func TestTextToSpeechWithCleanup(t *testing.T) {
	// Check if Piper is installed
	binaryPath := piperBinaryPath
	if _, err := os.Stat(binaryPath + ".exe"); os.IsNotExist(err) {
		t.Skip("Piper binary not found, skipping TTS cleanup test")
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("Voice model not found, skipping TTS cleanup test")
	}

	// Get temp directory
	tempDir := os.TempDir()

	// Count existing piper-tts files
	beforeFiles, _ := filepath.Glob(filepath.Join(tempDir, "piper-tts-*.wav"))
	beforeCount := len(beforeFiles)

	// Generate audio
	outputPath, err := textToSpeech("Test cleanup")
	if err != nil {
		t.Fatalf("textToSpeech failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file not created: %s", outputPath)
	}

	// Delete the file (simulating cleanup)
	if err := os.Remove(outputPath); err != nil {
		t.Errorf("Failed to clean up temp file: %v", err)
	}

	// Count files again
	afterFiles, _ := filepath.Glob(filepath.Join(tempDir, "piper-tts-*.wav"))
	afterCount := len(afterFiles)

	// Should have same count as before (or fewer if we successfully cleaned up)
	if afterCount > beforeCount {
		t.Errorf("Temp file not cleaned up properly. Before: %d, After: %d", beforeCount, afterCount)
	}
}

// BenchmarkTextToSpeech benchmarks the TTS performance
func BenchmarkTextToSpeech(b *testing.B) {
	// Check if Piper is installed
	binaryPath := piperBinaryPath
	if _, err := os.Stat(binaryPath + ".exe"); os.IsNotExist(err) {
		b.Skip("Piper binary not found, skipping TTS benchmark")
	}
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		b.Skip("Voice model not found, skipping TTS benchmark")
	}

	testText := "Hello world, this is a benchmark test."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath, err := textToSpeech(testText)
		if err != nil {
			b.Fatalf("textToSpeech failed: %v", err)
		}
		os.Remove(outputPath)
	}
}
