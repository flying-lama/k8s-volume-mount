package internal

import (
	"fmt"
	"os"
)

// Mounter defines the interface for mounting and unmounting volumes
type Mounter interface {
	// Name returns the name of the mounter
	Name() string

	// Mount mounts a remote volume to the specified directory
	Mount() (pid int, err error)

	// Unmount unmounts a volume from the specified directory
	Unmount() error
}

type BaseMounter struct {
	Metadata *Metadata
}

// EnsureMountDirExists creates the mount directory if it doesn't exist
func EnsureMountDirExists(mountDir string) error {
	// Check if directory exists
	if _, err := os.Stat(mountDir); os.IsNotExist(err) {
		// Create directory with appropriate permissions
		if err := os.MkdirAll(mountDir, 0755); err != nil {
			return fmt.Errorf("failed to create mount directory: %v", err)
		}
	}
	return nil
}
