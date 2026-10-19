// Package palworld integrates Palworld servers with Facade.
package palworld

import (
	"fmt"
	"net/http"
)

// Palworld provides Palworld-specific discovery and settings.
type Palworld struct {
	root          string
	restPort      int
	adminPassword string
	restClient    *http.Client
}

// String keeps credentials out of formatted logs.
func (a Palworld) String() string {
	return fmt.Sprintf("palworld.Palworld{restPort: %d, adminPassword: [REDACTED]}", a.restPort)
}

// GoString also redacts credentials when formatting with %#v.
func (a Palworld) GoString() string { return a.String() }
