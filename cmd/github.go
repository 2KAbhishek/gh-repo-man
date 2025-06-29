package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Repo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Ssh_url     string `json:"sshUrl"`
	StargazerCount int `json:"stargazerCount"`
	ForkCount      int `json:"forkCount"`
}

var ExecCommand = exec.Command

func GetRepos(user string) ([]Repo, error) {
	var cmd *exec.Cmd
	if user == "" {
		cmd = ExecCommand("gh", "repo", "list", "--limit", "1000", "--json", "name,description,sshUrl,stargazerCount,forkCount")
	} else {
		cmd = ExecCommand("gh", "repo", "list", user, "--json", "name,description,sshUrl,stargazerCount,forkCount")
	}

	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh command failed: %s", string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute gh command: %w", err)
	}

	var repos []Repo
	err = json.Unmarshal(out, &repos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gh output: %w", err)
	}

	return repos, nil
}

func CloneRepos(repos []Repo) error {
	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo.Name)
		cmd := ExecCommand("git", "clone", repo.Ssh_url)
		cmd.Stdout = nil
		cmd.Stderr = nil
		err := cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("failed to clone %s: %s", repo.Name, string(exitError.Stderr))
			}
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
		fmt.Printf("Successfully cloned %s\n", repo.Name)
	}
	return nil
}
