# Text-to-Speech Usage Examples

This document provides practical examples of using the Piper TTS integration.

## Quick Start

### 1. Install and Configure

```bash
# Install Piper (example for Linux)
wget https://github.com/rhasspy/piper/releases/latest/download/piper_linux_x86_64.tar.gz
tar xzf piper_linux_x86_64.tar.gz
sudo mv piper/piper /usr/local/bin/

# Download a voice model
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx
sudo mkdir -p /usr/local/share/piper
sudo mv en_US-lessac-medium.onnx /usr/local/share/piper/model.onnx

# Update tts.go constants to match your paths (if different)
```

### 2. Run the Demo

```bash
cd /path/to/pejelagarto-translator
go run tts_main.go tts.go
```

Expected output:
```
=== Piper Text-to-Speech Demonstration ===

Checking for Piper binary at: /usr/local/bin/piper
✓ Piper binary found
Checking for voice model at: /usr/local/share/piper/model.onnx
✓ Voice model found

Converting text to speech: "Hello, this is a test of the Piper text to speech system."
✓ Audio file generated successfully: /tmp/piper-tts-1234567.wav
  File size: 145632 bytes
  File location: /tmp
```

## Code Examples

### Example 1: Basic Usage

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Convert text to speech
    text := "Hello, world! This is a text to speech example."
    wavPath, err := textToSpeech(text)
    if err != nil {
        log.Fatalf("Failed to generate speech: %v", err)
    }
    
    fmt.Printf("Audio saved to: %s\n", wavPath)
}
```

### Example 2: With Error Handling

```go
package main

import (
    "fmt"
    "log"
    "os"
)

func generateAndVerify(text string) error {
    // Generate speech
    wavPath, err := textToSpeech(text)
    if err != nil {
        return fmt.Errorf("TTS generation failed: %w", err)
    }
    
    // Verify file was created
    info, err := os.Stat(wavPath)
    if err != nil {
        return fmt.Errorf("output file error: %w", err)
    }
    
    fmt.Printf("Generated %d bytes of audio at %s\n", info.Size(), wavPath)
    return nil
}

func main() {
    if err := generateAndVerify("Testing the system"); err != nil {
        log.Fatal(err)
    }
}
```

### Example 3: Convert Multiple Phrases

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    phrases := []string{
        "Welcome to the system.",
        "Please enter your password.",
        "Access granted.",
        "Goodbye.",
    }
    
    for i, phrase := range phrases {
        wavPath, err := textToSpeech(phrase)
        if err != nil {
            log.Printf("Failed to generate audio for phrase %d: %v", i, err)
            continue
        }
        fmt.Printf("Phrase %d saved to: %s\n", i+1, wavPath)
    }
}
```

### Example 4: Play Audio on Linux

```go
package main

import (
    "fmt"
    "log"
    "os/exec"
)

func textToSpeechAndPlay(text string) error {
    // Generate speech
    wavPath, err := textToSpeech(text)
    if err != nil {
        return fmt.Errorf("TTS failed: %w", err)
    }
    
    fmt.Printf("Playing audio: %s\n", wavPath)
    
    // Play using aplay (Linux)
    cmd := exec.Command("aplay", wavPath)
    return cmd.Run()
}

func main() {
    if err := textToSpeechAndPlay("Hello from Piper TTS"); err != nil {
        log.Fatal(err)
    }
}
```

### Example 5: Cross-Platform Audio Playback

```go
package main

import (
    "fmt"
    "log"
    "os/exec"
    "runtime"
)

func playAudio(wavPath string) error {
    var cmd *exec.Cmd
    
    switch runtime.GOOS {
    case "linux":
        cmd = exec.Command("aplay", wavPath)
    case "darwin":  // macOS
        cmd = exec.Command("afplay", wavPath)
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", wavPath)
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }
    
    return cmd.Run()
}

func speakText(text string) error {
    wavPath, err := textToSpeech(text)
    if err != nil {
        return err
    }
    
    fmt.Printf("Generated: %s\n", wavPath)
    return playAudio(wavPath)
}

func main() {
    if err := speakText("Cross-platform text to speech"); err != nil {
        log.Fatal(err)
    }
}
```

### Example 6: Integrating with Pejelagarto Translator

