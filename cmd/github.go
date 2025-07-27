package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
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

// TopicNames extracts topic names as strings
func (r *Repo) TopicNames() []string {
	names := make([]string, len(r.Topics))
	for i, topic := range r.Topics {
		names[i] = topic.Name
	}
	return names
}

// BuildRepoMap creates a name-to-repo lookup map
func BuildRepoMap(repos []Repo) map[string]Repo {
	repoMap := make(map[string]Repo, len(repos))
	for _, repo := range repos {
		repoMap[repo.Name] = repo
	}
	return repoMap
}

// SelectReposByNames filters repositories by name using map lookup
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

const (
	JSONFields            = "name,description,url,stargazerCount,forkCount,watchers,issues,owner,createdAt,updatedAt,diskUsage,homepageUrl,isFork,isArchived,isPrivate,isTemplate,repositoryTopics,primaryLanguage"
	DefaultRepoLimit      = "1000"
	MaxUsernameLength     = 39
	MinUsernameLength     = 1
	MaxConcurrentClones   = 3
	CloneTimeoutMinutes   = 10
	DefaultContextTimeout = 5 * time.Minute
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-_]*[a-zA-Z0-9])?$`)

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
	reposCacheTTL, err := ParseTTL(config.ReposCacheTTL)
	if err != nil {
		reposCacheTTL = 24 * time.Hour
	}

	cachePath, err := getReposCachePath(user)
	if err == nil && IsCacheValid(cachePath, reposCacheTTL) {
		if repos, err := LoadReposFromCache(user); err == nil {
			return repos, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout)
	defer cancel()
	repos, err := GetReposWithContext(ctx, user)
	if err != nil {
		return nil, err
	}

	SaveReposToCache(user, repos)

	return repos, nil
}

func buildRepoListArgs(user string) []string {
	args := []string{"repo", "list", "--limit", DefaultRepoLimit, "--json", JSONFields}
	if user != "" {
		args = append(args, user)
	}
	return args
}

func getUserContext(user string) string {
	if user != "" {
		return fmt.Sprintf("user '%s'", user)
	}
	return "current user"
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
			cmd.Process.Kill()
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

// CloneRepos clones repositories with default timeout and concurrency
func CloneRepos(repos []Repo) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(CloneTimeoutMinutes)*time.Minute*time.Duration(len(repos)))
	defer cancel()
	return CloneReposWithContext(ctx, repos)
}

// CloneReposWithContext clones repositories concurrently with context support for cancellation
func CloneReposWithContext(ctx context.Context, repos []Repo) error {
	if len(repos) == 0 {
		return nil
	}

	fmt.Printf("Cloning %d repositories with up to %d concurrent operations...\n", len(repos), MaxConcurrentClones)

	sem := make(chan struct{}, MaxConcurrentClones)
	errChan := make(chan error, len(repos))
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo Repo) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				errChan <- fmt.Errorf("clone of %s cancelled: %w", repo.Name, ctx.Err())
				return
			}
			defer func() { <-sem }()

			fmt.Printf("[%d/%d] %s Cloning %s...\n", i+1, len(repos), IconCloning, repo.Name)

			targetDir, err := GetProjectsDirForUser(repo.Owner.Login)
			if err != nil {
				errChan <- fmt.Errorf("failed to get target directory: %w", err)
				return
			}

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				errChan <- fmt.Errorf("failed to create target directory: %w", err)
				return
			}

			targetPath := filepath.Join(targetDir, repo.Name)
			sshURL := ConvertToSSHURL(repo.HTMLURL)
			cmd := ExecCommand("git", "clone", sshURL, targetPath)

			err = cmd.Start()
			if err != nil {
				errChan <- fmt.Errorf("failed to start clone for %s: %w", repo.Name, err)
				return
			}

			go func() {
				<-ctx.Done()
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			}()

			err = cmd.Wait()
			if err != nil {
				if ctx.Err() != nil {
					errChan <- fmt.Errorf("clone of %s cancelled: %w", repo.Name, ctx.Err())
					return
				}
				if exitError, ok := err.(*exec.ExitError); ok {
					errChan <- fmt.Errorf("failed to clone %s: %s", repo.Name, string(exitError.Stderr))
					return
				}
				errChan <- fmt.Errorf("failed to clone %s: %w", repo.Name, err)
				return
			}

			fmt.Printf("[%d/%d] %s Successfully cloned %s to %s\n", i+1, len(repos), IconSuccess, repo.Name, targetPath)
			errChan <- nil
		}(i, repo)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var firstError error
	completedCount := 0
	for err := range errChan {
		if err != nil && firstError == nil {
			firstError = err
		}
		completedCount++
	}

	if firstError != nil {
		return firstError
	}

	fmt.Printf("%s All %d repositories cloned successfully!\n", IconDone, len(repos))
	return nil
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

// ConvertToSSHURL converts an HTTPS GitHub URL to SSH format
func ConvertToSSHURL(httpsURL string) string {
	if strings.HasPrefix(httpsURL, "https://github.com/") {
		path := strings.TrimPrefix(httpsURL, "https://github.com/")
		if !strings.HasSuffix(path, ".git") {
			path = path + ".git"
		}
		return "git@github.com:" + path
	}
	return httpsURL
}

func GetReadme(repoFullName string) (string, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository name format: %s", repoFullName)
	}
	user, repoName := parts[0], parts[1]

	readmeCacheTTL, err := ParseTTL(config.ReadmeCacheTTL)
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
				SaveReadmeToCache(user, repoName, "")
				return "", nil
			}
			return "", fmt.Errorf("gh api failed: %s", stderr)
		}
		return "", fmt.Errorf("failed to execute gh api command: %w", err)
	}

	content := string(out)
	SaveReadmeToCache(user, repoName, content)

	return content, nil
}

// FilterRepositories filters repositories based on type and language
func FilterRepositories(repos []Repo, repoType, language string) []Repo {
	if repoType == "" && language == "" {
		return repos
	}

	var filtered []Repo
	for _, repo := range repos {
		// Filter by type
		if repoType != "" {
			switch strings.ToLower(repoType) {
			case "archived":
				if !repo.IsArchived {
					continue
				}
			case "forked":
				if !repo.IsFork {
					continue
				}
			case "private":
				if !repo.IsPrivate {
					continue
				}
			case "template":
				if !repo.IsTemplate {
					continue
				}
			default:
			}
		}

		if language != "" && strings.ToLower(repo.PrimaryLanguage.Name) != strings.ToLower(language) {
			continue
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// SortRepositories sorts repositories based on the specified criteria
func SortRepositories(repos []Repo, sortBy string) []Repo {
	if sortBy == "" {
		return repos
	}

	sorted := make([]Repo, len(repos))
	copy(sorted, repos)

	switch strings.ToLower(sortBy) {
	case "created":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
		})
	case "forks":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].ForkCount > sorted[j].ForkCount
		})
	case "issues":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Issues.TotalCount > sorted[j].Issues.TotalCount
		})
	case "language":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].PrimaryLanguage.Name) < strings.ToLower(sorted[j].PrimaryLanguage.Name)
		})
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		})
	case "pushed", "updated":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
		})
	case "size":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].DiskUsage > sorted[j].DiskUsage
		})
	case "stars":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].StargazerCount > sorted[j].StargazerCount
		})
	}

	return sorted
}
