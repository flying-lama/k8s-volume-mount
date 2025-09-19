package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"text/template"
)

//go:embed templates/rclone_deployment.yml.tmpl
var rcloneDeploymentTemplate string

// RcloneBaseProvider implements common functionality for Rclone-based providers
type RcloneBaseProvider struct {
	BaseProvider
	RcloneCommand string
	RcloneArgs    []string
}

// Deploy creates the necessary Kubernetes resources for an Rclone-based provider
func (p *RcloneBaseProvider) Deploy() error {
	pvcName := p.Metadata.PVCName
	namespace := p.Metadata.Namespace
	port := p.Metadata.LocalPort
	provisionerName := p.Metadata.ProvisionerName
	username := p.Metadata.MountUsername
	password, err := p.Metadata.GetDecodedPassword()
	if err != nil {
		return fmt.Errorf("error decoding password: %v", err)
	}
	manifestPath := p.GetManifestPath()
	logPath := p.GetLogFilePath()

	// Build command and args for the container
	commandArgs := append([]string{"rclone", "serve", p.RcloneCommand}, p.RcloneArgs...)
	commandArgs = append(commandArgs, "/data", "--addr", fmt.Sprintf(":%d", p.Metadata.RemotePort))

	// Add auth parameters if supported by this provider type
	if p.RcloneCommand == "webdav" || p.RcloneCommand == "http" {
		commandArgs = append(commandArgs, "--user", username, "--pass", password)
	}

	// Create manifest from template
	tmplData := struct {
		ProvisionerName string
		Command         string
		ContainerPort   int
		PVCName         string
		Namespace       string
		RemotePort      int
	}{
		ProvisionerName: provisionerName,
		Command:         formatStringArray(commandArgs),
		ContainerPort:   p.Metadata.RemotePort,
		PVCName:         pvcName,
		Namespace:       namespace,
	}

	// Parse embedded template
	tmpl, err := template.New("rclone_deployment").Parse(rcloneDeploymentTemplate)
	if err != nil {
		return fmt.Errorf("error parsing embedded template: %v", err)
	}

	// Execute template
	var manifestBuf bytes.Buffer
	if err := tmpl.Execute(&manifestBuf, tmplData); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	// Write manifest to file
	if err := os.MkdirAll(p.Metadata.ConfigDir, 0755); err != nil {
		return fmt.Errorf("error creating temp directory: %v", err)
	}
	if err := os.WriteFile(manifestPath, manifestBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing manifest file: %v", err)
	}

	// Apply manifest
	if err := ApplyManifest(manifestPath); err != nil {
		return fmt.Errorf("error applying manifest: %v", err)
	}

	// Wait for deployment to be ready
	fmt.Printf("Waiting for %s server for %s to be ready...\n", p.RcloneCommand, pvcName)
	if err := WaitForDeployment(provisionerName, namespace, 60); err != nil {
		fmt.Printf("Warning: Timeout waiting for %s server: %v\n", p.RcloneCommand, err)

		// Show logs for debugging
		logs, logErr := GetPodLogs(fmt.Sprintf("app=%s", provisionerName), p.Metadata.Namespace)
		if logErr == nil {
			fmt.Printf("Pod logs:\n%s\n", logs)
		}
		fmt.Println("Attempting to continue anyway...")
	}

	// Start port forwarding
	fmt.Printf("Starting port forwarding on port %d...\n", port)
	pid, err := StartPortForwarding(provisionerName, p.Metadata.Namespace, port, p.Metadata.RemotePort, logPath)
	if err != nil {
		return fmt.Errorf("error starting port forwarding: %v", err)
	}
	p.Metadata.PortForwardingPid = pid
	err = p.Metadata.Save()
	if err != nil {
		return fmt.Errorf("error saving metadata: %v", err)
	}

	// Check if port is reachable
	if CheckHostPort("127.0.0.1", port, 2000) == false {
		fmt.Printf("Warning: LocalPort %d does not seem to be reachable\n", port)
		fmt.Println("Attempting to continue anyway...")
	} else {
		fmt.Printf("LocalPort %d is reachable.\n", port)
	}

	return nil
}

// Helper function to format a string array for the manifest
func formatStringArray(arr []string) string {
	result := "["
	for i, s := range arr {
		result += fmt.Sprintf("%q", s)
		if i < len(arr)-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}
