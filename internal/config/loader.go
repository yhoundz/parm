package config

import (
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
)

var Cfg Config

func Init(homeDir string) error {
	var configDir string = filepath.Join(homeDir, ".config", "parm")
	var configFilePath string = filepath.Join(configDir, "config.toml")

	// TODO: create a default toml config file and have viper read that instead
	viper.SetConfigFile(configFilePath)
	viper.SetDefault("github_api_token", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfigAs(configFilePath); err != nil {
				return fmt.Errorf("ERROR: Cannot create config file %w", err)
			}
		} else {
			return fmt.Errorf("ERROR: Cannot read config file %w", err)
		}
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("ERROR: Cannot unmarshal config file %w", err)
	}

	// watch for live reload

	return nil
}
