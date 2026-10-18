package e2e_test

import (
	"context"
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
		dir        string
		wantExit   int
		wantOutput string
	}
	cases := []testCase{
		{name: "empty directory", dir: t.TempDir(), wantExit: 1, wantOutput: "no supported game detected"},
	}
	for _, dir := range palworldtest.Scenarios(t) {
		cases = append(cases, testCase{
			name:       filepath.Base(dir),
			dir:        dir,
			wantOutput: "detected game",
		})
	}
	disabledDir := palworldtest.Scenarios(t)[0]
	config := filepath.Join(disabledDir, "Pal", "Saved", "Config", "LinuxServer", "PalWorldSettings.ini")
	settings, err := os.ReadFile(config)
	require.NoError(t, err)
	settings = []byte(strings.ReplaceAll(string(settings), "RESTAPIEnabled=True", "RESTAPIEnabled=False"))
	require.NoError(t, os.WriteFile(config, settings, 0o644))
	cases = append(cases, testCase{
		name:       "REST API disabled",
		dir:        disabledDir,
		wantExit:   1,
		wantOutput: "set RESTAPIEnabled=True",
	})

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary)
			cmd.Dir = test.dir
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
