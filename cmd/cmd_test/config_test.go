package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
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
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
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
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := cmd.LoadConfig("~/test-config.yml")
	if config.UI.ShowReadmeInPreview != true {
		t.Errorf("LoadConfig with tilde path should load ShowReadmeInPreview=true, got %v", config.UI.ShowReadmeInPreview)
	}
}

func TestSetConfig(t *testing.T) {
	testConfig := cmd.Config{
		UI: cmd.UIConfig{
			ShowReadmeInPreview: true,
		},
	}

	cmd.SetConfig(testConfig)
}
