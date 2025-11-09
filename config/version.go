package config

// Version information for Pejelagarto Translator
// This constant is shared across all build modes.
//
// Build modes:
//   Backend server:  go build -tags "ngrok_default,downloadable" .
//   Frontend server: go build -tags "frontendserver,ngrok_default,downloadable" .
//   WASM module:     GOOS=js GOARCH=wasm go build -tags frontend .
//
// Changelog v1.2.3:
//   - FIXED: WASM loading issue on ngrok deployments ('Failed to load translation module')
//   - ENHANCED: server_frontend.go now searches multiple paths including executable directory
//   - IMPROVED: Build script copies WASM files to both bin/ and project root for reliability
//   - ADDED: path/filepath import for robust cross-platform path handling
//
// Changelog v1.2.2:
//   - FIXED: Version link CSS (better visibility in dark/light themes with badge style)
//   - FIXED: Pronunciation text box to always show Pejelagarto pronunciation
//   - FIXED: Removed incorrect number conversion in preprocessTextForTTS
//   - FIXED: Pronunciation updates correctly when swapping translation direction
//   - ADDED: RemoveTimestampSpecialCharacters exported function for TTS preprocessing
//
// Changelog v1.2.1:
//   - VALIDATED: All 13 build tag combinations compile and test successfully
//   - CONSOLIDATED: All garble+ngrok documentation into README.md
//   - CLEANUP: Removed separate MD files and log artifacts
//
// Changelog v1.2.0:
//   - FIXED: WASM file naming (translator.wasm)
//   - FIXED: ngrok domain configuration (empty for random URLs)
//   - DOCUMENTED: Garble+ngrok incompatibility (definitively proven)
//   - DEPRECATED: Garble production build for ngrok use
const Version = "v1.2.3"
