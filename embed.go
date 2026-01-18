package goilerplate

import "embed"

// JukeboxFS contains the built SvelteKit app files.
// Run "go run ./cmd/do build jukebox" to build.
//
//go:embed all:jukelab/build
var JukeboxFS embed.FS
