// Package proxy handles TCP and UDP forwarding.
package proxy

import (
	"context"
	"errors"
)

// Proxy will forward traffic between players and a game server.
type Proxy struct{}

// Run is a placeholder for network forwarding.
func (p *Proxy) Run(_ context.Context) error {
	return errors.ErrUnsupported
}
