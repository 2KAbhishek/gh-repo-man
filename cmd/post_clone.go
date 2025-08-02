package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// HandlePostClone handles post-cloning actions (tea integration or editor fallback)
func HandlePostClone(repos []Repo) error {
	if len(repos) == 0 {
		return nil
	}
	var cmdConfig = config.Integrations.PostClone

	if cmdConfig.Enabled {
		return OpenWithCommand(repos, cmdConfig)
	}

	return nil
}

// OpenWithCommand opens repositories with the configured command
func OpenWithCommand(repos []Repo, cmdConfig CommandConfig) error {
	command := os.ExpandEnv(cmdConfig.Command)

	if !isCommandAvailable(command) {
		return fmt.Errorf("command %s is not available in PATH", command)
	}

	fmt.Printf("%s Opening selected repos in %s\n", GetIcon("info"), command)
	for _, repo := range repos {
		targetDir, err := GetProjectsDirForUser(repo.Owner.Login)
		if err != nil {
			return fmt.Errorf("failed to get target directory for %s: %w", repo.Name, err)
		}
		repoPath := filepath.Join(targetDir, repo.Name)

		args := append(cmdConfig.Args, repoPath)
		cmd := ExecCommand(command, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Failed to open %s with %s: %v\n", repo.Name, command, err)
		}
	}

	return nil
}

// isCommandAvailable checks if tea command is available in PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
