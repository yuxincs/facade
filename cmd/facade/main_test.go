package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
)

type testGame struct{}

func (*testGame) Protocol() string         { return "tcp" }
func (*testGame) Command(uint16) *exec.Cmd { return nil }

func TestSelectGame(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	want := &testGame{}
	failure := errors.New("cannot inspect installation")
	match := adaptConstructor(func(got string) (*testGame, error) {
		require.Equal(t, dir, got)
		return want, nil
	})
	miss := adaptConstructor(func(string) (*testGame, error) { return nil, errors.ErrUnsupported })
	broken := adaptConstructor(func(string) (*testGame, error) { return nil, failure })
	partial := adaptConstructor(func(string) (*testGame, error) { return want, failure })
	for _, tt := range []struct {
		name    string
		games   []func(string) (engine.Game, error)
		wantErr string
		cause   error
	}{
		{name: "no constructors", wantErr: "no supported game detected"},
		{name: "no match", games: []func(string) (engine.Game, error){miss}, wantErr: "no supported game detected"},
		{name: "one match", games: []func(string) (engine.Game, error){miss, match, miss}},
		{name: "multiple matches", games: []func(string) (engine.Game, error){match, match}, wantErr: "multiple games matched"},
		{name: "constructor error", games: []func(string) (engine.Game, error){broken}, wantErr: "initialize game", cause: failure},
		{name: "error after match", games: []func(string) (engine.Game, error){match, broken}, wantErr: "initialize game", cause: failure},
		{name: "partial result", games: []func(string) (engine.Game, error){partial}, wantErr: "initialize game", cause: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			game, err := selectGame(t.Context(), dir, tt.games...)
			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Same(t, want, game)
				return
			}
			require.Nil(t, game)
			require.ErrorContains(t, err, tt.wantErr)
			if tt.cause != nil {
				require.ErrorIs(t, err, tt.cause)
			}
		})
	}
}

func TestSelectGameCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	game, err := selectGame(ctx, t.TempDir(), adaptConstructor(func(string) (*testGame, error) {
		t.Fatal("constructor called with canceled context")
		return nil, nil
	}))
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, game)
}
