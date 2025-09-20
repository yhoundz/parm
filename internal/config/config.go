package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	GitHubApiTokenFallback string `mapstructure:"github_api_token_fallback"`

	// where to store the packages
	ParmPkgDirPath string `mapstructure:"parm_pkg_dir_path"`

	// directory added to PATH where symlinked binaries reside
	ParmBinPath string `mapstructure:"parm_bin_path"`
}

func setEnvVars(v *viper.Viper) {
	v.BindEnv("github_api_token", "GITHUB_TOKEN", "GH_TOKEN", "PARM_GITHUB_TOKEN")
}

func setConfigDefaults(v *viper.Viper) {
	v.SetDefault("github_api_token_fallback", "")
	v.SetDefault("parm_pkg_dir_path", getDefaultPkgDir())
}

func getDefaultPkgDir() string {
	if d, ok := os.LookupEnv("XDG_DATA_HOME"); ok && d != "" {
		return filepath.Join(d, "parm")
	}
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		// TODO: change this?
		return filepath.Join(home, "Library", "Application Support", "parm")
	}
	if runtime.GOOS == "windows" {
		if d, ok := os.LookupEnv("APPDATA"); ok && d != "" {
			return filepath.Join(d, "parm")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "parm")
}
