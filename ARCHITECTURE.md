# Pejelagarto Translator Architecture

This document explains the architecture and build system for the Pejelagarto Translator.

## Build Modes

The project supports two build modes using Go build tags:

### 1. Normal Mode (Default)
**Command:** `go build`

**Architecture:**
```
┌─────────────────────────────────────┐
│         Go Binary (12-13MB)         │
├─────────────────────────────────────┤
│  • HTTP Server                      │
│  • Translation Logic (Backend)      │
│  • TTS Audio Generation             │
│  • HTML UI Serving                  │
└─────────────────────────────────────┘
         ↓
    All processing on server
```

**Files Compiled:**
- `main.go` - Core translation logic
- `server_backend.go` - HTTP server, handlers, TTS functions
- `obfuscation/*.go` - Configuration

**Use Case:** Single-user local installation, or traditional server deployment

---

### 2. Frontend Mode (WebAssembly)
**Command:** `./build-frontend.sh` then `go run server_frontend.go`

**Architecture:**
```
┌──────────────────┐         ┌──────────────────────┐
│   Browser        │         │   Go Server          │
├──────────────────┤         ├──────────────────────┤
│ • HTML/CSS/JS    │         │ • HTTP Server        │
│ • WASM Module    │◄────────┤ • Serve static files │
│   (2-3MB)        │  Load   │ • TTS endpoints only │
│ • Translation    │         └──────────────────────┘
│   (Client-side)  │
└──────────────────┘
         │
         └──► TTS requests only ────┐
                                     │
         ┌───────────────────────────┘
         ↓
    ┌──────────────┐
    │ Piper TTS    │
    │ (Server)     │
    └──────────────┘
```

**Files Compiled:**

WASM Module (`-tags frontend`):
- `main.go` - Core translation logic  
- `wasm_main.go` - WASM entry point, JS exports

Server (`server_frontend.go`):
- Lightweight HTTP server
- Serves HTML, WASM, and static files
- TTS endpoints only

**Use Case:** Web deployment with many concurrent users, reduced server load

---

## File Structure

```
pejelagarto-translator/
├── main.go                 # Core translation logic (shared)
├── server_backend.go       # Server mode: HTTP handlers, main()
├── wasm_main.go            # Frontend mode: WASM entry, JS exports
├── server_frontend.go      # Frontend mode: Lightweight server
├── build-frontend.sh       # Build script for frontend mode
├── obfuscation/
│   ├── constants_backend.go     # Build tag: !obfuscated
│   └── constants_obfuscated.go  # Build tag: obfuscated
└── bin/
    ├── pejelagarto-translator   # Backend build output
    ├── translator.wasm          # Frontend WASM module
    └── wasm_exec.js             # Go WASM runtime

```

## Build Tags

The project uses Go build tags to control compilation:

### `frontend` tag
- **Present:** Compiles WASM module
  - `main.go` ✓ (translation logic)
  - `wasm_main.go` ✓ (WASM entry point)
  - `server_backend.go` ✗ (excluded by `//go:build !frontend`)
  
- **Absent:** Compiles normal server
  - `main.go` ✓ (translation logic)
  - `server_backend.go` ✓ (server code)
  - `wasm_main.go` ✗ (excluded by `//go:build frontend`)

### `obfuscated` tag
- Controls project name and paths
- Used by both normal and frontend modes
- See `obfuscation/constants_*.go`

## Translation Pipeline

The translation logic in `main.go` is shared by both modes:

```go
// Available in both server and WASM builds
func TranslateToPejelagarto(input string) string
func TranslateFromPejelagarto(input string) string
```

### In Normal Mode:
Client → HTTP POST `/to` → `TranslateToPejelagarto()` → HTTP Response

### In Frontend Mode:
Client → `GoTranslateToPejelagarto()` (JS) → WASM → `TranslateToPejelagarto()` → Result in browser

## TTS (Text-to-Speech)

TTS always runs server-side in both modes because:
- Requires Piper binary and language models (~1.1GB)
- CPU-intensive audio generation
- Not suitable for browser execution

## Benefits of Frontend Mode

1. **Scalability:** Translation offloaded to client browsers
2. **Server Load:** Server only handles TTS requests
3. **Offline Capability:** Translation works offline after WASM loads
4. **Bandwidth:** ~2-3MB WASM vs streaming all translations through server
5. **Privacy:** Translation happens locally in user's browser

## Development Guidelines

### Adding New Translation Features

Add code to `main.go` - it will work in both modes automatically.

### Adding Server-Only Features

Add code to `server_backend.go` with `//go:build !frontend` constraint.

### Adding WASM Exports

Add functions to `wasm_main.go` and update HTML JavaScript to call them.

## Testing

```bash
# Test normal build
go test

# Test both builds compile
go build                                    # Normal
GOOS=js GOARCH=wasm go build -tags frontend # WASM
```

## Deployment

### Normal Mode
```bash
go build -o bin/pejelagarto-translator
./bin/pejelagarto-translator
```

### Frontend Mode
```bash
./build-frontend.sh
go run server_frontend.go
# Or build the server: go build -o bin/server-frontend server_frontend.go
```
