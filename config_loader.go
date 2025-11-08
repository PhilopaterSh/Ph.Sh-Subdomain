package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds the API keys for various services.
type Config struct {
	APIKeys struct {
		Urlscan        string `yaml:"urlscan"`
		DNSDumpster    string `yaml:"dnsdumpster"`
		VT             string `yaml:"vt"`
		SecurityTrails string `yaml:"securitytrails"`
		Shodan         string `yaml:"shodan"`
	} `yaml:"api_keys"`
}

const configFileName = "Ph.Sh_Sub_config.yaml"

// getDefaultConfigPath returns the path to the default configuration file.
func getDefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "Ph.Sh_Sub", configFileName), nil
}

// createDefaultConfigFile creates a template configuration file if it doesn't exist.
func createDefaultConfigFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		defaultConfig := Config{
			APIKeys: struct {
				Urlscan        string `yaml:"urlscan"`
				DNSDumpster    string `yaml:"dnsdumpster"`
				VT             string `yaml:"vt"`
				SecurityTrails string `yaml:"securitytrails"`
				Shodan         string `yaml:"shodan"`
			}{
				Urlscan:        "YOUR_URLSCAN_API_KEY",
				DNSDumpster:    "YOUR_DNSDUMPSTER_API_KEY",
				VT:             "YOUR_VT_API_KEY",
				SecurityTrails: "YOUR_SECURITYTRAILS_API_KEY",
				Shodan:         "YOUR_SHODAN_API_KEY",
			},
		}
		data, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		if err := ioutil.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("failed to write default config file: %w", err)
		}
		fmt.Printf("[*] Created default configuration file at: %s\n", path)
		fmt.Println("[*] Please edit this file with your API keys if needed.")
	}
	return nil
}

// loadConfig loads the configuration from the specified path.
func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}
	return &config, nil
}

// InitConfig initializes and loads the configuration.
func InitConfig() (*Config, error) {
	configPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, err
	}

	if err := createDefaultConfigFile(configPath); err != nil {
		return nil, err
	}

	return loadConfig(configPath)
}
