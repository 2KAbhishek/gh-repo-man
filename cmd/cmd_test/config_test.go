package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig_Default(t *testing.T) {
	cfg := cmd.LoadConfig("/tmp/does-not-exist-xyz.yml")
	if cfg.ShowReadmeInPreview != false {
		t.Errorf("expected default ShowReadmeInPreview=false, got %v", cfg.ShowReadmeInPreview)
	}
}

func TestLoadConfig_True(t *testing.T) {
	f, err := os.CreateTemp("", "gh-repo-man-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	enc := yaml.NewEncoder(f)
	cfg := cmd.Config{ShowReadmeInPreview: true}
	if err := enc.Encode(&cfg); err != nil {
		t.Fatal(err)
	}
	enc.Close()
	f.Close()
	loaded := cmd.LoadConfig(f.Name())
	if loaded.ShowReadmeInPreview != true {
		t.Errorf("expected ShowReadmeInPreview=true, got %v", loaded.ShowReadmeInPreview)
	}
}

func TestLoadConfig_False(t *testing.T) {
	f, err := os.CreateTemp("", "gh-repo-man-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	enc := yaml.NewEncoder(f)
	cfg := cmd.Config{ShowReadmeInPreview: false}
	if err := enc.Encode(&cfg); err != nil {
		t.Fatal(err)
	}
	enc.Close()
	f.Close()
	loaded := cmd.LoadConfig(f.Name())
	if loaded.ShowReadmeInPreview != false {
		t.Errorf("expected ShowReadmeInPreview=false, got %v", loaded.ShowReadmeInPreview)
	}
}

func TestLoadConfigWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-config.yml")

	invalidYAML := "show_readme_in_preview: true\ninvalid_yaml: [\n"
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	config := cmd.LoadConfig(configPath)
	if config.ShowReadmeInPreview != false {
		t.Errorf("LoadConfig with invalid YAML should return default ShowReadmeInPreview=false, got %v", config.ShowReadmeInPreview)
	}
}

func TestLoadConfigWithTildePath(t *testing.T) {
	homeDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	configContent := "show_readme_in_preview: true\n"
	configPath := filepath.Join(homeDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := cmd.LoadConfig("~/test-config.yml")
	if config.ShowReadmeInPreview != true {
		t.Errorf("LoadConfig with tilde path should load ShowReadmeInPreview=true, got %v", config.ShowReadmeInPreview)
	}
}

func TestSetConfig(t *testing.T) {
	testConfig := cmd.Config{
		ShowReadmeInPreview: true,
	}

	cmd.SetConfig(testConfig)
}
