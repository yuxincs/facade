// Package engine coordinates game adapters and server lifecycle management.
package engine

import (
	"context"
	"errors"
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

// Run is a placeholder for lifecycle coordination.
func (e *Engine) Run(_ context.Context) error {
	return errors.ErrUnsupported
}
