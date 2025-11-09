package config

// Version information for Pejelagarto Translator
// This constant is shared across all build modes.
//
// Build modes:
//   Backend server:  go build -tags "ngrok_default,downloadable" .
//   Frontend server: go build -tags "frontendserver,ngrok_default,downloadable" .
//   WASM module:     GOOS=js GOARCH=wasm go build -tags frontend .
//
// Changelog v1.2.0:
//   - FIXED: WASM file naming (translator.wasm)
//   - FIXED: ngrok domain configuration (empty for random URLs)
//   - DOCUMENTED: Garble+ngrok incompatibility (definitively proven)
//   - DEPRECATED: Garble production build for ngrok use
const Version = "v1.2.0"
