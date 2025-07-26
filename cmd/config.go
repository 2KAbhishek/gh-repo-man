package cmd

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	ShowReadmeInPreview bool `yaml:"show_readme_in_preview"`
}

const DefaultConfigPath = "~/.config/gh-repo-man.yml"

func LoadConfig(path string) Config {
	cfg := Config{ShowReadmeInPreview: false}
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home + path[1:]
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return cfg
	}
	defer f.Close()
	yaml.NewDecoder(f).Decode(&cfg)
	return cfg
}
