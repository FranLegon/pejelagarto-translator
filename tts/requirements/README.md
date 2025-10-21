# Piper TTS Requirements

This directory should contain the Piper TTS binary and language-specific voice model files.

## 🌍 Multi-Language Support

This application now supports **Portuguese, Spanish, English, and Russian** with language-specific models stored in subdirectories.

## Required Files

### For Linux/macOS:
```
tts/requirements/
├── piper                              (Piper TTS binary - must be executable)
├── espeak-ng-data/                    (Phoneme data directory)
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

### For Windows:
```
tts\requirements\
├── piper.exe                          (Piper TTS binary)
├── espeak-ng-data\                    (Phoneme data directory)
├── *.dll                              (Required DLL files)
└── piper\
    └── languages\
        ├── portuguese\
        │   ├── model.onnx
        │   └── model.onnx.json
        ├── spanish\
        │   ├── model.onnx
        │   └── model.onnx.json
        ├── english\
        │   ├── model.onnx
        │   └── model.onnx.json
        └── russian\
            ├── model.onnx
            └── model.onnx.json
```

## Installation Instructions

### 1. Download Piper TTS Binary

**Linux:**
```bash
# Download the latest release
wget https://github.com/rhasspy/piper/releases/latest/download/piper_linux_x86_64.tar.gz
tar xzf piper_linux_x86_64.tar.gz

# Copy binary to this directory
cp piper/piper tts/requirements/
chmod +x tts/requirements/piper
```

**macOS:**
```bash
# Download the latest release
wget https://github.com/rhasspy/piper/releases/latest/download/piper_macos_x86_64.tar.gz
tar xzf piper_macos_x86_64.tar.gz

# Copy binary to this directory
cp piper/piper tts/requirements/
chmod +x tts/requirements/piper
```

**Windows (PowerShell):**
```powershell
# Download the latest release
$url = "https://github.com/rhasspy/piper/releases/latest/download/piper_windows_amd64.zip"
Invoke-WebRequest -Uri $url -OutFile piper_windows_amd64.zip

# Extract and copy to requirements directory
Expand-Archive -Path piper_windows_amd64.zip -DestinationPath piper
Copy-Item -Path piper\piper.exe -Destination tts\requirements\
```

Or manually:
1. Visit [Piper Releases](https://github.com/rhasspy/piper/releases/latest)
2. Download the appropriate file for your platform:
   - Linux: `piper_linux_x86_64.tar.gz`
   - macOS: `piper_macos_x86_64.tar.gz`
   - Windows: `piper_windows_amd64.zip`
3. Extract the archive
4. Copy the binary to this directory:
   - Linux/macOS: Copy `piper` to `tts/requirements/piper`
   - Windows: Copy `piper.exe` to `tts\requirements\piper.exe`
5. Make executable (Linux/macOS only): `chmod +x tts/requirements/piper`

### 2. Download Voice Models

**🚀 Quick Method (Recommended):**

Use the automated download script to get all language models at once:

```powershell
cd tts\requirements\piper\languages
.\download_models.ps1
```

This will download:
- **Portuguese** (pt_BR-faber-medium) - ~63 MB
- **Spanish** (es_ES-davefx-medium) - ~63 MB
- **English** (en_US-lessac-medium) - ~63 MB
- **Russian** (ru_RU-irina-medium) - ~63 MB

**Total**: ~252 MB for all languages

---

**📖 Manual Method:**

See detailed instructions in: `piper/languages/README.md`

Or manually download from [Piper Voices on Hugging Face](https://huggingface.co/rhasspy/piper-voices):
1. Navigate to a voice (e.g., `pt/pt_BR/faber/medium/`)
2. Download both the `.onnx` and `.onnx.json` files
3. Place them in the appropriate language directory:
   - Portuguese: `piper/languages/portuguese/model.onnx`
   - Spanish: `piper/languages/spanish/model.onnx`
   - English: `piper/languages/english/model.onnx`
   - Russian: `piper/languages/russian/model.onnx`

## Verification

After installation, verify the files are in place:

**Windows (PowerShell):**
```powershell
# Check binary
Get-ChildItem tts\requirements\piper.exe

# Check language models
Get-ChildItem tts\requirements\piper\languages\*\model.onnx
```

**Linux/macOS:**
```bash
# Check binary
ls -lh tts/requirements/piper

# Check language models
ls -lh tts/requirements/piper/languages/*/model.onnx
```

**Test the installation:**

```bash
# Test with default Portuguese
.\pejelagarto-translator.exe -pronunciation_language=portuguese

# Test with different languages
.\pejelagarto-translator.exe -pronunciation_language=spanish
.\pejelagarto-translator.exe -pronunciation_language=english
.\pejelagarto-translator.exe -pronunciation_language=russian
```

Or test via API:
```powershell
# Portuguese
curl -X POST "http://localhost:8080/tts?lang=portuguese" -d "Olá mundo" -o test_pt.wav

# Spanish
curl -X POST "http://localhost:8080/tts?lang=spanish" -d "Hola mundo" -o test_es.wav

# English
curl -X POST "http://localhost:8080/tts?lang=english" -d "Hello world" -o test_en.wav

# Russian
curl -X POST "http://localhost:8080/tts?lang=russian" -d "Привет мир" -o test_ru.wav
```

## File Sizes

Typical file sizes:
- **Piper binary**: 10-30 MB
- **Voice models** (per language): 
  - Low quality: ~5-10 MB
  - Medium quality: ~60-65 MB
  - High quality: ~100+ MB
- **All 4 medium-quality models**: ~252 MB total

## Notes

- Each language has its own subdirectory: `piper/languages/{language}/`
- Model files must be named exactly `model.onnx` and `model.onnx.json`
- The binary must be named `piper` (Linux/macOS) or `piper.exe` (Windows)
- On Linux/macOS, the binary must have execute permissions
- On Windows, you may need to unblock files or allow through Windows Defender

## Multi-Language Usage

**Command-line flag:**
```bash
.\pejelagarto-translator.exe -pronunciation_language=spanish
```

**HTTP API with query parameter:**
```bash
curl -X POST "http://localhost:8080/tts?lang=english" -d "Hello" -o audio.wav
```

**Supported languages:** portuguese, spanish, english, russian

## Alternative Voice Models

You can replace any language model with different voices from [Piper Voices](https://huggingface.co/rhasspy/piper-voices):

- Different voice qualities (low, medium, high)
- Different speakers (male, female, various characteristics)
- Alternative regional accents (e.g., es_MX for Mexican Spanish)

Just download both `.onnx` and `.onnx.json` files and place them in the appropriate language directory.

Browse all available voices at: https://github.com/rhasspy/piper/blob/master/VOICES.md

## Troubleshooting

**"File not found" errors:**
- Ensure files are in this exact directory: `tts/requirements/`
- Check file names are correct (case-sensitive on Linux/macOS)
- Verify paths don't have extra spaces or special characters

**Permission errors (Linux/macOS):**
```bash
chmod +x tts/requirements/piper
```

**Windows Defender blocking:**
- Right-click `piper.exe` → Properties → Unblock (if checkbox appears)
- Or add an exception in Windows Security settings

For more help, see the main TTS_README.md in the parent directory.
