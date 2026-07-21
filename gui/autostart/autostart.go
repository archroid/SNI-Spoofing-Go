// Package autostart provides cross-platform management for starting the GUI application
// automatically when the user logs in / system boots.
package autostart

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrUnsupported = errors.New("autostart is not supported on this platform")

// AppName is the identifier used in autostart entries across operating systems.
const AppName = "sni-spoofing-gui"

// AppDisplayName is the human-readable name used in autostart files.
const AppDisplayName = "SNI Spoofing"

// ExecutablePath returns the clean, absolute path to the running executable.
func ExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate executable: %w", err)
	}
	evalExe, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = evalExe
	}
	return filepath.Clean(exe), nil
}
