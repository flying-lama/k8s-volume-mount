package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/utils/mount"
)

// DavFSMounter implements the Mounter interface using davfs2
type DavFSMounter struct {
	BaseMounter
	mounter mount.Interface
}

// NewDavFSMounter creates a new DavFSMounter
func NewDavFSMounter(metadata *Metadata) *DavFSMounter {
	return &DavFSMounter{
		mounter: mount.New(""),
		BaseMounter: BaseMounter{
			Metadata: metadata,
		},
	}
}

// Name returns the name of the mounter
func (m *DavFSMounter) Name() string {
	return "davfs2"
}

// Mount mounts a WebDAV volume using davfs2
func (m *DavFSMounter) Mount() (pid int, err error) {
	mountDir := m.Metadata.GetMountDir()
	host := m.Metadata.LocalHostname
	port := m.Metadata.LocalPort
	username := m.Metadata.MountUsername
	password, err := m.Metadata.GetDecodedPassword()
	if err != nil {
		err = fmt.Errorf("failed to decode password: %v", err)
		return
	}

	// Check if davfs2 is installed
	if _, err = exec.LookPath("mount.davfs"); err != nil {
		err = fmt.Errorf("davfs2 is not installed: %v", err)
		return
	}

	// Create mount directory if it doesn't exist
	if err = EnsureMountDirExists(mountDir); err != nil {
		return
	}

	// Prepare the source URL
	source := fmt.Sprintf("http://%s:%d/", host, port)

	// Prepare mount options
	uid := os.Getuid()
	gid := os.Getgid()
	options := []string{
		fmt.Sprintf("uid=%d", uid),
		fmt.Sprintf("gid=%d", gid),
	}

	// Use direct mount command with credentials in URL
	mountArgs := []string{
		"-t", "davfs",
		"-o", strings.Join(options, ","),
		source, mountDir,
	}

	mountCmd := exec.Command("mount", mountArgs...)

	stdin, err := mountCmd.StdinPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	// Start the command
	if err = mountCmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start mount command: %v", err)
	}

	// Write username and password to stdin
	_, err = fmt.Fprintf(stdin, "%s\n%s\n", username, password)
	if err != nil {
		return 0, fmt.Errorf("failed to write credentials: %v", err)
	}
	stdin.Close()

	// Wait for the command to complete
	if err = mountCmd.Wait(); err != nil {
		return 0, fmt.Errorf("mount command failed: %v", err)
	}

	// Check if the mount was successful
	if !IsMountPoint(mountDir) {
		unmountErr := m.mounter.Unmount(mountDir)
		if unmountErr != nil {
			fmt.Printf("failed to unmount davfs: %v\n", unmountErr)
		}
		err = fmt.Errorf("expected %s to be a mount point but it is not", mountDir)
		return
	}

	return
}

// Unmount unmounts a volume mounted with davfs2
func (m *DavFSMounter) Unmount() error {
	if err := m.mounter.Unmount(m.Metadata.GetMountDir()); err != nil {
		return fmt.Errorf("failed to unmount davfs: %v", err)
	}

	return nil
}
