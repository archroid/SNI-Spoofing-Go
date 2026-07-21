//go:build windows

package autostart

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

const runRegistryKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// IsEnabled returns true if the Windows Registry Run key contains an entry for SNI Spoofing.
func IsEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKey, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("open registry key: %w", err)
	}
	defer k.Close()

	val, _, err := k.GetStringValue(AppName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("query registry value: %w", err)
	}
	return val != "", nil
}

// Enable adds an entry for SNI Spoofing to the Windows Registry Run key.
func Enable() error {
	exe, err := ExecutablePath()
	if err != nil {
		return err
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open registry key for write: %w", err)
	}
	defer k.Close()

	cmd := fmt.Sprintf("\"%s\" --autostart", exe)
	if err := k.SetStringValue(AppName, cmd); err != nil {
		return fmt.Errorf("set registry value: %w", err)
	}
	return nil
}

// Disable removes the entry for SNI Spoofing from the Windows Registry Run key.
func Disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runRegistryKey, registry.SET_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("open registry key for write: %w", err)
	}
	defer k.Close()

	if err := k.DeleteValue(AppName); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("delete registry value: %w", err)
	}
	return nil
}
