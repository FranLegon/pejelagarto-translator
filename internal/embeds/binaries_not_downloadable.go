//go:build !downloadable && !ngrok_default

package embeds

import "embed"

var EmbeddedBinaries embed.FS
