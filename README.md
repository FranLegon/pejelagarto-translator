# Pejelagarto Translator

A complete, locally runnable web application that functions as a translator for the fictional "Pejelagarto" language. The translation is perfectly reversible.

## Features

- **Go Backend**: Single `main.go` file with local web server
- **HTMX Frontend**: Dynamic, interactive UI without external dependencies
- **Reversible Translation**: Perfectly bidirectional translation between Human and Pejelagarto
- **Live Translation**: Real-time translation as you type
- **Invert Function**: Swap input/output languages with one click

## Building and Running

### Build
```bash
go build -o pejelagarto-translator main.go
```

### Run
```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Test
```bash
go test -v
```

## Translation Logic

The translator uses three replacement maps applied in order:

1. **Word Map**: Replaces complete words/syllables (e.g., "hello" → "jetzo")
2. **Conjunction Map**: Replaces letter pairs (e.g., "ch" → "jj", "sh" → "xx")
3. **Letter Map**: Replaces single letters with invertible mappings (e.g., "a" ↔ "i", "e" ↔ "o")

### Key Implementation Details

- **Greedy Longest-Match**: At each position, the algorithm tries to match the longest possible pattern first (words before conjunctions before letters)
- **Case Preservation**: Capitalization of the first letter is preserved during translation
- **Perfect Reversibility**: All translations can be reversed back to the original text

## Project Structure

```
.
├── main.go          # Web server and translation logic
├── main_test.go     # Reversibility tests
├── go.mod           # Go module file
└── README.md        # This file
```

## UI Features

- **Two-panel interface**: Input on the left, output on the right
- **Translate button**: Manual translation trigger
- **Invert button (⇅)**: Swaps the languages and text content
- **Live Translation checkbox**: Enables real-time translation as you type
- **Responsive design**: Works on desktop and mobile devices

## Example Translations

| Human | Pejelagarto |
|---------|-------------|
| hello world | jetzo vorlag |
| thank you | zink yux |
| good morning | gux murneng |
| the quick check | ze kvakk jjokk |

## Technology Stack

- **Backend**: Go 1.x (standard library only)
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **HTMX**: 1.9.10 (CDN)
- **Testing**: Go testing package

## License

This is a demonstration project for educational purposes.
