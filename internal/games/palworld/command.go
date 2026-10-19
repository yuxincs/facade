package palworld

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Command builds an unstarted server command using the initialized installation root.
func (a *Palworld) Command(port uint16) *exec.Cmd {
	launcher := "PalServer.sh"
	if runtime.GOOS == "windows" {
		launcher = "PalServer.exe"
	}
	cmd := exec.Command(filepath.Join(a.root, launcher), fmt.Sprintf("-port=%d", port))
	cmd.Dir = a.root
	return cmd
}
