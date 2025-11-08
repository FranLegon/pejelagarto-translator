//go:build downloadable || ngrok_default

package main

import "embed"

//go:embed bin/pejelagarto-translator.exe bin/pejelagarto-translator
var embeddedBinaries embed.FS

const isDownloadable = true
