package internal

import (
	"fmt"
)

// NFSProvider implements the VolumeProvider interface for NFS
type NFSProvider struct {
	RcloneBaseProvider
}

// NewNFSProvider creates a new NFS provider
func NewNFSProvider(metadata *Metadata) *NFSProvider {
	return &NFSProvider{
		RcloneBaseProvider: RcloneBaseProvider{
			BaseProvider: BaseProvider{
				Metadata: metadata,
			},
			RcloneCommand: "nfs",
			RcloneArgs: []string{
				"--addr=0.0.0.0:" + fmt.Sprintf("%d", metadata.RemotePort),
				"--vfs-cache-mode=full",
			},
		},
	}
}

func (p *NFSProvider) Name() string {
	return p.Metadata.ProviderType
}

func (p *NFSProvider) GetMounter() (Mounter, error) {
	return NewNFSMounter(p.Metadata), nil
}

// Mount mounts the NFS volume to the specified directory
func (p *NFSProvider) Mount() error {
	mounterImpl, err := p.GetMounter()
	if err != nil {
		return err
	}
	p.Metadata.MountMethod = mounterImpl.Name()

	// Mount the volume
	fmt.Printf("Using %s...\n", p.Metadata.MountMethod)
	pid, err := mounterImpl.Mount()
	if err != nil {
		return err
	}

	p.Metadata.MountPid = pid

	// Save metadata
	if err := p.Metadata.Save(); err != nil {
		return fmt.Errorf("error saving metadata: %v", err)
	}

	return nil
}

// Cleanup unmounts the volume and removes all resources
func (p *NFSProvider) Cleanup() error {
	return p.CleanupResources()
}
