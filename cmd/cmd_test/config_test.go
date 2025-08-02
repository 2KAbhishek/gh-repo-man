package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestLoadConfig_Default(t *testing.T) {
	cfg := cmd.LoadConfig("/tmp/does-not-exist-xyz.yml")
	if cfg.UI.ShowReadmeInPreview != false {
		t.Errorf("expected default ShowReadmeInPreview=false, got %v", cfg.UI.ShowReadmeInPreview)
	}
}

func TestLoadConfig_True(t *testing.T) {
	configFile := createTempConfigFile(t, cmd.Config{UI: cmd.UIConfig{ShowReadmeInPreview: true}})
	loaded := cmd.LoadConfig(configFile)
	if loaded.UI.ShowReadmeInPreview != true {
		t.Errorf("expected ShowReadmeInPreview=true, got %v", loaded.UI.ShowReadmeInPreview)
	}
}

func TestLoadConfig_False(t *testing.T) {
	configFile := createTempConfigFile(t, cmd.Config{UI: cmd.UIConfig{ShowReadmeInPreview: false}})
	loaded := cmd.LoadConfig(configFile)
	if loaded.UI.ShowReadmeInPreview != false {
		t.Errorf("expected ShowReadmeInPreview=false, got %v", loaded.UI.ShowReadmeInPreview)
	}
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yml")

	invalidYAML := "ui:\n  show_readme_in_preview: true\ninvalid_yaml: [\n"
	err := os.WriteFile(configPath, []byte(invalidYAML), 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	config := cmd.LoadConfig(configPath)
	if config.UI.ShowReadmeInPreview != false {
		t.Errorf("LoadConfig with invalid YAML should return default ShowReadmeInPreview=false, got %v", config.UI.ShowReadmeInPreview)
	}
}

func TestLoadConfigWithTildePath(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	configContent := "ui:\n  show_readme_in_preview: true\n"
	configPath := filepath.Join(env.tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := cmd.LoadConfig("~/test-config.yml")
	if config.UI.ShowReadmeInPreview != true {
		t.Errorf("LoadConfig with tilde path should load ShowReadmeInPreview=true, got %v", config.UI.ShowReadmeInPreview)
	}
}

func TestProjectsDir(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	t.Run("default projects directory without per-user dir", func(t *testing.T) {
		config := cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  false,
			},
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDirForUser("testuser")
		if err != nil {
			t.Fatalf("GetProjectsDirForUser() returned error: %v", err)
		}

		expectedDir := filepath.Join(env.tmpDir, "Projects")
		if projectsDir != expectedDir {
			t.Errorf("GetProjectsDirForUser() = %v, want %v", projectsDir, expectedDir)
		}
	})

	t.Run("custom projects directory without per-user dir", func(t *testing.T) {
		config := cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/code",
				PerUserDir:  false,
			},
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDirForUser("testuser")
		if err != nil {
			t.Fatalf("GetProjectsDirForUser() returned error: %v", err)
		}

		expectedDir := filepath.Join(env.tmpDir, "code")
		if projectsDir != expectedDir {
			t.Errorf("GetProjectsDirForUser() = %v, want %v", projectsDir, expectedDir)
		}
	})

	t.Run("absolute path projects directory", func(t *testing.T) {
		absPath := filepath.Join(env.tmpDir, "workspace")
		config := cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: absPath,
				PerUserDir:  false,
			},
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDirForUser("testuser")
		if err != nil {
			t.Fatalf("GetProjectsDirForUser() returned error: %v", err)
		}

		if projectsDir != absPath {
			t.Errorf("GetProjectsDirForUser() = %v, want %v", projectsDir, absPath)
		}
	})

	t.Run("per-user directory enabled", func(t *testing.T) {
		config := cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
		}
		cmd.SetConfig(config)

		projectsDir, err := cmd.GetProjectsDirForUser("testuser")
		if err != nil {
			t.Fatalf("GetProjectsDirForUser() returned error: %v", err)
		}

		expectedDir := filepath.Join(env.tmpDir, "Projects", "testuser")
		if projectsDir != expectedDir {
			t.Errorf("GetProjectsDirForUser() = %v, want %v", projectsDir, expectedDir)
		}
	})
}

func TestConfigValidation(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	t.Run("valid config with projects directory", func(t *testing.T) {
		configPath := filepath.Join(env.tmpDir, "valid-projects-config.yml")
		configContent := `repos:
  projects_dir: ~/Projects`
		if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config := cmd.LoadConfig(configPath)
		if config.Repos.ProjectsDir != "~/Projects" {
			t.Errorf("Expected ProjectsDir to be ~/Projects, got %s", config.Repos.ProjectsDir)
		}
	})

	t.Run("config with invalid projects directory", func(t *testing.T) {
		configPath := filepath.Join(env.tmpDir, "invalid-projects-config.yml")
		configContent := `repos:
  projects_dir: ""`
		if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		config := cmd.LoadConfig(configPath)
		if config.Repos.ProjectsDir != "~/Projects" {
			t.Errorf("Expected fallback ProjectsDir to be ~/Projects, got %s", config.Repos.ProjectsDir)
		}
	})
}
