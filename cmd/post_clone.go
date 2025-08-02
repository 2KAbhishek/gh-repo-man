package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// HandlePostClone handles post-cloning actions (tea integration or editor fallback)
func HandlePostClone(repos []Repo) error {
	if len(repos) == 0 {
		return nil
	}

	if config.Integrations.Tea.Enabled && IsTeaAvailable() {
		if config.Integrations.Tea.AutoOpen {
			return OpenWithTea(repos)
		} else {
			fmt.Printf("🍵 Tea integration enabled but auto_open is disabled. Repositories cloned successfully.\n")
			return nil
		}
	}

	return OpenWithEditor(repos)
}

// IsTeaAvailable checks if tea command is available in PATH
func IsTeaAvailable() bool {
	cmd := ExecCommand("which", "tea")
	err := cmd.Run()
	return err == nil
}

// OpenWithTea opens repositories using tea tmux session manager
func OpenWithTea(repos []Repo) error {
	var paths []string
	for _, repo := range repos {
		targetDir, err := GetProjectsDirForUser(repo.Owner.Login)
		if err != nil {
			return fmt.Errorf("failed to get target directory for %s: %w", repo.Name, err)
		}
		repoPath := filepath.Join(targetDir, repo.Name)
		paths = append(paths, repoPath)
	}

	if len(paths) == 0 {
		return nil
	}

	if config.UI.ProgressIndicators {
		fmt.Printf("%s Opening %d repositories with tea...\n", GetIcon("tea"), len(repos))
	}

	cmd := ExecCommand("tea", paths...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// OpenWithEditor opens repositories with the configured editor
func OpenWithEditor(repos []Repo) error {
	editorCmd := config.Integrations.Editor.Command

	if editorCmd == "" {
		fmt.Println("No editor configured, skipping post-clone editor opening.")
		return nil
	}

	if config.UI.ProgressIndicators {
		fmt.Printf("%s Opening %d repositories with %s...\n", GetIcon("editor"), len(repos), editorCmd)
	}

	for _, repo := range repos {
		targetDir, err := GetProjectsDirForUser(repo.Owner.Login)
		if err != nil {
			return fmt.Errorf("failed to get target directory for %s: %w", repo.Name, err)
		}
		repoPath := filepath.Join(targetDir, repo.Name)

		if config.UI.ProgressIndicators {
			fmt.Printf("Opening %s in %s\n", repo.Name, editorCmd)
		}

		args := append(config.Integrations.Editor.Args, repoPath)
		cmd := ExecCommand(editorCmd, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Warning: Failed to open %s with %s: %v\n", repo.Name, editorCmd, err)
		}
	}

	return nil
}
