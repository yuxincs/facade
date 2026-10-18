// Package palworld provides the Palworld game adapter.
package palworld

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// Adapter provides Palworld-specific discovery and settings.
type Adapter struct{}

// Detect recognizes an installation from its root, Pal, Saved, or Config directory.
// It checks known layouts without recursively searching or parsing settings.
func (Adapter) Detect(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return false, fmt.Errorf("palworld: inspect directory: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("palworld: %q is not a directory", dir)
	}

	for _, marker := range []string{
		"PalServer.sh",
		"PalServer.exe",
		"Pal/Saved/Config/LinuxServer/PalWorldSettings.ini",
		"Pal/Saved/Config/WindowsServer/PalWorldSettings.ini",
		"Saved/Config/LinuxServer/PalWorldSettings.ini",
		"Saved/Config/WindowsServer/PalWorldSettings.ini",
		"Config/LinuxServer/PalWorldSettings.ini",
		"Config/WindowsServer/PalWorldSettings.ini",
		"LinuxServer/PalWorldSettings.ini",
		"WindowsServer/PalWorldSettings.ini",
	} {
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(marker)))
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("palworld: inspect %s: %w", marker, err)
		}
		if info.Mode().IsRegular() {
			return true, nil
		}
	}
	return false, nil
}
