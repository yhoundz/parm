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
		return fmt.Errorf("error: cannot find user config dir: \n%w", err)
	}
	cfgPath := filepath.Join(cfgDir, "parm")

	if err := os.MkdirAll(cfgPath, 0o700); err != nil {
		return fmt.Errorf("error: cannot create config dir: \n%w", err)
	}

	v := viper.GetViper()

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(cfgPath)

	err = setConfigDefaults(v)
	if err != nil {
		return err
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := v.SafeWriteConfig(); err != nil {
				return fmt.Errorf("error: cannot create config file: \n%w", err)
			}
		} else {
			return fmt.Errorf("error: Cannot read config file \n%w", err)
		}
	}

	setEnvVars(v)

	if err := v.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("error: Cannot unmarshal config file \n%w", err)
	}

	// watch for live reload ??
	return nil
}
