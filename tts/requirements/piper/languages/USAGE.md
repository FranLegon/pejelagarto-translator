# Multi-Language TTS Usage Guide

## Overview

The Pejelagarto Translator now supports Text-to-Speech (TTS) in multiple languages:
- **Portuguese** (Brazilian Portuguese - default)
- **Spanish** (Spain Spanish)
- **English** (US English)
- **Russian**

## Quick Start

### 1. Download Language Models

First, download the TTS models for the languages you want to use:

```powershell
cd tts\requirements\piper\languages
.\download_models.ps1
```

This script will download all four language models (~252 MB total).

### 2. Run the Application

#### Using Default Language (Portuguese)

```powershell
.\pejelagarto-translator.exe
```

#### Using a Specific Language

```powershell
# Spanish
.\pejelagarto-translator.exe -pronunciation_language=spanish

# English
.\pejelagarto-translator.exe -pronunciation_language=english

# Russian
.\pejelagarto-translator.exe -pronunciation_language=russian

# Portuguese (explicit)
.\pejelagarto-translator.exe -pronunciation_language=portuguese
```

## HTTP API Usage

### Setting Default Language

The language specified with the `-pronunciation_language` flag becomes the default for all TTS requests.

### Per-Request Language Override

You can override the default language for individual TTS requests using the `lang` query parameter:

#### Portuguese TTS Request
```bash
curl -X POST "http://localhost:8080/tts?lang=portuguese" \
     -H "Content-Type: text/plain" \
     -d "Olá, como você está?" \
     --output audio.wav
```

#### Spanish TTS Request
```bash
curl -X POST "http://localhost:8080/tts?lang=spanish" \
     -H "Content-Type: text/plain" \
     -d "Hola, ¿cómo estás?" \
     --output audio.wav
```

#### English TTS Request
```bash
curl -X POST "http://localhost:8080/tts?lang=english" \
     -H "Content-Type: text/plain" \
     -d "Hello, how are you?" \
     --output audio.wav
```

#### Russian TTS Request
```bash
curl -X POST "http://localhost:8080/tts?lang=russian" \
     -H "Content-Type: text/plain" \
     -d "Привет, как дела?" \
     --output audio.wav
```

### PowerShell Examples

```powershell
# Portuguese
Invoke-RestMethod -Uri "http://localhost:8080/tts?lang=portuguese" `
    -Method Post -Body "Bem-vindo ao tradutor Pejelagarto" `
    -OutFile "audio_pt.wav"

# Spanish
Invoke-RestMethod -Uri "http://localhost:8080/tts?lang=spanish" `
    -Method Post -Body "Bienvenido al traductor Pejelagarto" `
    -OutFile "audio_es.wav"

# English
Invoke-RestMethod -Uri "http://localhost:8080/tts?lang=english" `
    -Method Post -Body "Welcome to the Pejelagarto translator" `
    -OutFile "audio_en.wav"

# Russian
Invoke-RestMethod -Uri "http://localhost:8080/tts?lang=russian" `
    -Method Post -Body "Добро пожаловать в переводчик Pejelagarto" `
    -OutFile "audio_ru.wav"
```

## Text Preprocessing

The application automatically preprocesses text based on the selected language:

### Portuguese
- **Vowels**: a, e, i, o, u, á, é, í, ó, ú, â, ê, ô, ã, õ, à, ü
- **Consonants**: b, c, d, f, g, h, j, k, l, m, n, p, q, r, s, t, v, w, x, y, z, ç
- **Special**: Removes non-Portuguese characters
- **Consonant clusters**: Limits to max 2 consecutive consonants

### Spanish
- **Vowels**: a, e, i, o, u, á, é, í, ó, ú, ü
- **Consonants**: b, c, d, f, g, h, j, k, l, m, n, ñ, p, q, r, s, t, v, w, x, y, z
- **Special**: Supports ¡ and ¿ punctuation
- **Consonant clusters**: Limits to max 2 consecutive consonants

### English
- **Vowels**: a, e, i, o, u
- **Consonants**: b, c, d, f, g, h, j, k, l, m, n, p, q, r, s, t, v, w, x, y, z
- **Consonant clusters**: Limits to max 2 consecutive consonants

