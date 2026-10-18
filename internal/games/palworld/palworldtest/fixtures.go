// Package palworldtest provides shared Palworld test fixtures.
package palworldtest

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Files contains shared Palworld configuration files for tests.
//
//go:embed testdata
var Files embed.FS

// Scenarios creates independent, valid configuration layouts for Linux and Windows.
// Each returned working directory is named for its scenario and cleaned up with t.
func Scenarios(t *testing.T) []string {
	t.Helper()

	settings, err := Files.ReadFile("testdata/PalWorldSettings.ini")
	require.NoError(t, err)

	root := t.TempDir()
	var dirs []string
	for _, platform := range []struct {
		name string
		dir  string
	}{
		{name: "linux", dir: "LinuxServer"},
		{name: "windows", dir: "WindowsServer"},
	} {
		for _, layout := range []struct {
			name string
			path string
		}{
			{name: "root", path: "Pal/Saved/Config"},
			{name: "pal", path: "Saved/Config"},
			{name: "saved", path: "Config"},
			{name: "config", path: "."},
		} {
			dir := filepath.Join(root, platform.name+"-"+layout.name)
			config := filepath.Join(dir, filepath.FromSlash(layout.path), platform.dir)
			require.NoError(t, os.MkdirAll(config, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(config, "PalWorldSettings.ini"), settings, 0o644))
			dirs = append(dirs, dir)
		}
	}
	return dirs
}
