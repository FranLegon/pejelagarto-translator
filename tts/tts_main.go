// +build ignore

// TTS Main Program - A standalone demonstration of the Piper TTS functionality
//
// This is a separate main program to demonstrate the textToSpeech function.
// Build and run with: go run tts_main.go tts.go
//
// Note: This file uses build tag 'ignore' so it doesn't conflict with the
// main translator program. To use this, run:
//   go run tts_main.go tts.go
package main

func main() {
	demonstrateTTS()
}
