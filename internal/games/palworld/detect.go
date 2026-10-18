// Package palworld provides the Palworld game adapter.
package palworld

import "errors"

// Adapter provides Palworld-specific discovery and settings.
type Adapter struct{}

// Detect is a placeholder for recognizing a Palworld installation.
func (Adapter) Detect(_ string) (bool, error) {
	return false, errors.ErrUnsupported
}
