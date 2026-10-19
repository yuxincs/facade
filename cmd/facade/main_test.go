package main

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
	"github.com/yuxincs/facade/internal/games/palworld/palworldtest"
)

func TestUDPGamesImplementPlayerCounter(t *testing.T) {
	t.Parallel()
	entries, err := fs.ReadDir(palworldtest.Files, ".")
	require.NoError(t, err)
	if len(entries) == 0 {
		t.Skip("no game fixture for this platform")
	}
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, palworldtest.Files))
	matched := 0
	for _, newGame := range _games {
		game, err := newGame(dir)
		require.NoError(t, err)
		if game == nil {
			continue
		}
		matched++
		if game.Protocol() == "udp" {
			require.Implements(t, (*engine.PlayerCounter)(nil), game, "UDP game %T must support player counting", game)
		}
	}
	require.Positive(t, matched)
}
