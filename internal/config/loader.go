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

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	cfgPath := filepath.Join(cfgDir, "parm")
	viper.AddConfigPath(cfgPath)

	setConfigDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("ERROR: cannot create config file: %w", err)
			}
		} else {
			return fmt.Errorf("ERROR: Cannot read config file %w", err)
		}
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("ERROR: Cannot unmarshal config file %w", err)
	}

	if _, err := os.Stat(cfgPath); err == nil { // file exists
		_ = viper.WriteConfig() // ignore error; file already writable
	}

	// watch for live reload ??
	return nil
}
