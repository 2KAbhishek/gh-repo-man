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
}

var execCommand = exec.Command

func GetRepos(user string) ([]Repo, error) {
	var cmd *exec.Cmd
	if user == "" {
		cmd = execCommand("gh", "repo", "list", "--json", "name,description,sshUrl")
	} else {
		cmd = execCommand("gh", "repo", "list", user, "--json", "name,description,sshUrl")
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var repos []Repo
	err = json.Unmarshal(out, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func CloneRepos(repos []Repo) error {
	for _, repo := range repos {
		fmt.Printf("Cloning %s...\n", repo.Name)
		cmd := execCommand("git", "clone", repo.Ssh_url)
		cmd.Stdout = nil
		cmd.Stderr = nil
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to clone %s: %w", repo.Name, err)
		}
		fmt.Printf("Successfully cloned %s\n", repo.Name)
	}
	return nil
}