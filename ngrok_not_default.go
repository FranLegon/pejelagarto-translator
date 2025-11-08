//go:build !ngrok_default

package main

// No hardcoded ngrok configuration for regular builds
const (
	defaultNgrokToken  = ""
	defaultNgrokDomain = ""
	useNgrokDefault    = false
)
