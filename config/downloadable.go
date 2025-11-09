//go:build downloadable || ngrok_default

package config

import "embed"

//go:embed bin/pejelagarto-translator.exe bin/pejelagarto-translator
var EmbeddedBinaries embed.FS

const IsDownloadable = true
