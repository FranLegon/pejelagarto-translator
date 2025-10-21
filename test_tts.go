//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	fmt.Println("=== Testing Piper TTS Configuration ===\n")

	// Test 1: Check binary
	binaryPath := "tts/requirements/piper"
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	fmt.Printf("1. Checking binary at: %s\n", binaryPath)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Printf("   ❌ FAILED: Binary not found\n")
		os.Exit(1)
	}
	fmt.Printf("   ✅ Binary found\n\n")

	// Test 2: Check model
	modelPath := "tts/requirements/model.onnx"
	fmt.Printf("2. Checking model at: %s\n", modelPath)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		fmt.Printf("   ❌ FAILED: Model not found\n")
		os.Exit(1)
	}
	fmt.Printf("   ✅ Model found\n\n")

	// Test 3: Check DLLs (Windows only)
	if runtime.GOOS == "windows" {
		fmt.Println("3. Checking required DLLs:")
		dlls := []string{
			"tts/requirements/espeak-ng.dll",
			"tts/requirements/onnxruntime.dll",
			"tts/requirements/piper_phonemize.dll",
		}
		for _, dll := range dlls {
			if _, err := os.Stat(dll); os.IsNotExist(err) {
				fmt.Printf("   ❌ Missing: %s\n", dll)
			} else {
				fmt.Printf("   ✅ Found: %s\n", dll)
			}
		}
		fmt.Println()

		// Check espeak-ng-data
		fmt.Println("4. Checking espeak-ng-data directory:")
		if _, err := os.Stat("tts/requirements/espeak-ng-data"); os.IsNotExist(err) {
			fmt.Printf("   ❌ FAILED: espeak-ng-data directory not found\n")
		} else {
			fmt.Printf("   ✅ espeak-ng-data directory found\n")
		}
		fmt.Println()
	}

	// Test 4: Try to run Piper
	fmt.Println("5. Testing Piper execution:")

	// Create temp output file
	tempFile, err := os.CreateTemp("", "test-tts-*.wav")
	if err != nil {
		fmt.Printf("   ❌ FAILED: Could not create temp file: %v\n", err)
		os.Exit(1)
	}
	outputPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(outputPath)

	// Get absolute paths
	absOutputPath, _ := filepath.Abs(outputPath)
	absModelPath, _ := filepath.Abs(modelPath)

	binaryName := "piper"
	if runtime.GOOS == "windows" {
		binaryName = "piper.exe"
	}

	// Run Piper from its directory
	cmd := exec.Command(
		binaryName,
		"-m", absModelPath,
		"--output_file", absOutputPath,
		"--text", "Testing Piper TTS",
	)
	cmd.Dir = "tts/requirements"

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("   ❌ FAILED: Piper execution failed\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   Output: %s\n", string(output))
		os.Exit(1)
	}

	// Check output file
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		fmt.Printf("   ❌ FAILED: Output file not created\n")
		os.Exit(1)
	}
	if fileInfo.Size() == 0 {
		fmt.Printf("   ❌ FAILED: Output file is empty\n")
		os.Exit(1)
	}

	fmt.Printf("   ✅ Piper executed successfully\n")
	fmt.Printf("   Output file size: %d bytes\n", fileInfo.Size())
	fmt.Println("\n=== All tests passed! ===")
}
