package internal

import (
	"fmt"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"strings"
)

// NFSMounter implements the Mounter interface using NFS
type NFSMounter struct {
	BaseMounter
	mounter mount.Interface
}

// NewNFSMounter creates a new NFSMounter
func NewNFSMounter(metadata *Metadata) *NFSMounter {
	return &NFSMounter{
		BaseMounter: BaseMounter{
			Metadata: metadata,
		},
		mounter: mount.New(""),
	}
}

// Name returns the name of the mounter
func (m *NFSMounter) Name() string {
	return "nfs"
}

// Mount mounts an NFS volume
func (m *NFSMounter) Mount() (pid int, err error) {
	mountDir := m.Metadata.MountDir
	host := m.Metadata.LocalHostname
	port := m.Metadata.LocalPort

	// Create mount directory if it doesn't exist
	if err = EnsureMountDirExists(mountDir); err != nil {
		return
	}

	// Prepare the source URL
	source := fmt.Sprintf("%s:%d:/", host, port)

	// Check if we're on macOS or Linux
	isMacOS := false
	if _, err = exec.LookPath("sw_vers"); err == nil {
		isMacOS = true
	}

	// On macOS, use the mount command directly with sudo
	if isMacOS {
		// Prepare the source URL for macOS (without port in the host part)
		macSource := fmt.Sprintf("%s:/", host)

		// macOS specific options including port
		options := []string{
			"resvport",
			"noowners",
			"nolocks",
			fmt.Sprintf("port=%d", port),
			fmt.Sprintf("mountport=%d", port),
		}

		// Execute mount command with sudo
		cmd := exec.Command("sudo", "mount", "-t", "nfs", "-o", strings.Join(options, ","), macSource, mountDir)

		// Connect to standard input to allow password prompt
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Println("Mounting NFS volume. You may be prompted for your password.")
		if err = cmd.Run(); err != nil {
			return 0, fmt.Errorf("failed to mount NFS on macOS: %v", err)
		}
	} else {
		// Linux specific options
		options := []string{
			"nolock",
			"vers=3",
			"tcp",
			"rsize=1048576",
			"wsize=1048576",
			fmt.Sprintf("port=%d", port),
			fmt.Sprintf("mountport=%d", port),
		}

		// On Linux, use kubernetes mount utils
		if err = m.mounter.Mount(source, mountDir, "nfs", options); err != nil {
			err = fmt.Errorf("failed to mount with NFS: %v", err)
			return
		}
	}

	// Check if the mount was successful using a platform-independent method
	if !IsMountPoint(mountDir) {
		// Try to unmount in case of partial mount
		if isMacOS {
			_ = exec.Command("umount", mountDir).Run()
		} else {
			_ = m.mounter.Unmount(mountDir)
		}
		err = fmt.Errorf("expected %s to be a mount point but it is not", mountDir)
		return
	}

	return 0, nil // NFS mounts don't have a PID to track
}

// Unmount unmounts a volume mounted with NFS
func (m *NFSMounter) Unmount() error {
	mountDir := m.Metadata.MountDir

	// Check if the directory exists
	if _, err := os.Stat(mountDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to unmount
	}

	// Check if it's a mount point using our platform-independent function
	if !IsMountPoint(mountDir) {
		return nil // Not a mount point, nothing to unmount
	}

	// Unmount the directory
	if IsMacOs() {
		cmd := exec.Command("sudo", "umount", mountDir)

		// Connect to standard input to allow password prompt
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Println("Unmounting NFS volume. You may be prompted for your password.")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to unmount NFS on macOS: %v", err)
		}
	} else {
		if err := m.mounter.Unmount(mountDir); err != nil {
			return fmt.Errorf("failed to unmount NFS: %v", err)
		}
	}

	return nil
}
