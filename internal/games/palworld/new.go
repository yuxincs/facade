package palworld

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// New initializes a Palworld game from its installation root.
// It returns nil, nil if dir is unrelated, or an error for invalid settings.
func New(dir string) (*Palworld, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("palworld: inspect directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("palworld: %q is not a directory", dir)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("palworld: resolve installation root: %w", err)
	}

	var config string
	for _, marker := range []string{
		"Pal/Saved/Config/LinuxServer/PalWorldSettings.ini",
		"Pal/Saved/Config/WindowsServer/PalWorldSettings.ini",
	} {
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(marker)))
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("palworld: inspect %s: %w", marker, err)
		}
		if info.Mode().IsRegular() {
			if config != "" {
				return nil, fmt.Errorf("palworld: multiple active settings files found: %s and %s; expected one platform configuration per server root", config, marker)
			}
			config = filepath.Join(dir, filepath.FromSlash(marker))
		}
	}
	if config != "" {
		options, err := readOptions(config)
		if err != nil {
			return nil, fmt.Errorf("palworld: read %s: %w", config, err)
		}
		if !strings.EqualFold(options["restapienabled"], "True") {
			return nil, fmt.Errorf("palworld: REST API is disabled or not configured; set RESTAPIEnabled=True inside OptionSettings in %s, then restart the server if it is running", config)
		}
		port := 8212
		if value, found := options["restapiport"]; found {
			port, err = strconv.Atoi(value)
			if err != nil || port < 1 || port > 65535 {
				return nil, fmt.Errorf("palworld: RESTAPIPort in %s must be an integer between 1 and 65535", config)
			}
		}
		password := options["adminpassword"]
		if strings.HasPrefix(password, "\"") {
			password, err = strconv.Unquote(password)
			if err != nil {
				return nil, fmt.Errorf("palworld: invalid quoted AdminPassword in %s; check its quotes and escape sequences", config)
			}
		}
		return &Palworld{
			root:          dir,
			restPort:      port,
			adminPassword: password,
			restClient: &http.Client{
				Timeout: 5 * time.Second,
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse
				},
			},
		}, nil
	}
	for _, launcher := range []string{"PalServer.sh", "PalServer.exe"} {
		info, err := os.Stat(filepath.Join(dir, launcher))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("palworld: inspect %s: %w", launcher, err)
		}
		if info.Mode().IsRegular() {
			return nil, fmt.Errorf("palworld: no active PalWorldSettings.ini found; start the server once to create its configuration directories, copy DefaultPalWorldSettings.ini to Pal/Saved/Config/LinuxServer/PalWorldSettings.ini (WindowsServer on Windows), and set RESTAPIEnabled=True inside OptionSettings")
		}
	}
	return nil, nil
}
