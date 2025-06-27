package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"k8s-volume-mount/internal"
	"os"
)

// MountCommand handles the mount command execution
func MountCommand(args []string) error {
	// Parse command line flags
	mountCmd := flag.NewFlagSet("mount", flag.ExitOnError)
	pvcName := mountCmd.String("pvc", "", "Name of the PersistentVolumeClaim")
	port := mountCmd.Int("port", 0, "Specific port for port forwarding (optional)")
	providerType := mountCmd.String("provider", "webdav", "Provider type: webdav")
	pauseOnError := mountCmd.Bool("pause-on-error", false, "Wait for user input on error before cleanup")
	err := mountCmd.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing command line flags: %v", err)
	}

	// Validate arguments
	if *pvcName == "" {
		return fmt.Errorf("error: PVC name must be specified")
	}

	// Check if PVC exists
	exists := internal.CheckPVCExists(*pvcName)
	if !exists {
		return fmt.Errorf("error: PVC %s does not exist", *pvcName)
	}

	// Determine port
	selectedPort := *port
	if selectedPort == 0 {
		selectedPort, err = internal.FindFreePort(internal.PortRangeStart, internal.PortRangeEnd)
		if err != nil {
			return fmt.Errorf("error finding free port: %v", err)
		}
	}

	meta := internal.NewMetadata(*providerType, *pvcName, selectedPort)

	provider := internal.NewProviderFromMetadata(meta)
	if provider == nil {
		return fmt.Errorf("error: could not create provider for provider type: %s", *providerType)
	}

	// Create mount directory
	mountDir := provider.GetMetadata().MountDir
	if err := os.MkdirAll(mountDir, 0755); err != nil {
		return fmt.Errorf("error creating mount directory: %v", err)
	}

	// Check if mount directory is already mounted
	if _, err := os.Stat(meta.ConfigDir); err == nil {
		return fmt.Errorf("mount directory %s is already mounted", meta.MountDir)
	}

	// Deploy provider
	fmt.Printf("Creating %s provider for PVC %s...\n", *providerType, *pvcName)
	if err := provider.Deploy(); err != nil {
		return fmt.Errorf("error deploying provider: %v", err)
	}

	// Mount volume
	fmt.Printf("Mounting volume %s to %s...\n", *pvcName, mountDir)
	if err := provider.Mount(); err != nil {
		fmt.Printf("Error mounting volume: %v\n", err)

		if *pauseOnError {
			fmt.Println("Press Enter to continue with cleanup...")
			reader := bufio.NewReader(os.Stdin)
			_, _ = reader.ReadString('\n')
		}

		fmt.Println("Cleaning up resources...")
		cleanUpErr := provider.Cleanup()
		if cleanUpErr != nil {
			fmt.Printf("Error cleaning up resources: %v\n", cleanUpErr)
		}
		return fmt.Errorf("failed to mount volume")
	}

	fmt.Printf("Volume %s successfully mounted at %s using %s\n", *pvcName, mountDir, provider.Name())

	return nil
}
