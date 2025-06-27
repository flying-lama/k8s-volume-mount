package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VolumeProvider defines the interface for volume providers
type VolumeProvider interface {
	// Name returns the name of the provider
	Name() string

	// Deploy creates the necessary Kubernetes resources for the provider
	Deploy() error

	// Mount mounts the volume to the specified directory
	Mount() error

	// Cleanup unmounts the volume and removes all resources
	Cleanup() error

	// GetMetadata returns the metadata associated with the provider
	GetMetadata() *Metadata

	// GetMounter identifies which mounter can be used and returns the mounter
	GetMounter() (Mounter, error)
}

func NewProviderFromMetadata(metadata *Metadata) VolumeProvider {
	switch metadata.ProviderType {
	case "webdav":
		return NewWebDAVProvider(metadata)
	case "sftp":
		return NewSFTPProvider(metadata)
	case "nfs":
		return NewNFSProvider(metadata)
	default:
		return nil
	}
}

// BaseProvider implements common functionality for all providers
type BaseProvider struct {
	Metadata *Metadata
}

func (p *BaseProvider) GetMetadata() *Metadata {
	return p.Metadata
}

// GetManifestPath returns the path to the manifest file
func (p *BaseProvider) GetManifestPath() string {
	return filepath.Join(p.Metadata.ConfigDir, "kustomize.yaml")
}

// GetLogFilePath returns the path to the log file
func (p *BaseProvider) GetLogFilePath() string {
	return filepath.Join(p.Metadata.ConfigDir, "mount.log")
}

// cleanupMount unmounts a volume using the appropriate mounter
func (p *BaseProvider) cleanupMount() {
	mountDir := p.Metadata.MountDir
	if mountDir == "" {
		return
	}

	fmt.Printf("Unmounting %s...\n", mountDir)

	// we can't access the mounter directly, but we can re-init it to have access to the interface methods
	chP := NewProviderFromMetadata(p.Metadata)
	mounterImpl, err := chP.GetMounter()
	if err != nil {
		fmt.Printf("Warning: Error identifying mounter: %v\n", err)
		return
	}

	if err := mounterImpl.Unmount(); err != nil {
		fmt.Printf("Warning: Error unmounting volume: %v\n", err)
	}
}

func (p *BaseProvider) cleanupPortForwarding() {
	pid := p.Metadata.PortForwardingPid
	if pid != 0 {
		fmt.Printf("Stopping port forwarding (PID: %d)...\n", pid)
		err := exec.Command("bash", "-c", fmt.Sprintf("kill %d 2>/dev/null || true", pid)).Run()
		if err != nil {
			fmt.Printf("Warning: Failed to stop port forwarding: %v\n", err)
		}
	}
}

func (p *BaseProvider) cleanupKubernetes() {
	provisionerName := p.Metadata.ProvisionerName
	manifestPath := p.GetManifestPath()

	fmt.Printf("Deleting %s deployment %s...\n", p.Metadata.ProviderType, provisionerName)
	if _, err := os.Stat(manifestPath); err == nil {
		err := DeleteManifest(manifestPath)
		if err != nil {
			fmt.Printf("Warning: Error deleting manifest: %v\n", err)
		}

		return
	}

	fmt.Printf("No manifest found for %s deployment %s\n", p.Metadata.ProviderType, provisionerName)
}

// CleanupResources cleans up all resources associated with a provider
func (p *BaseProvider) CleanupResources() error {

	// Unmount volume if mount directory is known
	p.cleanupMount()

	// Stop port forwarding
	p.cleanupPortForwarding()

	// Delete Kubernetes resources
	p.cleanupKubernetes()

	// Delete metadata
	err := p.Metadata.Delete()
	if err != nil {
		fmt.Printf("Warning: Error deleting metadata: %v\n", err)
	}

	// Delete Mountpoint directory
	mountDir := p.Metadata.MountDir
	if mountDir != "" {
		err := os.Remove(mountDir)
		if err != nil {
			fmt.Printf("Warning: Failed to delete mount directory: %v\n", err)
		}
	}

	fmt.Printf("Volume %s successfully unmounted and resources cleaned up\n", p.Metadata.PVCName)
	return nil
}
