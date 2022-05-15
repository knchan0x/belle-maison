package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	configFileLocation_Debug      = "../../config.yaml"
	configFileLocation_Production = "./config.yaml"
)

var (
	FileNotFound = errors.New("config file not found")
)

// LoadConfig loads the configuations from config.yaml
func LoadConfig() error {

	path := ""
	if _, err := os.Stat(configFileLocation_Debug); err == nil {
		path = configFileLocation_Debug
	} else {
		path = configFileLocation_Production
	}

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return FileNotFound
		} else {
			return fmt.Errorf("fatal error config file: %s \n", err)
		}
	}

	return nil
}

// SetDebugFile sets the config file location for debug,
// default: ../../config.yaml
func SetDebugFile(path string) {
	configFileLocation_Debug = path
}

// SetDebugFile sets the config file location for production,
// default: ./config.yaml
func SetProductionFile(path string) {
	configFileLocation_Production = path
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool {
	return viper.GetBool(key)
}
