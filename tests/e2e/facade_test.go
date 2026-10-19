package e2e_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/games/palworld/palworldtest"
)

func TestFacade(t *testing.T) {
	t.Parallel()

	binary := filepath.Join(t.TempDir(), "facade")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "../../cmd/facade")
	output, err := build.CombinedOutput()
	require.NoError(t, err, "build facade: %s", output)

	type testCase struct {
		name       string
		args       []string
		files      fs.FS
		prepare    func(*testing.T, string)
		wantExit   int
		wantOutput string
	}
	cases := []testCase{
		{name: "help without installation", args: []string{"-help"}, wantOutput: "Usage: facade"},
		{name: "version without installation", args: []string{"-version"}, wantOutput: "facade dev"},
		{name: "palworld", files: palworldtest.Files, wantOutput: "detected game"},
		{name: "empty directory", wantExit: 1, wantOutput: "no supported game detected"},
		{
			name: "REST API disabled", files: palworldtest.Files,
			wantExit: 1, wantOutput: "set RESTAPIEnabled=True",
			prepare: func(t *testing.T, dir string) {
				configs, err := filepath.Glob(filepath.Join(dir, "Pal", "Saved", "Config", "*", "PalWorldSettings.ini"))
				require.NoError(t, err)
				require.Len(t, configs, 1)
				settings, err := os.ReadFile(configs[0])
				require.NoError(t, err)
				settings = []byte(strings.ReplaceAll(string(settings), "RESTAPIEnabled=True", "RESTAPIEnabled=False"))
				require.NoError(t, os.WriteFile(configs[0], settings, 0o644))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if test.files != nil {
				entries, err := fs.ReadDir(test.files, ".")
				require.NoError(t, err)
				if len(entries) == 0 {
					t.Skip("no fixture for this platform")
				}
				require.NoError(t, os.CopyFS(dir, test.files))
			}
			if test.prepare != nil {
				test.prepare(t, dir)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, test.args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			if test.wantExit == 0 {
				require.NoError(t, err, "%s", output)
			} else {
				var exitErr *exec.ExitError
				require.ErrorAs(t, err, &exitErr, "%s", output)
				require.Equal(t, test.wantExit, exitErr.ExitCode(), "%s", output)
			}
			require.Contains(t, string(output), test.wantOutput)
			require.NotContains(t, string(output), "test-only")
		})
	}
}
