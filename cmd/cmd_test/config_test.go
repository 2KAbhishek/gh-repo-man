package cmd_test

import (
	. "github.com/2KAbhishek/gh-repo-manager/cmd"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestLoadConfig_Default(t *testing.T) {
	cfg := LoadConfig("/tmp/does-not-exist-xyz.yml")
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
	cfg := Config{ShowReadmeInPreview: true}
	if err := enc.Encode(&cfg); err != nil {
		t.Fatal(err)
	}
	enc.Close()
	f.Close()
	loaded := LoadConfig(f.Name())
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
	cfg := Config{ShowReadmeInPreview: false}
	if err := enc.Encode(&cfg); err != nil {
		t.Fatal(err)
	}
	enc.Close()
	f.Close()
	loaded := LoadConfig(f.Name())
	if loaded.ShowReadmeInPreview != false {
		t.Errorf("expected ShowReadmeInPreview=false, got %v", loaded.ShowReadmeInPreview)
	}
}
