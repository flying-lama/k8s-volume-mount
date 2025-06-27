package cmd

import (
	"flag"
	"fmt"
	"k8s-volume-mount/internal"
)

// UnmountCommand handles the unmount command execution
func UnmountCommand(args []string) error {
	// Parse command line flags
	unmountCmd := flag.NewFlagSet("unmount", flag.ExitOnError)
	pvcName := unmountCmd.String("pvc", "", "Name of the PersistentVolumeClaim")
	err := unmountCmd.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing command line flags: %v", err)
	}

	// Validate arguments
	if *pvcName == "" {
		return fmt.Errorf("PVC name must be specified")
	}

	meta := internal.NewMetadata("", *pvcName, 0)
	if meta.ProviderType == "" {
		return fmt.Errorf("no mount information found for PVC: %s", *pvcName)
	}

	p := internal.NewProviderFromMetadata(meta)
	if p == nil {
		return fmt.Errorf("could not create provider for provider type: %s", meta.ProviderType)
	}

	// Unmount and cleanup
	fmt.Printf("Disconnecting volume %s...\n", *pvcName)
	if err := p.Cleanup(); err != nil {
		return fmt.Errorf("error during cleanup: %v", err)
	}

	return nil
}
