// Package palworld provides the Palworld game adapter.
package palworld

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Adapter provides Palworld-specific discovery and settings.
type Adapter struct{}

// Detect recognizes an installation from its root, Pal, Saved, or Config directory.
// It checks known layouts and requires the REST API to be enabled in active settings.
func (Adapter) Detect(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return false, fmt.Errorf("palworld: inspect directory: %w", err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("palworld: %q is not a directory", dir)
	}

	var config string
	for _, marker := range []string{
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
			if config != "" {
				return false, fmt.Errorf("palworld: multiple active settings files found: %s and %s; run Facade from the intended Config directory", config, marker)
			}
			config = filepath.Join(dir, filepath.FromSlash(marker))
		}
	}
	if config != "" {
		options, err := readOptions(config)
		if err != nil {
			return false, fmt.Errorf("palworld: read %s: %w", config, err)
		}
		if !strings.EqualFold(options["restapienabled"], "True") {
			return false, fmt.Errorf("palworld: REST API is disabled or not configured; set RESTAPIEnabled=True inside OptionSettings in %s, then restart the server if it is running", config)
		}
		return true, nil
	}
	for _, launcher := range []string{"PalServer.sh", "PalServer.exe"} {
		info, err := os.Stat(filepath.Join(dir, launcher))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("palworld: inspect %s: %w", launcher, err)
		}
		if info.Mode().IsRegular() {
			return false, fmt.Errorf("palworld: no active PalWorldSettings.ini found; start the server once to create its configuration directories, copy DefaultPalWorldSettings.ini to Pal/Saved/Config/LinuxServer/PalWorldSettings.ini (WindowsServer on Windows), and set RESTAPIEnabled=True inside OptionSettings")
		}
	}
	return false, nil
}
