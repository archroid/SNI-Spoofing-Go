//go:build linux

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func desktopFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", fmt.Errorf("locate config/home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "autostart", AppName+".desktop"), nil
}

// IsEnabled returns true if the Linux XDG autostart desktop file exists.
func IsEnabled() (bool, error) {
	path, err := desktopFilePath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Enable creates the XDG autostart desktop file for SNI Spoofing.
func Enable() error {
	path, err := desktopFilePath()
	if err != nil {
		return err
	}
	exe, err := ExecutablePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create autostart dir: %w", err)
	}

	content := strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=" + AppDisplayName,
		"Comment=DPI bypass via fake ClientHello injection",
		fmt.Sprintf("Exec=\"%s\" --autostart", exe),
		"Terminal=false",
		"X-GNOME-Autostart-enabled=true",
		"Categories=Network;Proxy;",
		"",
	}, "\n")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write autostart desktop file: %w", err)
	}
	return nil
}

// Disable removes the XDG autostart desktop file for SNI Spoofing.
func Disable() error {
	path, err := desktopFilePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove autostart desktop file: %w", err)
	}
	return nil
}
