# Testing: Garble vs Windows Defender with ngrok

This document provides step-by-step instructions to determine if garble obfuscation or Windows Defender is causing ngrok connection failures.

## Quick Test (Automated)

Run the automated test script:

```powershell
# Basic test (checks Defender status)
.\scripts\test\test-garble-ngrok.ps1

# Test with Defender exclusions (requires Admin)
.\scripts\test\test-garble-ngrok.ps1 -AddDefenderExclusions

# Test without checking Defender (faster)
.\scripts\test\test-garble-ngrok.ps1 -SkipDefenderCheck
```

The script will:
1. ✅ Build unobfuscated version (baseline test)
2. ✅ Build garble-obfuscated version
3. ✅ Build garble WITHOUT -literals flag
4. ✅ Test each binary with ngrok
5. ✅ Report which configuration works

## Manual Test Process

If you prefer to test manually:

### Test 1: Baseline (Unobfuscated Build)

**Expected Result:** Should work (already confirmed)

```powershell
# Build
.\scripts\helpers\build-prod-unobfuscated.ps1

# Test
.\bin\piper-server.exe

# Look for:
# ✓ "ngrok tunnel established successfully! ✓"
# ✓ "Public URL: https://..."
# ❌ "Failed to start ngrok listener"
# ❌ "remote gone away"
```

Press `Ctrl+C` to stop if working.

---

### Test 2: Disable Windows Defender (Temporarily)

**This isolates if Defender is interfering**

1. **Disable Defender Real-Time Protection:**
   - Open Windows Security → Virus & threat protection
   - Manage settings → Real-time protection → OFF
   - ⚠️ **Remember to re-enable after testing!**

2. **Build garble version:**
   ```powershell
   .\scripts\helpers\build-prod.ps1
   ```

3. **Test:**
   ```powershell
   .\bin\piper-server.exe
   ```

4. **Result interpretation:**
   - ✅ **Works now**: Defender was blocking execution
   - ❌ **Still fails**: Garble obfuscation is the issue

5. **Re-enable Defender!**

---

### Test 3: Add Defender Exclusions (Keep Protection On)

**This tests if exclusions help with Defender blocking**

1. **Add exclusions (run PowerShell as Administrator):**
   ```powershell
   # Exclude temp directory (where garble works)
   Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Temp"
   
   # Exclude output directory
   Add-MpPreference -ExclusionPath "C:\Users\francisco.legon\OneDrive - Open IT\Documentos\Personal\Go\pejelagarto-translator\bin"
   
   # Verify
   Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
   ```

2. **Clean and rebuild:**
   ```powershell
   Remove-Item bin\piper-server.exe -Force
   .\scripts\helpers\build-prod.ps1
   ```

3. **Test:**
   ```powershell
   .\bin\piper-server.exe
   ```

4. **Result interpretation:**
   - ✅ **Works now**: Defender was the issue (can use garble with exclusions)
   - ❌ **Still fails**: Garble obfuscation breaks ngrok SDK

---

### Test 4: Garble Without -literals Flag

**This tests if string literal obfuscation breaks ngrok**

The `-literals` flag obfuscates string constants, which could break:
- Protocol message strings
- API endpoint URLs
- Configuration keys
- Authentication tokens

**Manual build without -literals:**

```powershell
# Build WASM
$env:GOOS="js"; $env:GOARCH="wasm"
garble -tiny -seed=random build -tags "frontend" -o bin/main.wasm .

# Copy wasm_exec.js
Copy-Item "$(go env GOROOT)\misc\wasm\wasm_exec.js" bin\

# Build server WITHOUT -literals flag
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
garble -tiny -seed=random build `
    -tags "frontendserver,obfuscated,ngrok_default,downloadable" `
    -ldflags="-s -w -extldflags '-static'" `
    -trimpath `
    -o bin\piper-server-no-literals.exe `
    .
```

**Test:**
```powershell
.\bin\piper-server-no-literals.exe
```

**Result interpretation:**
- ✅ **Works**: The `-literals` flag is the culprit
- ❌ **Still fails**: Core garble obfuscation breaks ngrok

---

## Expected Test Results

### Scenario A: Garble is the Problem

| Build Configuration | Windows Defender | Result |
|---------------------|------------------|--------|
| Unobfuscated | ON | ✅ Works |
| Garble (full) | ON | ❌ Fails |
| Garble (full) | OFF | ❌ Fails |
| Garble (no -literals) | ON | ✅ Works or ❌ Fails |

**Conclusion:** Garble obfuscation breaks ngrok SDK (reflection/interface issues)

**Solution:** Use `build-prod-unobfuscated.ps1` for production

---

### Scenario B: Windows Defender is the Problem

| Build Configuration | Windows Defender | Result |
|---------------------|------------------|--------|
| Unobfuscated | ON | ✅ Works |
| Garble (full) | ON | ❌ Fails |
| Garble (full) | OFF | ✅ Works |
| Garble (full) | ON + Exclusions | ✅ Works |

**Conclusion:** Windows Defender blocks garble-obfuscated binaries

**Solution:** Add Defender exclusions or use unobfuscated build

---

### Scenario C: Both are Problems

| Build Configuration | Windows Defender | Result |
|---------------------|------------------|--------|
| Unobfuscated | ON | ✅ Works |
| Garble (full) | ON | ❌ Fails (Defender blocks) |
| Garble (full) | OFF | ❌ Fails (ngrok broken) |

**Conclusion:** Defender blocks the binary AND garble breaks ngrok

**Solution:** Must use unobfuscated build

---

## Why Garble Might Break ngrok

Garble obfuscation can break ngrok SDK through:

1. **String Literal Obfuscation (-literals flag):**
   - Protocol message strings modified
   - API endpoints scrambled
   - Authentication token constants changed
   - Configuration keys renamed

2. **Interface Method Obfuscation:**
   - ngrok SDK uses interfaces for extensibility
   - Method names must match exactly
   - Garble renames unexported methods

3. **Reflection Issues:**
   - ngrok likely uses reflection for config
   - Type names must be preserved
   - Field names must match JSON/protocol

4. **Import Path Obfuscation:**
   - Changes package initialization order
   - May break registration patterns

## Known Garble Incompatibilities

From garble's issue tracker, similar problems with:
- #849: Wails (reflection on types)
- #791: survey/v2 (broken by obfuscation)
- #962: GORM (database foreign keys)
- #966: Reflection-based tests fail

ngrok SDK likely has similar dependencies.

## Recommendations

### For Development/Testing:
```powershell
# Use unobfuscated build (fast, reliable)
.\scripts\helpers\build-prod-unobfuscated.ps1
```

### For Production (ngrok required):
```powershell
# Must use unobfuscated build
.\scripts\helpers\build-prod-unobfuscated.ps1
```

### For Production (no ngrok, max security):
```powershell
# Can use garble with Defender exclusions
.\scripts\helpers\build-prod.ps1
```

### If You Must Use Garble + ngrok:

Try building without `-literals` flag as shown in Test 4. This may preserve enough string constants for ngrok to work while still providing some obfuscation.

---

## Cleanup

Remove test binaries:
```powershell
Remove-Item bin\piper-server-garbled.exe -ErrorAction SilentlyContinue
Remove-Item bin\piper-server-no-literals.exe -ErrorAction SilentlyContinue
```

Remove Defender exclusions (if added):
```powershell
# Run as Administrator
Remove-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Temp"
Remove-MpPreference -ExclusionPath "C:\Users\francisco.legon\OneDrive - Open IT\Documentos\Personal\Go\pejelagarto-translator\bin"
```
