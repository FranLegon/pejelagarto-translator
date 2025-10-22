# Multi-Language TTS Models

This directory contains language-specific Piper TTS models for different pronunciation languages.

## Directory Structure

```
languages/
├── russian/     (North - Default)
│   ├── model.onnx
│   └── model.onnx.json
├── german/      (North-East)
│   ├── model.onnx
│   └── model.onnx.json
├── turkish/     (North-East-East)
│   ├── model.onnx
│   └── model.onnx.json
├── portuguese/  (East)
│   ├── model.onnx
│   └── model.onnx.json
├── french/      (Center)
│   ├── model.onnx
│   └── model.onnx.json
├── hindi/       (South-East)
│   ├── model.onnx
│   └── model.onnx.json
├── romanian/    (South)
│   ├── model.onnx
│   └── model.onnx.json
├── icelandic/   (South-South-East)
│   ├── model.onnx
│   └── model.onnx.json
├── arabic/      (South-West)
│   ├── model.onnx
│   └── model.onnx.json
├── swedish/     (South-West-West)
│   ├── model.onnx
│   └── model.onnx.json
├── vietnamese/  (South-South-West)
│   ├── model.onnx
│   └── model.onnx.json
├── czech/       (West)
│   ├── model.onnx
│   └── model.onnx.json
├── chinese/     (North-West)
│   ├── model.onnx
│   └── model.onnx.json
├── norwegian/   (North-West)
│   ├── model.onnx
│   └── model.onnx.json
├── hungarian/   (North-North-West)
│   ├── model.onnx
│   └── model.onnx.json
└── kazakh/      (North-North-East)
    ├── model.onnx
    └── model.onnx.json
```

## Download Models

You need to download the appropriate ONNX models and their configuration files for each language you want to support.

### Portuguese (Brazilian)

**Recommended Model**: `pt_BR-faber-medium`

```powershell
# Download model
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx" -OutFile "portuguese\model.onnx"

# Download config
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx.json" -OutFile "portuguese\model.onnx.json"
```



### Czech (West)

**Recommended Model**: `cs_CZ-jirka-medium`

```powershell
# Download model
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/cs/cs_CZ/jirka/medium/cs_CZ-jirka-medium.onnx" -OutFile "czech\model.onnx"

# Download config
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/cs/cs_CZ/jirka/medium/cs_CZ-jirka-medium.onnx.json" -OutFile "czech\model.onnx.json"
```

### Romanian (South)

**Recommended Model**: `ro_RO-mihai-medium`

```powershell
# Download model
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ro/ro_RO/mihai/medium/ro_RO-mihai-medium.onnx" -OutFile "romanian\model.onnx"

# Download config
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ro/ro_RO/mihai/medium/ro_RO-mihai-medium.onnx.json" -OutFile "romanian\model.onnx.json"
```

### Portuguese (East - Legacy)

**Recommended Model**: `ru_RU-irina-medium`

```powershell
# Download model
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ru/ru_RU/irina/medium/ru_RU-irina-medium.onnx" -OutFile "russian\model.onnx"

# Download config
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ru/ru_RU/irina/medium/ru_RU-irina-medium.onnx.json" -OutFile "russian\model.onnx.json"
```

## Alternative: Download All Models at Once

Run this PowerShell script to download all models:

```powershell
# Navigate to the languages directory
cd tts\requirements\piper\languages

# Russian (North - Default)
Write-Host "Downloading Russian model..." -ForegroundColor Green
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ru/ru_RU/irina/medium/ru_RU-irina-medium.onnx" -OutFile "russian\model.onnx"
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ru/ru_RU/irina/medium/ru_RU-irina-medium.onnx.json" -OutFile "russian\model.onnx.json"

# Portuguese (East)
Write-Host "Downloading Portuguese model..." -ForegroundColor Green
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx" -OutFile "portuguese\model.onnx"
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/pt/pt_BR/faber/medium/pt_BR-faber-medium.onnx.json" -OutFile "portuguese\model.onnx.json"

# Czech (West)
Write-Host "Downloading Czech model..." -ForegroundColor Green
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/cs/cs_CZ/jirka/medium/cs_CZ-jirka-medium.onnx" -OutFile "czech\model.onnx"
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/cs/cs_CZ/jirka/medium/cs_CZ-jirka-medium.onnx.json" -OutFile "czech\model.onnx.json"

# Romanian (South)
Write-Host "Downloading Romanian model..." -ForegroundColor Green
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ro/ro_RO/mihai/medium/ro_RO-mihai-medium.onnx" -OutFile "romanian\model.onnx"
Invoke-WebRequest -Uri "https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/ro/ro_RO/mihai/medium/ro_RO-mihai-medium.onnx.json" -OutFile "romanian\model.onnx.json"

Write-Host "All models downloaded successfully!" -ForegroundColor Cyan
```

## Model Information

All models are from the official Piper TTS repository on Hugging Face:
- **Repository**: https://huggingface.co/rhasspy/piper-voices
- **License**: Models have various licenses (mostly MIT, CC-BY-4.0)
- **Quality**: Medium quality models provide good balance between quality and file size

### Model Sizes (Approximate)

- Russian (ru_RU-irina-medium): ~63 MB - **North (Default)**
- German (de_DE-thorsten-medium): ~63 MB - **North-East**
- Turkish (tr_TR-dfki-medium): ~63 MB - **North-East-East**
- Portuguese (pt_BR-faber-medium): ~63 MB - **East**
- French (fr_FR-siwis-medium): ~63 MB - **Center**
- Hindi (hi_IN-pratham-medium): ~63 MB - **South-East**
- Romanian (ro_RO-mihai-medium): ~63 MB - **South**
- Icelandic (is_IS-bui-medium): ~63 MB - **South-South-East**
- Arabic (ar_JO-kareem-medium): ~63 MB - **South-West**
- Swedish (sv_SE-nst-medium): ~63 MB - **South-West-West**
- Vietnamese (vi_VN-vais1000-medium): ~63 MB - **South-South-West**
- Czech (cs_CZ-jirka-medium): ~63 MB - **West**
- Chinese (zh_CN-huayan-medium): ~63 MB - **North-West**
- Norwegian (no_NO-talesyntese-medium): ~63 MB - **North-West**
- Hungarian (hu_HU-anna-medium): ~63 MB - **North-North-West**
- Kazakh (kk_KZ-iseke-x_low): ~28 MB - **North-North-East** (x_low quality)

**Total**: ~988 MB for all 16 languages

## Usage

After downloading the models, you can use them by:

1. **Command-line flag**:
   ```bash
   .\pejelagarto-translator.exe -pronunciation_language=czech
   ```

2. **HTTP API query parameter**:
   ```bash
   curl -X POST "http://localhost:8080/tts?lang=romanian" -d "Bună ziua"
   ```

3. **Default language is now Russian (North)**

## Browse All Available Models

Visit the Hugging Face repository to see all available voices and languages:
https://huggingface.co/rhasspy/piper-voices/tree/v1.0.0

You can find alternative voices for each language with different qualities (low, medium, high) and different voice characteristics.
