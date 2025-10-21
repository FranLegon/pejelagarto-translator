# Piper Text-to-Speech Integration

This directory contains a complete, self-contained Golang implementation for executing the Piper Text-to-Speech binary.

## Overview

The TTS functionality provides a simple Go function to convert text to speech using the [Piper TTS](https://github.com/rhasspy/piper) engine. The implementation follows best practices for external process execution, error handling, and temporary file management.

## Files

- `tts.go` - Main TTS implementation with the `textToSpeech()` function
- `tts_main.go` - Standalone demonstration program (build tag: ignore)
- `tts_test.go` - Comprehensive test suite
- `TTS_README.md` - This documentation file

## Installation Requirements

### Quick Setup Summary

The Piper TTS requires the following files in the `tts/requirements/` directory:

**Windows:**
```
pejelagarto-translator/
└── tts/
    └── requirements/
        ├── piper.exe                    (Main executable)
        ├── model.onnx                   (Voice model file)
        ├── model.onnx.json             (Model configuration)
        ├── espeak-ng.dll               (Phoneme library)
        ├── onnxruntime.dll             (ONNX runtime)
        ├── piper_phonemize.dll         (Phonemization library)
        ├── onnxruntime_providers_shared.dll
        ├── libtashkeel_model.ort
        └── espeak-ng-data/             (Phoneme data directory)
```

**Linux/macOS:**
```
pejelagarto-translator/
└── tts/
    └── requirements/
        ├── piper                       (Main executable)
        ├── model.onnx                  (Voice model file)
        ├── model.onnx.json            (Model configuration)
        └── espeak-ng-data/            (Phoneme data directory)
```

---

## Complete Installation Steps

### Step 1: Download and Extract Piper TTS

**Windows (PowerShell):**
```powershell
# Navigate to project root
cd path\to\pejelagarto-translator

# Download Piper
$url = "https://github.com/rhasspy/piper/releases/latest/download/piper_windows_amd64.zip"
Invoke-WebRequest -Uri $url -OutFile "tts\requirements\piper_windows_amd64.zip"

# Extract to requirements directory
Expand-Archive -Path "tts\requirements\piper_windows_amd64.zip" -DestinationPath "tts\requirements" -Force

# Copy all DLLs and dependencies from extracted folder to requirements root
Copy-Item "tts\requirements\piper\*.dll" -Destination "tts\requirements\" -Force
Copy-Item "tts\requirements\piper\*.ort" -Destination "tts\requirements\" -Force
Copy-Item "tts\requirements\piper\espeak-ng-data" -Destination "tts\requirements\" -Recurse -Force

# Copy the executable
Copy-Item "tts\requirements\piper\piper.exe" -Destination "tts\requirements\" -Force
```

**Linux:**
```bash
# Navigate to project root
cd /path/to/pejelagarto-translator

# Download Piper
wget https://github.com/rhasspy/piper/releases/latest/download/piper_linux_x86_64.tar.gz

# Extract
tar xzf piper_linux_x86_64.tar.gz

# Move files to requirements directory
mkdir -p tts/requirements
mv piper/piper tts/requirements/
mv piper/espeak-ng-data tts/requirements/
chmod +x tts/requirements/piper

# Clean up
rm -rf piper piper_linux_x86_64.tar.gz
```

**macOS:**
```bash
# Navigate to project root
cd /path/to/pejelagarto-translator

# Download Piper
wget https://github.com/rhasspy/piper/releases/latest/download/piper_macos_x86_64.tar.gz

# Extract
tar xzf piper_macos_x86_64.tar.gz

# Move files to requirements directory
mkdir -p tts/requirements
mv piper/piper tts/requirements/
mv piper/espeak-ng-data tts/requirements/
chmod +x tts/requirements/piper

# Clean up
rm -rf piper piper_macos_x86_64.tar.gz
```

---

### Step 2: Download Voice Model and Configuration

**Windows (PowerShell):**
```powershell
# Download voice model (English US female voice - medium quality)
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx"
Invoke-WebRequest -Uri $modelUrl -OutFile "tts\requirements\model.onnx"

# Download model configuration (REQUIRED)
$configUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json"
Invoke-WebRequest -Uri $configUrl -OutFile "tts\requirements\model.onnx.json"
```

**Linux/macOS:**
```bash
# Download voice model (English US female voice - medium quality)
wget -O tts/requirements/model.onnx \
  https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx

# Download model configuration (REQUIRED)
wget -O tts/requirements/model.onnx.json \
  https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json
```

**⚠️ Important:** Both the `.onnx` model file AND the `.onnx.json` config file are required!

---

### Step 3: Verify Installation

**Windows (PowerShell):**
```powershell
# Check all required files exist
Get-ChildItem tts\requirements\ | Select-Object Name

# Expected output should include:
# - piper.exe
# - model.onnx
# - model.onnx.json
# - espeak-ng.dll
# - onnxruntime.dll
# - piper_phonemize.dll
# - espeak-ng-data (directory)

# Test Piper execution
cd tts\requirements
echo "test" | .\piper.exe --model model.onnx --output_file test.wav
cd ..\..

# If successful, you should see a test.wav file created
# Play it to verify: start test.wav (in Windows)
```

**Linux/macOS:**
```bash
# Check all required files exist
ls -la tts/requirements/

# Expected output should include:
# - piper (executable)
# - model.onnx
# - model.onnx.json
# - espeak-ng-data/ (directory)

# Test Piper execution
cd tts/requirements
echo "test" | ./piper --model model.onnx --output_file test.wav
cd ../..

# If successful, you should see a test.wav file created
# Play it to verify:
# Linux: aplay test.wav
# macOS: afplay test.wav
```

---

### Step 4: Test with the Application

**Run the tests:**
```bash
# Windows/Linux/macOS
go test -v -run TestHandleTextToSpeech
```

**All tests should pass:**
- ✅ Valid POST request
- ✅ GET request (should fail)
- ✅ Empty/minimal text
- ✅ Pejelagarto text with special characters

**Start the server:**
```bash
# Windows
go build
.\pejelagarto-translator.exe

# Linux/macOS
go build
./pejelagarto-translator
```

**Test in browser:**
1. Open http://localhost:8080
2. Type some text (e.g., "Hello world")
3. Click the "Play" button (speaker icon)
4. You should hear the text spoken aloud!

---

## Alternative Voice Models

You can use different voice models by replacing `model.onnx` and `model.onnx.json`:

**Other English voices:**
```powershell
# Windows - High quality male voice
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/libritts/high/en_US-libritts-high.onnx"
$configUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/libritts/high/en_US-libritts-high.onnx.json"
Invoke-WebRequest -Uri $modelUrl -OutFile "tts\requirements\model.onnx"
Invoke-WebRequest -Uri $configUrl -OutFile "tts\requirements\model.onnx.json"

# British English
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_GB/alan/medium/en_GB-alan-medium.onnx"
$configUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_GB/alan/medium/en_GB-alan-medium.onnx.json"
Invoke-WebRequest -Uri $modelUrl -OutFile "tts\requirements\model.onnx"
Invoke-WebRequest -Uri $configUrl -OutFile "tts\requirements\model.onnx.json"
```

Browse all available voices: [Piper Voices](https://github.com/rhasspy/piper/blob/master/VOICES.md)

**Remember:** Always download BOTH the `.onnx` and `.onnx.json` files for any voice model!

## Usage

### As a Library Function

```go
import "pejelagarto-translator"

func main() {
    // Convert text to speech
    outputPath, err := textToSpeech("Hello, world!")
    if err != nil {
        log.Fatalf("TTS failed: %v", err)
    }
    
    fmt.Printf("Audio generated: %s\n", outputPath)
    
    // TODO: Play the audio file
    // You can use libraries like github.com/hajimehoshi/oto
    // or system commands like aplay (Linux), afplay (macOS)
}
```

### Running the Demonstration

```bash
# Run the standalone demonstration
go run tts_main.go tts.go

# Or build and run
go build -o tts-demo tts_main.go tts.go
./tts-demo
```

### Expected Output

```
=== Piper Text-to-Speech Demonstration ===

Checking for Piper binary at: /usr/local/bin/piper
✓ Piper binary found
Checking for voice model at: /usr/local/share/piper/model.onnx
✓ Voice model found

Converting text to speech: "Hello, this is a test of the Piper text to speech system."
✓ Audio file generated successfully: /tmp/piper-tts-123456.wav
  File size: 145632 bytes
  File location: /tmp

// TODO: Use a library to play the generated .wav file
// For example, you could use: github.com/hajimehoshi/oto for audio playback
// Or use external commands like 'aplay' (Linux), 'afplay' (macOS), or 'start' (Windows)
```

## Function Specification

### `textToSpeech(input string) (outputPath string, err error)`

Converts text to speech using the Piper TTS engine.

**Parameters:**
- `input` (string): The text to convert to speech

**Returns:**
- `outputPath` (string): Full path to the generated WAV file
- `err` (error): Error if any step fails

**Process:**
1. Validates that the Piper binary exists at `piperBinaryPath`
2. Validates that the voice model exists at `modelPath`
3. Creates a unique temporary file with pattern `piper-tts-*.wav`
4. Executes: `piper -m <model_path> --output_file <temp_output_path> --text <input_text>`
5. Verifies the output file was created and contains data
6. Returns the path to the generated audio file

**Error Handling:**
- Returns error if binary not found
- Returns error if model not found
- Returns error if temp file creation fails
- Returns error if Piper command fails
- Returns error if output file is empty
- Automatically cleans up temp file on failure

## Testing

Run the test suite:

```bash
# Run all TTS tests
go test -v -run "TestTextToSpeech|TestPiperPath|TestTempFile"

# Run all tests including existing translator tests
go test -v

# Run tests with coverage
go test -cover
```

### Test Coverage

The test suite includes:

1. **Validation Tests** - Verifies error handling for missing binary/model
2. **Path Constant Tests** - Ensures paths are configured and absolute
3. **Temp File Tests** - Validates temporary file creation pattern
4. **Empty Input Tests** - Tests graceful handling of edge cases
5. **Mock Binary Tests** - Tests with simulated Piper installation

## Playing Generated Audio

The `textToSpeech()` function generates a WAV file but doesn't play it. Here are some options for playback:

### Option 1: System Commands (Simple)

```go
import "os/exec"
import "runtime"

func playAudio(filepath string) error {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        cmd = exec.Command("aplay", filepath)
    case "darwin":
        cmd = exec.Command("afplay", filepath)
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", filepath)
    default:
        return fmt.Errorf("unsupported platform")
    }
    return cmd.Run()
}
```

### Option 2: Go Audio Library (Recommended)

```go
import "github.com/hajimehoshi/oto/v2"
import "github.com/youpy/go-wav"

// See github.com/hajimehoshi/oto documentation for complete example
```

### Option 3: External Player

You can manually play the generated file using:
- **Linux**: `aplay /tmp/piper-tts-*.wav`
- **macOS**: `afplay /tmp/piper-tts-*.wav`
- **Windows**: Open the file in Windows Media Player or VLC

## Architecture

The implementation follows Go best practices:

- **Package-level constants** for configuration
- **os/exec** for subprocess execution
- **os.CreateTemp** for secure temporary file creation
- **Comprehensive error handling** with wrapped errors
- **Resource cleanup** on failure paths
- **Clear documentation** and examples
- **Separation of concerns** (library vs demonstration)

## Command Line Reference

The Piper TTS binary is invoked as:

```bash
piper -m <model_path> --output_file <output_path> --text <input_text>
```

**Arguments:**
- `-m` or `--model` - Path to the ONNX voice model file
- `--output_file` - Path where the WAV file will be written
- `--text` - The text string to convert to speech

**Alternative:** You can also pipe text via stdin:
```bash
echo "Hello world" | piper -m model.onnx --output_file output.wav
```

## Security Considerations

1. **No Shell Injection**: Uses `exec.Command()` with explicit arguments (not shell execution)
2. **Path Validation**: Checks file existence before execution
3. **Temp File Safety**: Uses `os.CreateTemp()` which creates files with secure permissions
4. **Error Messages**: Includes command output in errors for debugging
5. **Resource Cleanup**: Removes temp files on failure

## Troubleshooting

### "Piper binary not found"

**Linux/macOS:**
- Verify Piper is in the requirements directory: `ls -l tts/requirements/piper`
- Ensure the binary has execute permissions: `chmod +x tts/requirements/piper`
- If you placed it elsewhere, update `piperBinaryPath` in the code

**Windows:**
- Verify `piper.exe` exists: `Get-ChildItem tts\requirements\piper.exe` (PowerShell)
- Or check in File Explorer: Navigate to `tts\requirements\` and look for `piper.exe`
- Make sure the file is named exactly `piper.exe` (not `piper.exe.exe`)
- If Windows Defender blocked the file, you may need to allow it

### "Voice model not found"
- Verify model file exists in requirements directory:
  - Linux/macOS: `ls -l tts/requirements/model.onnx`
  - Windows: `Get-ChildItem tts\requirements\model.onnx` (PowerShell)
- The file must be named exactly `model.onnx`
- Download a model from [Piper voices](https://github.com/rhasspy/piper/blob/master/VOICES.md)

### "Output file is empty"
- Check Piper error messages in the error output
- Verify the model file is not corrupted
- Ensure sufficient disk space in the temp directory
- Try running Piper manually to test:
  - Linux/macOS: `./tts/requirements/piper -m tts/requirements/model.onnx --text "test" --output_file test.wav`
  - Windows: `.\tts\requirements\piper.exe -m tts\requirements\model.onnx --text "test" --output_file test.wav`

### Permission Errors

**Linux/macOS:**
- Ensure the binary has execute permissions: `chmod +x tts/requirements/piper`
- Verify write permissions for the temp directory: `ls -ld /tmp`
- Check file ownership if needed: `ls -l tts/requirements/`

**Windows:**
- Right-click `piper.exe` → Properties → Unblock (if present)
- Run as Administrator if needed
- Check Windows Defender or antivirus hasn't quarantined the file
- Ensure your user has write permissions to the temp directory

### Windows-Specific Issues

**"This app can't run on your PC":**
- Make sure you downloaded the correct architecture (amd64 for 64-bit Windows)
- Try downloading a different release version
- Check if you need Visual C++ Redistributable

**File path issues:**
- Use backslashes (`\`) in Windows paths, not forward slashes
- Or use forward slashes throughout (Go handles both on Windows)
- Avoid spaces in file paths or quote them properly

**Antivirus blocking:**
- Some antivirus software may block unknown executables
- Add `piper.exe` to your antivirus whitelist/exceptions
- Or temporarily disable antivirus to test (re-enable after confirming it works)

## Performance Considerations

- **File Size**: WAV files are uncompressed (~176 KB per second of audio)
- **Generation Speed**: Typically faster than real-time (depends on model and hardware)
- **Temp Storage**: Files accumulate in temp directory - clean up when done
- **Model Loading**: First run may be slower as the model loads into memory

## License

This TTS integration follows the same license as the main Pejelagarto Translator project.
Piper TTS itself is licensed under the MIT License.

## Further Reading

- [Piper GitHub Repository](https://github.com/rhasspy/piper)
- [Available Voice Models](https://github.com/rhasspy/piper/blob/master/VOICES.md)
- [Piper Documentation](https://github.com/rhasspy/piper/blob/master/README.md)
- [ONNX Runtime](https://onnxruntime.ai/) (used by Piper)
