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

### Quick Setup

The easiest way to set up Piper TTS is to place the binary and model files in the `tts/requirements/` directory:

```
pejelagarto-translator/
└── tts/
    └── requirements/
        ├── piper           (Linux/macOS binary)
        ├── piper.exe       (Windows binary)
        └── model.onnx      (Voice model file)
```

### 1. Install Piper TTS

Follow the instructions at [Piper GitHub](https://github.com/rhasspy/piper) to install the Piper binary for your platform:

**Linux:**
```bash
# Download pre-built binary
wget https://github.com/rhasspy/piper/releases/latest/download/piper_linux_x86_64.tar.gz
tar xzf piper_linux_x86_64.tar.gz

# Move to requirements directory
mkdir -p tts/requirements
mv piper/piper tts/requirements/
chmod +x tts/requirements/piper
```

**macOS:**
```bash
# Download manually (Homebrew version may not work with relative paths)
wget https://github.com/rhasspy/piper/releases/latest/download/piper_macos_x86_64.tar.gz
tar xzf piper_macos_x86_64.tar.gz

# Move to requirements directory
mkdir -p tts/requirements
mv piper/piper tts/requirements/
chmod +x tts/requirements/piper
```

**Windows:**
```powershell
# Download from releases page or use PowerShell
$url = "https://github.com/rhasspy/piper/releases/latest/download/piper_windows_amd64.zip"
Invoke-WebRequest -Uri $url -OutFile piper_windows_amd64.zip

# Extract the archive
Expand-Archive -Path piper_windows_amd64.zip -DestinationPath piper

# Move to requirements directory
New-Item -ItemType Directory -Force -Path tts\requirements
Move-Item -Path piper\piper.exe -Destination tts\requirements\
```

Alternatively on Windows, you can:
1. Download the Windows release from [Piper releases page](https://github.com/rhasspy/piper/releases)
2. Extract the `.zip` file
3. Copy `piper.exe` to `tts\requirements\` directory
4. Ensure the file is named `piper.exe` (the application automatically adds `.exe` on Windows)

### 2. Download a Voice Model

Piper requires a voice model file (.onnx) to generate speech. Download one from the [Piper voices repository](https://github.com/rhasspy/piper/blob/master/VOICES.md):

**Linux/macOS:**
```bash
# Example: Download English US female voice
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx

# Move to requirements directory
mv en_US-lessac-medium.onnx tts/requirements/model.onnx
```

**Windows (PowerShell):**
```powershell
# Download voice model
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx"
Invoke-WebRequest -Uri $modelUrl -OutFile en_US-lessac-medium.onnx

# Move to requirements directory
Move-Item -Path en_US-lessac-medium.onnx -Destination tts\requirements\model.onnx
```

Alternatively, you can:
1. Visit [Piper voices on Hugging Face](https://huggingface.co/rhasspy/piper-voices)
2. Browse available voices and download your preferred `.onnx` file
3. Rename it to `model.onnx` and place it in `tts\requirements\`

### 3. Verify Installation

The application expects the following file structure:

**Linux/macOS:**
```
tts/requirements/
├── piper         (executable)
└── model.onnx
```

**Windows:**
```
tts\requirements\
├── piper.exe
└── model.onnx
```

You can verify the setup by checking:

**Linux/macOS:**
```bash
ls -l tts/requirements/
# Should show: piper (executable) and model.onnx
```

**Windows (PowerShell):**
```powershell
Get-ChildItem tts\requirements\
# Should show: piper.exe and model.onnx
```

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
