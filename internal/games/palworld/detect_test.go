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

	t.Run("embedded installation", func(t *testing.T) {
		t.Parallel()
		entries, err := fs.ReadDir(palworldtest.Files, ".")
		require.NoError(t, err)
		if len(entries) == 0 {
			t.Skip("no Palworld fixture for this platform")
		}
		dir := t.TempDir()
		require.NoError(t, os.CopyFS(dir, palworldtest.Files))
		got, err := palworld.New().Detect(dir)
		require.NoError(t, err)
		require.True(t, got)
	})

	for _, settings := range []string{
		"RESTAPIEnabled=True",
		"RESTAPIEnabled=true,ServerName=\"A, B; # (C)\",CrossplayPlatforms=(Steam,Xbox)",
		"ServerName=\"Escaped \\\"quote\\\"\",RESTAPIEnabled=False,RESTAPIEnabled=True",
		"\n; a comment\nRESTAPIEnabled = True,\nCrossplayPlatforms=(Steam,Xbox)\n",
	} {
		t.Run(settings, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			config := filepath.Join(dir, "Pal", "Saved", "Config", "LinuxServer")
			require.NoError(t, os.MkdirAll(config, 0o755))
			ini := "\ufeff[/Script/Pal.PalGameWorldSettings]\r\nOptionSettings=(" + settings + ")\r\n"
			require.NoError(t, os.WriteFile(filepath.Join(config, "PalWorldSettings.ini"), []byte(ini), 0o644))
			got, err := palworld.New().Detect(dir)
			require.NoError(t, err)
			require.True(t, got)
		})
	}

}

func TestDetectFalse(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	entries, err := fs.ReadDir(palworldtest.Files, ".")
	require.NoError(t, err)
	require.NoError(t, os.CopyFS(root, palworldtest.Files))
	for _, subdir := range []string{"Pal", "Pal/Saved", "Pal/Saved/Config"} {
		t.Run(subdir, func(t *testing.T) {
			t.Parallel()
			if len(entries) == 0 {
				t.Skip("no Palworld fixture for this platform")
			}
			matched, err := palworld.New().Detect(filepath.Join(root, filepath.FromSlash(subdir)))
			require.NoError(t, err)
			require.False(t, matched)
		})
	}

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
			got, err := palworld.New().Detect(dir)
			require.NoError(t, err)
			require.False(t, got)
		})
	}
}

func TestDetectError(t *testing.T) {
	t.Parallel()

	for _, launcher := range []string{"PalServer.sh", "PalServer.exe"} {
		t.Run(launcher+" without active configuration", func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, launcher), nil, 0o644))
			got, err := palworld.New().Detect(dir)
			require.False(t, got)
			require.ErrorContains(t, err, "no active PalWorldSettings.ini")
			require.ErrorContains(t, err, "RESTAPIEnabled=True")
		})
	}

	for _, test := range []struct {
		name     string
		section  string
		settings string
		wantErr  string
	}{
		{name: "disabled", settings: "RESTAPIEnabled=False", wantErr: "set RESTAPIEnabled=True"},
		{name: "missing", settings: "ServerName=\"Test\"", wantErr: "set RESTAPIEnabled=True"},
		{name: "wrong section", section: "Other", settings: "RESTAPIEnabled=True", wantErr: "set RESTAPIEnabled=True"},
		{name: "invalid boolean", settings: "RESTAPIEnabled=maybe", wantErr: "set RESTAPIEnabled=True"},
		{name: "quoted text is not a flag", settings: "ServerName=\"RESTAPIEnabled=True\"", wantErr: "set RESTAPIEnabled=True"},
		{name: "nested flag is not a flag", settings: "Other=(RESTAPIEnabled=True)", wantErr: "set RESTAPIEnabled=True"},
		{name: "commented flag", settings: "\n; RESTAPIEnabled=True\n", wantErr: "set RESTAPIEnabled=True"},
		{name: "last flag wins", settings: "RESTAPIEnabled=True,RESTAPIEnabled=False", wantErr: "set RESTAPIEnabled=True"},
		{name: "missing assignment", settings: "RESTAPIEnabled", wantErr: "invalid field"},
		{name: "unclosed parentheses", settings: "Other=(", wantErr: "unterminated OptionSettings"},
		{name: "unclosed quote", settings: "AdminPassword=\"never-log-secret,RESTAPIEnabled=True", wantErr: "unterminated OptionSettings"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			config := filepath.Join(dir, "Pal", "Saved", "Config", "LinuxServer")
			require.NoError(t, os.MkdirAll(config, 0o755))
			section := test.section
			if section == "" {
				section = "/Script/Pal.PalGameWorldSettings"
			}
			ini := "[" + section + "]\nOptionSettings=(" + test.settings + ")\n"
			path := filepath.Join(config, "PalWorldSettings.ini")
			require.NoError(t, os.WriteFile(path, []byte(ini), 0o644))
			// A launcher must not bypass validation of the active configuration.
			require.NoError(t, os.WriteFile(filepath.Join(dir, "PalServer.sh"), nil, 0o644))
			got, err := palworld.New().Detect(dir)
			require.False(t, got)
			require.ErrorContains(t, err, test.wantErr)
			require.ErrorContains(t, err, path)
			require.NotContains(t, err.Error(), "never-log-secret")
		})
	}

	t.Run("ambiguous configuration", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		config := filepath.Join(dir, "Pal", "Saved", "Config")
		for _, platform := range []string{"LinuxServer", "WindowsServer"} {
			require.NoError(t, os.MkdirAll(filepath.Join(config, platform), 0o755))
			settings := "[/Script/Pal.PalGameWorldSettings]\nOptionSettings=(RESTAPIEnabled=True)\n"
			require.NoError(t, os.WriteFile(filepath.Join(config, platform, "PalWorldSettings.ini"), []byte(settings), 0o644))
		}
		got, err := palworld.New().Detect(dir)
		require.False(t, got)
		require.ErrorContains(t, err, "multiple active settings files")
	})

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()
		got, err := palworld.New().Detect(filepath.Join(t.TempDir(), "missing"))
		require.ErrorIs(t, err, fs.ErrNotExist)
		require.False(t, got)
	})

	t.Run("file instead of directory", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(path, nil, 0o644))
		got, err := palworld.New().Detect(path)
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
		got, err := palworld.New().Detect(dir)
		require.ErrorIs(t, err, fs.ErrPermission)
		require.False(t, got)
	})
}
