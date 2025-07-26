package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Owner struct {
	Login string `json:"login"`
}

type Count struct {
	TotalCount int `json:"totalCount"`
}

type Topic struct {
	Name string `json:"name"`
}

type Language struct {
	Name string `json:"name"`
}

type Repo struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"url"`
	StargazerCount  int       `json:"stargazerCount"`
	ForkCount       int       `json:"forkCount"`
	Watchers        Count     `json:"watchers"`
	Issues          Count     `json:"issues"`
	Owner           Owner     `json:"owner"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	DiskUsage       int       `json:"diskUsage"`
	HomepageURL     string    `json:"homepageUrl"`
	IsFork          bool      `json:"isFork"`
	IsArchived      bool      `json:"isArchived"`
	IsPrivate       bool      `json:"isPrivate"`
	IsTemplate      bool      `json:"isTemplate"`
	Topics          []Topic   `json:"repositoryTopics"`
	PrimaryLanguage Language  `json:"primaryLanguage"`
}

// TopicNames returns a slice of topic names from the repository topics
func (r *Repo) TopicNames() []string {
	names := make([]string, len(r.Topics))
	for i, topic := range r.Topics {
		names[i] = topic.Name
	}
	return names
}

// BuildRepoMap creates a map for efficient repository lookup by name
func BuildRepoMap(repos []Repo) map[string]Repo {
	repoMap := make(map[string]Repo, len(repos))
	for _, repo := range repos {
		repoMap[repo.Name] = repo
	}
	return repoMap
}

// SelectReposByNames efficiently selects repositories by their names using a map lookup
func SelectReposByNames(repoMap map[string]Repo, selectedNames []string) []Repo {
	var selectedRepos []Repo
	for _, name := range selectedNames {
		if name != "" {
			if repo, exists := repoMap[name]; exists {
				selectedRepos = append(selectedRepos, repo)
			}
		}
	}
	return selectedRepos
}

// Constants for GitHub API and validation
const (
	// JSON fields to request from GitHub API
	JSONFields = "name,description,url,stargazerCount,forkCount,watchers,issues,owner,createdAt,updatedAt,diskUsage,homepageUrl,isFork,isArchived,isPrivate,isTemplate,repositoryTopics,primaryLanguage"
	
	// Default limit for repository listing
	DefaultRepoLimit = "1000"
	
	// GitHub username constraints
	MaxUsernameLength = 39 // GitHub's maximum username length
	MinUsernameLength = 1  // GitHub's minimum username length
)

// Username validation regex - allows alphanumeric, hyphens, and underscores
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-_]*[a-zA-Z0-9])?$`)

var ExecCommand = exec.Command

// ValidateUsername validates that a username is safe and follows GitHub username rules
func ValidateUsername(username string) error {
	if username == "" {
		return nil // empty username is valid (means current user)
	}
	
	// Check length constraints
	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username too long: maximum %d characters allowed", MaxUsernameLength)
	}
	
	if len(username) < MinUsernameLength {
		return fmt.Errorf("username too short: minimum %d character required", MinUsernameLength)
	}
	
	// Check for shell metacharacters that could be dangerous for command injection
	if strings.ContainsAny(username, ";|&$`(){}[]<>\"'\\") {
		return fmt.Errorf("username contains invalid characters that could be unsafe")
	}
	
	// Check GitHub username format rules:
	// - Must start and end with alphanumeric character
	// - Can contain hyphens and underscores in the middle
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username format is invalid: must start and end with alphanumeric character, may contain hyphens and underscores")
	}
	
	return nil
}

func GetRepos(user string) ([]Repo, error) {
	// Validate username input
	if err := ValidateUsername(user); err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	// Build command arguments
	args := []string{"repo", "list"}
	if user != "" {
		args = append(args, user)
	} else {
		args = append(args, "--limit", DefaultRepoLimit)
	}
	args = append(args, "--json", JSONFields)
	
	cmd := ExecCommand("gh", args...)

	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			userContext := "current user"
			if user != "" {
				userContext = fmt.Sprintf("user '%s'", user)
			}
			return nil, fmt.Errorf("failed to fetch repositories for %s: %s", userContext, string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute gh repo list command: %w", err)
	}

	var repos []Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
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
