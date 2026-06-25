//go:build embedspa

package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// DistFS returns the embedded SPA (the contents of ./dist), populated by the
// production build (Dockerfile / `make build`, which copies frontend/dist here).
func DistFS() (fs.FS, bool) {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, false
	}
	return sub, true
}
