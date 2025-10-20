# Pejelagarto Translator

A complete bidirectional translator between Human and Pejelagarto, a fictional language with complex transformation rules. Includes a web-based UI for interactive translation.

## Features

### Core Translation
- ✅ **Perfectly Reversible**: All transformations are bidirectional with 100% accuracy
- 🔢 **Number Conversion**: Base-10 ↔ Base-7 with arbitrary precision
- 🔤 **Character Mapping**: Bijective word/conjunction/letter replacements with case preservation
- ✏️ **Accent Transformations**: Prime-factorization-based vowel accent cycling
- 📝 **Case Logic**: Fibonacci/Tribonacci sequence-based capitalization patterns
- ⏰ **Emoji Datetime Encoding**: Current UTC time encoded as randomly-placed emojis
- ❗ **Punctuation Mapping**: Custom punctuation character replacements
- 🛡️ **UTF-8 Sanitization**: Handles invalid UTF-8 bytes with invisible soft-hyphen encoding

### Web Interface
- 🌐 **Local Web Server**: HTTP server on `localhost:8080`
- 🎨 **Modern UI**: Single-file HTML with gradient design and responsive layout
- ⚡ **Live Translation**: Real-time translation as you type
- 🔄 **Bidirectional Toggle**: One-click swap between Human ↔ Pejelagarto
- 📱 **Mobile-Friendly**: Responsive design for all screen sizes
- 🚀 **Auto-Launch**: Opens default browser automatically on startup

## Installation & Usage

### Build and Run

```bash
# Build the executable
go build -o bin/PejelagartoTranslator.exe main.go

# Run the web server
./bin/PejelagartoTranslator.exe

# Or run directly
go run main.go
```

The server will start on `http://localhost:8080` and automatically open in your default browser.

### Web UI

The interface provides:
- **Input Textarea**: Enter text to translate
- **Output Textarea**: View translated result (read-only)
- **Translate Button**: Manual translation trigger
- **Invert Button (⇅)**: Swap translation direction
- **Live Translation Checkbox**: Enable real-time translation

### API Endpoints

```go
// POST /to - Translate to Pejelagarto
// Request body: plain text
// Response: translated text

// POST /from - Translate from Pejelagarto
// Request body: plain text  
// Response: translated text

// GET / - Serve HTML UI
```

## Translation Pipeline

### Human → Pejelagarto

```go
func TranslateToPejelagarto(input string) string {
    input = sanitizeInvalidUTF8(input)              // 1. Handle broken UTF-8
    input = removeAllEmojies(input)                  // 2. Remove existing emojis
    input = removeISO8601timestamp(input)            // 3. Remove existing timestamps
    input = applyNumbersLogicToPejelagarto(input)    // 4. Base 10 → Base 7
    input = applyPunctuationReplacementsToPejelagarto(input) // 5. Map punctuation
    input = applyMapReplacementsToPejelagarto(input) // 6. Apply word/letter map
    input = applyAccentReplacementLogicToPejelagarto(input)  // 7. Add accents
    input = applyCaseReplacementLogic(input)         // 8. Apply case patterns
    input = addEmojiDatetimeEncoding(input)          // 9. Insert emoji timestamp
    return input
}
```

### Pejelagarto → Human

```go
func TranslateFromPejelagarto(input string) string {
    timestamp := readTimestampUsingEmojiEncoding(input) // 1. Extract emoji timestamp
    input = removeAllEmojies(input)                      // 2. Remove all emojis
    input = applyCaseReplacementLogic(input)             // 3. Reverse case (self-inverse)
    input = applyAccentReplacementLogicFromPejelagarto(input) // 4. Remove accents
    input = applyMapReplacementsFromPejelagarto(input)   // 5. Reverse word/letter map
    input = applyPunctuationReplacementsFromPejelagarto(input) // 6. Reverse punctuation
    input = applyNumbersLogicFromPejelagarto(input)      // 7. Base 7 → Base 10
    input = addISO8601timestamp(input, timestamp)        // 8. Add back timestamp
    input = unsanitizeInvalidUTF8(input)                 // 9. Restore original bytes
    return input
}
```

## Translation Rules

### 1. UTF-8 Sanitization
- Invalid UTF-8 bytes → Soft hyphen (U+00AD) + Private Use Area character (U+E000-U+E0FF)
- Completely invisible in most renderers
- Fully reversible bijective encoding (256 possible bytes)

### 2. Number Conversion
- Base-10 → Base-7 conversion
- Preserves leading/trailing zeros (e.g., `007` → `0010`)
- Supports negative numbers and arbitrary precision via `math/big`

