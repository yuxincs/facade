package engine_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yuxincs/facade/internal/engine"
)

type testGame struct{}

func (testGame) Protocol() string         { return "tcp" }
func (testGame) Command(uint16) *exec.Cmd { return nil }

func TestRun(t *testing.T) {
	t.Parallel()
	e := engine.Engine{Game: testGame{}}
	require.NoError(t, e.Run(t.Context()))
}

func TestRunWithoutGame(t *testing.T) {
	t.Parallel()
	require.ErrorContains(t, new(engine.Engine).Run(t.Context()), "no game configured")
}

func TestRunCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	e := engine.Engine{Game: testGame{}}
	require.ErrorIs(t, e.Run(ctx), context.Canceled)
}
