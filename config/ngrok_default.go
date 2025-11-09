//go:build ngrok_default

package config

// Hardcoded ngrok configuration for ngrok_default builds
// Note: Empty domain means ngrok will use a random URL (recommended)
// Get a reserved domain at: https://dashboard.ngrok.com/cloud-edge/domains
const (
	DefaultNgrokToken  = "34QfuhfXXNQmIe0TbFH67RmNZZZ_7TtoYMAdwwgdYV1JFE1z6"
	DefaultNgrokDomain = "" // Empty = use random URL (prevents "domain already in use" errors)
	UseNgrokDefault    = true
)
