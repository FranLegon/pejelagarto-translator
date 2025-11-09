//go:build !ngrok_default

package config

// No hardcoded ngrok configuration for regular builds
const (
	DefaultNgrokToken  = ""
	DefaultNgrokDomain = ""
	UseNgrokDefault    = false
)
