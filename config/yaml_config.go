package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// YAMLConfig mirrors the CLI struct but with YAML struct tags.
// This structure is used for parsing configuration from YAML files.
type YAMLConfig struct {
	Port             string `yaml:"metrics-port"`
	ProtectedMetrics bool   `yaml:"metrics-protected"`
	MetricsUsername  string `yaml:"metrics-username"`
	MetricsPassword  string `yaml:"metrics-password"`
	UpdateInterval   int    `yaml:"update-interval"`
	TimeZone         string `yaml:"timezone"`
	BaseURL          string `yaml:"panel-base-url"`
	ApiUsername      string `yaml:"panel-username"`
	ApiPassword      string `yaml:"panel-password"`
	Version          string `yaml:"version,omitempty"`
}

// LoadYAMLConfig loads and parses the YAML configuration file at the given path.
func LoadYAMLConfig(path string) (*YAMLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config YAMLConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ToCLI converts a YAMLConfig to a CLI structure.
// This is used to merge YAML configuration with command-line flags.
func (y *YAMLConfig) ToCLI() CLI {
	return CLI{
		Port:             y.Port,
		ProtectedMetrics: y.ProtectedMetrics,
		MetricsUsername:  y.MetricsUsername,
		MetricsPassword:  y.MetricsPassword,
		UpdateInterval:   y.UpdateInterval,
		TimeZone:         y.TimeZone,
		BaseURL:          y.BaseURL,
		ApiUsername:      y.ApiUsername,
		ApiPassword:      y.ApiPassword,
		// ConfigFile is not included as it's a CLI-specific field
		// Version field is ignored as it's handled differently in CLI
	}
}
