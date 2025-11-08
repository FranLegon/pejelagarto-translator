//go:build !downloadable && !ngrok_default

package main

import "embed"

var embeddedBinaries embed.FS

const isDownloadable = false
