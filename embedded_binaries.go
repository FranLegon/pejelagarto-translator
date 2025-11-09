//go:build downloadable || ngrok_default

package main

import "embed"

//go:embed bin/pejelagarto-translator-windows-amd64.exe bin/pejelagarto-translator-linux-amd64 bin/pejelagarto-translator-darwin-amd64 bin/pejelagarto-translator.apk
var embeddedBinaries embed.FS
