package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
)

func TestUDPGamesImplementPlayerCounter(t *testing.T) {
	t.Parallel()
	for _, game := range _games {
		if game.Protocol() == "udp" {
			require.Implements(t, (*engine.PlayerCounter)(nil), game, "UDP game %T must support player counting", game)
		}
	}
}
