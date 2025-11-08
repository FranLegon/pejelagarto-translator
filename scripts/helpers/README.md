# Helper Scripts

Production build scripts for creating optimized, obfuscated releases.

## build-prod Scripts

These scripts build the application with all production tags combined:
- `obfuscated` - Code obfuscation via garble
- `frontend` - WASM-based client-side translation
- `ngrok_default` - Hardcoded ngrok credentials (no flags needed)

### Features

✅ **Code Obfuscation**: Uses garble with `-tiny -literals -seed=random`  
✅ **WASM Frontend**: Client-side translation reduces server load  
✅ **Hardcoded Ngrok**: No need to pass ngrok token/domain flags  
✅ **Downloadable Binaries**: Embedded Windows/Linux executables  
✅ **Stripped Binaries**: `-ldflags="-s -w"` for minimal size  
✅ **Static Linking**: `-extldflags '-static'` for portability  

### Usage

#### Windows (PowerShell)

```powershell
# Build for Windows (default)
.\scripts\helpers\build-prod.ps1

# Build for Linux
.\scripts\helpers\build-prod.ps1 -OS linux

# Build for macOS
.\scripts\helpers\build-prod.ps1 -OS darwin

# Build for ARM64
.\scripts\helpers\build-prod.ps1 -OS linux -Arch arm64
```

#### Linux/macOS (Bash)

```bash
# Build for Linux (default)
./scripts/helpers/build-prod.sh

# Build for Windows
./scripts/helpers/build-prod.sh windows

# Build for macOS
./scripts/helpers/build-prod.sh darwin amd64

# Build for ARM64
./scripts/helpers/build-prod.sh linux arm64
```

### Output Files

All files are created in the `bin/` directory:

- **piper-server** (or piper-server.exe) - Obfuscated server binary
- **main.wasm** - WASM translation module
- **wasm_exec.js** - Go WASM runtime
- **checksums-prod.txt** - SHA256 checksums and build info

### Build Process

1. **Check garble**: Verifies garble is installed (installs if missing)
2. **Build WASM**: Compiles frontend translation module
3. **Copy Runtime**: Copies wasm_exec.js from Go installation
4. **Build Server**: Compiles obfuscated server with all tags
5. **Generate Checksums**: Creates SHA256 hashes for verification
6. **Summary**: Displays build information and file sizes

### Requirements

- **Go**: 1.24.2 or higher
- **garble**: Installed automatically if missing
  - Manual install: `go install mvdan.cc/garble@latest`
- **Windows Defender Exclusion** (Windows only):
  - Add exclusion for Go temp directory: `C:\Users\<YourUser>\AppData\Local\Temp\go-build*`
  - Or temporarily disable real-time protection during build
  - Garble-obfuscated binaries may trigger false positives

### Running the Build

After building, the server runs with:

```bash
# The binary name is "piper-server" (obfuscated build name)
./bin/piper-server

# No flags needed! ngrok credentials are hardcoded
# Server will:
# - Start with automatic port selection (8080-8090)
# - Serve WASM frontend for client-side translation
# - Connect to ngrok with embedded credentials
# - Provide downloadable binaries via /download/* endpoints
```

### Build Configuration

The scripts use optimal flags for production:

```bash
garble -tiny -literals -seed=random build \
    -tags "obfuscated,frontend,ngrok_default" \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o bin/piper-server \
    .
```

**Flags Explained:**
- `-tiny`: Maximum obfuscation
- `-literals`: Obfuscate string literals
- `-seed=random`: Random obfuscation seed each build
- `-s -w`: Strip debug info and symbol tables
- `-extldflags '-static'`: Static linking
- `-trimpath`: Remove absolute paths from binary

### Security Notes

⚠️ **Obfuscated Build Behavior:**
- Minimal console output (logs suppressed)
- Binary named "piper-server" (not "pejelagarto-translator")
- Hardcoded ngrok credentials (no environment variables)
- Download endpoints exposed at `/download/windows` and `/download/linux`

### Example Build Session

```
======================================
  Production Build Script (Garble)
======================================

Build Configuration:
  - Tags: obfuscated, frontend, ngrok_default
  - OS: linux
  - Architecture: amd64
  - Obfuscation: garble

[1/6] Checking for garble...
✓ garble found: /home/user/go/bin/garble

[2/6] Building WASM module...
  Output: bin/main.wasm
✓ WASM built successfully (2.51 MB)

[3/6] Copying wasm_exec.js...
✓ wasm_exec.js copied

[4/6] Building obfuscated server...
  Tags: obfuscated,frontend,ngrok_default
  Output: bin/piper-server
✓ Server built successfully (15.23 MB)

[5/6] Generating checksums...
✓ Checksums saved to bin/checksums-prod.txt

[6/6] Build Summary
======================================
Build Type: Production (Obfuscated)
Features:
  ✓ Code obfuscation (garble)
  ✓ WASM frontend
  ✓ Hardcoded ngrok credentials
  ✓ Embedded binaries (downloadable)

Output Files:
  Server:  bin/piper-server (15.23 MB)
  WASM:    bin/main.wasm (2.51 MB)
  Runtime: bin/wasm_exec.js

✓ Production build complete!
======================================
```

### Comparison with Other Build Scripts

| Script | Tags | Use Case |
|--------|------|----------|
| `build-prod.ps1/sh` | obfuscated, frontend, ngrok_default | **Production deployment** |
| `build-frontend.sh` | frontend | Development (WASM only) |
| `obfuscation/build-obfuscated.*` | obfuscated | Backend obfuscation only |
| `go build` | none | Regular development build |

### Troubleshooting

**garble not found:**
```bash
go install mvdan.cc/garble@latest
```

**WASM build fails:**
- Ensure Go 1.24.2+ is installed
- Check GOROOT is set correctly: `go env GOROOT`

**Server won't start:**
- Check if ports 8080-8090 are available
- Verify ngrok credentials in `obfuscation/constants_obfuscated.go`

**Binary too large:**
- This is normal for obfuscated + embedded builds (15-20 MB)
- Includes WASM module, TTS requirements scripts, and downloadable binaries

**Windows Defender blocking build:**
```powershell
# Option 1: Add temp directory exclusion
Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Temp"

# Option 2: Temporarily disable (run as Administrator)
Set-MpPreference -DisableRealtimeMonitoring $true
.\scripts\helpers\build-prod.ps1
Set-MpPreference -DisableRealtimeMonitoring $false

# Option 3: Build on Linux/WSL (no antivirus issues)
wsl ./scripts/helpers/build-prod.sh windows
```

**Garble fails on server_frontend.go:**
- This is expected - garble may have issues with standalone files
- The script already handles this by building server_frontend.go directly
- If issues persist, ensure all dependencies are in go.mod
