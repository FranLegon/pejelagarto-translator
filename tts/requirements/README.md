# Piper TTS Requirements

This directory should contain the Piper TTS binary and voice model files.

## Required Files

### For Linux/macOS:
```
tts/requirements/
├── piper         (Piper TTS binary - must be executable)
└── model.onnx    (Voice model file)
```

### For Windows:
```
tts\requirements\
├── piper.exe     (Piper TTS binary)
└── model.onnx    (Voice model file)
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

### 2. Download Voice Model

**Any Platform:**

Visit [Piper Voices on Hugging Face](https://huggingface.co/rhasspy/piper-voices) and download a voice model.

**Recommended voice models:**
- **English (US)**: `en_US-lessac-medium.onnx` - Clear, natural female voice
- **English (US)**: `en_US-libritts-high.onnx` - High quality multi-speaker
- **English (GB)**: `en_GB-alan-medium.onnx` - British English male voice

**Download example (English US female voice):**

**Linux/macOS:**
```bash
# Download voice model
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx

# Rename and move to this directory
mv en_US-lessac-medium.onnx tts/requirements/model.onnx
```

**Windows (PowerShell):**
```powershell
# Download voice model
$modelUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx"
Invoke-WebRequest -Uri $modelUrl -OutFile en_US-lessac-medium.onnx

# Rename and move to this directory
Move-Item -Path en_US-lessac-medium.onnx -Destination tts\requirements\model.onnx
```

Or manually:
1. Visit [Piper Voices on Hugging Face](https://huggingface.co/rhasspy/piper-voices)
2. Navigate to a voice (e.g., `en/en_US/lessac/medium/`)
3. Download the `.onnx` file
4. Rename it to `model.onnx`
5. Place it in this directory

## Verification

After installation, verify the files are in place:

**Linux/macOS:**
```bash
ls -lh tts/requirements/
# Should show: piper (executable) and model.onnx
```

**Windows (PowerShell):**
```powershell
Get-ChildItem tts\requirements\
# Should show: piper.exe and model.onnx
```

**Test the installation:**

**Linux/macOS:**
```bash
./tts/requirements/piper -m tts/requirements/model.onnx --text "Hello world" --output_file test.wav
```

**Windows:**
```powershell
.\tts\requirements\piper.exe -m tts\requirements\model.onnx --text "Hello world" --output_file test.wav
```

If successful, you'll have a `test.wav` file you can play to hear the generated speech.

## File Sizes

Typical file sizes:
- **Piper binary**: 10-30 MB
- **Voice models**: 
  - Low quality: ~5-10 MB
  - Medium quality: ~20-50 MB
  - High quality: ~50-100 MB

## Notes

- The `.onnx` file must be named exactly `model.onnx`
- The binary must be named `piper` (Linux/macOS) or `piper.exe` (Windows)
- On Linux/macOS, the binary must have execute permissions
- On Windows, you may need to unblock the file or allow it through Windows Defender

## Alternative Voice Models

You can use different voice models by replacing `model.onnx` with your preferred voice. Available options:

- Multiple languages supported (English, Spanish, French, German, etc.)
- Different voice qualities (low, medium, high)
- Different speakers (male, female, various tones)

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
