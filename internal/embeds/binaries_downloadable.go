//go:build downloadable || ngrok_default

package embeds

import "embed"

//go:embed ../../bin/pejelagarto-translator.exe ../../bin/pejelagarto-translator ../../bin/pejelagarto-translator.apk ../../bin/pejelagarto-translator-webview.apk
var EmbeddedBinaries embed.FS
