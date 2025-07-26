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
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
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
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
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
}

func TestCloneReposWithContextCancellation(t *testing.T) {
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		if command == "git" && len(args) >= 2 && args[0] == "clone" {
			return exec.Command("sleep", "30")
		}
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
	}

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

	repos := make([]cmd.Repo, 6)
	for i := range 6 {
		repos[i] = cmd.Repo{
			Name:    fmt.Sprintf("repo%d", i+1),
			HTMLURL: fmt.Sprintf("https://github.com/user/repo%d", i+1),
		}
	}

	err := cmd.CloneReposWithContext(ctx, repos)
	if err != nil {
		t.Errorf("CloneReposWithContext() returned error: %v", err)
	}
}
