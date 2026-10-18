// Package engine coordinates game adapters and server lifecycle management.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Adapter provides game-specific installation detection.
type Adapter interface {
	// Detect reports whether dir looks like an installation of this game.
	Detect(dir string) (bool, error)
}

// Engine coordinates the registered game adapters.
type Engine struct {
	Adapters []Adapter
}

// Run detects the game in the working directory and reports the selected adapter.
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
	for _, adapter := range e.Adapters {
		if err := ctx.Err(); err != nil {
			return err
		}
		matched, err := adapter.Detect(dir)
		if err != nil {
			return fmt.Errorf("detect %T: %w", adapter, err)
		}
		if matched {
			matches = append(matches, fmt.Sprintf("%T", adapter))
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	switch len(matches) {
	case 0:
		return fmt.Errorf("no supported game detected in %q", dir)
	case 1:
		slog.Info("detected game", "adapter", matches[0], "directory", dir)
		return nil
	default:
		return fmt.Errorf("multiple game adapters matched %q: %s", dir, strings.Join(matches, ", "))
	}
}
