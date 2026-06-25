//go:build !embedspa

// Package web optionally embeds the built SPA. The default (dev) build does NOT
// embed it, so `go run`/tests never require a frontend build. Production builds
// with `-tags embedspa` compile embed_prod.go instead.
package web

import "io/fs"

// DistFS returns the embedded SPA filesystem and whether it is present.
func DistFS() (fs.FS, bool) { return nil, false }
