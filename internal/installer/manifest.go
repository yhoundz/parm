package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const ManifestFileName string = ".parmfile.json"

type InstallType string

const (
	Release InstallType = "binary"
	Commit  InstallType = "commit"
	Branch  InstallType = "branch"
	Source  InstallType = "source"
)

type Manifest struct {
	Owner       string      `json:"owner"`
	Repo        string      `json:"repo"`
	LastUpdated string      `json:"last_updated"`
	Executables []string    `json:"executables"`
	InstallType InstallType `json:"install_type"`
}

func (m *Manifest) WriteManifest(installDir string) error {
	path := filepath.Join(installDir, ManifestFileName)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ReadManifest(installDir string) (*Manifest, error) {
	path := filepath.Join(installDir, ManifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	err = json.Unmarshal(data, &m)
	return &m, err
}
