package engine_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
)

type gameFunc func(string) (bool, error)

func (f gameFunc) Detect(dir string) (bool, error) { return f(dir) }

func (gameFunc) Protocol() string { return "tcp" }

func TestRun(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	require.NoError(t, err)
	failure := errors.New("cannot inspect installation")
	match := gameFunc(func(got string) (bool, error) {
		return got == dir, nil
	})
	miss := gameFunc(func(string) (bool, error) { return false, nil })
	broken := gameFunc(func(string) (bool, error) { return false, failure })

	for _, test := range []struct {
		name    string
		games   []engine.Game
		wantErr string
		cause   error
	}{
		{name: "no games", wantErr: "no supported game detected"},
		{name: "no match", games: []engine.Game{miss}, wantErr: "no supported game detected"},
		{name: "one match", games: []engine.Game{miss, match, miss}},
		{name: "multiple matches", games: []engine.Game{match, match}, wantErr: "multiple games matched"},
		{name: "detection error", games: []engine.Game{broken}, wantErr: "detect engine_test.gameFunc", cause: failure},
		{name: "error after match", games: []engine.Game{match, broken}, wantErr: "detect engine_test.gameFunc", cause: failure},
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
	e := engine.Engine{Games: []engine.Game{
		gameFunc(func(string) (bool, error) {
			t.Fatal("Detect called with canceled context")
			return false, nil
		}),
	}}
	require.ErrorIs(t, e.Run(ctx), context.Canceled)
}
