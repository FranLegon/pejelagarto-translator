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

**Backend Build (Default - Server-Side Translation):**

Windows:
```powershell
go build -o bin/pejelagarto-translator.exe main.go
```

Linux/macOS:
```bash
go build -o bin/pejelagarto-translator main.go
```

**Frontend Build (NEW - Client-Side Translation via WebAssembly):**

This mode compiles the translation logic to WebAssembly, allowing translation to run entirely in the browser. Only TTS (text-to-speech) uses the server.

Linux/macOS:
```bash
./build-frontend.sh
go run server_frontend.go
```

Windows:
```powershell
# Manual build (build-frontend.sh equivalent)
$env:GOOS="js"; $env:GOARCH="wasm"
go build -tags frontend -o bin/translator.wasm
Copy-Item "$(go env GOROOT)\lib\wasm\wasm_exec.js" bin\
go run server_frontend.go
```

**Frontend Mode Features:**
- ✅ Translation runs in your browser (no server calls for text operations)
- ✅ Significantly reduced server load
- ✅ Works offline after initial WASM module load
- ✅ TTS still uses server (for Piper audio generation)
- ✅ Same translation quality and features
- 📦 WASM module: ~2-3MB (one-time download)

**Obfuscated Build** (for server deployment):

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
| `downloadable` | Embeds Windows/Linux binaries | `pejelagarto-translator` + embedded bins | All languages |
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

**Verify all combinations:** Run `./test-build-combinations.sh` to test all 12 build tag combinations

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

### Testing Strategy

- **Unit Tests**: Test individual transformation functions
- **Fuzz Tests**: Verify reversibility with random inputs
- **Integration Tests**: Test full translation pipeline

Always ensure your changes maintain 100% reversibility.

## Implementation Details

### Project Structure

```
pejelagarto-translator/
├── main.go              # Core translator + web server + TTS (~3350 lines)
├── main_test.go         # Comprehensive test suite with fuzz testing
├── test_tts.go          # TTS-specific tests
├── README.md            # This documentation
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── get-requirements.ps1 # Embedded in binary - downloads TTS dependencies
├── .gitignore           # Git ignore patterns
├── coverage/            # Test coverage reports
├── obfuscation/         # Obfuscation and service deployment scripts
│   ├── constants_backend.go             # Constants for backend build
│   ├── constants_obfuscated.go          # Constants for obfuscated build
│   ├── build-obfuscated.ps1             # Build script with garble
│   └── create-obfuscated-server-service.ps1  # Service creation script
├── bin/                 # Built executables and scripts
│   ├── pejelagarto-translator.exe  # Main executable (~12MB)
│   ├── piper-server.exe            # Obfuscated executable (optional)
│   └── runWithNgrok.ps1            # Helper script for ngrok
├── testdata/            # Fuzz test corpus
│   └── fuzz/            # Fuzz testing data
└── tts/requirements/    # Auto-downloaded at runtime to temp directory
    └── (Not tracked in Git - downloaded automatically on first run)
        ├── piper.exe           # Piper TTS binary
        ├── *.dll               # Required DLLs
        ├── espeak-ng-data/     # Phoneme data (~2MB)
        └── piper/languages/    # 16 language models (~1008MB total)
            ├── russian/        # ru_RU-dmitri-medium (~63MB)
            ├── german/         # de_DE-thorsten-medium (~63MB)
            ├── turkish/        # tr_TR-dfki-medium (~63MB)
            ├── portuguese/     # pt_BR-faber-medium (~63MB)
            ├── french/         # fr_FR-siwis-medium (~63MB)
            ├── hindi/          # hi_HI-medium (~63MB)
            ├── romanian/       # ro_RO-mihai-medium (~63MB)
            ├── icelandic/      # is_IS-bui-medium (~63MB)
            ├── swahili/        # sw_CD-lanfrica-medium (~63MB)
            ├── swedish/        # sv_SE-nst-medium (~63MB)
            ├── vietnamese/     # vi_VN-vivos-medium (~63MB)
            ├── czech/          # cs_CZ-jirka-medium (~63MB)
            ├── chinese/        # zh_CN-huayan-medium (~63MB)
            ├── norwegian/      # no_NO-talesyntese-medium (~63MB)
            ├── hungarian/      # hu_HU-anna-medium (~63MB)
            └── kazakh/         # kk_KZ-iseke-x_low (~28MB)
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

## Current Status

**Version**: 2.4.9 (Production Ready)

**Key Achievements:**
- ✅ **99% Size Reduction**: Binary reduced from 1.14GB → 12-13MB!
- ✅ **Full Linux/macOS Support**: Native scripts for all platforms
- ✅ **16 Language TTS Support**: Full multi-language audio support with compass-based organization
- ✅ **Runtime Dependency Management**: OS-specific scripts download all dependencies
- ✅ **Smart Caching**: Dependencies cached in temp directory, no re-download needed
- ✅ **Dual-Speed Audio**: Normal and slowed (0.5x) playback with automatic caching
- ✅ **Simplified Build**: No more `build.ps1`, just `go build`
- ✅ **80,000+ Fuzz Tests**: Proven reliability with comprehensive fuzzing
- ✅ **Modern UI**: Dark/light theme with responsive design
- ✅ **Full Reversibility**: All transformations are bidirectional (except timestamp encoding)

**Recent Updates (v2.4.9):**
- 🐧 **Full Linux/macOS Support**: Native shell scripts for all platforms
- 🔧 Created `get-requirements.sh` for Linux/macOS dependency downloads
- 🔧 Created `build-obfuscated.sh` for cross-platform obfuscated builds
- 🔧 Created `Run-Server.sh` for Linux/macOS server launch
- 🔄 Updated main.go to automatically detect OS and use appropriate scripts
- ✅ Removed Windows-only restriction from dependency download
- 📦 Supports both x86_64 and ARM64 architectures on Linux/macOS
- ✅ Tested and verified on Linux (Ubuntu 20.04+)

**Previous Updates (v2.4.8):**
- 🚀 Embedded PowerShell script instead of large binary files
- 📦 Binary size reduced by 99% (1,135MB → 12MB)
- 🔄 Runtime dependency downloading with automatic caching
- 🗑️ Removed redundant `build.ps1` script
- 🧹 Cleaned up repository (removed 760 large files, ~1.15GB)
- ✅ Successfully pushed to GitHub (no more file size issues!)

**Tested Configurations:**
- **Windows**: Windows 10/11 with PowerShell
- **Linux**: Ubuntu 20.04+ with Bash (x86_64 and ARM64)
- **macOS**: macOS 12+ with Bash (x86_64 and ARM64)
- Go 1.24.2+
- All 16 languages verified working
- Build size: ~12-13MB
- First run download: ~1.1GB (3-5 minutes)
- Subsequent runs: instant startup

## License

MIT License

Copyright (c) 2025 Francisco Legón

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
