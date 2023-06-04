package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// NewDefaultConfig create a default configuration structure with
// the same values as the default configuration in config.yaml
func NewDefaultConfig() Config {
	return Config{
		Log: Log{
			Level: "warn",
			Path:  "logs/error.log",
		},
		Port: Port{
			Mode:   "static",
			Static: &[]uint16{9999, 9998, 9997},
			TOTP: &PortTOTP{
				Secret: "secret",
				Prefix: "999",
			},
			SkipLoopback: true,
		},
	}
}

// SetupConfig reads the configuration from the specified file
func SetupConfig(c *Config, file string) error {
	// setup config file path
	viper.SetConfigFile(file)

	// setup env config search
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// load config
	if err := viper.ReadInConfig(); err != nil {
		return errors.Errorf("failed to read configuration: %v", err)
	}

	// unmarshal config
	err := viper.Unmarshal(c)
	if err != nil {
		return errors.Errorf("failed to unmarshal configuration: %s, err: %v", file, err)
	}

	return nil
}
