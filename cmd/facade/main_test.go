package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
)

func TestUDPAdaptersImplementPlayerCounter(t *testing.T) {
	t.Parallel()
	for _, adapter := range _registeredAdapters {
		if adapter.Protocol() == "udp" {
			require.Implements(t, (*engine.PlayerCounter)(nil), adapter, "UDP adapter %T must support player counting", adapter)
		}
	}
}
