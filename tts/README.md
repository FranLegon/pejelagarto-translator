# Text-to-Speech (TTS) Integration

Complete multi-language text-to-speech support using the [Piper TTS](https://github.com/rhasspy/piper) engine.

## Features

- 🗣️ **4 Languages**: Portuguese, Spanish, English, Russian
- 🎙️ **High-Quality Neural TTS**: Piper with ONNX models
- 🌍 **Automatic Text Preprocessing**: Language-specific character filtering
- 🔀 **Per-Request Language Selection**: Override default via HTTP API
- 📦 **Fully Embedded**: All dependencies included in executable
- 🚀 **Zero Configuration**: Extracts and runs automatically

## Quick Start

### 1. Download Models (First Time Only)

All required TTS files are automatically downloaded by the build script:

```powershell
.\get-requirements.ps1
```

Or manually download specific language models:

```powershell
cd tts\requirements\piper\languages
.\download_models.ps1
```

### 2. Build the Application

```powershell
.\build.ps1
```

### 3. Use TTS

**Via Command Line:**
```bash
# Default Portuguese
.\pejelagarto-translator.exe

# Spanish
.\pejelagarto-translator.exe -pronunciation_language spanish

# English
.\pejelagarto-translator.exe -pronunciation_language english

# Russian
.\pejelagarto-translator.exe -pronunciation_language russian
```

**Via HTTP API:**
```bash
curl -X POST "http://localhost:8080/tts?lang=portuguese" -d "Olá mundo" -o audio.wav
curl -X POST "http://localhost:8080/tts?lang=spanish" -d "Hola mundo" -o audio.wav
curl -X POST "http://localhost:8080/tts?lang=english" -d "Hello world" -o audio.wav
curl -X POST "http://localhost:8080/tts?lang=russian" -d "Привет мир" -o audio.wav
```

## How It Works

### Embedded Dependencies

All TTS dependencies are embedded in the executable during build:

**Embedded Files (~260MB):**
- Piper TTS binary (`piper.exe` on Windows, `piper` on Linux/macOS)
- DLL dependencies (Windows only)
- espeak-ng-data directory (phoneme data)
- 4 language models:
  - Portuguese: `pt_BR-faber-medium` (~63MB)
  - Spanish: `es_ES-davefx-medium` (~63MB)
  - English: `en_US-lessac-medium` (~63MB)
  - Russian: `ru_RU-irina-medium` (~63MB)

**Runtime Extraction:**

On first run, dependencies extract to temp directory:
- Windows: `C:\Windows\Temp\pejelagarto-translator\requirements\`
- Linux/macOS: `/tmp/pejelagarto-translator/requirements/`

Extraction happens once and is cached. Subsequent runs start instantly.

### Language-Specific Text Preprocessing

The `preprocessTextForTTS()` function automatically adapts text for natural pronunciation:

**1. Character Filtering**

Each language has its own allowed character set:

- **Portuguese**: a-z, á, é, í, ó, ú, â, ê, ô, ã, õ, à, ü, ç + punctuation
- **Spanish**: a-z, á, é, í, ó, ú, ü, ñ, ¡, ¿ + punctuation
- **English**: a-z + standard punctuation
- **Russian**: Cyrillic alphabet (а-я, ь, ъ) + punctuation

All other characters (emoji, special symbols, etc.) are removed.

**2. Consonant Cluster Limiting**

Limits consecutive consonants to maximum 2 for better pronunciation:

- `"tkr"` → `"tk"` (3rd consonant removed)
- `"strp"` → `"st"` (3rd and 4th removed)
- `"blá"` → `"blá"` (unchanged, only 2 consonants)

This prevents the TTS engine from spelling out unpronounceable combinations.

## File Structure

```
tts/
├── README.md                    # This file
├── tts.go                       # TTS implementation
├── tts_main.go                  # Standalone demo
├── tts_test.go                  # Test suite
└── requirements/                # TTS dependencies (embedded)
    ├── piper.exe / piper        # TTS binary
    ├── espeak-ng-data/          # Phoneme data
    ├── *.dll                    # Windows DLLs
    └── piper/
        └── languages/
            ├── portuguese/
            │   ├── model.onnx
            │   └── model.onnx.json
            ├── spanish/
            │   ├── model.onnx
            │   └── model.onnx.json
            ├── english/
            │   ├── model.onnx
            │   └── model.onnx.json
            └── russian/
                ├── model.onnx
                └── model.onnx.json
```

## API Reference

### Function: `textToSpeech(input, language string)`

Converts text to speech and returns path to generated WAV file.

**Parameters:**
- `input` (string): Text to convert to speech
- `language` (string): Language for pronunciation (portuguese, spanish, english, russian)

**Returns:**
- `outputPath` (string): Full path to generated WAV file
- `err` (error): Error if conversion fails

**Process:**
1. Preprocesses text for selected language
2. Validates Piper binary exists
3. Validates voice model exists for language
4. Creates unique temporary WAV file
5. Executes Piper with appropriate model
6. Verifies output file was created
7. Returns path to audio file

**Example:**
```go
wavPath, err := textToSpeech("Hello world", "english")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Audio saved to: %s\n", wavPath)
```

### HTTP Endpoint: `/tts`

**Method:** POST

**Query Parameters:**
- `lang` (optional): Language override (portuguese, spanish, english, russian)
  - Default: Uses `-pronunciation_language` flag value

**Request Body:** Plain text to convert

**Response:** audio/wav file

**Example:**
```bash
curl -X POST "http://localhost:8080/tts?lang=spanish" \
  -d "Hola, ¿cómo estás?" \
  -o saludo.wav
```

## Testing

```bash
# Run all TTS tests
go test -v ./tts/

# Run specific test
go test -v -run TestTextToSpeech

# Run with coverage
go test -cover ./tts/
```

## Adding New Languages

To add support for a new language:

### 1. Download Voice Model

Find a voice model at [Piper Voices](https://huggingface.co/rhasspy/piper-voices):

```powershell
# Example: Adding German
$lang = "german"
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/de/de_DE/thorsten/medium/de_DE-thorsten-medium.onnx"
$configUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/de/de_DE/thorsten/medium/de_DE-thorsten-medium.onnx.json"

$dir = "tts\requirements\piper\languages\$lang"
New-Item -ItemType Directory -Path $dir -Force
Invoke-WebRequest -Uri $modelUrl -OutFile "$dir\model.onnx"
Invoke-WebRequest -Uri $configUrl -OutFile "$dir\model.onnx.json"
```

### 2. Update Text Preprocessing

Edit `preprocessTextForTTS()` in `main.go`:

```go
case "german":
    vowels = "aeiouäöü"
    consonants = "bcdfghjklmnpqrsßtvwxyz"
```

### 3. Rebuild

```powershell
.\build.ps1
```

## Troubleshooting

### "Binary not found" Error

**Cause:** Piper binary not extracted or missing

**Solution:**
```powershell
# Delete temp directory to force re-extraction
Remove-Item "$env:TEMP\pejelagarto-translator" -Recurse -Force

# Restart application
.\pejelagarto-translator.exe
```

### "Model not found" Error

**Cause:** Voice model missing for selected language

**Solution:**
- Verify model files exist in `tts\requirements\piper\languages\{language}\`
- Re-run `.\get-requirements.ps1` to download missing models
- Rebuild with `.\build.ps1`

### "Permission denied" (Linux/macOS)

**Cause:** Piper binary not executable

**Solution:**
```bash
chmod +x /tmp/pejelagarto-translator/requirements/piper
```

### Windows Defender Blocking

**Cause:** Antivirus blocking `piper.exe`

**Solution:**
- Check Windows Security → Virus & threat protection → Protection history
- Allow the file if blocked
- Add exception if needed

### Poor Audio Quality

**Cause:** Text contains unpronounceable character sequences

**Solution:**
- Check that text uses characters from the selected language's character set
- Avoid excessive special characters or emoji
- Try preprocessing text manually before TTS

### "Output file is empty"

**Cause:** Piper failed to generate audio

**Solution:**
- Check that model files are not corrupted
- Verify sufficient disk space
- Try manual Piper execution to see error:
  ```bash
  cd C:\Windows\Temp\pejelagarto-translator\requirements
  echo "test" | .\piper.exe --model piper\languages\english\model.onnx --output_file test.wav
  ```

## Performance

- **Generation Speed**: Typically faster than real-time
- **File Size**: ~176 KB per second of audio (uncompressed WAV)
- **Temp Storage**: WAV files accumulate in OS temp directory
- **Model Loading**: First TTS call loads model into memory (~1-2 seconds)

## Security

- ✅ No shell injection (uses `exec.Command` with explicit args)
- ✅ Path validation before execution
- ✅ Secure temporary file creation
- ✅ Automatic resource cleanup on errors
- ✅ No sensitive data in error messages

## Alternative Voice Models

You can use different voices by replacing model files:

**Browse Available Voices:** [Piper Voices List](https://github.com/rhasspy/piper/blob/master/VOICES.md)

**Quality Levels:**
- **x-low**: Fastest, smallest (~10-20MB), lower quality
- **low**: Fast, small (~20-40MB), acceptable quality
- **medium**: Balanced (~60-80MB), good quality ⭐ **Current**
- **high**: Slower, large (~100-150MB), excellent quality

**Example: Switching to High-Quality English:**

```powershell
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/libritts/high/en_US-libritts-high.onnx"
$configUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/libritts/high/en_US-libritts-high.onnx.json"

Invoke-WebRequest -Uri $modelUrl -OutFile "tts\requirements\piper\languages\english\model.onnx"
Invoke-WebRequest -Uri $configUrl -OutFile "tts\requirements\piper\languages\english\model.onnx.json"

# Rebuild to embed new model
.\build.ps1
```

## Code Examples

### Basic Usage

```go
import "pejelagarto-translator/tts"

func main() {
    wavPath, err := textToSpeech("Hello world", "english")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Audio: %s\n", wavPath)
}
```

### Multiple Languages

```go
texts := map[string]string{
    "portuguese": "Olá mundo",
    "spanish":    "Hola mundo",
    "english":    "Hello world",
    "russian":    "Привет мир",
}

for lang, text := range texts {
    wavPath, err := textToSpeech(text, lang)
    if err != nil {
        log.Printf("Failed %s: %v", lang, err)
        continue
    }
    fmt.Printf("%s: %s\n", lang, wavPath)
}
```

### With Cleanup

```go
wavPath, err := textToSpeech("Test", "english")
if err != nil {
    return err
}
defer os.Remove(wavPath)  // Clean up when done

// Use the audio file...
```

### Cross-Platform Playback

```go
func playAudio(wavPath string) error {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        cmd = exec.Command("aplay", wavPath)
    case "darwin":
        cmd = exec.Command("afplay", wavPath)
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", wavPath)
    default:
        return fmt.Errorf("unsupported OS")
    }
    return cmd.Run()
}
```

## Further Reading

- [Piper TTS GitHub](https://github.com/rhasspy/piper)
- [Piper Voice Models](https://huggingface.co/rhasspy/piper-voices)
- [espeak-ng Documentation](https://github.com/espeak-ng/espeak-ng)
- [ONNX Runtime](https://onnxruntime.ai/)

## License

This TTS integration follows the same license as the main Pejelagarto Translator project (MIT).
Piper TTS is licensed under the MIT License.
