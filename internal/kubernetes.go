package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// ApplyManifest applies a Kubernetes manifest file
func ApplyManifest(manifestPath string) error {
	cmd := exec.Command("kubectl", "apply", "-f", manifestPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply manifest: %v\nOutput: %s", err, string(output))
	}
	return nil
}

// WaitForDeployment waits for a deployment to be ready
func WaitForDeployment(deploymentName string, timeoutSeconds int) error {
	cmd := exec.Command("kubectl", "wait", "--for=condition=Available",
		fmt.Sprintf("deployment/%s", deploymentName),
		fmt.Sprintf("--timeout=%ds", timeoutSeconds))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deployment not ready: %v\nOutput: %s", err, string(output))
	}
	return nil
}

// GetPodLogs retrieves logs from pods matching the given selector
func GetPodLogs(selector string) (string, error) {
	// First get pod names
	cmd := exec.Command("kubectl", "get", "pods", "-l", selector, "-o", "name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get pods: %v", err)
	}

	podNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(podNames) == 0 || (len(podNames) == 1 && podNames[0] == "") {
		return "", fmt.Errorf("no pods found with selector: %s", selector)
	}

	// Get logs from the first pod
	podName := strings.TrimPrefix(podNames[0], "pod/")
	cmd = exec.Command("kubectl", "logs", podName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}

	return string(output), nil
}

// StartPortForwarding starts port forwarding to a Kubernetes service
func StartPortForwarding(serviceName string, localPort int, remotePort int, logPath string) (pid int, err error) {
	// Create a command to run port forwarding
	cmd := exec.Command("kubectl", "port-forward",
		fmt.Sprintf("svc/%s", serviceName),
		fmt.Sprintf("%d:%d", localPort, remotePort))

	// Redirect output to log file
	logFile, err := os.Create(logPath)
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			fmt.Printf("warning: failed to close log file %s: %v\n", logPath, err)
		}
	}(logFile)
	if err != nil {
		return pid, fmt.Errorf("failed to create log file: %v", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Set the command to run in its own process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return pid, fmt.Errorf("failed to start port forwarding: %v", err)
	}
	pid = cmd.Process.Pid

	// Release the process so it continues running after this program exits
	if err := cmd.Process.Release(); err != nil {
		return pid, fmt.Errorf("failed to release port forwarding process: %v", err)
	}

	isReachable := CheckHostPort("127.0.0.1", localPort, 5000)
	if !isReachable {
		return pid, fmt.Errorf("port forwarding not reachable")
	}

	return pid, nil
}

// CheckPVCExists checks if a PVC exists in the cluster
func CheckPVCExists(pvcName string) bool {
	cmd := exec.Command("kubectl", "get", "pvc", pvcName)
	err := cmd.Run()
	return err == nil
}

// DeleteManifest deletes Kubernetes resources defined in a manifest file
func DeleteManifest(manifestPath string) error {
	cmd := exec.Command("kubectl", "delete", "-f", manifestPath, "--wait=false")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete resources: %v\nOutput: %s", err, string(output))
	}
	return nil
}
