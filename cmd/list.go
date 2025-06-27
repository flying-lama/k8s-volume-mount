package cmd

import (
	"flag"
	"fmt"
	"k8s-volume-mount/internal"
	"os"
	"path/filepath"
	"strings"
)

// ListCommand handles the list command execution
func ListCommand(args []string) error {
	// Parse command line flags (no flags for list command currently)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	err := listCmd.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing command line flags: %v", err)
	}

	fmt.Println("Mounted Kubernetes Volumes:")
	fmt.Println("---------------------------")

	// Check if temp directory exists
	if _, err := os.Stat(internal.TempDir); os.IsNotExist(err) {
		fmt.Println("No mounted volumes found")
		return nil
	}

	// Read all mount files
	configFiles, err := filepath.Glob(filepath.Join(internal.TempDir, "**/config.json"))
	if err != nil {
		return fmt.Errorf("error reading mount information: %v", err)
	}

	if len(configFiles) == 0 {
		fmt.Println("No mounted volumes found")
		return nil
	}

	// Process each mount file
	var errors []string
	for _, mountFile := range configFiles {
		meta := internal.Metadata{}
		if err := meta.Load(mountFile); err != nil {
			errors = append(errors, fmt.Sprintf("Error loading metadata from %s: %v", mountFile, err))
			continue
		}

		mountDir := meta.MountDir

		// Display volume information
		fmt.Printf("PVC: %s\n", meta.PVCName)
		fmt.Printf("  Mount Directory: %s\n", mountDir)
		fmt.Printf("  Provider: %s\n", meta.ProviderType)
		fmt.Printf("  Mount Method: %s\n", meta.MountMethod)
		fmt.Printf("  LocalPort: %d\n", meta.LocalPort)

		// Check if volume is still accessible
		if _, err := os.Stat(mountDir); os.IsNotExist(err) {
			fmt.Printf("  Status: Mount directory not found\n")
		} else {
			// Try to read directory to check if mount is still active
			_, err := os.ReadDir(mountDir)
			if err != nil {
				fmt.Printf("  Status: Mount may be stale (error: %v)\n", err)
			} else {
				fmt.Printf("  Status: Active\n")
			}
		}
		fmt.Println("---------------------------")
	}

	if len(errors) > 0 {
		fullErr := fmt.Sprintf("Errors loading metadata: %s", strings.Join(errors, "\n"))
		return fmt.Errorf(fullErr)
	}

	return nil
}
