package main

// Version information for Pejelagarto Translator
// This constant is shared across all build modes.
//
// Build modes:
//   Backend server:  go build -tags "ngrok_default,downloadable" .
//   Frontend server: go build -tags "frontendserver,ngrok_default,downloadable" .
//   WASM module:     GOOS=js GOARCH=wasm go build -tags frontend .
const Version = "v1.0.7"
