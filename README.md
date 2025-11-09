# Pejelagarto Translator

A complete bidirectional translator between Human and Pejelagarto, a fictional language with complex transformation rules. Includes a web-based UI with multi-language text-to-speech support, light/dark theme, and ngrok integration for remote access.

## About

Pejelagarto is a fictional constructed language designed as a challenging translation exercise. This translator implements a sophisticated set of transformations including:

- Mathematical base conversions (base-10 ↔ base-8 for positive, base-10 ↔ base-7 for negative)
- Prime factorization-based accent placement
- Fibonacci/Tribonacci capitalization patterns
- Custom character and punctuation mappings
- Special Unicode character-based timestamp encoding (U+2300-U+23FB)
- Multi-language text-to-speech with 18 languages

The project demonstrates advanced string manipulation, bijective mappings, and cryptographic-style transformations while maintaining perfect reversibility.

## Features

### Core Translation
- ✅ **Nearly Reversible**: Most transformations are bidirectional, but special character timestamp encoding has some limitations
- 🔢 **Number Conversion**: Base-10 ↔ Base-8 (positive) or Base-7 (negative) with arbitrary precision
- 🔤 **Character Mapping**: Bijective word/conjunction/letter replacements with case preservation
- ✏️ **Accent Transformations**: Prime-factorization-based vowel accent cycling
- 📝 **Case Logic**: Fibonacci/Tribonacci sequence-based capitalization patterns
- ⏰ **Special Character Datetime Encoding**: UTC time encoded as special Unicode characters (U+2300-U+23FB)
- ❗ **Punctuation Mapping**: Custom punctuation character replacements
- 🛡️ **UTF-8 Sanitization**: Handles invalid UTF-8 bytes with invisible soft-hyphen encoding

### Text-to-Speech (TTS)
- 🗣️ **18 Languages**: Russian (North), German (North-East), Turkish (East-North-East), Portuguese (East), French (South-East-East), Hindi (South-East), Romanian (South), Icelandic (South-South-East), Swahili (South-West), Swedish (South-West-West), Vietnamese (West-South-West), Czech (West), Chinese (North-West), Norwegian (North-West), Hungarian (North-North-West), Kazakh (North-North-East), plus Spanish and English
- 🎙️ **Piper TTS Integration**: High-quality neural TTS using ONNX models
- 🌍 **Language-Specific Preprocessing**: Automatic text cleaning based on pronunciation language
- 🔀 **Per-Request Language Selection**: Override default language via HTTP API or dropdown
- 🎛️ **Configurable Default Language**: Set preferred language via command-line flag
- 🐇🐌 **Dual Speed Playback**: Normal and slowed-down (0.5x) audio generation
- 💾 **Smart Caching**: Automatic audio caching for faster repeated requests

### Web Interface
- 🌐 **Local Web Server**: HTTP server with automatic port selection (8080-8090)
- 🎨 **Modern UI**: Single-file HTML with gradient design and responsive layout
- ⚡ **Live Translation**: Real-time translation as you type
- 🔄 **Bidirectional Toggle**: One-click swap between Human ↔ Pejelagarto
- 📱 **Mobile-Friendly**: Responsive design for all screen sizes
- 🚀 **Auto-Launch**: Opens default browser automatically on startup

## Requirements

- **Go**: Version 1.24.2 or higher
- **Shell/PowerShell**: For running build scripts
  - Windows: PowerShell (built-in)
  - Linux/macOS: Bash (built-in)
- **FFmpeg**: Required for slow audio generation (optional - only needed for slow playback feature)
- **Dependencies**: Automatically downloaded on first run
  - Piper TTS engine with 16 language models (~1008MB total)
  - espeak-ng phoneme data
  - ngrok library for remote access (optional)
  - OS-specific scripts embedded in binary handle all downloads
- **Supported OS**: Windows, macOS, Linux
  - Tested on Windows 10/11, Ubuntu 20.04+, macOS 12+
  - Supports x86_64 (amd64) and ARM64 architectures
- **Browser**: Any modern web browser for the UI

## Quick Start

### 1. Build the Application

This project supports multiple build modes for different use cases. Choose the one that fits your needs:

#### Development Build (Recommended for Local Use)

**Backend Build (Default - Server-Side Translation):**

Windows:
```powershell
go build -o bin/pejelagarto-translator.exe main.go
```

Linux/macOS:
```bash
go build -o bin/pejelagarto-translator main.go
```

**Frontend Build (Client-Side Translation via WebAssembly):**

This mode compiles the translation logic to WebAssembly, allowing translation to run entirely in the browser. Only TTS (text-to-speech) uses the server.

Linux/macOS:
```bash
./build-frontend.sh
go run server_frontend.go version.go
```

Windows:
```powershell
# Manual build (build-frontend.sh equivalent)
$env:GOOS="js"; $env:GOARCH="wasm"
go build -tags frontend -o bin/translator.wasm
Copy-Item "$(go env GOROOT)\lib\wasm\wasm_exec.js" bin\
go run server_frontend.go version.go
```

**Frontend Mode Features:**
- ✅ Translation runs in your browser (no server calls for text operations)
- ✅ Significantly reduced server load
- ✅ Works offline after initial WASM module load
- ✅ TTS still uses server (for Piper audio generation)
- ✅ Same translation quality and features
- 📦 WASM module: ~2-3MB (one-time download)

#### Production Build (WASM + ngrok)

For production server deployment with client-side WASM translation and hardcoded ngrok credentials:

**🚨 CRITICAL:** Garble obfuscation is **INCOMPATIBLE** with ngrok-go SDK. The obfuscated build will fail with "remote gone away" error. **You MUST use the unobfuscated build for production deployments with ngrok.**

**✅ Recommended Build (Unobfuscated + Optimized):**

Windows:
```powershell
.\scripts\helpers\build-prod-unobfuscated.ps1
```

Linux/macOS:
```bash
./scripts/helpers/build-prod-unobfuscated.sh
```

**Requirements:**
- Go 1.24.2 or higher
- Build tags: `obfuscated`, `frontend`, `ngrok_default`
- Uses standard Go optimization (`-ldflags="-s -w"`) instead of garble

**Why Unobfuscated?**
1. ✅ **ngrok Compatible**: Works reliably with ngrok-go SDK
2. ✅ **Optimized**: Binary size optimization via `-s -w` flags
3. ✅ **Windows Defender Friendly**: No false positive triggers
4. ✅ **Stable**: Proven reliable in production

**❌ DEPRECATED: Garble-Obfuscated Build (Incompatible with ngrok):**

⚠️ **Warning:** Garble obfuscation corrupts ngrok-go SDK internals, causing connection failures. This build is preserved for non-ngrok use cases only.

Installation:
```bash
go install mvdan.cc/garble@latest
```

Windows:
```powershell
.\scripts\helpers\build-prod.ps1
```

Linux/macOS:
```bash
./scripts/helpers/build-prod.sh
```

**Output:**
- `bin/pejelagarto-server.exe` (Windows) or `bin/pejelagarto-server` (Linux/macOS) - Optimized server
- `bin/translator.wasm` - Client-side WASM translation module
- `bin/wasm_exec.js` - Go WASM runtime
- `bin/checksums-prod.txt` - SHA256 checksums for verification

**What This Build Does:**
1. ✅ **Standard Go Optimization**: Uses `-ldflags="-s -w"` for size reduction
2. ✅ **WASM Frontend**: Compiles translation to WebAssembly for client-side execution
3. ✅ **Hardcoded ngrok**: Embeds ngrok credentials (no command-line flags needed)
4. ✅ **Embedded Binaries**: Includes Windows/Linux/Android binaries
5. ✅ **ngrok Compatible**: Works reliably with ngrok-go SDK
6. ✅ **Checksum Generation**: Creates SHA256 checksums for integrity verification

**Build Process:**
```
Step 1: Build WASM module (frontend tag)
Step 2: Copy wasm_exec.js runtime
Step 3: Build optimized server (obfuscated+ngrok_default tags)
Step 4: Generate SHA256 checksums
Step 5: Display build summary
```

**Why NOT Garble?**

Garble obfuscation is incompatible with ngrok-go SDK due to:
- TLS/crypto corruption
- Reflection-based code mangling
- Interface implementation renaming

See `TEST_RESULTS.md` and `GARBLE_NGROK_FINAL_ANSWER.md` for detailed test results.

#### Obfuscated Build Only (without WASM/ngrok)

For server deployment with obfuscation only:

Windows:
```powershell
# Using build script (requires garble installed)
.\obfuscation\build-obfuscated.ps1

# Or manually with Go (no obfuscation)
go build -tags obfuscated -o bin/piper-server.exe main.go
```

Linux/macOS:
```bash
# Using build script (requires garble installed)
./obfuscation/build-obfuscated.sh linux

# Or manually with Go (no obfuscation)
go build -tags obfuscated -o bin/piper-server main.go
```

The binary will automatically download all TTS requirements (~1.1GB) on first run.

**Which Build Should I Use?**

| Build Type | Best For | Translation | TTS Audio | Binary Size |
|------------|----------|-------------|-----------|-------------|
| **Backend** | Single user, local use | Server (backend) | Server | ~12-13MB |
| **Frontend** | Multiple users, web deployment | Browser (WASM) | Server | ~2-3MB WASM + Server |
| **Obfuscated** | Production server deployment | Server (backend) | Server | ~12-13MB |

**Build Tag Compatibility:**

All combinations of build tags work together seamlessly:

