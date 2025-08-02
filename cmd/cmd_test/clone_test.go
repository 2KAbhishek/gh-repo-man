package cmd_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestCloneRepos(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	t.Run("successful cloning", func(t *testing.T) {
		reposToClone := []cmd.Repo{
			{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
			{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
		}

		err := cmd.CloneRepos(reposToClone)
		if err != nil {
			t.Errorf("CloneRepos() returned an error for successful cloning: %v", err)
		}
	})

	t.Run("failed cloning", func(t *testing.T) {
		reposToClone := []cmd.Repo{
			{Name: "fail_repo", HTMLURL: "fail_clone_url"},
		}

		err := cmd.CloneRepos(reposToClone)
		if err == nil {
			t.Error("CloneRepos() did not return an error for failed cloning")
		}
	})
}

func TestCloneReposWithContext(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
	})

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("echo", "mock clone")
	}

	ctx := context.Background()
	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
	}

	err := cmd.CloneReposWithContext(ctx, repos)
	if err != nil {
		t.Errorf("CloneReposWithContext() returned error: %v", err)
	}
}

func TestCloneReposWithContextCancellation(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
	})

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("sleep", "10")
	}

	ctx, cancel := context.WithCancel(context.Background())
	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := cmd.CloneReposWithContext(ctx, repos)
	if err == nil {
		t.Error("CloneReposWithContext() should have returned error due to context cancellation")
		return
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("CloneReposWithContext() error should mention cancellation, got: %v", err)
	}
}

func TestCloneReposEmptyList(t *testing.T) {
	ctx := context.Background()
	err := cmd.CloneReposWithContext(ctx, []cmd.Repo{})
	if err != nil {
		t.Errorf("CloneReposWithContext() with empty list should not return error, got: %v", err)
	}
}

func TestConcurrentCloning(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
	})

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		return exec.Command("echo", "mock clone")
	}

	ctx := context.Background()
	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
		{Name: "repo3", HTMLURL: "https://github.com/user/repo3"},
		{Name: "repo4", HTMLURL: "https://github.com/user/repo4"},
		{Name: "repo5", HTMLURL: "https://github.com/user/repo5"},
	}

	err := cmd.CloneReposWithContext(ctx, repos)
	if err != nil {
		t.Errorf("CloneReposWithContext() returned error: %v", err)
	}
}

func TestCloneReposWithExistingDirectories(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
	})

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	var executedCommands [][]string
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		executedCommands = append(executedCommands, append([]string{command}, args...))
		return exec.Command("echo", "mock command")
	}

	projectsDir := filepath.Join(env.tmpDir, "Projects", "user")
	err := os.MkdirAll(projectsDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create projects directory: %v", err)
	}

	existingRepoDir := filepath.Join(projectsDir, "existing-repo")
	err = os.MkdirAll(existingRepoDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create existing repo directory: %v", err)
	}

	repos := []cmd.Repo{
		{Name: "existing-repo", HTMLURL: "https://github.com/user/existing-repo", Owner: cmd.Owner{Login: "user"}},
		{Name: "new-repo", HTMLURL: "https://github.com/user/new-repo", Owner: cmd.Owner{Login: "user"}},
	}

	err = cmd.CloneRepos(repos)
	if err != nil {
		t.Errorf("CloneRepos() returned error: %v", err)
	}

	gitCloneCommands := 0
	for _, cmd := range executedCommands {
		if len(cmd) >= 2 && cmd[0] == "git" && cmd[1] == "clone" {
			gitCloneCommands++
		}
	}

	if gitCloneCommands != 1 {
		t.Errorf("Expected 1 git clone command, got %d", gitCloneCommands)
	}
}

func TestConvertToSSHURL(t *testing.T) {
	tests := []struct {
		name     string
		httpsURL string
		expected string
	}{
		{
			name:     "basic GitHub URL",
			httpsURL: "https://github.com/user/repo",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "GitHub URL with .git suffix",
			httpsURL: "https://github.com/user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "non-GitHub URL",
			httpsURL: "https://gitlab.com/user/repo",
			expected: "https://gitlab.com/user/repo",
		},
		{
			name:     "SSH URL already",
			httpsURL: "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.ConvertToSSHURL(tt.httpsURL)
			if result != tt.expected {
				t.Errorf("ConvertToSSHURL(%q) = %q, want %q", tt.httpsURL, result, tt.expected)
			}
		})
	}
}