### Russian
- **Vowels**: а, е, ё, и, о, у, ы, э, ю, я
- **Consonants**: б, в, г, д, ж, з, й, к, л, м, н, п, р, с, т, ф, х, ц, ч, ш, щ
- **Special**: Supports ь (soft sign) and ъ (hard sign)
- **Consonant clusters**: Limits to max 2 consecutive consonants

## Command-Line Flags

```
-pronunciation_language string
    TTS pronunciation language (default "portuguese")
    Allowed values: portuguese, spanish, english, russian

-ngrok_token string
    Optional ngrok auth token to expose server publicly

-ngrok_domain string
    Optional ngrok persistent domain (e.g., your-domain.ngrok-free.app)
```

## Examples

### Example 1: Run with Spanish TTS
```powershell
.\pejelagarto-translator.exe -pronunciation_language=spanish
```

### Example 2: Run with English TTS and ngrok
```powershell
.\pejelagarto-translator.exe `
    -pronunciation_language=english `
    -ngrok_token=YOUR_TOKEN `
    -ngrok_domain=your-domain.ngrok-free.app
```

### Example 3: Test Different Languages in One Session
```powershell
# Start server with default Portuguese
.\pejelagarto-translator.exe

# In another terminal, test each language:
curl -X POST "http://localhost:8080/tts?lang=portuguese" -d "Olá!" -o pt.wav
curl -X POST "http://localhost:8080/tts?lang=spanish" -d "¡Hola!" -o es.wav
curl -X POST "http://localhost:8080/tts?lang=english" -d "Hello!" -o en.wav
curl -X POST "http://localhost:8080/tts?lang=russian" -d "Привет!" -o ru.wav
```

## Error Handling

### Invalid Language Error
If you specify an invalid language, you'll get an error:
```
Invalid pronunciation language 'french'. Allowed: portuguese, spanish, english, russian
```

### Missing Model Error
If a model file is missing:
```
voice model not found at tts/requirements/piper/languages/spanish/model.onnx
```

**Solution**: Run the download script to get the missing models:
```powershell
cd tts\requirements\piper\languages
.\download_models.ps1
```

## Model Files

Each language directory should contain:
- `model.onnx` - The neural network model file
- `model.onnx.json` - Configuration file for the model

### File Structure
```
tts/requirements/piper/languages/
├── portuguese/
│   ├── model.onnx (63 MB)
│   └── model.onnx.json (1 KB)
├── spanish/
│   ├── model.onnx (63 MB)
│   └── model.onnx.json (1 KB)
├── english/
│   ├── model.onnx (63 MB)
│   └── model.onnx.json (1 KB)
└── russian/
    ├── model.onnx (63 MB)
    └── model.onnx.json (1 KB)
```

## Advanced: Custom Models

You can replace the models with other voices from the Piper repository:

1. Browse available models: https://huggingface.co/rhasspy/piper-voices/tree/v1.0.0
2. Download your preferred `.onnx` and `.onnx.json` files
3. Rename them to `model.onnx` and `model.onnx.json`
4. Place them in the appropriate language directory

### Example: Using a Different Spanish Voice
```powershell
# Download alternative Spanish voice (Mexican Spanish)
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/es/es_MX/claude/medium/es_MX-claude-medium.onnx" -OutFile "spanish\model.onnx"
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/es/es_MX/claude/medium/es_MX-claude-medium.onnx.json" -OutFile "spanish\model.onnx.json"
```

## Troubleshooting

### Issue: "Model not found" error
**Solution**: Download the models using the provided script

### Issue: Audio quality is poor
**Solution**: Try downloading a "high" quality model instead of "medium"

### Issue: Language characters not rendering correctly
**Solution**: Ensure your terminal supports UTF-8 encoding

### Issue: Application won't start with custom language
**Solution**: Verify the language name is exactly: `portuguese`, `spanish`, `english`, or `russian` (lowercase)

## Performance Notes

- Model loading happens once per TTS request
- First request may be slightly slower due to model initialization
- Average processing time: 1-3 seconds for short texts
- Model size: ~63 MB per language in memory when active

## See Also

- [README.md](README.md) - Model download instructions
- [download_models.ps1](download_models.ps1) - Automatic model downloader
- [TTS_README.md](../../TTS_README.md) - General TTS documentation
