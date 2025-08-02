package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ExecCommand = exec.Command

// ValidateUsername ensures username is safe and follows GitHub rules
func ValidateUsername(username string) error {
	if username == "" {
		return nil
	}

	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username too long: maximum %d characters allowed", MaxUsernameLength)
	}

	if len(username) < MinUsernameLength {
		return fmt.Errorf("username too short: minimum %d character required", MinUsernameLength)
	}

	if strings.ContainsAny(username, ";|&$`(){}[]<>\"'\\") {
		return fmt.Errorf("username contains invalid characters that could be unsafe")
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username format is invalid: must start and end with alphanumeric character, may contain hyphens and underscores")
	}

	return nil
}

// GetRepos fetches repositories for a user with caching support
func GetRepos(user string) ([]Repo, error) {
	reposCacheTTL, err := ParseTTL(config.Performance.Cache.Repos)
	if err != nil {
		reposCacheTTL = 24 * time.Hour
	}

	cachePath, err := getReposCachePath(user)
	if err == nil && IsCacheValid(cachePath, reposCacheTTL) {
		if repos, err := LoadReposFromCache(user); err == nil {
			return repos, nil
		}
	}

	if config.UI.ProgressIndicators {
		if user == "" {
			fmt.Printf("%s Fetching your repositories from GitHub...\n", GetIcon("info"))
		} else {
			fmt.Printf("%s Fetching repositories for %s from GitHub...\n", GetIcon("info"), user)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout)
	defer cancel()
	repos, err := GetReposWithContext(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := SaveReposToCache(user, repos); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save repos to cache: %v\n", err)
	}

	return repos, nil
}

// GetReposWithContext fetches repositories with context support for cancellation
func GetReposWithContext(ctx context.Context, user string) ([]Repo, error) {
	if err := ValidateUsername(user); err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	cmd := ExecCommand("gh", buildRepoListArgs(user)...)

	type result struct {
		output []byte
		err    error
	}

	resultChan := make(chan result, 1)
	go func() {
		out, err := cmd.Output()
		resultChan <- result{output: out, err: err}
	}()

	select {
	case <-ctx.Done():
		if cmd.Process != nil {
			if killErr := cmd.Process.Kill(); killErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to kill process: %v\n", killErr)
			}
		}
		return nil, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case res := <-resultChan:
		if res.err != nil {
			if exitError, ok := res.err.(*exec.ExitError); ok {
				return nil, fmt.Errorf("failed to fetch repositories for %s: %s", getUserContext(user), string(exitError.Stderr))
			}
			return nil, fmt.Errorf("failed to execute gh repo list command: %w", res.err)
		}

		var repos []Repo
		if err := json.Unmarshal(res.output, &repos); err != nil {
			return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
		}

		return repos, nil
	}
}

// GetCurrentUsername fetches the current authenticated user's username
func GetCurrentUsername() (string, error) {
	cmd := ExecCommand("gh", "api", "user")
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh api user failed: %s", string(exitError.Stderr))
		}
		return "", fmt.Errorf("failed to execute gh api user command: %w", err)
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(out, &user); err != nil {
		return "", fmt.Errorf("failed to parse user API response: %w", err)
	}

	return user.Login, nil
}

// GetReadme fetches README content for a repository
func GetReadme(repoFullName string) (string, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository name format: %s", repoFullName)
	}
	user, repoName := parts[0], parts[1]

	readmeCacheTTL, err := ParseTTL(config.Performance.Cache.Readme)
	if err != nil {
		readmeCacheTTL = 24 * time.Hour
	}

	cachePath, err := getReadmeCachePath(user, repoName)
	if err == nil && IsCacheValid(cachePath, readmeCacheTTL) {
		if content, err := LoadReadmeFromCache(user, repoName); err == nil {
			return content, nil
		}
	}

	cmd := ExecCommand("gh", "api", fmt.Sprintf("repos/%s/readme", repoFullName), "-H", "Accept: application/vnd.github.v3.raw")
	out, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			if exitError.ExitCode() == 1 && (strings.Contains(stderr, "Not Found") || strings.Contains(stderr, "404")) {
				if err := SaveReadmeToCache(user, repoName, ""); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to save empty README to cache: %v\n", err)
				}
				return "", nil
			}
			return "", fmt.Errorf("gh api failed: %s", stderr)
		}
		return "", fmt.Errorf("failed to execute gh api command: %w", err)
	}

	content := string(out)
	if err := SaveReadmeToCache(user, repoName, content); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save README to cache: %v\n", err)
	}

	return content, nil
}

// ConvertToSSHURL converts an HTTPS GitHub URL to SSH format
func ConvertToSSHURL(httpsURL string) string {
	if after, ok := strings.CutPrefix(httpsURL, "https://github.com/"); ok {
		path := after
		if !strings.HasSuffix(path, ".git") {
			path = path + ".git"
		}
		return "git@github.com:" + path
	}
	return httpsURL
}

// buildRepoListArgs builds command arguments for fetching repositories
func buildRepoListArgs(user string) []string {
	repoLimit := config.Performance.RepoLimit
	if repoLimit == "" {
		repoLimit = DefaultRepoLimit
	}
	args := []string{"repo", "list", "--limit", repoLimit, "--json", JSONFields}
	if user != "" {
		args = append(args, user)
	}
	return args
}

// getUserContext returns user context string for error messages
func getUserContext(user string) string {
	if user != "" {
		return fmt.Sprintf("user '%s'", user)
	}
	return "current user"
}
