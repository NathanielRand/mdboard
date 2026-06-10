package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configDir = ".mdboard"
const configFile = "config.yaml"

// Config holds all mdboard settings. Stored per-project at .mdboard/config.yaml.
type Config struct {
	GitUser        string   `yaml:"git_user,omitempty"`
	DefaultColumns []string `yaml:"default_columns,omitempty"`
	Board          string   `yaml:"board,omitempty"`
}

// Default returns hardcoded defaults (no file I/O).
func Default() *Config {
	return &Config{
		DefaultColumns: []string{
			"Backlog",
			"In Progress",
			"Testing",
			"Done",
		},
	}
}

// LoadProject walks up from dir looking for .mdboard/config.yaml with a non-empty
// board field. Returns (projectDir, cfg, nil) on success, or nil cfg if not found.
func LoadProject(dir string) (string, *Config, error) {
	home, _ := os.UserHomeDir()
	cur := filepath.Clean(dir)
	for {
		path := filepath.Join(cur, configDir, configFile)
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.Board != "" {
				return cur, &cfg, nil
			}
		}
		parent := filepath.Dir(cur)
		// Stop at filesystem root or home dir to avoid the old global location
		if cur == parent || cur == home {
			break
		}
		cur = parent
	}
	return "", nil, nil
}

// LoadProjectAt loads .mdboard/config.yaml from the given dir only (no walk-up).
// Returns an empty Config if the file doesn't exist.
func LoadProjectAt(dir string) (*Config, error) {
	path := filepath.Join(dir, configDir, configFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading project config: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing project config: %w", err)
	}
	return cfg, nil
}

// SaveProject reads any existing .mdboard/config.yaml in dir, merges in non-zero
// fields from updates, and writes back. Returns an error if updates.Board conflicts
// with an already-set board value.
func SaveProject(dir string, updates *Config) error {
	path := filepath.Join(dir, configDir, configFile)

	existing, err := LoadProjectAt(dir)
	if err != nil {
		return err
	}

	// Merge non-zero fields; detect board conflict
	if updates.Board != "" {
		if existing.Board != "" && existing.Board != updates.Board {
			return fmt.Errorf("project config already has board %q — use --file to target a different board", existing.Board)
		}
		existing.Board = updates.Board
	}
	if updates.GitUser != "" {
		existing.GitUser = updates.GitUser
	}
	if len(updates.DefaultColumns) > 0 {
		existing.DefaultColumns = updates.DefaultColumns
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating project config dir: %w", err)
	}
	data, err := yaml.Marshal(existing)
	if err != nil {
		return fmt.Errorf("marshaling project config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ProjectConfigPath returns the .mdboard/config.yaml path for a given dir.
func ProjectConfigPath(dir string) string {
	return filepath.Join(dir, configDir, configFile)
}
