package internal

import (
	"fmt"
	"os/exec"
)

// WebDAVProvider implements the VolumeProvider interface for WebDAV
type WebDAVProvider struct {
	RcloneBaseProvider
}

// NewWebDAVProvider creates a new WebDAV provider
func NewWebDAVProvider(metadata *Metadata) *WebDAVProvider {
	passw, err := metadata.GetDecodedPassword()
	if err != nil {
		fmt.Printf("Error decoding password: %v", err)
		return nil
	}

	return &WebDAVProvider{
		RcloneBaseProvider: RcloneBaseProvider{
			BaseProvider: BaseProvider{
				Metadata: metadata,
			},
			RcloneCommand: "webdav",
			RcloneArgs: []string{
				"--user", metadata.MountUsername,
				"--pass", passw,
			},
		},
	}
}

func (p *WebDAVProvider) Name() string {
	return p.Metadata.ProviderType
}

func (p *WebDAVProvider) GetMounter() (Mounter, error) {
	if _, err := exec.LookPath("mount.davfs"); err == nil {
		return NewDavFSMounter(p.Metadata), nil
	}

	if _, err := exec.LookPath("rclone"); err == nil {
		return NewRcloneMounter(p.Metadata), nil
	} else {
		return nil, fmt.Errorf("no WebDAV mount method available. Please install rclone: https://rclone.org/install/")
	}
}

// Mount mounts the WebDAV volume to the specified directory
func (p *WebDAVProvider) Mount() error {
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
func (p *WebDAVProvider) Cleanup() error {
	return p.CleanupResources()
}
