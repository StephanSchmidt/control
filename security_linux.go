//go:build linux

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// checkNotRoot ensures the program is not running as root (UID 0)
func checkNotRoot() error {
	if os.Getuid() == 0 {
		return fmt.Errorf("this program must not be run as root (UID=0) for security reasons")
	}
	return nil
}

// dropCapabilities drops all unnecessary Linux capabilities to reduce attack surface
func dropCapabilities() error {
	// Drop all bounding capabilities
	// This limits what the process can do even if it tries to gain privileges
	for i := 0; i <= unix.CAP_LAST_CAP; i++ {
		// Attempt to drop each capability from the bounding set
		// Ignore errors for capabilities we don't have
		_ = unix.Prctl(unix.PR_CAPBSET_DROP, uintptr(i), 0, 0, 0)
	}

	// Set no-new-privileges flag to prevent gaining new privileges
	// This prevents the process from gaining privileges via setuid binaries or similar
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("failed to set no-new-privs: %w", err)
	}

	return nil
}
