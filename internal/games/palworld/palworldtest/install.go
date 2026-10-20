// Package palworldtest prepares Palworld server installations for tests.
package palworldtest

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//go:embed server/main.go
var _serverSource string

//go:embed testdata
var _files embed.FS

// Install prepares a temporary Palworld installation and returns its root.
// It copies the settings fixture, chooses a REST port, and builds a fake server.
// The server echoes UDP packets, reports an empty player list, implements Save,
// and appends started/saved/stopped events to events.log in the installation.
// Its first stdout line is its UDP address, printed after both listeners bind.
// macOS uses the Linux fixture and launcher. Other unsupported platforms skip.
// Install does not start the server; t manages removal of the installation.
func Install(t *testing.T) string {
	t.Helper()
	platform := runtime.GOOS
	switch platform {
	case "darwin":
		platform = "linux"
	case "linux", "windows":
	default:
		t.Skip("no fake Palworld installation for this platform")
	}
	fixture, err := fs.Sub(_files, "testdata/"+platform)
	require.NoError(t, err)
	root := filepath.Join(t.TempDir(), "Palworld server")
	require.NoError(t, os.CopyFS(root, fixture))
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	configs, err := filepath.Glob(filepath.Join(root, "Pal", "Saved", "Config", "*", "PalWorldSettings.ini"))
	require.NoError(t, err)
	require.Len(t, configs, 1)
	settings, err := os.ReadFile(configs[0])
	require.NoError(t, err)
	pattern := regexp.MustCompile(`RESTAPIPort=\d+`)
	require.Len(t, pattern.FindAll(settings, -1), 1)
	settings = pattern.ReplaceAll(settings, []byte(fmt.Sprintf("RESTAPIPort=%d", port)))
	require.NoError(t, os.WriteFile(configs[0], settings, 0o644))

	source := filepath.Join(t.TempDir(), "main.go")
	require.NoError(t, os.WriteFile(source, []byte(_serverSource), 0o644))
	binary := filepath.Join(root, "PalServer")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, source)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "build fake Palworld server: %s", output)
	if runtime.GOOS != "windows" {
		launcher := "#!/bin/sh\ncd \"$(dirname \"$0\")\" || exit 1\nexec ./PalServer \"$@\"\n"
		require.NoError(t, os.WriteFile(filepath.Join(root, "PalServer.sh"), []byte(launcher), 0o755))
	}
	return root
}
