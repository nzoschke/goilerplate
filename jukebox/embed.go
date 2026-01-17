package jukebox

import "embed"

// JukeboxFS contains the built SvelteKit app files.
// Run "go run ./cmd/do build jukebox" to build and populate this directory.
//
//go:embed all:*
var JukeboxFS embed.FS