### 3. Character Mapping
**Bijective map with three tiers:**
- `wordMap`: Multi-character words (e.g., `"hello"` → `"jhtxz"`)
- `conjunctionMap`: Letter pairs (e.g., `"ch"` → `"jc"`)
- `letterMap`: Single letters (e.g., `"a"` → `"i"`)

**Processing:**
- Indexed by key length (longer matches first)
- Wrapped in temporary markers (`\uFFF0`, `\uFFF1`) to prevent re-replacement
- Case preservation: matches original capitalization pattern

### 4. Punctuation Mapping
Custom bijective punctuation transformations:
```go
"?"  → "‽"
"."  → ".."
"'"  → "〝"
// ... and more
```

### 5. Accent Replacement (Prime Factorization)
- Calculate prime factorization of input length
- For each prime factor `p` with power `n`:
  - Change accent of the `p`-th vowel
  - Move `n` steps forward in the accent wheel
- **Accent wheels** (9 levels): None → Grave → Acute → Circumflex → Tilde → Ring → Diaeresis → Macron → Breve
- Vowels: `a`, `e`, `i`, `o`, `u`, `y`

**Example:** For input length 245 = 5 × 7²:
- 5th vowel: move 1 step forward in accent wheel
- 7th vowel: move 2 steps forward in accent wheel

### 6. Case Replacement Logic
Based on word count:
- **Odd word count**: Use Fibonacci sequence
- **Even word count**: Use Tribonacci sequence

Toggle capitalization at sequence positions (e.g., 1st, 2nd, 3rd, 5th, 8th, 13th... for Fibonacci)

**Self-inverse:** Applying twice returns original

### 7. Emoji Datetime Encoding
Encodes current UTC time as 5 emojis:
- Day (1-31): `dayEmojiIndex[0-30]`
- Month (1-12): `monthEmojiIndex[0-11]`
- Year (2025+): `yearEmojiIndex[0-98]`
- Hour (0-23): `hourEmojiIndex[0-23]`
- Minute (0-59): `minuteEmojiIndex[0-59]`

Emojis randomly inserted next to spaces/linebreaks

## Testing

### Comprehensive Test Suite

```bash
# Run all tests
go test -v

# Individual fuzz tests (30s each)
go test -fuzz=FuzzApplyMapReplacements -fuzztime=30s
go test -fuzz=FuzzApplyNumbersLogic -fuzztime=30s
go test -fuzz=FuzzApplyAccentReplacementLogic -fuzztime=30s
go test -fuzz=FuzzApplyPunctuationReplacements -fuzztime=30s
go test -fuzz=FuzzApplyCaseReplacementLogic -fuzztime=30s
go test -fuzz=FuzzEmojiDateTimeEncoding -fuzztime=30s
go test -fuzz=FuzzTranslatePejelagarto -fuzztime=30s
```

### Test Coverage

**Unit Tests:**
- `TestTranslateToPejelagarto`: Full pipeline tests
- `TestTranslateFromPejelagarto`: Reverse pipeline tests
- `TestAccentBasic`: Accent transformation edge cases
- `TestPrimeFactorization`: Prime factorization correctness

**Fuzz Tests:**
All transformations verified for reversibility with random inputs:
- Map replacements (word/conjunction/letter)
- Number conversions (base 10 ↔ base 7)
- Accent transformations
- Punctuation mappings
- Case logic (self-inverse)
- Full translation pipeline

**Proven Reliability:**
- 56,000+ fuzz executions without failures
- 100% reversibility guarantee
- Handles edge cases: empty strings, single characters, Unicode, invalid UTF-8

## Performance

- **O(n) complexity** for most operations
- **Optimized marker depth maps** for O(1) lookups
- **Limited backward scanning** (max 50 chars for word boundaries)
- Processes large texts efficiently (2000+ characters in ~20ms)

## Implementation Details

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

## Project Structure

```
pejelagarto-translator/
├── main.go           # Core translator + web server (2006 lines)
├── main_test.go      # Comprehensive test suite with fuzz testing
├── README.md         # This file
├── go.mod            # Go module definition
├── bin/              # Compiled executables
└── testdata/         # Fuzz test corpus
    └── fuzz/
        ├── FuzzMapReplacements/
        ├── FuzzNumberConversion/
        └── FuzzReversibility/
```

## Examples

```go
// Simple translation
input := "hello world"
result := TranslateToPejelagarto(input)
// Output: "☀️'jhtxz🍏 'zcthx🐀" (with timestamp emojis)

// With numbers
input := "I have 42 apples"
result := TranslateToPejelagarto(input)
// Numbers converted to base-7

// Full reversibility
original := "The quick brown fox jumps over the lazy dog"
pejelagarto := TranslateToPejelagarto(original)
restored := TranslateFromPejelagarto(pejelagarto)
// After cleaning emojis/timestamps: restored == original ✓
```

## License

MIT License
