// Package engine coordinates games and server lifecycle management.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// Game provides game-specific detection, transport, and startup details.
type Game interface {
	// Detect reports whether dir looks like an installation of this game.
	Detect(dir string) (bool, error)
	// Protocol returns the gameplay transport: "tcp" or "udp".
	Protocol() string
	// Command builds an unstarted server command after successful detection.
	// The port is the backend gameplay port chosen by the engine.
	Command(port uint16) *exec.Cmd
}

// PlayerCounter is an optional capability required by UDP games.
type PlayerCounter interface {
	// PlayerCount returns the connected player count. An error means the count is unknown.
	PlayerCount(ctx context.Context) (int, error)
}

// Engine coordinates the registered games.
type Engine struct {
	Games []Game
}

// Run detects the game in the working directory and reports the selected game.
// Server lifecycle management is not implemented yet.
func (e *Engine) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	var matches []string
	for _, game := range e.Games {
		if err := ctx.Err(); err != nil {
			return err
		}
		matched, err := game.Detect(dir)
		if err != nil {
			return fmt.Errorf("detect %T: %w", game, err)
		}
		if matched {
			matches = append(matches, fmt.Sprintf("%T", game))
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	switch len(matches) {
	case 0:
		return fmt.Errorf("no supported game detected in %q", dir)
	case 1:
		slog.Info("detected game", "game", matches[0], "directory", dir)
		return nil
	default:
		return fmt.Errorf("multiple games matched %q: %s", dir, strings.Join(matches, ", "))
	}
}
