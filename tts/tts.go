// Text-to-Speech functionality using Piper TTS
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Package-level constants for Piper TTS configuration
// The binary and model should be placed in the tts/requirements/ directory
const (
	piperBinaryPath = "tts/requirements/piper"       // Path to the Piper TTS binary
	modelPath       = "tts/requirements/model.onnx"  // Path to the voice model file
)

// textToSpeech executes the Piper Text-to-Speech binary to convert text to audio.
// It generates a unique temporary WAV file for the output audio.
//
// Parameters:
//   - input: The text string to convert to speech
//
// Returns:
//   - outputPath: The path to the generated WAV file
//   - err: Any error encountered during execution
//
// The function performs the following:
//  1. Validates that the Piper binary exists
//  2. Validates that the voice model file exists
//  3. Creates a unique temporary WAV file for output
//  4. Executes: piper -m <model_path> --output_file <temp_output_path> --text <input_text>
//  5. Returns the path to the generated audio file
func textToSpeech(input string) (outputPath string, err error) {
	// Check if the Piper binary exists
	if _, err := os.Stat(piperBinaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("piper binary not found at %s: %w", piperBinaryPath, err)
	}

	// Check if the voice model exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("voice model not found at %s: %w", modelPath, err)
	}

	// Create a unique temporary file for the output audio
	// Using os.CreateTemp with pattern ensures uniqueness
	tempFile, err := os.CreateTemp("", "piper-tts-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary output file: %w", err)
	}
	outputPath = tempFile.Name()
	
	// Close the file handle immediately as Piper will write to it
	tempFile.Close()

	// Build the command to execute Piper
	// Command format: piper -m <model_path> --output_file <output_path> --text <input_text>
	cmd := exec.Command(
		piperBinaryPath,
		"-m", modelPath,
		"--output_file", outputPath,
		"--text", input,
	)

	// Capture both stdout and stderr for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up the temporary file if command fails
		os.Remove(outputPath)
		return "", fmt.Errorf("piper command failed: %w\nOutput: %s", err, string(output))
	}

	// Verify that the output file was actually created and has content
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

// demonstrateTTS shows how to use the textToSpeech function
func demonstrateTTS() {
	fmt.Println("=== Piper Text-to-Speech Demonstration ===")
	fmt.Println()

	// Check if binary exists
	fmt.Printf("Checking for Piper binary at: %s\n", piperBinaryPath)
	if _, err := os.Stat(piperBinaryPath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: Piper binary not found at %s\n", piperBinaryPath)
		fmt.Println("   Please update the piperBinaryPath constant in tts.go")
		return
	}
	fmt.Println("✓ Piper binary found")

	// Check if model exists
	fmt.Printf("Checking for voice model at: %s\n", modelPath)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Printf("❌ Error: Voice model not found at %s\n", modelPath)
		fmt.Println("   Please update the modelPath constant in tts.go")
		return
	}
	fmt.Println("✓ Voice model found")

	// Sample text to convert to speech
	sampleText := "Hello, this is a test of the Piper text to speech system."
	fmt.Printf("\nConverting text to speech: %q\n", sampleText)

	// Call the textToSpeech function
	outputPath, err := textToSpeech(sampleText)
	if err != nil {
		fmt.Printf("❌ Error generating speech: %v\n", err)
		return
	}

	fmt.Printf("✓ Audio file generated successfully: %s\n", outputPath)
	
	// Get file size for confirmation
	fileInfo, _ := os.Stat(outputPath)
	fmt.Printf("  File size: %d bytes\n", fileInfo.Size())
	fmt.Printf("  File location: %s\n", filepath.Dir(outputPath))
	
	fmt.Println("\n// TODO: Use a library to play the generated .wav file")
	fmt.Println("// For example, you could use: github.com/hajimehoshi/oto for audio playback")
	fmt.Println("// Or use external commands like 'aplay' (Linux), 'afplay' (macOS), or 'start' (Windows)")
}