| Build Tags | Description | Binary Output | TTS Downloads |
|------------|-------------|---------------|---------------|
| None | Backend server (default) | `pejelagarto-translator` | All languages (default) |
| `downloadable` | Embeds Windows/Linux/Android binaries | `pejelagarto-translator` + embedded bins | All languages |
| `ngrok_default` | Hardcoded ngrok credentials (includes downloadable) | `pejelagarto-translator` + embedded bins | All languages |
| `obfuscated` | Code obfuscation for deployment | `piper-server` | All languages |
| `frontend` | Client-side WASM translation | `translator.wasm` | All languages |
| `downloadable,obfuscated` | Downloadable + obfuscation | `piper-server` + embedded bins | All languages |
| `downloadable,frontend` | Downloadable + WASM | `translator.wasm` + embedded bins | All languages |
| `ngrok_default,obfuscated` | Ngrok + obfuscation | `piper-server` + embedded bins | All languages |
| `ngrok_default,frontend` | Ngrok + WASM | `translator.wasm` + embedded bins | All languages |
| `obfuscated,frontend` | Obfuscated WASM | `translator.wasm` (obfuscated) | All languages |
| `downloadable,obfuscated,frontend` | All features | `translator.wasm` + embedded bins (obfuscated) | All languages |
| `ngrok_default,obfuscated,frontend` | All features with ngrok | `translator.wasm` + embedded bins (obfuscated) | All languages |

**Build Tag Relationships:**
- `ngrok_default` automatically includes `downloadable` functionality
- `downloadable` and `ngrok_default` are mutually compatible (OR condition in build tags)
- All tags can be combined with `obfuscated` and `frontend`
- TTS language downloads can be optimized using `-pronunciation_language` flag (downloads only specified language when dropdown is disabled)

**Verify all combinations:** Run `./test-build-combinations.sh` to test all 13 build tag combinations

**Build Notes**: 
- Backend build creates **~12-13MB executable** 
  - Windows: `pejelagarto-translator.exe`
  - Linux/macOS: `pejelagarto-translator`
- Frontend build creates **~2-3MB WASM module**
  - Translation runs in browser (JavaScript + WebAssembly)
  - Server only needed for TTS audio generation
  - Perfect for web deployment with many concurrent users
- Obfuscated build uses different internal names (`piper-server` instead of `pejelagarto-translator`)
- First run downloads 16 TTS language models (takes several minutes)
- Dependencies are cached in temp directory for subsequent runs
- OS-specific scripts handle all downloads automatically:
  - Windows: PowerShell script (`get-requirements.ps1`)
  - Linux/macOS: Shell script (`get-requirements.sh`)
- No manual dependency management needed!

