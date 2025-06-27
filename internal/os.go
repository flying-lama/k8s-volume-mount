package internal

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// IsPortListening checks if a port is listening on the specified host
func IsPortListening(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// FindFreePort finds an available port in the specified range
func FindFreePort(startPort, endPort int) (int, error) {
	for port := startPort; port <= endPort; port++ {
		if !IsPortListening("127.0.0.1", port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found in range %d-%d", startPort, endPort)
}

func IsMacOs() bool {
	isMacOS := false
	if _, err := exec.LookPath("sw_vers"); err == nil {
		isMacOS = true
	}

	return isMacOS
}

// IsMountPoint checks if a directory is a mount point
func IsMountPoint(path string) bool {
	if IsMacOs() {
		// On macOS, use mount command
		cmd := exec.Command("mount")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		// Check if the path appears in the mount output
		return strings.Contains(string(output), path)
	} else {
		// On Linux, use findmnt
		cmd := exec.Command("findmnt", "--mountpoint", path)
		err := cmd.Run()
		return err == nil
	}
}
