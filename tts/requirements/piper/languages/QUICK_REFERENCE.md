# Multi-Language TTS Quick Reference

## Command-Line Flags

```bash
-pronunciation_language string
    Default TTS language (default: "portuguese")
    Allowed: portuguese, spanish, english, russian
```

## Quick Start

### 1. Download Models (One-Time Setup)
```powershell
cd tts\requirements\piper\languages
.\download_models.ps1
```

### 2. Run Application
```powershell
# With default Portuguese
.\pejelagarto-translator.exe

# With specific language
.\pejelagarto-translator.exe -pronunciation_language=spanish
```

## API Usage

### Default Language (from flag)
```bash
curl -X POST http://localhost:8080/tts -d "Text here" -o audio.wav
```

### Override Language (per request)
```bash
# Portuguese
curl -X POST "http://localhost:8080/tts?lang=portuguese" -d "Olá" -o pt.wav

# Spanish
curl -X POST "http://localhost:8080/tts?lang=spanish" -d "Hola" -o es.wav

# English
curl -X POST "http://localhost:8080/tts?lang=english" -d "Hello" -o en.wav

# Russian
curl -X POST "http://localhost:8080/tts?lang=russian" -d "Привет" -o ru.wav
```

## Language-Specific Character Sets

### Portuguese
- Vowels: a e i o u á é í ó ú â ê ô ã õ à ü
- Special: ç

### Spanish
- Vowels: a e i o u á é í ó ú ü
- Special: ñ, ¡, ¿

### English
- Vowels: a e i o u
- Standard ASCII

### Russian
- Cyrillic alphabet
- Vowels: а е ё и о у ы э ю я
- Special: ь (soft sign), ъ (hard sign)

## Directory Structure

```
tts/requirements/piper/languages/
├── README.md                    # Model download instructions
├── USAGE.md                     # Detailed usage guide
├── QUICK_REFERENCE.md          # This file
├── download_models.ps1         # Auto-download script
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

## Model Information

| Language   | Model Name              | Size  | Voice Type |
|------------|-------------------------|-------|------------|
| Portuguese | pt_BR-faber-medium      | 63 MB | Male       |
| Spanish    | es_ES-davefx-medium     | 63 MB | Male       |
| English    | en_US-lessac-medium     | 63 MB | Female     |
| Russian    | ru_RU-irina-medium      | 63 MB | Female     |

**Total**: ~252 MB for all languages

## Code Changes Summary

### Functions Modified

1. **`preprocessTextForTTS(input string, pronunciationLanguage string)`**
   - Now accepts language parameter
   - Switches vowels/consonants based on language
   - Supports: portuguese, spanish, english, russian

2. **`textToSpeech(input string, pronunciationLanguage string)`**
   - Now accepts language parameter
   - Uses language-specific model path
   - Calls updated preprocessTextForTTS

3. **`handleTextToSpeech(w http.ResponseWriter, r *http.Request)`**
   - Reads `lang` query parameter
   - Validates language
   - Falls back to global flag value

4. **`getModelPath(language string) string`** (NEW)
   - Returns language-specific model path
   - Format: `tts/requirements/piper/languages/{language}/model.onnx`

### Global Variables Added

```go
var pronunciationLanguage string
```

### Constants Modified

```go
// Removed fixed modelPath constant
// Now using dynamic getModelPath(language) function
```

## Error Messages

### Invalid Language
```
Invalid pronunciation language 'french'. 
Allowed: portuguese, spanish, english, russian
```

### Missing Model
```
voice model not found at tts/requirements/piper/languages/spanish/model.onnx
```

**Solution**: Download models using `download_models.ps1`

## Testing

```bash
# Test different languages
echo "Olá mundo" | curl -X POST "http://localhost:8080/tts?lang=portuguese" --data-binary @- -o test_pt.wav
echo "Hola mundo" | curl -X POST "http://localhost:8080/tts?lang=spanish" --data-binary @- -o test_es.wav
echo "Hello world" | curl -X POST "http://localhost:8080/tts?lang=english" --data-binary @- -o test_en.wav
echo "Привет мир" | curl -X POST "http://localhost:8080/tts?lang=russian" --data-binary @- -o test_ru.wav
```

## PowerShell Examples

```powershell
# Test all languages
$languages = @{
    "portuguese" = "Bem-vindo"
    "spanish"    = "Bienvenido"
    "english"    = "Welcome"
    "russian"    = "Добро пожаловать"
}

foreach ($lang in $languages.Keys) {
    $text = $languages[$lang]
    Write-Host "Testing $lang : $text"
    Invoke-RestMethod -Uri "http://localhost:8080/tts?lang=$lang" `
        -Method Post -Body $text -OutFile "test_$lang.wav"
}
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Model not found | Run `download_models.ps1` |
| Invalid language error | Use: portuguese, spanish, english, or russian (lowercase) |
| Characters not rendering | Ensure UTF-8 terminal encoding |
| Poor audio quality | Try downloading "high" quality model instead |

## Additional Resources

- **Piper TTS Repository**: https://huggingface.co/rhasspy/piper-voices
- **Model Browser**: https://huggingface.co/rhasspy/piper-voices/tree/v1.0.0
- **Project README**: `../../README.md`
- **TTS Documentation**: `../../tts/TTS_README.md`
