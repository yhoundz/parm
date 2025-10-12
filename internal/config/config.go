package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Config struct {
	GitHubApiTokenFallback string `mapstructure:"github_api_token_fallback"`

	// where to store the packages
	ParmPkgPath string `mapstructure:"parm_pkg_path"`

	// directory added to PATH where symlinked binaries reside
	ParmBinPath string `mapstructure:"parm_bin_path"`
}

var defaultPkgDir = getOrCreateDefaultPkgDir()
var defaultBinDir = getOrCreateDefaultBinDir()
var DefaultCfg = &Config{
	GitHubApiTokenFallback: "",
	ParmPkgPath:            defaultPkgDir,
	ParmBinPath:            defaultBinDir,
}

func setEnvVars(v *viper.Viper) {
	v.BindEnv("github_api_token", "PARM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN")
}

func setConfigDefaults(v *viper.Viper) error {
	var cfgMap map[string]any
	if err := mapstructure.Decode(DefaultCfg, &cfgMap); err != nil {
		return err
	}

	for k, val := range cfgMap {
		v.SetDefault(k, val)
	}

	return nil
}

func GetDefaultPrefixDir() (string, error) {
	var path string
	switch runtime.GOOS {
	case "linux":
		if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok && dir != "" {
			path = filepath.Join(dir, "parm")
			return path, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, ".local", "share", "parm")
		return path, nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "parm"), nil
	case "windows":
		var err error
		if pf, err := os.UserHomeDir(); err != nil && pf != "" {
			return filepath.Join(pf, ".parm"), nil
		}
		return "", err
	default:
		return "", fmt.Errorf("error: os not supported")
	}
}

func getOrCreateDefaultPkgDir() string {
	return getOrCreatDefaultDir("pkg")
}

func getOrCreateDefaultBinDir() string {
	return getOrCreatDefaultDir("bin")
}

// TODO: return err?
func getOrCreatDefaultDir(addedPath string) string {
	prefix, err := GetDefaultPrefixDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(prefix, filepath.Clean(addedPath))

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(path, 0o700); mkErr != nil {
				return ""
			}
			return path
		}
		return ""
	}
	if !fi.IsDir() {
		return ""
	}
	return path
}
