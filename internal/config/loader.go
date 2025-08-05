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
	homeDir, _ := os.UserConfigDir()
	var configFilePath string = filepath.Join(homeDir, "parm", "config.toml")

	viper.SetConfigFile(configFilePath)
	setConfigDefaults()

	// TODO: figure out this bs
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
