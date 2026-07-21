package autostart_test

import (
	"testing"

	"sni-spoofing-gui/autostart"
)

func TestExecutablePath(t *testing.T) {
	exe, err := autostart.ExecutablePath()
	if err != nil {
		t.Fatalf("ExecutablePath failed: %v", err)
	}
	if exe == "" {
		t.Fatal("ExecutablePath returned empty string")
	}
}

func TestAutostartToggle(t *testing.T) {
	initial, err := autostart.IsEnabled()
	if err != nil {
		t.Fatalf("IsEnabled failed: %v", err)
	}

	// Test enabling
	if err := autostart.Enable(); err != nil {
		t.Fatalf("Enable failed: %v", err)
	}

	enabled, err := autostart.IsEnabled()
	if err != nil {
		t.Fatalf("IsEnabled after Enable failed: %v", err)
	}
	if !enabled {
		t.Errorf("expected IsEnabled() to be true after Enable()")
	}

	// Test disabling
	if err := autostart.Disable(); err != nil {
		t.Fatalf("Disable failed: %v", err)
	}

	disabled, err := autostart.IsEnabled()
	if err != nil {
		t.Fatalf("IsEnabled after Disable failed: %v", err)
	}
	if disabled {
		t.Errorf("expected IsEnabled() to be false after Disable()")
	}

	// Restore initial state if it was enabled initially
	if initial {
		_ = autostart.Enable()
	}
}
