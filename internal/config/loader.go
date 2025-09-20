package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var Cfg Config

func Init() error {
	// will have to be figured out by the install script
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("ERROR: cannot find user config dir: %w", err)
	}
	cfgPath := filepath.Join(cfgDir, "parm")

	if err := os.MkdirAll(cfgPath, 0o700); err != nil {
		return fmt.Errorf("ERROR: cannot create config dir: %w", err)
	}

	v := viper.GetViper()

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(cfgPath)

	setConfigDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := v.SafeWriteConfig(); err != nil {
				return fmt.Errorf("error: cannot create config file: %w", err)
			}
		} else {
			return fmt.Errorf("error: Cannot read config file %w", err)
		}
	}

	setEnvVars(v)

	if err := v.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("error: Cannot unmarshal config file %w", err)
	}

	// watch for live reload ??
	return nil
}
