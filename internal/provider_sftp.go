package internal

import (
	"fmt"
	"os/exec"
)

// SFTPProvider implements the VolumeProvider interface for SFTP
type SFTPProvider struct {
	RcloneBaseProvider
}

// NewSFTPProvider creates a new SFTP provider
func NewSFTPProvider(metadata *Metadata) *SFTPProvider {
	passw, err := metadata.GetDecodedPassword()
	if err != nil {
		fmt.Printf("Error decoding password: %v", err)
		return nil
	}

	return &SFTPProvider{
		RcloneBaseProvider: RcloneBaseProvider{
			BaseProvider: BaseProvider{
				Metadata: metadata,
			},
			RcloneCommand: "sftp",
			RcloneArgs: []string{
				"--addr=0.0.0.0:" + fmt.Sprintf("%d", metadata.RemotePort),
				"--user=" + metadata.MountUsername,
				"--pass=" + passw,
			},
		},
	}
}

func (p *SFTPProvider) Name() string {
	return p.Metadata.ProviderType
}

func (p *SFTPProvider) GetMounter() (Mounter, error) {
	if _, err := exec.LookPath("rclone"); err == nil {
		return NewRcloneMounter(p.Metadata), nil
	} else {
		return nil, fmt.Errorf("no SFTP mount method available. Please install rclone")
	}
}

// Mount mounts the SFTP volume to the specified directory
func (p *SFTPProvider) Mount() error {
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
func (p *SFTPProvider) Cleanup() error {
	return p.CleanupResources()
}
