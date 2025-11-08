# Usage Examples

## Building and Running Different Modes

### 1. Normal Mode (Default - All Server-Side)

**Build:**
```bash
go build -o bin/pejelagarto-translator
```

**Run:**
```bash
./bin/pejelagarto-translator

# With options:
./bin/pejelagarto-translator \
  -pronunciation_language=portuguese \
  -pronunciation_language_dropdown=true
```

**What happens:**
- Server starts on http://localhost:8080
- Browser opens automatically
- All translation happens on the server
- All TTS happens on the server
- Good for: Single user, local installation

---

### 2. Frontend Mode (Client-Side Translation via WASM)

**Build:**
```bash
./build-frontend.sh
```

**Run:**
```bash
go run server_frontend.go

# With options:
go run server_frontend.go \
  -pronunciation_language=russian \
  -pronunciation_language_dropdown=true
```

**What happens:**
- Server starts on http://localhost:8080
- Browser loads ~2-3MB WASM module
- Translation runs in your browser (JavaScript + WebAssembly)
- Only TTS requests go to server
- Good for: Web deployment, multiple users, reduced server load

---

### 3. Obfuscated Mode (Production Deployment)

**Build:**
```bash
# Linux/macOS
./obfuscation/build-obfuscated.sh linux

# Windows
.\obfuscation\build-obfuscated.ps1
```

**Run:**
```bash
./bin/piper-server
```

**What happens:**
- Code is obfuscated using garble
- Project name changes to "piper-server"
- Browser doesn't auto-open
- All translation and TTS happen on server
- Good for: Production server deployment

---

## Testing the Frontend Build

After building and starting the frontend server:

1. **Open Browser:** Navigate to http://localhost:8080
2. **Wait for WASM:** You'll see "✓ Ready - Translation happens in your browser!"
3. **Test Translation:**
   - Type "Hello World" in the input box
   - Click "Translate"
   - See result immediately (no server delay for translation)
4. **Test TTS:**
   - Click the 🔊 button next to output
   - Audio is generated on server (will take a moment)

## Comparison

### Translation Speed

**Normal Mode:**
```
User types → Server receives → Server translates → Server responds → Browser displays
Latency: ~50-200ms (network + processing)
```

**Frontend Mode:**
```
User types → Browser translates (WASM) → Browser displays
Latency: <10ms (local processing only)
```

### Server Load

**Normal Mode** - 100 concurrent users:
- 100 users × typing constantly = High translation load
- 100 users × occasional TTS = Moderate TTS load
- Total: High server load

**Frontend Mode** - 100 concurrent users:
- 0 translation load (happens in browsers)
- 100 users × occasional TTS = Moderate TTS load  
- Total: Low server load

### File Sizes

**Normal Mode:**
- Binary: 12-13 MB
- Client downloads: HTML only (~50 KB)

**Frontend Mode:**
- Server: Minimal (server_frontend.go)
- WASM module: 2-3 MB (one-time download)
- Total client download: ~2.5 MB

## Development Workflow

### Working on Translation Logic

Edit `main.go` - it's shared by both modes:

```bash
# Test changes in normal mode
go build && ./bin/pejelagarto-translator

# Test changes in frontend mode  
./build-frontend.sh && go run server_frontend.go
```

### Working on Server Features

Edit `server_main.go`:

```bash
go build && ./bin/pejelagarto-translator
```

### Working on WASM Integration

Edit `wasm_main.go` or `server_frontend.go`:

```bash
./build-frontend.sh && go run server_frontend.go
```

## Deployment

### Local Network

```bash
# Normal mode - all on server
./bin/pejelagarto-translator

# Frontend mode - translation in browser
./build-frontend.sh
go run server_frontend.go
```

Access from other devices: `http://<your-ip>:8080`

### Internet (with ngrok)

```bash
# Normal mode
./bin/pejelagarto-translator \
  -ngrok_token=YOUR_TOKEN \
  -ngrok_domain=your-domain.ngrok-free.app

# Frontend mode - same server load reduction applies
./build-frontend.sh
# Edit server_frontend.go to add ngrok support (or use reverse proxy)
```

## Troubleshooting

### WASM Module Not Loading

**Symptom:** "Failed to load WebAssembly module" error

**Solutions:**
1. Check that `bin/translator.wasm` exists
2. Check that `bin/wasm_exec.js` exists
3. Make sure server is running
4. Check browser console for errors

### Translation Not Working in Frontend Mode

**Symptom:** Clicking translate does nothing

**Solutions:**
1. Wait for "✓ Ready" status message
2. Check browser console for JavaScript errors
3. Verify WASM module loaded successfully

### TTS Not Working

**Symptom:** Audio doesn't play

**Solutions:**
1. Check that TTS dependencies are downloaded (first run takes time)
2. Look at server logs for error messages
3. Ensure the selected language model is installed
4. This is the same in both normal and frontend modes
