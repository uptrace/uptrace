package uptrace

import (
	"embed"
	"io/fs"
)

//go:embed vue/dist
var distFS embed.FS

func DistFS() fs.FS {
	if fs, err := fs.Sub(distFS, "vue/dist"); err == nil {
		return fs
	}
	return distFS
}