```go
package main

import (
    "fmt"
    "log"
)

func translateAndSpeak(humanText string) error {
    // Translate to Pejelagarto
    pejelagarto := TranslateToPejelagarto(humanText)
    fmt.Printf("Human:       %s\n", humanText)
    fmt.Printf("Pejelagarto: %s\n", pejelagarto)
    
    // Generate speech for both versions
    humanWav, err := textToSpeech(humanText)
    if err != nil {
        return fmt.Errorf("human TTS failed: %w", err)
    }
    
    pejeWav, err := textToSpeech(pejelagarto)
    if err != nil {
        return fmt.Errorf("pejelagarto TTS failed: %w", err)
    }
    
    fmt.Printf("Human audio:       %s\n", humanWav)
    fmt.Printf("Pejelagarto audio: %s\n", pejeWav)
    
    return nil
}

func main() {
    if err := translateAndSpeak("hello world"); err != nil {
        log.Fatal(err)
    }
}
```

### Example 7: Batch Processing with Cleanup

```go
package main

import (
    "fmt"
    "log"
    "os"
)

func batchConvert(texts []string) ([]string, error) {
    var wavFiles []string
    
    for i, text := range texts {
        wavPath, err := textToSpeech(text)
        if err != nil {
            return wavFiles, fmt.Errorf("failed at index %d: %w", i, err)
        }
        wavFiles = append(wavFiles, wavPath)
        fmt.Printf("Generated %d/%d: %s\n", i+1, len(texts), wavPath)
    }
    
    return wavFiles, nil
}

func cleanup(wavFiles []string) {
    for _, file := range wavFiles {
        if err := os.Remove(file); err != nil {
            log.Printf("Failed to remove %s: %v", file, err)
        }
    }
    fmt.Printf("Cleaned up %d files\n", len(wavFiles))
}

func main() {
    texts := []string{
        "First sentence.",
        "Second sentence.",
        "Third sentence.",
    }
    
    wavFiles, err := batchConvert(texts)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process the files...
    fmt.Println("Processing complete")
    
    // Clean up temporary files
    defer cleanup(wavFiles)
}
```

## Common Patterns

### Pattern 1: Deferred Cleanup

```go
func processWithCleanup(text string) error {
    wavPath, err := textToSpeech(text)
    if err != nil {
        return err
    }
    defer os.Remove(wavPath)  // Cleanup when done
    
    // Use the audio file...
    return nil
}
```

### Pattern 2: Error Logging

```go
func robustTTS(text string) string {
    wavPath, err := textToSpeech(text)
    if err != nil {
        log.Printf("TTS warning: %v", err)
        return ""  // Return empty on error
    }
    return wavPath
}
```

### Pattern 3: Retry Logic

```go
func ttsWithRetry(text string, maxRetries int) (string, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        wavPath, err := textToSpeech(text)
        if err == nil {
            return wavPath, nil
        }
        lastErr = err
        log.Printf("Retry %d/%d: %v", i+1, maxRetries, err)
    }
    
    return "", fmt.Errorf("all retries failed: %w", lastErr)
}
```

## Testing Examples

### Test with Mock Setup

```go
func TestMyTTSFunction(t *testing.T) {
    // Skip if Piper not installed
    if _, err := os.Stat(piperBinaryPath); os.IsNotExist(err) {
        t.Skip("Piper not installed, skipping test")
    }
    
    wavPath, err := textToSpeech("test")
    if err != nil {
        t.Fatalf("TTS failed: %v", err)
    }
    defer os.Remove(wavPath)
    
    // Verify file exists and has content
    info, err := os.Stat(wavPath)
    if err != nil {
        t.Fatalf("Output file error: %v", err)
    }
    if info.Size() == 0 {
        t.Error("Output file is empty")
    }
}
```

## Troubleshooting Common Issues

### Issue: "Binary not found"

**Solution:**
```bash
# Find where Piper is installed
which piper  # Linux/macOS
where piper  # Windows

# Update tts.go constant to match:
const piperBinaryPath = "/actual/path/to/piper"
```

### Issue: "Model not found"

**Solution:**
```bash
# Verify model exists
ls -l /usr/local/share/piper/model.onnx

# Or download a model
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx
```

### Issue: "Permission denied"

**Solution:**
```bash
# Make binary executable
chmod +x /path/to/piper

# Verify permissions
ls -l /path/to/piper
```

## Performance Tips

1. **Reuse voice model**: Keep the Piper process running for multiple conversions (not implemented in basic version)
2. **Batch processing**: Convert multiple texts in one session
3. **Async processing**: Use goroutines for parallel TTS generation
4. **Cleanup**: Remove temporary files promptly to save disk space

## Additional Resources

- [Piper TTS Documentation](https://github.com/rhasspy/piper)
- [Voice Models](https://github.com/rhasspy/piper/blob/master/VOICES.md)
- [Go os/exec Documentation](https://pkg.go.dev/os/exec)
- See `TTS_README.md` for detailed setup instructions
