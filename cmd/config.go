package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ShowReadmeInPreview bool   `yaml:"show_readme_in_preview"`
	ReposCacheTTL       string `yaml:"repos_cache_ttl"`
	ReadmeCacheTTL      string `yaml:"readme_cache_ttl"`
	UsernameCacheTTL    string `yaml:"username_cache_ttl"`
	ProjectsDir         string `yaml:"projects_dir"`
}

const DefaultConfigPath = "~/.config/gh-repo-man.yml"

// LoadConfig loads configuration from the specified path with proper error handling
func LoadConfig(path string) Config {
	cfg := getDefaultConfig()

	expandedPath, err := expandPath(path)
	if err != nil {
		return cfg
	}

	f, err := os.Open(expandedPath)
	if err != nil {
		return cfg
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to parse config file %s: %v\n", expandedPath, err)
		return getDefaultConfig()
	}

	cfg = applyDefaults(cfg)

	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid config in %s: %v\n", expandedPath, err)
		return getDefaultConfig()
	}

	return cfg
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() Config {
	return Config{
		ShowReadmeInPreview: false,
		ReposCacheTTL:       "24h",
		ReadmeCacheTTL:      "24h",
		UsernameCacheTTL:    "90d",
		ProjectsDir:         "~/Projects",
	}
}

// applyDefaults fills in empty fields with default values
func applyDefaults(cfg Config) Config {
	defaults := getDefaultConfig()

	if cfg.ReposCacheTTL == "" {
		cfg.ReposCacheTTL = defaults.ReposCacheTTL
	}
	if cfg.ReadmeCacheTTL == "" {
		cfg.ReadmeCacheTTL = defaults.ReadmeCacheTTL
	}
	if cfg.UsernameCacheTTL == "" {
		cfg.UsernameCacheTTL = defaults.UsernameCacheTTL
	}
	if cfg.ProjectsDir == "" {
		cfg.ProjectsDir = defaults.ProjectsDir
	}

	return cfg
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// validateConfig validates the configuration values
func validateConfig(cfg Config) error {
	if _, err := ParseTTL(cfg.ReposCacheTTL); err != nil {
		return fmt.Errorf("invalid repos_cache_ttl: %w", err)
	}
	if _, err := ParseTTL(cfg.ReadmeCacheTTL); err != nil {
		return fmt.Errorf("invalid readme_cache_ttl: %w", err)
	}
	if _, err := ParseTTL(cfg.UsernameCacheTTL); err != nil {
		return fmt.Errorf("invalid username_cache_ttl: %w", err)
	}
	if _, err := expandPath(cfg.ProjectsDir); err != nil {
		return fmt.Errorf("invalid projects_dir: %w", err)
	}
	return nil
}

// GetProjectsDir returns the expanded projects directory path
func GetProjectsDir() (string, error) {
	return expandPath(config.ProjectsDir)
}
