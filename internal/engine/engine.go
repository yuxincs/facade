// Package engine coordinates games and server lifecycle management.
package engine

import (
	"context"
	"fmt"
	"os/exec"
)

// Game provides transport and startup details for an initialized game.
type Game interface {
	// Protocol returns the gameplay transport: "tcp" or "udp".
	Protocol() string
	// Command builds an unstarted server command.
	// The port is the backend gameplay port chosen by the engine.
	Command(port uint16) *exec.Cmd
}

// PlayerCounter is an optional capability required by UDP games.
type PlayerCounter interface {
	// PlayerCount returns the connected player count. An error means the count is unknown.
	PlayerCount(ctx context.Context) (int, error)
}

// Saver is an optional capability for persisting game state before suspension or shutdown.
type Saver interface {
	Save(ctx context.Context) error
}

// Engine coordinates an initialized game.
type Engine struct {
	Game Game
}

// Run validates the initialized game.
// Server lifecycle management is not implemented yet.
func (e *Engine) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if e.Game == nil {
		return fmt.Errorf("no game configured")
	}
	return nil
}
