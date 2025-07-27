package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
)

func TestCloneReposWithExistingDirectories(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		ProjectsDir: "~/Projects",
		PerUserDir:  true,
	})

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	var executedCommands [][]string
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		fullCmd := append([]string{command}, args...)
		executedCommands = append(executedCommands, fullCmd)

		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	repos := []cmd.Repo{
		{
			Name:    "existing-repo",
			Owner:   cmd.Owner{Login: "testuser"},
			HTMLURL: "https://github.com/testuser/existing-repo",
		},
		{
			Name:    "new-repo",
			Owner:   cmd.Owner{Login: "testuser"},
			HTMLURL: "https://github.com/testuser/new-repo",
		},
	}

	existingRepoPath := filepath.Join(env.tmpDir, "Projects", "testuser", "existing-repo")
	err := os.MkdirAll(existingRepoPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create existing repo directory: %v", err)
	}

	err = cmd.CloneRepos(repos)
	if err != nil {
		t.Errorf("CloneRepos() returned error: %v", err)
	}

	gitCloneCommands := 0
	for _, cmd := range executedCommands {
		if len(cmd) >= 2 && cmd[0] == "git" && cmd[1] == "clone" {
			gitCloneCommands++
			if len(cmd) >= 4 && cmd[3] == existingRepoPath {
				t.Errorf("Should not have attempted to clone existing repo at %s", existingRepoPath)
			}
		}
	}

	if gitCloneCommands != 1 {
		t.Errorf("Expected 1 git clone command, got %d", gitCloneCommands)
	}
}

func TestHandlePostCloneWithExistingRepos(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		ProjectsDir:    "~/Projects",
		PerUserDir:     true,
		TeaIntegration: true,
		Editor:         "nvim",
	})

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	var executedCommands [][]string
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		fullCmd := append([]string{command}, args...)
		executedCommands = append(executedCommands, fullCmd)

		if command == "which" && len(args) == 1 && args[0] == "tea" {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "MOCK_TEA_AVAILABLE=true"}
			return cmd
		}

		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	repos := []cmd.Repo{
		{
			Name:  "existing-repo",
			Owner: cmd.Owner{Login: "testuser"},
		},
		{
			Name:  "another-repo",
			Owner: cmd.Owner{Login: "testuser"},
		},
	}

	for _, repo := range repos {
		repoPath := filepath.Join(env.tmpDir, "Projects", "testuser", repo.Name)
		err := os.MkdirAll(repoPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create repo directory %s: %v", repoPath, err)
		}
	}

	err := cmd.HandlePostClone(repos)
	if err != nil {
		t.Errorf("HandlePostClone() returned error: %v", err)
	}

	if len(executedCommands) < 2 {
		t.Errorf("Expected at least 2 commands, got %d", len(executedCommands))
		return
	}

	if executedCommands[0][0] != "which" || executedCommands[0][1] != "tea" {
		t.Errorf("Expected first command to be 'which tea', got %v", executedCommands[0])
	}

	if executedCommands[1][0] != "tea" {
		t.Errorf("Expected tea command, got %v", executedCommands[1])
	}

	expectedPaths := []string{
		filepath.Join(env.tmpDir, "Projects", "testuser", "existing-repo"),
		filepath.Join(env.tmpDir, "Projects", "testuser", "another-repo"),
	}

	for i, expectedPath := range expectedPaths {
		if len(executedCommands[1]) <= i+1 || executedCommands[1][i+1] != expectedPath {
			t.Errorf("Expected path %s at position %d, got %s", expectedPath, i+1, executedCommands[1][i+1])
		}
	}
}
