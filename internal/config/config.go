package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configDir = ".mdboard"
const configFile = "config.yaml"

// Config holds user preferences
type Config struct {
	GitHubUser     string   `yaml:"github_user"`
	DefaultColumns []string `yaml:"default_columns"`
}

// Default returns a sensible default config
func Default() *Config {
	return &Config{
		GitHubUser: "",
		DefaultColumns: []string{
			"Backlog",
			"In Progress",
			"Testing",
			"Done",
		},
	}
}

// Load reads config from ~/.mdboard/config.yaml, creating it if absent
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := Default()
		if err := Save(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// Save writes config to ~/.mdboard/config.yaml
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home dir: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Path returns the config file path for display
func Path() string {
	p, _ := configPath()
	return p
}
