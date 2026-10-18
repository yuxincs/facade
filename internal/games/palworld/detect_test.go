package palworld_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/games/palworld"
	"github.com/yuxincs/facade/internal/games/palworld/palworldtest"
)

func TestDetect(t *testing.T) {
	t.Parallel()

	for _, dir := range palworldtest.Scenarios(t) {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			t.Parallel()

			got, err := (palworld.Adapter{}).Detect(dir)
			require.NoError(t, err)
			require.True(t, got)
		})
	}

	for _, launcher := range []string{"PalServer.sh", "PalServer.exe"} {
		t.Run(launcher+" before first start", func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, launcher), nil, 0o644))

			got, err := (palworld.Adapter{}).Detect(dir)
			require.NoError(t, err)
			require.True(t, got)
		})
	}

}

func TestDetectFalse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		prepare func(*testing.T, string)
	}{
		{name: "empty directory"},
		{
			name: "different game",
			prepare: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "server.properties"), []byte("server-port=25565\n"), 0o644))
			},
		},
		{
			name: "settings in wrong location",
			prepare: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "PalWorldSettings.ini"), nil, 0o644))
			},
		},
		{
			name: "directory named like launcher",
			prepare: func(t *testing.T, dir string) {
				require.NoError(t, os.Mkdir(filepath.Join(dir, "PalServer.sh"), 0o755))
			},
		},
		{
			name: "non-directory in settings path",
			prepare: func(t *testing.T, dir string) {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "Pal"), []byte("not a directory"), 0o644))
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if test.prepare != nil {
				test.prepare(t, dir)
			}
			got, err := (palworld.Adapter{}).Detect(dir)
			require.NoError(t, err)
			require.False(t, got)
		})
	}
}

func TestDetectError(t *testing.T) {
	t.Parallel()

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()
		got, err := (palworld.Adapter{}).Detect(filepath.Join(t.TempDir(), "missing"))
		require.ErrorIs(t, err, fs.ErrNotExist)
		require.False(t, got)
	})

	t.Run("file instead of directory", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(path, nil, 0o644))
		got, err := (palworld.Adapter{}).Detect(path)
		require.Error(t, err)
		require.False(t, got)
	})

	t.Run("permission denied", func(t *testing.T) {
		t.Parallel()
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("requires Unix permission enforcement and a non-root user")
		}
		dir := t.TempDir()
		locked := filepath.Join(dir, "Pal")
		require.NoError(t, os.Mkdir(locked, 0o000))
		t.Cleanup(func() {
			require.NoError(t, os.Chmod(locked, 0o755))
		})
		got, err := (palworld.Adapter{}).Detect(dir)
		require.ErrorIs(t, err, fs.ErrPermission)
		require.False(t, got)
	})
}
