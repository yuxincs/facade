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

// Game provides transport and startup details for an initialized game.
type Game interface {
	// Protocol returns the gameplay transport: "tcp" or "udp".
	Protocol() string
	// Command builds an unstarted server command.
	// The port is the backend gameplay port chosen by the engine.
	Command(port uint16) *exec.Cmd
}

// GameFactory initializes a game from dir, returning nil, nil for an unrelated directory.
type GameFactory func(dir string) (Game, error)

// PlayerCounter is an optional capability required by UDP games.
type PlayerCounter interface {
	// PlayerCount returns the connected player count. An error means the count is unknown.
	PlayerCount(ctx context.Context) (int, error)
}

// Engine coordinates the registered games.
type Engine struct {
	Games []GameFactory
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
	for _, newGame := range e.Games {
		if err := ctx.Err(); err != nil {
			return err
		}
		game, err := newGame(dir)
		if err != nil {
			return fmt.Errorf("initialize game: %w", err)
		}
		if game != nil {
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
