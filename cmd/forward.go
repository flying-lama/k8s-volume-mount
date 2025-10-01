package cmd

import (
	"flag"
	"fmt"
	"k8s-volume-mount/internal"
)

// ForwardCommand handles the forward-port command execution
func ForwardCommand(args []string) error {
	// Parse command line flags
	forwardCmd := flag.NewFlagSet("forward", flag.ExitOnError)
	pvcName := forwardCmd.String("pvc", "", "Name of the PersistentVolumeClaim")
	port := forwardCmd.Int("port", 0, "Specific port for port forwarding (optional)")
	providerType := forwardCmd.String("provider", "webdav", "Provider type: webdav")
	namespace := forwardCmd.String("namespace", "", "Namespace (optional)")
	err := forwardCmd.Parse(args)
	if err != nil {
		return fmt.Errorf("error parsing command line flags: %v", err)
	}

	// Validate arguments
	if *pvcName == "" {
		return fmt.Errorf("error: PVC name must be specified")
	}

	// Check if PVC exists
	exists := internal.CheckPVCExists(*pvcName, *namespace)
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
	meta.Namespace = *namespace

	provider := internal.NewProviderFromMetadata(meta)
	if provider == nil {
		return fmt.Errorf("error: could not create provider for provider type: %s", *providerType)
	}

	// Deploy provider
	fmt.Printf("Creating %s provider for PVC %s...\n", *providerType, *pvcName)
	if err := provider.Deploy(); err != nil {
		return fmt.Errorf("error deploying provider: %v", err)
	}

	fmt.Printf("Volume %s available at port %d via %s server\n", *pvcName, meta.LocalPort, provider.Name())
	fmt.Printf("k8s-volume-mount config file: %s\n", meta.GetConfigFilePath())

	rcloneMounter := internal.NewRcloneMounter(meta)
	confPath, err := rcloneMounter.WriteRcloneConfig()
	if err != nil {
		return fmt.Errorf("error generating rclone config: %v", err)
	}

	fmt.Printf("rclone config file: %s\n", confPath)

	return nil
}