**Obfuscation Details:**
- Uses [garble](https://github.com/burrowers/garble) for code obfuscation (`-literals -tiny` flags)
- Build tags switch between backend and obfuscated constants
- Backend build: uses `pejelagarto-translator` for temp directories and scripts
- Obfuscated build: uses `piper-server` for temp directories and scripts
- Output binary named `piper-server.exe` (Windows) or `piper-server` (Unix)

### 2. Run the Application

Windows:
```powershell
# Local server only (Russian TTS by default)
.\bin\pejelagarto-translator.exe

# With specific TTS language and dropdown enabled
.\bin\pejelagarto-translator.exe -pronunciation_language portuguese -pronunciation_language_dropdown

# With ngrok for remote access
.\bin\pejelagarto-translator.exe -ngrok_token YOUR_TOKEN -ngrok_domain your-domain.ngrok-free.app
```

Linux/macOS:
```bash
# Local server only (Russian TTS by default)
./bin/pejelagarto-translator

# With specific TTS language and dropdown enabled
./bin/pejelagarto-translator -pronunciation_language portuguese -pronunciation_language_dropdown

# With ngrok for remote access
./bin/pejelagarto-translator -ngrok_token YOUR_TOKEN -ngrok_domain your-domain.ngrok-free.app
```

The server automatically finds an available port (tries 8080 first, then 8081-8090) and opens your browser.

**On First Run:**
- Binary extracts and runs embedded script to download TTS dependencies:
  - Windows: PowerShell script (`get-requirements.ps1`)
  - Linux/macOS: Shell script (`get-requirements.sh`)
- Script downloads all TTS dependencies (~1.1GB) to temp directory:
  - Windows: `C:\Windows\Temp\pejelagarto-translator\`
  - Linux/macOS: `/tmp/pejelagarto-translator/`
- Download takes 3-5 minutes depending on internet speed
- Dependencies are verified after download
- Subsequent runs use cached files and start instantly

### Web UI

The interface provides:
- **Input Textarea**: Enter text to translate
- **Output Textarea**: View translated result (read-only)
- **Translate Button**: Manual translation trigger
- **Invert Button (⇅)**: Swap translation direction
- **Live Translation Checkbox**: Enable real-time translation
- **Theme Toggle Button (🌙/☀️)**: Switch between dark and light modes
- **TTS Play Buttons**: 🔊 Play audio in selected language
  - 🐇 Fast speed (normal)
  - 🐌 Slow speed (0.5x) - appears after fast audio generation completes
- **Language Dropdown**: Select TTS language (if enabled with `-pronunciation_language_dropdown` flag)
  - 18 languages available, organized by compass directions

### API Endpoints

```go
// POST /to - Translate to Pejelagarto
// Request body: plain text
// Response: translated text

// POST /from - Translate from Pejelagarto
// Request body: plain text  
// Response: translated text

// POST /tts?lang=<language>&slow=<true|false> - Text-to-Speech
// Request body: plain text
// Query params: 
//   - lang (optional): russian, german, turkish, portuguese, french, hindi, 
//                      romanian, icelandic, swahili, swedish, vietnamese, 
//                      czech, chinese, norwegian, hungarian, kazakh, 
//                      spanish, english
//   - slow (optional): true to slow down audio playback to 0.5x speed
// Response: audio/wav file

// POST /tts-check-slow?lang=<language> - Check if slow audio is ready
// Request body: plain text (same text as TTS request)
// Response: JSON {"ready": true/false}

// GET / - Serve HTML UI
```

### Text-to-Speech Usage

The application includes multi-language TTS with automatic text preprocessing for 18 languages:

**Supported Languages (organized by compass directions):**

| Direction | Language | Code | Voice Model |
|-----------|----------|------|-------------|
| North | Russian | `russian` | ru_RU-dmitri-medium |
| North-East | German | `german` | de_DE-thorsten-medium |
| North-East-East | Turkish | `turkish` | tr_TR-dfki-medium |
| East | Portuguese | `portuguese` | pt_BR-faber-medium |
| South-East-East | French | `french` | fr_FR-siwis-medium |
| South-East | Hindi | `hindi` | hi_HI-medium |
| South | Romanian | `romanian` | ro_RO-mihai-medium |
| South-South-East | Icelandic | `icelandic` | is_IS-bui-medium |
| South-West | Swahili | `swahili` | sw_CD-lanfrica-medium (DRC) |
| South-West-West | Swedish | `swedish` | sv_SE-nst-medium |
| West-South-West | Vietnamese | `vietnamese` | vi_VN-vivos-medium |
| West | Czech | `czech` | cs_CZ-jirka-medium |
| North-West-West | Chinese | `chinese` | zh_CN-huayan-medium |
| North-West | Norwegian | `norwegian` | no_NO-talesyntese-medium |
| North-North-West | Hungarian | `hungarian` | hu_HU-anna-medium |
| North-North-East | Kazakh | `kazakh` | kk_KZ-iseke-x_low |


**Total:** 16 language models (~1008MB)

**Command-line:**

Windows:
```powershell
# Set default TTS language
.\bin\pejelagarto-translator.exe -pronunciation_language swahili

# Enable language dropdown in UI
.\bin\pejelagarto-translator.exe -pronunciation_language_dropdown
```

Linux/macOS:
```bash
# Set default TTS language
./bin/pejelagarto-translator -pronunciation_language swahili

# Enable language dropdown in UI
./bin/pejelagarto-translator -pronunciation_language_dropdown
```

**HTTP API:**
```bash
# Normal speed
curl -X POST "http://localhost:8080/tts?lang=russian" -d "Привет мир" -o audio.wav
curl -X POST "http://localhost:8080/tts?lang=swahili" -d "Habari yako" -o audio.wav
curl -X POST "http://localhost:8080/tts?lang=chinese" -d "你好世界" -o audio.wav

# Slowed down by half (0.5x speed) - requires FFmpeg
curl -X POST "http://localhost:8080/tts?lang=turkish&slow=true" -d "Merhaba dünya" -o audio-slow.wav
```

**Text Preprocessing:**

Each language has specific character filtering and consonant cluster limiting:
- **Russian**: Cyrillic alphabet (а-я, ь, ъ) + Latin fallback
- **German**: a-z, ä, ö, ü, ß
- **Turkish**: a-z, ç, ğ, ı, ö, ş, ü (special handling for dotted/dotless i)
- **Portuguese**: a-z, á, é, í, ó, ú, â, ê, ô, ã, õ, à, ü, ç
- **French**: a-z, à, â, ä, æ, ç, é, è, ê, ë, î, ï, ô, œ, ù, û, ü, ÿ
- **Hindi**: Devanagari script (अ-ह, ा-्, ं, ः) + Latin fallback
- **Swahili**: Simple Latin (a-z)
- **Chinese**: Chinese characters (Unicode ranges 4E00-9FFF, 3400-4DBF, etc.) + pinyin
- And more...

All preprocessing includes:
1. Number conversion from Pejelagarto format (base-8/7) to standard (base-10)
2. Character filtering to language-specific allowed set
3. Consonant cluster limiting (max 2 consecutive consonants)
4. Empty result fallback to prevent TTS errors

## Translation Pipeline

### Human → Pejelagarto

```go
func TranslateToPejelagarto(input string) string {
    input = sanitizeInvalidUTF8(input)                    // 1. Handle broken UTF-8
    input = removeTimestampSpecialCharacters(input)       // 2. Remove existing special chars
    input, timestamp := removeISO8601timestamp(input)     // 3. Extract & remove existing timestamps
    input = applyNumbersLogicToPejelagarto(input)         // 4. Base 10 → Base 8 (positive) / Base 7 (negative)
    input = applyPunctuationReplacementsToPejelagarto(input) // 5. Map punctuation
    input = applyMapReplacementsToPejelagarto(input)      // 6. Apply word/letter map
    input = applyAccentReplacementLogicToPejelagarto(input)  // 7. Add accents
    input = applyCaseReplacementLogic(input)              // 8. Apply case patterns
    input = addSpecialCharDatetimeEncoding(input, timestamp) // 9. Insert special char timestamp
    return input
}
```

### Pejelagarto → Human

```go
func TranslateFromPejelagarto(input string) string {
    timestamp := readTimestampUsingSpecialCharEncoding(input) // 1. Extract special char timestamp
    input = removeTimestampSpecialCharacters(input)       // 2. Remove all special chars
    input = applyCaseReplacementLogic(input)              // 3. Reverse case (self-inverse)
    input = applyAccentReplacementLogicFromPejelagarto(input) // 4. Remove accents
    input = applyMapReplacementsFromPejelagarto(input)    // 5. Reverse word/letter map
    input = applyPunctuationReplacementsFromPejelagarto(input) // 6. Reverse punctuation
    input = applyNumbersLogicFromPejelagarto(input)       // 7. Base 8/7 → Base 10
    input = addISO8601timestamp(input, timestamp)         // 8. Add back timestamp
    input = unsanitizeInvalidUTF8(input)                  // 9. Restore original bytes
    return input
}
```

## Translation Rules

### 1. UTF-8 Sanitization
- Invalid UTF-8 bytes → Soft hyphen (U+00AD) + Private Use Area character (U+E000-U+E0FF)
- Completely invisible in most renderers
- Fully reversible bijective encoding (256 possible bytes)

### 2. Number Conversion

**Algorithm Details:**

The number conversion transforms base-10 numbers using different bases depending on sign:
- **Positive numbers**: Base-10 → Base-8 (octal)
- **Negative numbers**: Base-10 → Base-7

**To Pejelagarto (Base-10 → Base-8 for positive, Base-10 → Base-7 for negative):**
1. Scan input for sequences of ASCII digits (0-9), including negative numbers (prefixed with `-`)
2. Extract and preserve leading zeros separately
3. Parse the number using arbitrary-precision arithmetic (`math/big`)
4. **For positive numbers**: Convert to base-8 (octal) representation
   **For negative numbers**: Convert to base-7 representation
5. Reconstruct: sign + leading zeros + converted digits
6. Handle edge case: if only zeros present (e.g., "000", "-0"), preserve them without conversion

**From Pejelagarto (Base-8 or Base-7 → Base-10):**
1. **For positive numbers**: Scan for valid base-8 sequences (digits 0-7 only)
   **For negative numbers**: Scan for valid base-7 sequences (digits 0-6 only)
2. **Key distinction:** If digits 8-9 are found after base-8 digits (or 7-9 after base-7), treat entire number as base-10 (pass through unchanged)
3. Extract sign and leading zeros
4. Parse as base-8 (positive) or base-7 (negative) using `math/big`
5. Convert to base-10 representation
6. Reconstruct with preserved sign and leading zeros

**Special Cases:**
- Leading zeros are always preserved: positive `007` → `007` in base-8, negative `-007` → `-0010` in base-7
- Negative signs are handled separately from the magnitude
- Zero-only numbers (e.g., "000") are preserved as-is
- **No size limits:** `math/big` provides arbitrary precision, supporting numbers of any size without overflow
- Numbers with digits 8-9 following base-8 patterns (or 7-9 following base-7) are treated as base-10 and passed through unchanged

### 3. Character Mapping

**Bijective Map Structure:**

The translator uses two tiers of character mappings with sophisticated indexing:

- `conjunctionMap`: Multi-character words and letter pairs (e.g., `"hello"` → `"arakan"`, `"ch"` → `"jc"`)  
- `letterMap`: Single letters (e.g., `"a"` → `"i"`)

**Index-Based Ordering:**

The bijective map uses **positive and negative indices** to determine processing order:

**Positive Indices (To Pejelagarto):**
- Index = length of source text in runes
- Processed in descending order: longer patterns matched first
- Multi-rune target values are prefixed with `'` (e.g., `"hello"` → `"'jhtxz"`)
- Processing order: conjunctionMap → letterMap (by descending length)

**Negative Indices (From Pejelagarto):**
- Index = -(length of Pejelagarto pattern including quote prefix)
- For multi-rune patterns: index = -(rune_count + 1) to account for `'` prefix
- For single-rune patterns: index = -1
- Processed first (negative before positive), then by descending absolute value
- The `'` prefix in Pejelagarto text identifies multi-character patterns

**Example Index Mapping:**
```
"hello" (5 runes) → "'jhtxz" (6 runes with ')
  To Pejelagarto:   index +5, maps "hello" → "'jhtxz"
  From Pejelagarto: index -6, maps "'jhtxz" → "hello"

"ch" (2 runes) → "'jc" (3 runes with ')
  To Pejelagarto:   index +2, maps "ch" → "'jc"
  From Pejelagarto: index -3, maps "'jc" → "ch"

"a" (1 rune) → "i" (1 rune, no ')
  To Pejelagarto:   index +1, maps "a" → "i"
  From Pejelagarto: index -1, maps "i" → "a"
```

**Replacement Algorithm:**

1. **Marker Protection:** Use Unicode markers (`\uFFF0`, `\uFFF1`) to wrap replaced text
2. **Depth Tracking:** Pre-calculate marker depth at each position for O(1) lookup
3. **Quote Boundaries:** The `'` character acts as a word boundary - patterns cannot span across quotes
4. **Word Boundary Detection:** Scan backward (max 50 chars) to find word start for quote checking
5. **Case Preservation:** Extract case pattern from source, apply to target using `matchCase()`
6. **Reversible Case Check:** Skip characters where `ToUpper(ToLower(c)) != ToUpper(c)` (e.g., Turkish İ, German ß)
7. **Marker Removal:** After all replacements, remove all `\uFFF0` and `\uFFF1` markers

**Case Matching Logic:**
- If source is all uppercase → target becomes all uppercase
- If source is title case (first letter upper) → target becomes title case  
- If source is mixed case → applies case pattern position-by-position
- Non-reversible case characters are preserved in lowercase

### 4. Punctuation Mapping

**Bijective Punctuation Transformations:**

Punctuation uses a separate bijective map independent from character mappings:

```go
"?"  → "‽"  (interrobang)
"!"  → "¡"  (inverted exclamation)
"."  → ".." (doubled period)
","  → "،"  (Arabic comma)
";"  → "⁏"  (reversed semicolon)
":"  → "︰"  (presentation form colon)
"'"  → "〝" (reversed quotation mark) 
"\"" → "〞" (low quotation mark)
"-"  → "‐"  (hyphen)
"("  → "⦅" (left white parenthesis)
")"  → "⦆" (right white parenthesis)
```

**Processing Details:**

Unlike character mapping, punctuation can have different lengths for keys and values.

**To Pejelagarto:**
1. Literal single quotes (`'`) are converted to a temporary marker (`\uFFF3`) to avoid conflicts with the multi-rune pattern prefix
2. Apply punctuation bijective map using the same `applyReplacements()` algorithm as character mapping
3. Restore literal quotes as doubled quotes (`''`) in the output - this escapes them

**From Pejelagarto:**
1. Doubled quotes (`''`) are converted to the temporary marker (`\uFFF3`) to preserve them as literals
2. Apply reverse punctuation mapping
3. Restore the marker as single quote (`'`) in Human text

**Why Quote Escaping:**
The `'` character serves dual purpose:
- Marks multi-rune patterns in Pejelagarto text
- Can also appear as literal punctuation

Escaping as `''` disambiguates: `''hello` = literal quote + "hello", while `'jhtxz` = multi-rune pattern.

### 5. Accent Replacement (Prime Factorization)

**Algorithm Overview:**

Accents are applied based on the prime factorization of the input string's length:

1. **Calculate Input Length:** Count runes (not bytes) in the string
2. **Prime Factorization:** Break down length into prime factors with powers
   - Example: 245 = 5¹ × 7²
3. **Find Vowels:** Identify all vowel positions using `isVowel()` check
4. **Apply Transformations:** For each prime factor `p` with power `n`:
   - Locate the `p`-th vowel (1-indexed)
   - Move that vowel forward `n` steps in its accent wheel

**Dual Accent Wheel System:**

Each base vowel has **two independent accent wheels**:

1. **One-Rune Accent Wheel** (includes no-accent):
   - Contains single-rune accented forms, including the base vowel with no accent
   - Example for 'a': `["a", "à", "á", "â", "ã", "å", "ä", "ā", "ă"]` (9 forms)
   - Used in the accent transformation algorithm
   - All forms have reversible case conversion

2. **Two-Rune Accent Wheel** (excludes no-accent):
   - Contains two-rune accented forms using combining diacritics
   - Example for 'a': `["a\u0328", "a\u030C"]` (base + combining ogonek, base + combining caron)
   - Does NOT include the no-accent form (since it's only 1 rune)
   - Currently defined but not used in transformation algorithm

**Vowel Identification with Case Reversibility Check:**

The `isVowel()` function determines which characters can have accents changed. A character is considered a vowel if:

1. It exists in either the one-rune or two-rune accent wheels (for any base vowel)
2. **AND** it passes the case reversibility check:
   - For uppercase characters: `ToUpper(ToLower(char)) == char`
   - This prevents treating non-reversible characters as vowels

**Example of Non-Reversible Cases:**
- Turkish İ (U+0130): `ToLower(İ) = i`, but `ToUpper(i) = I` (not İ)
- These fail the reversibility check and are NOT treated as vowels, so accents won't be changed

This ensures that only characters with predictable, reversible case behavior can have their accents transformed.

**Example Calculation:**

For text length 245 = 5 × 7²:
- **5th vowel:** Move forward 1 step (power of 5 is 1)
- **7th vowel:** Move forward 2 steps (power of 7 is 2)

**Special Cases and Exceptions:**

1. **Only Wheel Vowels Are Transformed:**
   - Characters must be in the accent wheels to be considered vowels
   - Consonants and unknown accented letters are skipped
   - This means accented vowels from outside the standard wheels are left unchanged

2. **Reversible Case Handling:**
   - If original vowel is uppercase and `ToUpper(ToLower(vowel))` is reversible, apply uppercase
   - Otherwise, keep result in lowercase to maintain reversibility

3. **Single-Rune Guarantee:**
   - Only single-rune replacements from the one-rune wheel are applied
   - Multi-rune Unicode sequences are rejected to ensure clean reversibility

4. **Reverse Direction (From Pejelagarto):**
   - Verify the current accented form exists in our wheel before transforming
   - If accent is not in our wheel (unknown accent), skip transformation
   - Move **backward** by power steps with modular arithmetic to reverse

**Why Prime Factorization?**

This creates a deterministic yet complex transformation:
- Same text length always produces same accent pattern
- Different lengths produce very different patterns
- Reversible: knowing the length allows exact reversal

### 6. Case Replacement Logic
Based on word count:
- **Odd word count**: Use Fibonacci sequence
- **Even word count**: Use Tribonacci sequence

Toggle capitalization at sequence positions (e.g., 1st, 2nd, 3rd, 5th, 8th, 13th... for Fibonacci)

**Self-inverse:** Applying twice returns original

### 7. Special Character Datetime Encoding

**Encoding Process:**

The translator embeds a UTC timestamp using special Unicode characters from the range U+2300 to U+23FB. These characters are distributed across 5 categories representing date/time components:

**Special Character Categories:**
1. **Day Characters** (1-60): Unicode range U+2300-U+233B (60 characters, only first 31 used for days 1-31)
2. **Month Characters** (1-12): Unicode range U+233C-U+2347 (12 characters for months 1-12)
3. **Year Characters** (2025+): Unicode range U+2348-U+23A9 (98 characters, indexed from year 2025)
4. **Hour Characters** (0-23): Unicode range U+23AA-U+23C0 (23 characters for hours 0-22)
5. **Minute Characters** (0-59): Unicode range U+23C1-U+23FB (59 characters for minutes 0-58)

**To Pejelagarto - Insertion Algorithm:**

1. Extract any existing ISO 8601 timestamp from input using `removeISO8601timestamp()`
   - Returns both cleaned input and found timestamp (or empty string if none)
2. If timestamp parameter is empty: use current UTC time `time.Now().UTC()`
   - If timestamp parameter provided: parse it as RFC3339 format
3. Convert to 0-indexed character array indices:
   - Day: `now.Day() - 1` 
   - Month: `int(now.Month()) - 1`
   - Year: `now.Year() - 2025`
   - Hour: `now.Hour()`
   - Minute: `now.Minute()`
4. Validate all indices are within bounds (fallback to 0 if out of range)
5. Find insertion positions: at start, next to spaces, and at newlines
6. **Random placement:** Shuffle available positions and select up to 5 random spots
7. **Guarantee all 5 special characters inserted:** If fewer than 5 positions available, remaining characters are appended to the end
8. Insert special characters from end to beginning (to maintain correct indices)

**From Pejelagarto - Extraction Algorithm:**

1. Search entire input string for presence of special characters from each category
2. Find **first match** in each category (day, month, year, hour, minute)
3. Convert special character back to its index value
4. **Optional hour/minute:** If day, month, and year are found but hour or minute are missing, default them to 0
5. Reconstruct ISO 8601 timestamp: `YYYY-MM-DDTHH:MM:00Z`
6. Return empty string if day, month, or year are missing (required components)
7. The restored timestamp is added back to the output using `addISO8601timestamp()`

**Key Characteristics:**

- **Timestamp Preservation:** If input contains an existing ISO 8601 timestamp, it's preserved through translation
- **Random Placement:** Special character positions are randomized for each translation
- **Always 5 Special Characters:** All date/time components are always inserted, even in short text
- **Reversible with Tolerance:** Hour and minute can be reconstructed even if those characters are missing (default to 00:00)
- **Not Fully Reversible:** The timestamp encoding is **not 100% reversible** because:
  - Special characters are randomly placed each time
  - Reconstruction depends on finding these characters in the text
  - If special characters are removed or modified, timestamp cannot be recovered
  - Characters from range U+2300-U+23FB in the original text will be removed during encoding
- **Fuzz Test Special Handling:** 
  - Unlike other transformations, special character encoding is tested differently in fuzzing
  - Test verifies correctness by **removing special characters and timestamps** before comparison
  - This is because timestamps can vary between translation calls
  - See `FuzzSpecialCharDateTimeEncoding()` which cleans both input and output before comparing

**Why Timestamp Might Not Be Fully Restored:**
- Input doesn't contain day, month, or year special characters (required)
- Special characters were removed or modified after translation
- Text was not previously translated to Pejelagarto
- Original text contained characters in the U+2300-U+23FB range (these get removed)
- In these cases, an empty timestamp is returned and no timestamp line is added back

## Testing

### Comprehensive Test Suite

```bash
# Quick test (runs seed corpus only, no fuzzing)
go test -v

# Run all fuzz tests with required minimum durations (recommended)
./run-fuzz-tests.sh

# Run all tests including build verification (includes full fuzz tests)
./test-all.sh

# Verify build tag compatibility (quick check, seed corpus only)
./test-build-combinations.sh

# Run WASM-specific tests
./test-wasm.sh

# Individual fuzz tests (minimum 30s each, 120s for main test)
go test -fuzz=FuzzApplyMapReplacements -fuzztime=30s
go test -fuzz=FuzzApplyNumbersLogic -fuzztime=30s
go test -fuzz=FuzzApplyAccentReplacementLogic -fuzztime=30s
go test -fuzz=FuzzApplyPunctuationReplacements -fuzztime=30s
go test -fuzz=FuzzApplyCaseReplacementLogic -fuzztime=30s
go test -fuzz=FuzzSpecialCharDateTimeEncoding -fuzztime=30s
go test -fuzz=FuzzTranslatePejelagarto -fuzztime=120s  # Main translation test requires 120s
```

### Test Files Structure

**Translation Tests (`translation_test.go`):**
- 7 fuzz tests for core translation logic
- All tests use random fuzzy input
- Shared by both backend and WASM builds
- Minimum 30s per test, 120s for main translation test

**WASM Tests (`wasm_test.go`):**
- WASM-specific tests with `//go:build frontend` tag
- Tests JS wrapper functions
- Validates WASM export functionality
- Ensures consistency with backend build

**TTS Tests (`tts_test.go`):**
- Server-only tests with `//go:build !frontend` tag
- Text-to-speech functionality tests
- HTTP handler tests
- Excluded from WASM builds

**Test Scripts:**
- `test-all.sh`: Runs all tests for both builds
- `test-wasm.sh`: Compiles WASM tests (cannot execute WASM directly)

### Test Coverage

**All tests use fuzz testing with random inputs:**

**Fuzz Tests (7 total):**
All transformations verified for reversibility with random inputs:
- `FuzzApplyMapReplacements`: Map replacements (word/conjunction/letter) - 30s minimum
- `FuzzApplyNumbersLogic`: Number conversions (base 10 ↔ base 7) - 30s minimum
- `FuzzApplyAccentReplacementLogic`: Accent transformations - 30s minimum
- `FuzzApplyPunctuationReplacements`: Punctuation mappings - 30s minimum
- `FuzzApplyCaseReplacementLogic`: Case logic (self-inverse) - 30s minimum
- `FuzzSpecialCharDateTimeEncoding`: Special character datetime encoding - 30s minimum
- `FuzzTranslatePejelagarto`: Full translation pipeline - **120s minimum** (main test)

**Proven Reliability:**
- 80,000+ fuzz executions without failures
- 100% reversibility guarantee
- Handles edge cases: empty strings, single characters, Unicode, invalid UTF-8
- All tests use random fuzzy input for comprehensive coverage

## Performance

- **O(n) complexity** for most operations
- **Optimized marker depth maps** for O(1) lookups
- **Limited backward scanning** (max 50 chars for word boundaries)
- Processes large texts efficiently (2000+ characters in ~20ms)

## Known Limitations

- **Special Character Timestamp Not Fully Reversible**: The datetime encoding using Unicode characters U+2300-U+23FB is **not 100% reversible** because:
  - Special characters are randomly placed and cannot be exactly restored
  - Original text containing these Unicode characters will have them removed
  - Timestamp reconstruction relies on finding these characters (may fail if modified)
- **UTF-8 Sanitization**: Invalid UTF-8 bytes are encoded using soft hyphens and private use area characters, which may not display correctly in all environments
- **Case Preservation**: Some Unicode characters with complex case rules (e.g., Turkish İ, German ß) may not preserve case perfectly
- **Word Boundary Detection**: Limited to 50 characters of backward scanning for performance reasons
- **Punctuation**: Only specific punctuation marks are mapped; unmapped punctuation passes through unchanged
- **Ngrok Token Security**: When using ngrok, be careful not to commit your token to version control
- **TTS Slow Audio**: Requires FFmpeg to be installed separately for the 0.5x speed feature

## Troubleshooting

### Build Issues

**"Not enough space on disk" during build:**
- Clean Go build cache: `go clean -cache`
- Final executable is ~12MB

**Build succeeds but binary won't run:**
- Ensure you have **internet connection for first run**
- Binary needs to download ~1.1GB of TTS dependencies
- Check you have **~2GB free space** in temp directory
- Windows: Ensure PowerShell is available (comes with Windows by default)
- Linux/macOS: Ensure Bash and curl are installed (usually pre-installed)

### Windows Defender Blocking Builds

**Problem:** Garble-obfuscated binaries may be blocked by Windows Defender

**Why This Happens:**
- Garble heavily obfuscates code structure, control flow, and strings
- Windows Defender's heuristics flag unknown obfuscation patterns as suspicious
- This affects the `build-prod.ps1` and `build-obfuscated.ps1` scripts

**Recommended Solution:** Use the unobfuscated production build to avoid antivirus false positives:

Windows:
```powershell
.\scripts\helpers\build-prod-unobfuscated.ps1
```

Linux/macOS:
```bash
./scripts/helpers/build-prod-unobfuscated.sh
```

This build includes all production features (WASM, hardcoded ngrok, embedded binaries) but uses standard Go compiler optimization (`-ldflags="-s -w"`) instead of garble, which:
- ✅ Bypasses Windows Defender false positives
- ✅ Strips debug symbols and symbol table (still optimized)
- ❌ Does not obfuscate code structure or strings

**Alternative Solution - Add Exclusions (If Using Garble, Run PowerShell as Administrator):**

If you must use garble obfuscation, add Windows Defender exclusions:

```powershell
# Exclude temp directory where garble works
Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Temp"

# Exclude output directory
Add-MpPreference -ExclusionPath "C:\Users\REDACTED_USER\OneDrive - REDACTED_COMPANY\Documentos\Personal\Go\pejelagarto-translator\bin"

# Verify exclusions were added
Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
```

**Alternative - Temporarily Disable Real-time Protection:**

Windows Settings → Windows Security → Virus & threat protection → Manage settings → Real-time protection (OFF)

⚠️ **Remember to re-enable after building!**

**What Gets Flagged:**
- Obfuscated server binaries (`piper-server.exe`)
- Intermediate build artifacts in `%LOCALAPPDATA%\Temp`

**What's Safe:**
- Non-obfuscated builds (`pejelagarto-translator.exe`) - no issues
- WASM modules (`main.wasm`) - no issues
- Scripts themselves - no issues

### Production Build Troubleshooting

**"garble: command not found":**
```bash
# Install garble
go install mvdan.cc/garble@latest

# Verify installation
garble version
```

**Build fails with garble errors:**
- Ensure Go 1.24.2+ is installed
- Clean Go cache: `go clean -cache -modcache`
- Try building without garble (tags only): `go build -tags "obfuscated,frontend,ngrok_default" -o bin/test.exe server_frontend.go`

**WASM build fails:**
```bash
# Check GOOS/GOARCH are set correctly
$env:GOOS="js"
$env:GOARCH="wasm"

# Verify wasm_exec.js exists in Go installation
ls "$(go env GOROOT)\lib\wasm\wasm_exec.js"
```

**Checksums don't match after rebuild:**
- This is expected! Garble uses `-seed=random`
- Each build produces different obfuscated code
- Only compare checksums for the same build artifacts

### Runtime Issues

**"Language model not installed" or "Piper binary not found":**
- Delete temp directory to force re-download:
  - Windows: `Remove-Item "$env:TEMP\pejelagarto-translator" -Recurse -Force`
  - Linux/macOS: `rm -rf /tmp/pejelagarto-translator`
- Restart application (will automatically re-download dependencies)
- Windows: Check PowerShell execution policy if download fails: `Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy Bypass`
- Linux/macOS: Ensure downloaded binary has execute permissions: `chmod +x /tmp/pejelagarto-translator/requirements/piper`

**Language not working (exit status 0xc0000409):**
- This is a Piper crash, usually due to incompatible text encoding
- Try a different language model
- Check if text preprocessing is working correctly
- Report the issue with specific language and input text

**Slow audio not generating:**
- Install FFmpeg and ensure it's in system PATH
- Test: `ffmpeg -version` should show version info
- Without FFmpeg, only normal-speed audio works

**UI not showing language dropdown:**
- Add `-pronunciation_language_dropdown` flag when starting the application
- By default, dropdown is hidden and uses command-line language setting

### Performance Issues

**First run is slow:**
- Normal: downloading ~1.1GB of TTS models takes 3-5 minutes
- PowerShell script extracts and verifies all dependencies
- Subsequent runs are instant (uses cached files in temp directory)

**Audio generation is slow:**
- First TTS request per language loads model into memory (~1-2 seconds)
- Subsequent requests are much faster
- Slow audio (0.5x speed) takes extra time for FFmpeg processing

**Small executable size:**
- **v2.4.8**: ~12MB (99% reduction from v2.4.7's 1.14GB!)
- TTS models downloaded at runtime instead of embedded
- Cached in temp directory after first download
- Trade-off: requires internet connection on first run

### Common Errors

**"There is not enough space on the disk" during Go build:**
```powershell
go clean -cache
go clean -modcache
# Then rebuild
go build -o bin/pejelagarto-translator.exe main.go
```

**"Permission denied" on Linux/macOS:**
```bash
chmod +x /tmp/pejelagarto-translator/requirements/piper
```

**Windows Defender blocking Piper:**
- Check Windows Security → Protection history
- Add exception for temp directory if needed

## Development

### Setup for Contributors

Windows:
```powershell
# Clone the repository
git clone https://github.com/FranLegon/pejelagarto-translator.git
cd pejelagarto-translator

# Build the project
go build -o bin/pejelagarto-translator.exe main.go

# Run tests
go test -v

# Run with live reload during development
# (dependencies will auto-download on first run)
go run main.go
```

Linux/macOS:
```bash
# Clone the repository
git clone https://github.com/FranLegon/pejelagarto-translator.git
cd pejelagarto-translator

# Build the project
go build -o bin/pejelagarto-translator main.go

# Run tests
go test -v

# Run with live reload during development
# (dependencies will auto-download on first run)
go run main.go
```

### Development Workflow

#### Typical Development Cycle

```bash
# 1. Make code changes
vim main.go

# 2. Run tests to verify
go test -v

# 3. Test with live reload
go run main.go

# 4. Build and test executable
go build -o bin/test.exe main.go
./bin/test.exe

# 5. Run fuzz tests (recommended)
./run-fuzz-tests.sh

# 6. Build all configurations to verify compatibility
./test-build-combinations.sh
```

#### Testing Different Build Modes

**Backend Mode (default):**
```bash
go build -o bin/test.exe main.go
./bin/test.exe
```

**Frontend Mode (WASM):**
```bash
./build-frontend.sh
go run server_frontend.go version.go
```

**Obfuscated Mode:**
```bash
./obfuscation/build-obfuscated.ps1
./bin/piper-server.exe
```

**Production Mode (all features):**
```bash
./scripts/helpers/build-prod.ps1
./bin/piper-server.exe
```

#### Quick Comparison Test

To verify translation consistency across different builds:

```bash
# Build all versions
go build -o bin/backend.exe main.go
./build-frontend.sh
./obfuscation/build-obfuscated.ps1

# Test translation output
echo "hello world" | ./bin/backend.exe
# Compare with WASM/obfuscated versions
```

### Adding New Transformations

To extend the translator with new rules:

1. **Word/Letter Mappings**: Edit `conjunctionMap` or `letterMap` in `main.go`
   - Ensure all mappings are bijective (one-to-one)
   - Keys and values must have the same rune count
   - Avoid collisions between different map types

2. **Punctuation**: Add entries to `punctuationMap`
   - Can have different lengths for keys and values

3. **New Transformation Stage**: Add your function to both pipelines
   - `TranslateToPejelagarto`: Add transformation step
   - `TranslateFromPejelagarto`: Add reverse transformation at mirror position

4. **Add Tests**: Create fuzz tests for new transformations
   - Test reversibility with random inputs
   - Add seed corpus in `testdata/fuzz/`
   - Minimum 30s fuzz time, 120s for full pipeline

### Testing Strategy

- **Unit Tests**: Test individual transformation functions
- **Fuzz Tests**: Verify reversibility with random inputs (80,000+ executions)
- **Integration Tests**: Test full translation pipeline
- **Build Tests**: Verify all build tag combinations work (`test-build-combinations.sh`)
- **WASM Tests**: Compile WASM tests to verify JS exports (`test-wasm.sh`)

Always ensure your changes maintain 100% reversibility.

### Debugging Tips

**Translation Issues:**
```go
// Add debug logging in transformation functions
fmt.Printf("DEBUG: Before transformation: %q\n", input)
result := yourTransformation(input)
fmt.Printf("DEBUG: After transformation: %q\n", result)
```

**WASM Issues:**
```bash
# Check WASM compilation
GOOS=js GOARCH=wasm go build -tags frontend -o bin/test.wasm wasm_main.go

# Check for syscall/js errors
# (These indicate frontend tag is incorrectly applied to backend code)
```

**Build Tag Issues:**
```bash
# List files included in build
go list -tags frontend -f '{{.GoFiles}}'
go list -tags obfuscated -f '{{.GoFiles}}'

# Check build constraints
go list -tags frontend -f '{{.IgnoredGoFiles}}'
```

## Architecture & Implementation

### Build System Architecture

The project uses Go build tags to create different build configurations:

#### Build Tags

| Tag | Purpose | Effects |
|-----|---------|---------|
| `frontend` | Client-side WASM translation | Compiles to WebAssembly, excludes syscall/js incompatible code |
| `obfuscated` | Code obfuscation for production | Uses `piper-server` naming, changes temp directories |
| `ngrok_default` | Hardcoded ngrok credentials | Embeds auth token and domain, includes `downloadable` |
| `downloadable` | Embed binaries | Includes Windows/Linux/Android binaries in executable |

#### Build Tag Relationships

```
ngrok_default
    └── Includes: downloadable (automatic)
    
downloadable
    └── OR condition with ngrok_default
    
obfuscated
    └── Compatible with all tags
    
frontend
    └── Compatible with all tags
    └── Changes compilation target to WASM
```

#### Build Files

| File | Build Tags | Purpose |
|------|-----------|---------|
| `main.go` | None (all builds) | Core translation logic (shared by all builds) |
| `tts.go` | `//go:build !frontend` | TTS functions (backend + frontendserver) |
| `server_backend.go` | `//go:build !frontend && !frontendserver` | Backend HTTP server with server-side translation |
| `server_frontend.go` | `//go:build frontendserver` | Frontend HTTP server (WASM client-side translation) |
| `wasm_main.go` | `//go:build frontend` | WASM entry point with JS exports |
| `translation_test.go` | None (all builds) | Shared translation tests |
| `wasm_test.go` | `//go:build frontend` | WASM-specific tests |
| `tts_test.go` | `//go:build !frontend` | Server-only TTS tests |
| `embedded_binaries.go` | `//go:build downloadable \|\| ngrok_default` | Embeds binaries for download |
| `embedded_binaries_not_downloadable.go` | `//go:build !downloadable && !ngrok_default` | Empty binary embed |
| `config/version.go` | None (all builds) | Version constant (v1.0.8) |
| `config/downloadable.go` | `//go:build downloadable \|\| ngrok_default` | IsDownloadable constant (true) |
| `config/not_downloadable.go` | `//go:build !downloadable && !ngrok_default` | IsDownloadable constant (false) |
| `config/ngrok_default.go` | `//go:build ngrok_default` | Hardcoded ngrok credentials |
| `config/ngrok_not_default.go` | `//go:build !ngrok_default` | No hardcoded credentials |
| `config/constants_backend.go` | `//go:build !obfuscated` | Normal build constants |
| `config/constants_obfuscated.go` | `//go:build obfuscated` | Obfuscated build constants |

#### Build Process Flow

**Backend Build (Default):**
```bash
go build .
# Includes: main.go, tts.go, server_backend.go, version.go
# Excludes: server_frontend.go, wasm_main.go (via build tags)
# Output: pejelagarto-translator binary (~40MB)
```

**Frontend Server Build:**
```bash
# Step 1: Build WASM module
GOOS=js GOARCH=wasm go build -tags frontend -o bin/main.wasm .
# Includes: main.go, wasm_main.go, version.go
# Output: main.wasm (~3MB)

# Step 2: Build frontend server
go build -tags frontendserver .
# Includes: main.go, tts.go, server_frontend.go, version.go
# Excludes: server_backend.go (via build tags)
# Output: server binary (~40MB)
```

**Production Build (scripts/helpers/build-prod.*):**
```bash
# All builds now use tags exclusively - no explicit file lists
1. Check garble installation
2. Build WASM: garble build -tags frontend
3. Copy wasm_exec.js runtime
4. Build server: garble build -tags "frontendserver,obfuscated,ngrok_default,downloadable"
5. Generate SHA256 checksums
```

**Key Improvements in v1.0.3:**
- **Tag-based builds**: All builds now use `-tags` with `.` - no more explicit file lists
- **No code duplication**: Removed 629 lines of duplicate code from server_frontend.go
- **Shared modules**: main.go (translation), tts.go (TTS), version.go shared across builds
- **Mutual exclusion**: server_backend.go and server_frontend.go are mutually exclusive via tags

### Project Structure

```
pejelagarto-translator/
├── main.go                  # HTML template and embed directives only (~840 lines)
├── server_backend.go        # Backend HTTP server with server-side translation
├── server_frontend.go       # Frontend HTTP server (WASM client-side translation)
├── wasm_main.go             # WASM entry point with JS exports
├── wasm_test.go             # WASM-specific tests
├── embedded_binaries.go     # Embed binaries for downloadable builds (main package)
├── embedded_binaries_not_downloadable.go  # Empty embed for non-downloadable builds
├── README.md                # This documentation
├── USAGE_EXAMPLES.md        # Usage examples and workflow (legacy - merged into README)
├── ARCHITECTURE.md          # Build system details (legacy - merged into README)
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── build-frontend.sh        # WASM build helper script
├── .gitignore               # Git ignore patterns
├── coverage/                # Test coverage reports
├── config/                  # Configuration files and build-tag-controlled constants
│   ├── version.go           # Version constant (v1.0.9)
│   ├── constants_backend.go         # Constants for normal build
│   ├── constants_obfuscated.go      # Constants for obfuscated build
│   ├── downloadable.go              # IsDownloadable constant (true)
│   ├── not_downloadable.go          # IsDownloadable constant (false)
│   ├── ngrok_default.go             # Hardcoded ngrok credentials
│   └── ngrok_not_default.go         # No hardcoded ngrok credentials
├── internal/                # Internal packages (not for external import)
│   ├── translator/          # Core translation logic package (~1850 lines)
│   │   ├── translator.go    # Translation engine implementation
│   │   └── translator_test.go  # Comprehensive test suite with fuzz testing
│   └── tts/                 # Text-to-speech package (~780 lines)
│       ├── tts.go           # TTS functionality and audio processing
│       └── tts_test.go      # TTS-specific tests (server-only)
├── scripts/
│   ├── requirements/
│   │   ├── get-requirements.ps1     # Embedded in binary - downloads TTS dependencies (Windows)
│   │   └── get-requirements.sh      # Embedded in binary - downloads TTS dependencies (Linux/macOS)
│   ├── helpers/
│   │   ├── build-prod.ps1           # Production build script (Windows)
│   │   ├── build-prod.sh            # Production build script (Linux/macOS)
│   │   ├── build-prod-unobfuscated.ps1  # Unobfuscated production build (Windows)
│   │   ├── build-prod-unobfuscated.sh   # Unobfuscated production build (Linux/macOS)
│   │   ├── build-obfuscated.ps1     # Build script with garble
│   │   ├── build-obfuscated.sh      # Build script with garble (Linux/macOS)
│   │   ├── create-obfuscated-server-service.ps1  # Service creation script
│   │   ├── Run-Server.ps1           # Helper script for Windows
│   │   └── Run-Server.sh            # Helper script for Linux/macOS
│   └── test/
│       ├── test-build-combinations.ps1  # Test all build tag combinations (Windows)
│       ├── test-build-combinations.sh   # Test all build tag combinations (Linux/macOS)
│       ├── run-fuzz-tests.ps1       # Run fuzz tests (Windows)
│       └── run-fuzz-tests.sh        # Run fuzz tests (Linux/macOS)
├── bin/                     # Built executables and scripts
│   ├── pejelagarto-translator       # Main executable (~12MB, Linux/macOS)
│   ├── pejelagarto-translator.exe   # Main executable (~12MB, Windows)
│   ├── piper-server                 # Obfuscated executable (Linux/macOS, optional)
│   ├── piper-server.exe             # Obfuscated executable (Windows, optional)
│   ├── main.wasm                    # WASM translation module (~2-3MB)
│   ├── wasm_exec.js                 # Go WASM runtime
│   ├── checksums-prod.txt           # Production build checksums
│   ├── Run-Server.ps1               # Helper script for Windows
│   └── Run-Server.sh                # Helper script for Linux/macOS
├── testdata/                # Fuzz test corpus
│   └── fuzz/                # Fuzz testing data (80,000+ test cases)
└── tts/requirements/        # Auto-downloaded at runtime to temp directory
    └── (Not tracked in Git - downloaded automatically on first run)
        ├── piper / piper.exe       # Piper TTS binary
        ├── *.dll / *.so            # Required libraries
        ├── espeak-ng-data/         # Phoneme data (~2MB)
        └── piper/languages/        # 16 language models (~1008MB total)
            ├── russian/            # ru_RU-dmitri-medium (~63MB)
            ├── german/             # de_DE-thorsten-medium (~63MB)
            ├── turkish/            # tr_TR-dfki-medium (~63MB)
            ├── portuguese/         # pt_BR-faber-medium (~63MB)
            ├── french/             # fr_FR-siwis-medium (~63MB)
            ├── hindi/              # hi_HI-medium (~63MB)
            ├── romanian/           # ro_RO-mihai-medium (~63MB)
            ├── icelandic/          # is_IS-bui-medium (~63MB)
            ├── swahili/            # sw_CD-lanfrica-medium (~63MB)
            ├── swedish/            # sv_SE-nst-medium (~63MB)
            ├── vietnamese/         # vi_VN-vivos-medium (~63MB)
            ├── czech/              # cs_CZ-jirka-medium (~63MB)
            ├── chinese/            # zh_CN-huayan-medium (~63MB)
            ├── norwegian/          # no_NO-talesyntese-medium (~63MB)
            ├── hungarian/          # hu_HU-anna-medium (~63MB)
            └── kazakh/             # kk_KZ-iseke-x_low (~28MB)
```

### Unicode Markers (Private Use Area)
- `\uFFF0`: Start replacement marker
- `\uFFF1`: End replacement marker
- `\uFFF2`: Internal quote marker
- `\u00AD`: Soft hyphen (invisible UTF-8 sanitization marker)
- `\uE000-\uE0FF`: Private use characters for byte encoding

### Bijective Map Construction
1. Combine all maps into unified structure
2. Index by key length (descending)
3. Prefix multi-rune values with `'` marker
4. Create inverse map with negative indices
5. Process in deterministic order

### Case Preservation Algorithm
- Extract uppercase positions from original
- Apply to replacement maintaining pattern
- Handle non-reversible case characters (Turkish İ, German ß, Greek Σ)

## Examples

```go
// Simple translation
input := "hello world"
result := TranslateToPejelagarto(input)
// Output: "'jhtxz 'zcthx" (with random special characters U+2300-U+23FB inserted)

// With numbers
input := "I have 42 apples"
result := TranslateToPejelagarto(input)
// Positive numbers converted to base-8: 42 (decimal) → 52 (octal)

// Nearly full reversibility
original := "The quick brown fox jumps over the lazy dog"
pejelagarto := TranslateToPejelagarto(original)
restored := TranslateFromPejelagarto(pejelagarto)
// After cleaning special characters/timestamps: restored == original ✓
// Note: Special character timestamp encoding is not fully reversible
```

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes ensuring:
   - All tests pass (`go test -v`)
   - Code follows Go best practices
   - New features include tests
   - Reversibility is maintained
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

### Code Quality Guidelines

- Maintain 100% reversibility for all transformations
- Add fuzz tests for new transformation functions
- Document complex algorithms with comments
- Keep functions focused and testable
- Follow existing code style and patterns

## Distribution

The built executable is **lightweight and portable**:
- ✅ No installation required
- ✅ **Only ~12-13MB** executable size (99% smaller than v2.4.7!)
- ✅ No external dependencies (except FFmpeg for slow audio, optional)
- ✅ Auto-downloads 16 TTS language models on first run (~1.1GB)
- ✅ Copy to any machine and run
- ⚠️ Requires internet connection on first run only
- ✅ Dependencies cached in temp directory for offline use afterward
- ✅ OS-specific scripts embedded in binary handle all downloads automatically

### Normal Distribution

Just share the binary file:
- Windows: `bin\pejelagarto-translator.exe`
- Linux/macOS: `bin/pejelagarto-translator`

**First Run Requirements:**
- Internet connection for downloading TTS models
- Windows: PowerShell (built-in)
- Linux/macOS: Bash and curl (usually pre-installed)
- ~2GB free space in temp directory

**Subsequent Runs:**
- No internet needed (uses cached dependencies)
- Starts instantly

**Note**: For slow audio playback feature, FFmpeg must be installed separately and available in system PATH.

### Android Distribution

An Android APK is available for mobile devices. The APK includes native libraries for all Android architectures (ARM, ARM64, x86, x86_64).

**Download the APK:**
- Android: `bin/pejelagarto-translator.apk` (~14MB)

**Installation:**
1. Enable "Install from unknown sources" in Android settings
2. Transfer the APK to your Android device
3. Open the APK file to install
4. The app provides translation functionality with a simple interface

**Building the Android APK:**

Prerequisites:
```bash
# Install gomobile
go install golang.org/x/mobile/cmd/gomobile@latest

# Initialize gomobile (downloads Android SDK/NDK if needed)
gomobile init
```

Build command:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
export ANDROID_NDK_HOME=/path/to/android/sdk/ndk/26.3.11579264
export ANDROID_HOME=/path/to/android/sdk
gomobile build -androidapi=21 -target=android -o bin/pejelagarto-translator.apk ./cmd/androidapp
```

See `android/README.md` and `cmd/androidapp/README.md` for detailed documentation.

### Server Deployment (Obfuscated)

For production server deployment with obfuscation:

1. **Build the obfuscated binary:**
   
   Windows:
   ```powershell
   .\obfuscation\build-obfuscated.ps1
   ```
   
   Linux/macOS:
   ```bash
   ./obfuscation/build-obfuscated.sh linux
   ```
   
   This creates `bin/piper-server` (or `bin/piper-server.exe` on Windows) with obfuscated code and different internal names.

2. **Copy to server:**
   
   Windows:
   ```powershell
   # Copy the obfuscated binary
   Copy-Item bin\piper-server.exe \\server\path\
   
   # Copy the service creation script
   Copy-Item obfuscation\create-obfuscated-server-service.ps1 \\server\path\
   ```
   
   Linux/macOS:
   ```bash
   # Copy the obfuscated binary
   scp bin/piper-server user@server:/path/to/deploy/
   
   # Copy the service creation script
   scp obfuscation/create-obfuscated-server-service.ps1 user@server:/path/to/deploy/
   ```

3. **Create system service on server:**
   
   Windows (run as Administrator):
   ```powershell
   cd \\server\path
   .\create-obfuscated-server-service.ps1
   ```
   
   Linux (run with sudo, requires PowerShell Core):
   ```bash
   cd /path/to/deploy
   sudo pwsh create-obfuscated-server-service.ps1
   ```
   
   macOS (run with sudo, requires PowerShell Core):
   ```bash
   cd /path/to/deploy
   sudo pwsh create-obfuscated-server-service.ps1
   ```
   
   **Note**: On Linux/macOS, you need to install PowerShell Core (`pwsh`) first:
   - Ubuntu/Debian: `sudo apt-get install -y powershell`
   - macOS: `brew install --cask powershell`
   - Or use the distribution's native service management tools

The service creation script automatically:
- Detects the operating system (Windows/Linux/macOS)
- Finds the `piper-server` binary in the current directory
- Creates appropriate service configuration:
  - **Windows**: Scheduled Task running at startup with SYSTEM privileges
  - **Linux**: systemd service with automatic restart
  - **macOS**: LaunchDaemon with KeepAlive
- Enables and starts the service automatically

**Service Management:**

Windows:
```powershell
Start-ScheduledTask -TaskName "PiperServer"
Stop-ScheduledTask -TaskName "PiperServer"
Unregister-ScheduledTask -TaskName "PiperServer"
```

Linux:
```bash
sudo systemctl status PiperServer.service
sudo systemctl restart PiperServer.service
journalctl -u PiperServer.service -f
```

macOS:
```bash
sudo launchctl list | grep PiperServer
sudo launchctl unload /Library/LaunchDaemons/com.PiperServer.plist
tail -f /var/log/PiperServer.log
```

## Future Enhancements

Potential areas for expansion:

- Additional TTS language models (currently 18 languages supported)
- Translation history and caching
- Batch file processing
- Additional transformation rules
- Performance optimizations for very large texts
- CLI interface for command-line translation
- Audio format options (MP3, OGG, etc.)
- Voice customization (pitch, speed, etc.)
- TTS queue management for multiple requests

## Project Status

**Production Ready**

**Key Features:**
- ✅ **Compact Binary**: Optimized ~12-13MB executable with runtime dependency management
- ✅ **Cross-Platform**: Full support for Windows, Linux, and macOS (x86_64 and ARM64)
- ✅ **18 Language TTS**: Multi-language audio support with compass-based organization
- ✅ **Smart Dependency Management**: OS-specific scripts automatically download all dependencies
- ✅ **Intelligent Caching**: Dependencies cached in temp directory for instant subsequent startups
- ✅ **Dual-Speed Audio**: Normal and slowed (0.5x) playback with automatic caching
- ✅ **Production Build System**: Automated garble obfuscation + WASM + ngrok builds
- ✅ **Dynamic Port Selection**: Automatic fallback from 8080 to 8090
- ✅ **Pronunciation Display**: Real-time TTS preprocessing visibility
- ✅ **Comprehensive Testing**: 80,000+ fuzz test executions proving reliability
- ✅ **Modern UI**: Dark/light theme with responsive design
- ✅ **High Reversibility**: All transformations are bidirectional (except timestamp encoding)

**Tested Configurations:**
- **Windows**: Windows 10/11 with PowerShell
- **Linux**: Ubuntu 20.04+ with Bash (x86_64 and ARM64)
- **macOS**: macOS 12+ with Bash (x86_64 and ARM64)
- **Go**: 1.24.2+
- **Languages**: All 18 TTS languages verified working
- **Binary Size**: ~12-13MB
- **First Run**: Downloads ~1.1GB dependencies (3-5 minutes)
- **Subsequent Runs**: Instant startup from cache

## AI Training Restriction

**⚠️ IMPORTANT NOTICE: This repository and its contents are NOT available for use in training artificial intelligence models, machine learning systems, or large language models without explicit written permission and compensation.**

### Prohibited Uses Without Authorization

The following uses of this repository's code, documentation, algorithms, or any other content are **strictly prohibited** without prior written authorization and financial compensation:

- Training, fine-tuning, or improving AI models, including but not limited to:
  - Large Language Models (LLMs)
  - Machine Learning models
  - Neural networks
  - Code generation models
  - Documentation generation systems
- Creating derivative datasets for AI training purposes
- Using this code as training data for automated code completion tools
- Incorporating algorithms or implementation patterns into AI training corpora
- Using documentation or comments as natural language training data

### Authorization Process

Organizations or individuals interested in using this repository for AI training purposes must:

1. Contact the copyright holder: Francisco Legón
2. Negotiate terms and compensation
3. Obtain explicit written authorization
4. Comply with any additional terms specified in the authorization agreement

### Enforcement

Unauthorized use of this repository for AI training purposes constitutes a violation of the license terms and may result in legal action. The copyright holder reserves all rights to pursue damages for unauthorized AI training use.

### General Use

This restriction does **not** affect normal software development activities, including:
- Using the software as intended (translation, TTS, web server)
- Reading the code for educational purposes
- Contributing to the project
- Forking and modifying for personal or commercial use (subject to MIT License terms)
- Code review and analysis by human developers

The MIT License below applies to all uses **except** AI training, which requires separate authorization.

---

## Versioning

### Current Version
**v1.1.0**

### Versioning System
This project follows [Semantic Versioning](https://semver.org/) (SemVer):

```
MAJOR.MINOR.PATCH (v1.2.3)
```

- **MAJOR** (v**X**.0.0): Incompatible API changes or major feature overhauls
- **MINOR** (vX.**Y**.0): New features added in a backward-compatible manner
- **PATCH** (vX.Y.**Z**): Backward-compatible bug fixes

### How to Update Version

#### 1. Update version.go
Edit `version.go` and change the Version constant:

```go
const Version = "vX.Y.Z"  // Update this
```

**Note:** The Version constant is now defined only once in `version.go` and shared across all build modes. No need to update multiple files.

#### 2. Commit Message Format
All commit messages **MUST** contain the version number. Format:

```
vX.Y.Z: <commit message>
```

**Examples:**
- `v1.0.1: Fix TTS audio cache memory leak`
- `v1.1.0: Add support for Italian language`
- `v2.0.0: Complete UI redesign with new translation engine`

#### 3. Git Tag (Optional but Recommended)
After committing, create a git tag for the version:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Version Display
The version is displayed in the bottom-right corner of the UI:
- Desktop: 12px font
- Mobile: 10px font
- Styled with monospace font, low opacity
- Non-interactive (user-select: none, pointer-events: none)

### Release Checklist

When preparing a new release:

1. [ ] Update `version.go` (single source of truth)
2. [ ] Test build: `go build .`
3. [ ] Test frontend build: `go build server_frontend.go version.go`
4. [ ] Run all tests: `go test ./...`
5. [ ] Run fuzz tests: `.\scripts\test\run-fuzz-tests.ps1`
6. [ ] Run build combinations: `.\scripts\test\test-build-combinations.ps1`
7. [ ] Update CHANGELOG.md (if exists)
8. [ ] Commit with version: `git commit -m "vX.Y.Z: <description>"`
9. [ ] Create git tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
10. [ ] Push with tags: `git push && git push --tags`

### Version History

#### v1.1.0 (Current - Test Coverage Enhancement)
- **Enhanced test coverage**: Added 13th build combination test (ngrok+downloadable+obfuscated+frontend)
- **Complete validation**: All 13 build tag combinations now explicitly tested
- **Build system completeness**: Ensures comprehensive coverage of all possible tag combinations
- **Minor version bump**: Following SemVer for backward-compatible enhancement

#### v1.0.9 (Major Refactoring: Package Structure)
- **New package structure**: Created `internal/translator/` and `internal/tts/` packages
- **Translation logic**: Moved ~1850 lines from main.go to internal/translator/
- **TTS logic**: Moved ~780 lines from root to internal/tts/
- **Reduced main.go**: Now only contains HTML template and embed directives (~840 lines)
- **Translator exports**: TranslateToPejelagarto, TranslateFromPejelagarto, maps, constants
- **TTS exports**: HandleTextToSpeech, HandleCheckSlowAudio, HandlePronunciation, config variables
- **Test migration**: Moved translation_test.go and tts_test.go to respective packages
- **Import updates**: Updated 9 files to use new package structure
- **Architecture decisions**: Server files, WASM files, and embed files remain at root (build tag constraints)
- **Code organization**: ~2630 lines moved to internal packages while preserving all 13 build configurations
- **Validated builds**: All 13 build tag combinations tested and passing after each refactoring step
- **Gradual approach**: Incremental refactoring strategy with continuous testing between changes

#### v1.0.8 (Project Reorganization)
- **Major restructuring**: Moved configuration files to `config/` package
- **Requirements organization**: Moved scripts to `scripts/requirements/`
- **Build tag improvements**: Cleaner separation of concerns
- **Fixed embed paths**: Binary embeds now in main package (embed constraint)
- **Tested all builds**: 13 build tag combinations verified working
- **Updated documentation**: Comprehensive project structure documentation

#### v1.0.7 (Documentation Corrections)
- Corrected README about garble/ngrok/Windows Defender relationship
- Updated GitHub release notes

#### v1.0.0 (Initial Release)
- Initial versioning system implementation
- Version display in UI (bottom-right corner)
- Semantic versioning adopted

---

## Garble + ngrok Compatibility Investigation

### 🚨 CRITICAL: Garble IS INCOMPATIBLE with ngrok-go SDK

After exhaustive testing, we have definitively proven that **garble obfuscation breaks ngrok-go SDK functionality**. This is not a configuration issue or a Windows Defender problem—it's a fundamental incompatibility.

### Test Results Summary

| Build Configuration | ngrok Domain | Result |
|---------------------|--------------|--------|
| Standard Go (unobfuscated) | empty (random URL) | ✅ **WORKS** |
| Garble (full: -tiny -literals) | empty (random URL) | ❌ **FAILS** |
| Garble (no -literals) | empty (random URL) | ❌ **FAILS** |
| Garble (-tiny only) | empty (random URL) | ❌ **FAILS** |
| Garble (no flags) | empty (random URL) | ❌ **FAILS** |

**Result:** All garble builds fail with "remote gone away" error regardless of obfuscation flags or configuration.

### Technical Root Cause

Garble's code obfuscation corrupts critical ngrok-go SDK functionality:

1. **TLS/Crypto Operations:** Garble mangles cryptographic implementations used by ngrok
2. **Reflection-Based Code:** ngrok SDK uses reflection which garble obfuscates
3. **Interface Implementations:** Garble renames interfaces breaking ngrok's internal API
4. **Package Initialization:** Critical setup routines get corrupted

### Error Observed

```
Using random ngrok domain
Establishing tunnel (this may take a few seconds)...
Attempt 1 failed: failed to start tunnel: remote gone away
Attempt 2 failed: failed to start tunnel: remote gone away
Attempt 3 failed: failed to start tunnel: remote gone away
```

### Production Recommendation

**✅ USE:** `scripts\helpers\build-prod-unobfuscated.ps1`
- Works reliably with ngrok
- Provides excellent optimization via `-s -w` flags (32.96 MB binary)
- Windows Defender friendly
- Stable and production-ready

**❌ AVOID:** `scripts\helpers\build-prod.ps1` (garble-obfuscated)
- Incompatible with ngrok-go SDK
- No workaround exists
- Script now shows deprecation warning

### Fixes Applied in v1.2.0

1. **WASM File Naming:** Changed from `main.wasm` to `translator.wasm`
2. **ngrok Domain:** Set to empty string for random URLs (prevents "domain already in use" errors)
3. **Documentation:** Updated all build scripts and README with compatibility warnings

### Testing Methodology

The garble+ngrok compatibility was tested systematically:

1. **Initial Test:** All builds (obfuscated and unobfuscated) failed with hardcoded domain
2. **Configuration Fix:** Updated domain to empty string and fixed WASM naming
3. **Comprehensive Testing:** Tested 5 different garble configurations with fixed config
4. **Definitive Result:** Unobfuscated works, all garble builds fail identically

**Test Script:** `scripts\test\test-garble-ngrok.ps1`

### Related GitHub Issues

Similar issues have been reported in other projects using ngrok:
- Expo framework experienced "remote gone away" errors (expo/expo#22186)
- Common ngrok error indicating SDK/network layer issues

---

## License

MIT License (with AI Training Restriction)

Copyright (c) 2025 Francisco Legón

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

**AI Training Restriction**: The use of this Software, its source code, documentation, algorithms, or any other content for the purpose of training, fine-tuning, or improving artificial intelligence models, machine learning systems, or large language models is strictly prohibited without prior written authorization and compensation from the copyright holder.

The above copyright notice, this permission notice, and the AI Training Restriction shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
