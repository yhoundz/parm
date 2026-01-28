package manifest

import (
	"encoding/json"
	"os"
	"parm/internal/parmutil"
	"parm/pkg/sysutil"
	"path/filepath"
	"time"
)

const ManifestFileName string = ".curdfile.json"
const CurrentSchemaVersion int = 2

type InstallType string

const (
	Release    InstallType = "release"
	PreRelease InstallType = "pre-release"
)

type Manifest struct {
	SchemaVersion int         `json:"schema_version"`
	Owner         string      `json:"owner"`
	Repo          string      `json:"repo"`
	LastUpdated   string      `json:"last_updated"`
	Executables   []string    `json:"executables"`
	InstallType   InstallType `json:"install_type"`
	Version       string      `json:"version"`
	Pinned        bool        `json:"pinned"`
}

// TODO: create manifest options struct??
func New(owner, repo, version string, installType InstallType, installDir string) (*Manifest, error) {
	m := &Manifest{
		SchemaVersion: CurrentSchemaVersion,
		Owner:         owner,
		Repo:          repo,
		LastUpdated:   time.Now().UTC().Format(time.DateTime),
		Executables:   []string{},
		InstallType:   installType,
		Version:       version,
		Pinned:        false,
	}

	binM, err := getBinExecutables(installDir)
	if err != nil {
		// just return no bins
		m.Executables = nil
		return m, err
	}
	m.Executables = binM
	return m, nil
}

func (m *Manifest) GetFullExecPaths() []string {
	var res []string
	for _, path := range m.Executables {
		srcPath := parmutil.GetInstallDir(m.Owner, m.Repo)
		newPath := filepath.Join(srcPath, path)
		res = append(res, newPath)
	}
	return res
}

func (m *Manifest) Write(installDir string) error {
	path := filepath.Join(installDir, ManifestFileName)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func Read(installDir string) (*Manifest, error) {
	path := filepath.Join(installDir, ManifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	migrate(&raw)

	var m Manifest
	err = json.Unmarshal(data, &m)
	return &m, err
}

// INFO: should modify the raw param
func migrate(raw *(map[string]any)) {
	version, _ := (*raw)["schema_version"].(float64)

	if version < float64(CurrentSchemaVersion) {
		(*raw)["schema_version"] = CurrentSchemaVersion
		(*raw)["pinned"] = false
	}
}

// TODO: move this out of here? shouldn't really be here
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

		isExec, err := sysutil.IsValidBinaryExecutable(path)
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
