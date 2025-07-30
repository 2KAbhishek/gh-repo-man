package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type IconConfig struct {
	General   map[string]string `yaml:"general"`
	Languages map[string]string `yaml:"languages"`
}

type ReposConfig struct {
	ProjectsDir string `yaml:"projects_dir"`
	PerUserDir  bool   `yaml:"per_user_dir"`
	SortBy      string `yaml:"sort_by"`
	RepoType    string `yaml:"repo_type"`
	Language    string `yaml:"language"`
}

type UIConfig struct {
	ShowReadmeInPreview bool       `yaml:"show_readme_in_preview"`
	ColorOutput         bool       `yaml:"color_output"`
	ProgressIndicators  bool       `yaml:"progress_indicators"`
	Icons               IconConfig `yaml:"icons"`
}

type CacheConfig struct {
	Repos    string `yaml:"repos"`
	Readme   string `yaml:"readme"`
	Username string `yaml:"username"`
}

type PerformanceConfig struct {
	RepoLimit           string      `yaml:"repo_limit"`
	MaxConcurrentClones int         `yaml:"max_concurrent_clones"`
	Cache               CacheConfig `yaml:"cache"`
}

type TeaConfig struct {
	Enabled  bool `yaml:"enabled"`
	AutoOpen bool `yaml:"auto_open"`
}

type EditorConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

type GitConfig struct {
	CloneDepth int      `yaml:"clone_depth"`
	CloneArgs  []string `yaml:"clone_args"`
}

type IntegrationsConfig struct {
	Tea    TeaConfig    `yaml:"tea"`
	Editor EditorConfig `yaml:"editor"`
	Git    GitConfig    `yaml:"git"`
}

type Config struct {
	Repos        ReposConfig        `yaml:"repos"`
	UI           UIConfig           `yaml:"ui"`
	Performance  PerformanceConfig  `yaml:"performance"`
	Integrations IntegrationsConfig `yaml:"integrations"`
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

func SetConfigAndUpdateIcons(cfg Config) {
	config = cfg
}

// getDefaultConfig returns the default configuration
func getDefaultConfig() Config {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	return Config{
		Repos: ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
			SortBy:      "",
			RepoType:    "",
			Language:    "",
		},
		UI: UIConfig{
			ShowReadmeInPreview: false,
			ColorOutput:         true,
			ProgressIndicators:  true,
			Icons: IconConfig{
				General:   GeneralIcons,
				Languages: LanguageIcons,
			},
		},
		Performance: PerformanceConfig{
			RepoLimit:           "1000",
			MaxConcurrentClones: 8,
			Cache: CacheConfig{
				Repos:    "24h",
				Readme:   "24h",
				Username: "90d",
			},
		},
		Integrations: IntegrationsConfig{
			Tea: TeaConfig{
				Enabled:  true,
				AutoOpen: true,
			},
			Editor: EditorConfig{
				Command: editor,
				Args:    []string{},
			},
			Git: GitConfig{
				CloneDepth: 0,
				CloneArgs:  []string{},
			},
		},
	}
}

// applyDefaults fills in empty fields with default values
func applyDefaults(cfg Config) Config {
	defaults := getDefaultConfig()

	if cfg.Repos.ProjectsDir == "" {
		cfg.Repos.ProjectsDir = defaults.Repos.ProjectsDir
	}
	if cfg.Performance.RepoLimit == "" {
		cfg.Performance.RepoLimit = defaults.Performance.RepoLimit
	}
	if cfg.Performance.MaxConcurrentClones == 0 {
		cfg.Performance.MaxConcurrentClones = defaults.Performance.MaxConcurrentClones
	}
	if cfg.Performance.Cache.Repos == "" {
		cfg.Performance.Cache.Repos = defaults.Performance.Cache.Repos
	}
	if cfg.Performance.Cache.Readme == "" {
		cfg.Performance.Cache.Readme = defaults.Performance.Cache.Readme
	}
	if cfg.Performance.Cache.Username == "" {
		cfg.Performance.Cache.Username = defaults.Performance.Cache.Username
	}
	if cfg.Integrations.Editor.Command == "" {
		cfg.Integrations.Editor.Command = defaults.Integrations.Editor.Command
	}

	if cfg.UI.Icons.General == nil {
		cfg.UI.Icons = defaults.UI.Icons
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
	if _, err := ParseTTL(cfg.Performance.Cache.Repos); err != nil {
		return fmt.Errorf("invalid performance.cache.repos: %w", err)
	}
	if _, err := ParseTTL(cfg.Performance.Cache.Readme); err != nil {
		return fmt.Errorf("invalid performance.cache.readme: %w", err)
	}
	if _, err := ParseTTL(cfg.Performance.Cache.Username); err != nil {
		return fmt.Errorf("invalid performance.cache.username: %w", err)
	}
	if _, err := expandPath(cfg.Repos.ProjectsDir); err != nil {
		return fmt.Errorf("invalid repos.projects_dir: %w", err)
	}

	return nil
}

// GetProjectsDirForUser returns the target directory for a specific user's repositories
func GetProjectsDirForUser(username string) (string, error) {
	projectsDir, err := expandPath(config.Repos.ProjectsDir)
	if err != nil {
		return "", fmt.Errorf("failed to expand projects directory: %w", err)
	}

	if config.Repos.PerUserDir {
		return filepath.Join(projectsDir, username), nil
	}

	return projectsDir, nil
}
