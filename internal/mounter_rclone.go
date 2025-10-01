package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// RcloneMounter implements the Mounter interface using rclone
type RcloneMounter struct {
	BaseMounter
}

// NewRcloneMounter creates a new RcloneMounter
func NewRcloneMounter(metadata *Metadata) *RcloneMounter {
	return &RcloneMounter{
		BaseMounter{
			Metadata: metadata,
		},
	}
}

// Name returns the name of the mounter
func (m *RcloneMounter) Name() string {
	return "rclone"
}

// Mount mounts a volume using rclone
func (m *RcloneMounter) Mount() (pid int, err error) {
	mountDir := m.Metadata.GetMountDir()
	providerType := m.Metadata.ProviderType

	// Check if rclone is installed
	if _, err = exec.LookPath("rclone"); err != nil {
		err = fmt.Errorf("rclone is not installed: %v", err)
		return
	}

	// Create mount directory if it doesn't exist
	if err = EnsureMountDirExists(mountDir); err != nil {
		return
	}

	configDir := m.Metadata.ConfigDir

	// Create log file path
	logFile := filepath.Join(configDir, "rclone.log")

	// Create rclone config file with obscured password
	configFile, err := m.WriteRcloneConfig()
	if err != nil {
		return
	}

	// Determine remote name based on provider type
	remoteName := providerType + ":/"

	// Execute the command directly without bash
	cmd := exec.Command("rclone", "mount", remoteName, mountDir,
		"--config", configFile,
		"--vfs-cache-mode", "writes",
		"--log-file", logFile)

	// Set the command to run in its own process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("failed to start rclone mount command: %v", err)
		return
	}

	pid = cmd.Process.Pid
	fmt.Printf("DEBUG: Rclone process started with pid: %d\n", pid)

	// Release the process so it continues running after this program exits
	if err = cmd.Process.Release(); err != nil {
		err = fmt.Errorf("failed to release rclone process: %v", err)
		return
	}

	// Give it a moment to start up
	time.Sleep(2 * time.Second)

	return
}

// Unmount unmounts a volume mounted with rclone
func (m *RcloneMounter) Unmount() error {
	if m.Metadata.MountPid != 0 {
		killCmd := exec.Command("kill", strconv.Itoa(m.Metadata.MountPid))
		killErr := killCmd.Run()
		if killErr != nil {
			fmt.Printf("Warning: Failed to kill rclone process: %v\n", killErr)
		}
	} else {
		fmt.Printf("WARNING: No pid set\n")
	}

	return nil
}

func (m *RcloneMounter) GetConfig() (content string, err error) {
	host := m.Metadata.LocalHostname
	port := m.Metadata.LocalPort
	username := m.Metadata.MountUsername
	password, err := m.Metadata.GetDecodedPassword()
	if err != nil {
		err = fmt.Errorf("failed to decode password: %v", err)
		return
	}
	providerType := m.Metadata.ProviderType

	// Obscure the password using rclone
	obscuredCmd := exec.Command("rclone", "obscure", password)
	obscuredOutput, err := obscuredCmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("failed to obscure password: %v", err)
		return
	}
	obscuredPassword := strings.TrimSpace(string(obscuredOutput))

	switch providerType {
	case "webdav":
		content = fmt.Sprintf(`[webdav]
type = webdav
url = http://%s:%d
vendor = other
user = %s
pass = %s
`, host, port, username, obscuredPassword)
	case "sftp":
		content = fmt.Sprintf(`[sftp]
type = sftp
hostname = %s
port = %d
vendor = other
user = %s
pass = %s
`, host, port, username, obscuredPassword)
	default:
		err = fmt.Errorf("unsupported provider type for rclone: %s", providerType)
		return
	}

	return
}

func (m *RcloneMounter) GetRcloneConfigFilePath() string {
	return filepath.Join(m.Metadata.ConfigDir, "rclone.conf")
}

func (m *RcloneMounter) WriteRcloneConfig() (configFile string, err error) {
	configFile = m.GetRcloneConfigFilePath()
	configContent, err := m.GetConfig()
	if err != nil {
		err = fmt.Errorf("failed to generate rclone config: %v", err)
		return
	}

	if err = os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		err = fmt.Errorf("failed to write rclone config: %v", err)
		return
	}

	return
}
