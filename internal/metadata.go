package internal

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Metadata represents the structure for storing mount metadata
type Metadata struct {
	ProviderType      string `json:"providerType"`
	PVCName           string `json:"pvcName"`
	Namespace         string `json:"namespace,omitempty"`
	MountDir          string `json:"mountDir"`
	ConfigDir         string `json:"configDir"`
	LocalHostname     string `json:"localHostname"`
	LocalPort         int    `json:"localPort"`
	RemotePort        int    `json:"remotePort"`
	PortForwardingPid int    `json:"portForwardingPid,omitempty"`
	MountMethod       string `json:"mountMethod"`
	MountPid          int    `json:"mountPid,omitempty"`
	MountUsername     string `json:"mountUsername"`
	MountPassword     string `json:"mountPassword"`
	ProvisionerName   string `json:"provisionerName"`
}

// NewMetadata creates a new metadata instance for a specific provisioner
func NewMetadata(providerType string, pvcName string, port int, namespace string) *Metadata {
	username, err := GenerateRandomString(8)
	if err != nil {
		err = fmt.Errorf("failed to generate random username: %v", err)
		panic(err)
	}

	password, err := GenerateRandomString(40)
	if err != nil {
		err = fmt.Errorf("failed to generate random password: %v", err)
		panic(err)
	}

	// Encode password in base64
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

	meta := &Metadata{
		ProviderType:    providerType,
		ProvisionerName: fmt.Sprintf("%s-%s-%d", providerType, pvcName, port),
		ConfigDir:       GetConfigDir(pvcName),
		PVCName:         pvcName,
		Namespace:       namespace,
		LocalHostname:   "127.0.0.1",
		LocalPort:       port,
		RemotePort:      8090,
		MountDir:        filepath.Join(MountBaseDir, pvcName),
		MountUsername:   username,
		MountPassword:   encodedPassword,
	}

	configFilePath := meta.GetConfigFilePath()
	_ = meta.Load(configFilePath)

	return meta
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(n int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

func GetConfigDir(pvcName string) string {
	return filepath.Join(TempDir, pvcName)
}

// GetConfigFilePath returns the path to the metadata configuration file
func (m *Metadata) GetConfigFilePath() string {
	return filepath.Join(m.ConfigDir, "config.json")
}

// Save stores metadata to a JSON file at the default location
func (m *Metadata) Save() error {
	path := m.GetConfigFilePath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling metadata: %v", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error saving metadata: %v", err)
	}

	return nil
}

// Load retrieves metadata from a JSON file
func (m *Metadata) Load(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("metadata file not found: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading metadata file: %v", err)
	}

	if err := json.Unmarshal(data, m); err != nil {
		return fmt.Errorf("error unmarshaling metadata: %v", err)
	}

	return nil
}

// Delete removes the metadata file
func (m *Metadata) Delete() error {
	return os.RemoveAll(m.ConfigDir)
}

// GetLogFilePath returns the path to the log file
func (m *Metadata) GetLogFilePath() string {
	return filepath.Join(TempDir, m.ProvisionerName+".log")
}

// GetDecodedPassword returns the decoded password
func (m *Metadata) GetDecodedPassword() (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(m.MountPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode password: %v", err)
	}
	return string(decodedBytes), nil
}
