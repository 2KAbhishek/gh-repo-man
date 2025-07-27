package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
)

func TestProjectsDir(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	t.Run("default projects directory", func(t *testing.T) {
		config := cmd.Config{
			ProjectsDir: "~/Projects",
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDir()
		if err != nil {
			t.Fatalf("GetProjectsDir() returned error: %v", err)
		}

		expectedDir := filepath.Join(env.tmpDir, "Projects")
		if projectsDir != expectedDir {
			t.Errorf("GetProjectsDir() = %v, want %v", projectsDir, expectedDir)
		}
	})

	t.Run("custom projects directory", func(t *testing.T) {
		config := cmd.Config{
			ProjectsDir: "~/code",
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDir()
		if err != nil {
			t.Fatalf("GetProjectsDir() returned error: %v", err)
		}

		expectedDir := filepath.Join(env.tmpDir, "code")
		if projectsDir != expectedDir {
			t.Errorf("GetProjectsDir() = %v, want %v", projectsDir, expectedDir)
		}
	})

	t.Run("absolute path projects directory", func(t *testing.T) {
		absPath := filepath.Join(env.tmpDir, "workspace")
		config := cmd.Config{
			ProjectsDir: absPath,
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDir()
		if err != nil {
			t.Fatalf("GetProjectsDir() returned error: %v", err)
		}

		if projectsDir != absPath {
			t.Errorf("GetProjectsDir() = %v, want %v", projectsDir, absPath)
		}
	})
}

func TestConfigValidation(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	t.Run("valid config with projects directory", func(t *testing.T) {
		configPath := filepath.Join(env.tmpDir, "valid-projects-config.yml")
		configContent := `projects_dir: ~/Projects`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config := cmd.LoadConfig(configPath)
		if config.ProjectsDir != "~/Projects" {
			t.Errorf("Expected ProjectsDir to be ~/Projects, got %v", config.ProjectsDir)
		}
	})

	t.Run("config with invalid projects directory", func(t *testing.T) {
		configPath := filepath.Join(env.tmpDir, "invalid-projects-config.yml")
		configContent := `projects_dir: ""`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config := cmd.LoadConfig(configPath)
		if config.ProjectsDir != "~/Projects" {
			t.Errorf("Expected fallback ProjectsDir to be ~/Projects, got %v", config.ProjectsDir)
		}
	})
}
