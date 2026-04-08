package webassets

import "embed"

// Assets contains the embedded browser client assets.
//
//go:embed index.html static/*
var Assets embed.FS
