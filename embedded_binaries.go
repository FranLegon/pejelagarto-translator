//go:build downloadable || ngrok_default

package main

import "embed"

//go:embed bin/pejelagarto-translator.exe bin/pejelagarto-translator bin/pejelagarto-translator-webview.apk
var embeddedBinaries embed.FS
