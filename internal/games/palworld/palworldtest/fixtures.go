// Package palworldtest provides shared Palworld test fixtures.
package palworldtest

import (
	"embed"
	"io/fs"
	"runtime"
)

//go:embed testdata
var files embed.FS

// Files exposes the platform's server root, or an empty filesystem if unsupported.
var Files = func() fs.FS {
	switch runtime.GOOS {
	case "linux", "windows":
		root, err := fs.Sub(files, "testdata/"+runtime.GOOS)
		if err != nil {
			panic(err)
		}
		return root
	default:
		return embed.FS{}
	}
}()
