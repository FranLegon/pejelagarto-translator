//go:build !frontend

package tts

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzTextToSpeech uses fuzzing to test text-to-speech functionality with random inputs
func FuzzTextToSpeech(f *testing.F) {
	// Check if Piper is installed
	binaryPath := getPiperBinaryPath()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		f.Skip("Piper binary not found, skipping TTS test")
	}
	modelPath := getModelPath("english")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		f.Skip("Voice model not found, skipping TTS test")
	}

	// Seed corpus with basic cases
	f.Add("")
	f.Add("Hello world")
	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(input) {
			return
		}

		outputPath, err := textToSpeech(input, "english")

		// TTS should never error for valid UTF-8 input
		if err != nil {
			t.Errorf("textToSpeech() unexpected error: %v\nInput: %q", err, input)
			return
		}

		// Verify the output file exists
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("textToSpeech() output file not created: %s", outputPath)
		}

		// Clean up the output file
		os.Remove(outputPath)
	})
}

// TestHandleTextToSpeech tests the HTTP handler for text-to-speech
func TestHandleTextToSpeech(t *testing.T) {
	// Check if Piper is installed
	binaryPath := getPiperBinaryPath()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Piper binary not found, skipping TTS handler test")
	}
	modelPath := getModelPath("russian")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("Voice model not found, skipping TTS handler test")
	}

	// Set global variables
	PronunciationLanguage = "russian"

	testCases := []struct {
		name           string
		method         string
		body           string
		lang           string
		expectedStatus int
		expectAudio    bool
	}{
		{
			name:           "Valid request",
			method:         http.MethodPost,
			body:           "test",
			lang:           "russian",
			expectedStatus: http.StatusOK,
			expectAudio:    true,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			body:           "",
			lang:           "",
			expectedStatus: http.StatusMethodNotAllowed,
			expectAudio:    false,
		},
		{
			name:           "Invalid language",
			method:         http.MethodPost,
			body:           "test",
			lang:           "invalid",
			expectedStatus: http.StatusBadRequest,
			expectAudio:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/tts?lang="+tc.lang, strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			HandleTextToSpeech(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.expectAudio && w.Header().Get("Content-Type") != "audio/wav" {
				t.Errorf("Expected Content-Type audio/wav, got %s", w.Header().Get("Content-Type"))
			}
		})
	}
}

// TestTextToSpeechWithCleanup tests that temporary files are cleaned up properly
func TestTextToSpeechWithCleanup(t *testing.T) {
	// Check if Piper is installed
	binaryPath := getPiperBinaryPath()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Piper binary not found, skipping TTS cleanup test")
	}
	modelPath := getModelPath("russian")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("Voice model not found, skipping TTS cleanup test")
	}

	// Get initial temp file count
	tempDir := os.TempDir()
	initialFiles, err := filepath.Glob(filepath.Join(tempDir, "piper-tts-*.wav"))
	if err != nil {
		t.Fatalf("Failed to list temp files: %v", err)
	}
	initialCount := len(initialFiles)

	// Generate TTS
	outputPath, err := textToSpeech("test", "russian")
	if err != nil {
		t.Fatalf("textToSpeech() error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file not created: %s", outputPath)
	}

	// Clean up
	os.Remove(outputPath)

	// Verify cleanup
	afterFiles, err := filepath.Glob(filepath.Join(tempDir, "piper-tts-*.wav"))
	if err != nil {
		t.Fatalf("Failed to list temp files: %v", err)
	}
	afterCount := len(afterFiles)

	if afterCount > initialCount {
		t.Errorf("Temp files not cleaned up properly. Initial: %d, After: %d", initialCount, afterCount)
	}
}

// BenchmarkTextToSpeech benchmarks text-to-speech performance
func BenchmarkTextToSpeech(b *testing.B) {
	binaryPath := getPiperBinaryPath()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		b.Skip("Piper binary not found, skipping TTS benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath, err := textToSpeech("Hello world", "russian")
		if err != nil {
			b.Fatalf("textToSpeech failed: %v", err)
		}
		os.Remove(outputPath)
	}
}
