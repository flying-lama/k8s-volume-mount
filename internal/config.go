package internal

import (
	"os"
	"path/filepath"
)

// Default configuration values
const (
	// PortRangeStart LocalPort range for port forwarding
	PortRangeStart = 10000
	PortRangeEnd   = 10100

	// DefaultTempDir Default temporary directory
	DefaultTempDir = "/tmp/k8s-volume-mount"

	// DefaultMountBaseDir Default mount base directory
	DefaultMountBaseDir = "k8s-mounts"
)

// Paths that can be overridden by environment variables
var (
	// TempDir is the directory for temporary files
	TempDir = getEnvOrDefault("K8S_VOLUME_MOUNT_TEMP_DIR", DefaultTempDir)

	// MountBaseDir is the base directory for mounting volumes
	MountBaseDir = filepath.Join(os.Getenv("HOME"), getEnvOrDefault("K8S_VOLUME_MOUNT_MOUNT_DIR", DefaultMountBaseDir))
)

// getEnvOrDefault returns the value of the environment variable or the default value
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

// Initialize creates necessary directories
func Initialize() error {
	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(TempDir, 0755); err != nil {
		return err
	}

	// Create mount base directory if it doesn't exist
	if err := os.MkdirAll(MountBaseDir, 0755); err != nil {
		return err
	}

	return nil
}
