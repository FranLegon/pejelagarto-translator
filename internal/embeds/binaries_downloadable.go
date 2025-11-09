//go:build downloadable || ngrok_default

package embeds

import "embed"

//go:embed bin/pejelagarto-translator.exe bin/pejelagarto-translator
var EmbeddedBinaries embed.FS
