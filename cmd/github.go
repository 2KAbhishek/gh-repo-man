package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Owner struct {
	Login string `json:"login"`
}

type Repo struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	HTMLURL        string    `json:"url"`
	StargazerCount int       `json:"stargazerCount"`
	ForkCount      int       `json:"forkCount"`
	WatchersCount  int       `json:"watchers"`
	Issues struct {
		TotalCount int `json:"totalCount"`
	} `json:"issues"`
	Owner          Owner     `json:"owner"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	DiskUsage      int       `json:"diskUsage"`
	HomepageURL    string    `json:"homepageUrl"`
	IsFork         bool      `json:"isFork"`
	IsArchived     bool      `json:"isArchived"`
	IsPrivate      bool      `json:"isPrivate"`
	IsTemplate     bool      `json:"isTemplate"`
	Topics         []string  `json:"repositoryTopics"`
	PrimaryLanguage struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
}

var ExecCommand = exec.Command

func GetRepos(user string) ([]Repo, error) {
	var cmd *exec.Cmd
	if user == "" {
		cmd = ExecCommand("gh", "repo", "list", "--limit", "1000", "--json", "name,description,url,stargazerCount,forkCount,watchers,issues.totalCount,owner,createdAt,updatedAt,diskUsage,homepageUrl,isFork,isArchived,isPrivate,isTemplate,repositoryTopics,primaryLanguage")
	} else {
		cmd = ExecCommand("gh", "repo", "list", user, "--json", "name,description,sshUrl,url,stargazerCount,forkCount,watchersCount,openIssuesCount,owner,createdAt,updatedAt,diskUsage,homepageUrl,isFork,isArchived,isPrivate,isTemplate,repositoryTopics,primaryLanguage")
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
		cmd := ExecCommand("git", "clone", repo.HTMLURL)
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

func GetReadme(repoFullName string) (string, error) {
	cmd := ExecCommand("gh", "api", fmt.Sprintf("repos/%s/readme", repoFullName), "-H", "Accept: application/vnd.github.v3.raw")
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			if exitError.ExitCode() == 1 && (strings.Contains(stderr, "Not Found") || strings.Contains(stderr, "404")) {
				return "", nil // README not found, return empty string and no error
			}
			return "", fmt.Errorf("gh api failed: %s", stderr)
		}
		return "", fmt.Errorf("failed to execute gh api command: %w", err)
	}
	return string(out), nil
}
