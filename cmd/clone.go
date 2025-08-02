package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

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

	maxConcurrent := getMaxConcurrentClones()
	fmt.Printf("Cloning %d repositories with up to %d concurrent operations...\n", len(repos), maxConcurrent)

	sem := make(chan struct{}, maxConcurrent)
	errChan := make(chan error, len(repos))
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo Repo) {
			defer wg.Done()
			cloneSingleRepo(ctx, i, repo, len(repos), sem, errChan)
		}(i, repo)
	}

	return waitForCompletion(&wg, errChan, len(repos))
}

// getMaxConcurrentClones returns maximum concurrent clone operations
func getMaxConcurrentClones() int {
	maxConcurrent := config.Performance.MaxConcurrentClones
	if maxConcurrent == 0 {
		maxConcurrent = MaxConcurrentClones
	}
	return maxConcurrent
}

// cloneSingleRepo handles cloning of a single repository
func cloneSingleRepo(ctx context.Context, index int, repo Repo, totalRepos int, sem chan struct{}, errChan chan error) {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		errChan <- fmt.Errorf("clone of %s cancelled: %w", repo.Name, ctx.Err())
		return
	}
	defer func() { <-sem }()

	targetPath, err := prepareTargetDirectory(repo, index, totalRepos)
	if err != nil {
		errChan <- err
		return
	}
	if targetPath == "" {
		return
	}

	err = executeGitClone(ctx, repo, targetPath, index, totalRepos)
	errChan <- err
}

// prepareTargetDirectory prepares the target directory for cloning
func prepareTargetDirectory(repo Repo, index, totalRepos int) (string, error) {
	targetDir, err := GetProjectsDirForUser(repo.Owner.Login)
	if err != nil {
		return "", fmt.Errorf("failed to get target directory: %w", err)
	}

	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	targetPath := filepath.Join(targetDir, repo.Name)
	if _, err := os.Stat(targetPath); err == nil {
		fmt.Printf("[%d/%d] %s %s already exists in %s, skipping clone\n", index+1, totalRepos, GetIcon("info"), repo.Name, targetPath)
		return "", nil
	}

	return targetPath, nil
}

// executeGitClone executes the actual git clone command
func executeGitClone(ctx context.Context, repo Repo, targetPath string, index, totalRepos int) error {
	sshURL := ConvertToSSHURL(repo.HTMLURL)
	args := buildGitCloneArgs(sshURL, targetPath)
	cmd := ExecCommand("git", args...)

	fmt.Printf("[%d/%d] %s Cloning %s...\n", index+1, totalRepos, GetIcon("cloning"), repo.Name)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start clone for %s: %w", repo.Name, err)
	}

	go func() {
		<-ctx.Done()
		if cmd.Process != nil {
			if killErr := cmd.Process.Kill(); killErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to kill process: %v\n", killErr)
			}
		}
	}()

	err := cmd.Wait()
	if err != nil {
		return handleCloneError(ctx, err, repo.Name)
	}

	fmt.Printf("[%d/%d] %s Successfully cloned %s to %s\n", index+1, totalRepos, GetIcon("success"), repo.Name, targetPath)
	return nil
}

// buildGitCloneArgs builds git clone command arguments
func buildGitCloneArgs(sshURL, targetPath string) []string {
	args := []string{"clone"}

	if config.Integrations.Git.CloneDepth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", config.Integrations.Git.CloneDepth))
	}

	args = append(args, config.Integrations.Git.CloneArgs...)
	args = append(args, sshURL, targetPath)
	return args
}

// handleCloneError handles clone operation errors
func handleCloneError(ctx context.Context, err error, repoName string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("clone of %s cancelled: %w", repoName, ctx.Err())
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("failed to clone %s: %s", repoName, string(exitError.Stderr))
	}
	return fmt.Errorf("failed to clone %s: %w", repoName, err)
}

// waitForCompletion waits for all clone operations to complete
func waitForCompletion(wg *sync.WaitGroup, errChan chan error, totalRepos int) error {
	go func() {
		wg.Wait()
		close(errChan)
	}()

	var firstError error
	for err := range errChan {
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	if firstError != nil {
		return firstError
	}

	return nil
}
