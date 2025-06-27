package internal

import (
	"fmt"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"path/filepath"
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
	mountDir := m.Metadata.MountDir
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

	// Create temporary credentials file
	credsFile := filepath.Join(TempDir, fmt.Sprintf("davfs-%s-%d.creds", filepath.Base(mountDir), port))
	credsContent := fmt.Sprintf("%s %s", username, password)
	if err = os.WriteFile(credsFile, []byte(credsContent), 0600); err != nil {
		err = fmt.Errorf("failed to write credentials file: %v", err)
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
		"conf=/dev/null",
		fmt.Sprintf("secrets=%s", credsFile),
	}

	// Use kubernetes mount utils to mount the filesystem
	if err = m.mounter.Mount(source, mountDir, "davfs", options); err != nil {
		removeErr := os.Remove(credsFile)
		if removeErr != nil {
			err = fmt.Errorf("failed to remove credentials file: %v", removeErr)
			return
		}
		err = fmt.Errorf("failed to mount with davfs2: %v", err)
		return
	}

	// Check if the mount was successful
	notMnt, err := m.mounter.IsLikelyNotMountPoint(mountDir)
	if err != nil {
		unmountErr := m.mounter.Unmount(mountDir)
		if unmountErr != nil {
			fmt.Printf("failed to unmount davfs: %v\n", unmountErr)
		}
		credsRemErr := os.Remove(credsFile)
		if credsRemErr != nil {
			fmt.Printf("failed to remove credentials file: %v\n", credsRemErr)
		}
		err = fmt.Errorf("error checking mount point: %v", err)
		return
	}
	if notMnt {
		unmountErr := m.mounter.Unmount(mountDir)
		if unmountErr != nil {
			fmt.Printf("failed to unmount davfs: %v\n", unmountErr)
		}
		credsRemErr := os.Remove(credsFile)
		if credsRemErr != nil {
			fmt.Printf("failed to remove credentials file: %v\n", credsRemErr)
		}
		err = fmt.Errorf("expected %s to be a mount point but it is not", mountDir)
		return
	}

	return
}

// Unmount unmounts a volume mounted with davfs2
func (m *DavFSMounter) Unmount() error {
	if err := m.mounter.Unmount(m.Metadata.MountDir); err != nil {
		return fmt.Errorf("failed to unmount davfs: %v", err)
	}

	// Clean up credentials file
	credsFile := filepath.Join(TempDir, fmt.Sprintf("davfs-%s.creds", filepath.Base(m.Metadata.MountDir)))
	removeErr := os.Remove(credsFile)
	if removeErr != nil {
		return fmt.Errorf("failed to remove credentials file: %v", removeErr)
	}

	return nil
}
