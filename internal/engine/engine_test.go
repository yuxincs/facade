package engine_test

import (
	"context"
	"errors"
	"os"
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
	dir, err := os.Getwd()
	require.NoError(t, err)
	failure := errors.New("cannot inspect installation")
	match := engine.GameFactory(func(got string) (engine.Game, error) {
		require.Equal(t, dir, got)
		return testGame{}, nil
	})
	miss := engine.GameFactory(func(string) (engine.Game, error) { return nil, nil })
	broken := engine.GameFactory(func(string) (engine.Game, error) { return nil, failure })

	for _, test := range []struct {
		name    string
		games   []engine.GameFactory
		wantErr string
		cause   error
	}{
		{name: "no games", wantErr: "no supported game detected"},
		{name: "no match", games: []engine.GameFactory{miss}, wantErr: "no supported game detected"},
		{name: "one match", games: []engine.GameFactory{miss, match, miss}},
		{name: "multiple matches", games: []engine.GameFactory{match, match}, wantErr: "multiple games matched"},
		{name: "detection error", games: []engine.GameFactory{broken}, wantErr: "initialize game", cause: failure},
		{name: "error after match", games: []engine.GameFactory{match, broken}, wantErr: "initialize game", cause: failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			e := engine.Engine{Games: test.games}
			err := e.Run(t.Context())
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
			if test.cause != nil {
				require.ErrorIs(t, err, test.cause)
			}
		})
	}
}

func TestRunCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	e := engine.Engine{Games: []engine.GameFactory{
		func(string) (engine.Game, error) {
			t.Fatal("constructor called with canceled context")
			return nil, nil
		},
	}}
	require.ErrorIs(t, e.Run(ctx), context.Canceled)
}
