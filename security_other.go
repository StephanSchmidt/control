//go:build !linux

package main

// checkNotRoot is a no-op on non-Linux platforms.
func checkNotRoot() error {
	return nil
}

// dropCapabilities is a no-op on non-Linux platforms.
func dropCapabilities() error {
	return nil
}
