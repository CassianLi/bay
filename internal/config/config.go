package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Roots          []string `yaml:"roots"`
	Ignore         []string `yaml:"ignore"`
	StaleAfterDays int      `yaml:"stale_after_days"`
	PMDir          string   `yaml:"pm_dir"`
}

func Default() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	return Config{
		Roots: []string{
			filepath.Join(home, "CursorProjects"),
			filepath.Join(home, "DeepseekProjects"),
		},
		Ignore:         nil,
		StaleAfterDays: 90,
		PMDir:          "_pm",
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), err
		}
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadOrDefault(path string) (Config, error) {
	cfg, err := Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bay", "config.yaml"), nil
}

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "bay", "ledger.db"), nil
}
