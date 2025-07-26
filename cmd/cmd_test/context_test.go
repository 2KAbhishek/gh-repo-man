package cmd_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
)

func TestGetReposWithContext(t *testing.T) {
	// Mock the exec command
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repos, err := cmd.GetReposWithContext(ctx, "")
	if err != nil {
		t.Errorf("GetReposWithContext() returned error: %v", err)
	}

	if len(repos) == 0 {
		t.Error("GetReposWithContext() returned no repositories")
	}
}

func TestGetReposWithContextCancellation(t *testing.T) {
	// Mock the exec command to simulate a long-running operation
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		// Create a command that will run for a while
		cmd := exec.Command("sleep", "10")
		return cmd
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := cmd.GetReposWithContext(ctx, "")
	if err == nil {
		t.Error("GetReposWithContext() should have returned error due to context cancellation")
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("GetReposWithContext() error should mention cancellation, got: %v", err)
	}
}

func TestCloneReposWithContext(t *testing.T) {
	// Mock the exec command
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	clonedRepos := make([]string, 0)
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		if command == "git" && len(args) > 2 && args[0] == "clone" {
			// Mock successful git clone
			clonedRepos = append(clonedRepos, args[2]) // URL is the 3rd argument
			return exec.Command("true")                // Command that always succeeds
		}
		return exec.Command("false") // Command that always fails for other cases
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
		{Name: "repo3", HTMLURL: "https://github.com/user/repo3"},
	}

	err := cmd.CloneReposWithContext(ctx, repos)
	if err != nil {
		t.Errorf("CloneReposWithContext() returned error: %v", err)
	}

	if len(clonedRepos) != 3 {
		t.Errorf("Expected 3 repos to be cloned, got %d", len(clonedRepos))
	}
}

func TestCloneReposWithContextCancellation(t *testing.T) {
	// Mock the exec command to simulate long-running clones
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		if command == "git" && len(args) > 2 && args[0] == "clone" {
			// Create a command that will run for a while
			return exec.Command("sleep", "10")
		}
		return exec.Command("false")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
	}

	err := cmd.CloneReposWithContext(ctx, repos)
	if err == nil {
		t.Error("CloneReposWithContext() should have returned error due to context cancellation")
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
	// Mock the exec command to track concurrency
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	activeConcurrency := make(chan int, 10)
	maxConcurrency := 0

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		if command == "git" && len(args) > 2 && args[0] == "clone" {
			// Track concurrency
			activeConcurrency <- 1
			go func() {
				time.Sleep(50 * time.Millisecond) // Simulate some work
				<-activeConcurrency
			}()
			return exec.Command("true")
		}
		return exec.Command("false")
	}

	// Monitor maximum concurrency
	go func() {
		for {
			current := len(activeConcurrency)
			if current > maxConcurrency {
				maxConcurrency = current
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create more repos than the max concurrent limit
	repos := make([]cmd.Repo, 6)
	for i := 0; i < 6; i++ {
		repos[i] = cmd.Repo{
			Name:    fmt.Sprintf("repo%d", i+1),
			HTMLURL: fmt.Sprintf("https://github.com/user/repo%d", i+1),
		}
	}

	err := cmd.CloneReposWithContext(ctx, repos)
	if err != nil {
		t.Errorf("CloneReposWithContext() returned error: %v", err)
	}

	// Give some time for concurrency monitoring
	time.Sleep(100 * time.Millisecond)

	if maxConcurrency > cmd.MaxConcurrentClones {
		t.Errorf("Expected max concurrency to be <= %d, got %d", cmd.MaxConcurrentClones, maxConcurrency)
	}
}
