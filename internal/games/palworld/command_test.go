package palworld

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "server with spaces")
	config := filepath.Join(dir, "Pal", "Saved", "Config", "LinuxServer")
	require.NoError(t, os.MkdirAll(config, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(config, "PalWorldSettings.ini"), []byte(
		"[/Script/Pal.PalGameWorldSettings]\nOptionSettings=(RESTAPIEnabled=True,PublicPort=8211)\n",
	), 0o644))
	game, err := New(dir)
	require.NoError(t, err)
	require.NotNil(t, game)

	launcher := "PalServer.sh"
	if runtime.GOOS == "windows" {
		launcher = "PalServer.exe"
	}
	path := filepath.Join(dir, launcher)
	cmd := game.Command(9000)
	require.Equal(t, path, cmd.Path)
	require.Equal(t, []string{path, "-port=9000"}, cmd.Args)
	require.Equal(t, dir, cmd.Dir)
	require.Nil(t, cmd.Process)

	other := game.Command(9001)
	require.NotSame(t, cmd, other)
	require.Equal(t, []string{path, "-port=9001"}, other.Args)
}
