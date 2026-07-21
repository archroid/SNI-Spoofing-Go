//go:build darwin

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", "com.snispoofing.gui.plist"), nil
}

// IsEnabled returns true if the macOS LaunchAgent plist exists.
func IsEnabled() (bool, error) {
	path, err := plistPath()
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

// Enable creates the macOS LaunchAgent plist file.
func Enable() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	exe, err := ExecutablePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}

	content := strings.Join([]string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`,
		`<plist version="1.0">`,
		`<dict>`,
		`    <key>Label</key>`,
		`    <string>com.snispoofing.gui</string>`,
		`    <key>ProgramArguments</key>`,
		`    <array>`,
		fmt.Sprintf(`        <string>%s</string>`, exe),
		`        <string>--autostart</string>`,
		`    </array>`,
		`    <key>RunAtLoad</key>`,
		`    <true/>`,
		`</dict>`,
		`</plist>`,
		``,
	}, "\n")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write launchagent plist: %w", err)
	}
	return nil
}

// Disable removes the macOS LaunchAgent plist file.
func Disable() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove launchagent plist: %w", err)
	}
	return nil
}
