//go:build !downloadable

package main

import "embed"

var embeddedBinaries embed.FS

const isDownloadable = false
