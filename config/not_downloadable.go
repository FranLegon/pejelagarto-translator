//go:build !downloadable && !ngrok_default

package config

import "embed"

var EmbeddedBinaries embed.FS

const IsDownloadable = false
