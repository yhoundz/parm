package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var Cfg Config

func GetParmConfigDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot find XDG_CONFIG_HOME or APPDATA: \n%w", err)
	}
	cfgPath := filepath.Join(cfgDir, "parm")
	return cfgPath, nil
}

func Init() error {
	cfgPath, err := GetParmConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfgPath, 0o700); err != nil {
		return fmt.Errorf("cannot create config dir: \n%w", err)
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
				return fmt.Errorf("cannot create config file: \n%w", err)
			}
		} else {
			return fmt.Errorf("cannot read config file \n%w", err)
		}
	}

	setEnvVars(v)

	if err := v.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("cannot unmarshal config file \n%w", err)
	}

	// watch for live reload ??
	return nil
}
