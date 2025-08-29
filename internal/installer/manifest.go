package installer

import (
	"encoding/json"
	"os"
	"parm/internal/utils"
	"path/filepath"
	"time"
)

const ManifestFileName string = ".parmfile.json"

type InstallType string

const (
	Release    InstallType = "release"
	Commit     InstallType = "commit"
	Branch     InstallType = "branch"
	PreRelease InstallType = "pre-release"
)

type Manifest struct {
	Owner       string      `json:"owner"`
	Repo        string      `json:"repo"`
	LastUpdated string      `json:"last_updated"`
	Executables []string    `json:"executables"`
	InstallType InstallType `json:"install_type"`
	IsSource    bool        `json:"is_source"`
	Version     string      `json:"version"`
}

// TODO: create manifest options struct??
func NewManifest(owner, repo, version string, installType InstallType, isSource bool, installDir string) (*Manifest, error) {
	m := &Manifest{
		Owner:       owner,
		Repo:        repo,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
		Executables: []string{},
		InstallType: installType,
		Version:     version,
	}

	if installType == Release || installType == PreRelease {
		binM, err := getBinExecutables(installDir)
		if err != nil {
			return nil, err
		}
		// FIX: optimize?
		m.Executables = binM
	}
	return m, nil
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

func getBinExecutables(installDir string) ([]string, error) {
	var paths []string
	scanDir := installDir
	binDir := filepath.Join(installDir, "bin")
	if info, err := os.Stat(binDir); err == nil && info.IsDir() {
		scanDir = binDir
	}

	err := filepath.WalkDir(scanDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == ManifestFileName || d.IsDir() {
			return nil
		}

		isExec, err := utils.IsBinaryExecutable(path)
		if err != nil {
			return err
		}
		if isExec {
			relPath, _ := filepath.Rel(installDir, path)
			paths = append(paths, filepath.ToSlash(relPath))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return paths, err
}
