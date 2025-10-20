# Pejelagarto Translator

A Go implementation of a bidirectional translator between English and Pejelagarto language.

## Features

- **Bidirectional Translation**: Convert text from English to Pejelagarto and back
- **Number Conversion**: Translates numbers from base-10 to base-7 with custom offset
- **Character Mapping**: Maps English letters, conjunctions, and words to Pejelagarto equivalents
- **Quote Handling**: Properly escapes quotes to avoid ambiguity
- **High Performance**: Optimized for O(n) complexity with pre-calculated marker depth maps

## API

### Main Functions

```go
// Translate from English to Pejelagarto
func TranslateToPejelagarto(input string) string

// Translate from Pejelagarto to English  
func TranslateFromPejelagarto(input string) string
```

### Examples

```go
// Simple translation
TranslateToPejelagarto("hello world")
// Output: "'jhtxz 'zcthx"

// With numbers
TranslateToPejelagarto("I have 42 apples")
// Output: "A hiqo 3333333333333333423 ibbgos"

// With quotes
TranslateToPejelagarto("It's a beautiful day")
// Output: "At''s i poirtadrg fiu"

// Reversibility
original := "The quick brown fox"
pejelagarto := TranslateToPejelagarto(original)
reversed := TranslateFromPejelagarto(pejelagarto)
// reversed == original (true)
```

## Performance

The translator has been optimized for high performance:

- **O(n) complexity** using pre-calculated marker depth maps
- **Limited backward scanning** (max 50 characters)
- **300x speedup** compared to naive O(n²) implementation
- Processes 2300+ character strings in ~20ms (previously 6 seconds)

## Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
go test -v

# Run fuzz tests
go test -fuzz=FuzzReversibility -fuzztime=30s
go test -fuzz=FuzzNumberConversion -fuzztime=30s
go test -fuzz=FuzzMapReplacements -fuzztime=30s
```

## Known Limitations

### Unicode Case Conversion

The translator has a known limitation with certain Unicode characters that have locale-specific case conversion rules:

- **Turkish İ (U+0130)**: Capital I with dot above converts to lowercase i, but reverses to regular I without dot
- **Other affected characters**: German ß, Greek Σ/ς may also behave unexpectedly

This is due to Go's standard `unicode.ToLower/ToUpper` functions which don't handle all locale-specific case rules.

**Workaround**: For text containing these characters, consider normalizing to ASCII equivalents before translation.

## Translation Rules

### Number Conversion
- Base-10 numbers → Base-7 with offset `5699447592686571`
- Supports negative numbers and arbitrary precision using `math/big`

### Character Mapping
The translator uses a bijective mapping system with:
- **Positive indices**: English → Pejelagarto
- **Negative indices**: Pejelagarto → English
- Processing order by index magnitude: 6, 5, 4, 3, 2, 1

### Quote Escaping
- Single quotes in English are doubled in Pejelagarto: `'` → `''`
- Internal representation uses Unicode marker `\uFFF2` to avoid ambiguity
- Ensures reversibility: `'quoted'` → `''vretof''` → `'quoted'`

## Implementation Details

### Unicode Markers (Private Use Area)
- `\uFFF0`: Start marker for replacements
- `\uFFF1`: End marker for replacements  
- `\uFFF2`: Quote marker for escaping

### Optimization Techniques
1. **Marker Depth Map**: Pre-calculate nesting depth for O(1) lookups
2. **Limited Backward Scanning**: Only scan up to 50 chars for word boundaries
3. **Single Pass Processing**: Process all indices in one iteration

## License

MIT License
