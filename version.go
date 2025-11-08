package main

// Version information for Pejelagarto Translator
// This constant is shared across all build modes.
//
// Note: server_frontend.go has //go:build ignore and must be built with:
//   go run server_frontend.go version.go
//   go build -o bin/frontend.exe server_frontend.go version.go
const Version = "v1.0.2"
