//go:build !windows && !linux && !darwin

package autostart

// IsEnabled returns false on unsupported operating systems.
func IsEnabled() (bool, error) {
	return false, nil
}

// Enable returns ErrUnsupported on unsupported operating systems.
func Enable() error {
	return ErrUnsupported
}

// Disable returns ErrUnsupported on unsupported operating systems.
func Disable() error {
	return ErrUnsupported
}
