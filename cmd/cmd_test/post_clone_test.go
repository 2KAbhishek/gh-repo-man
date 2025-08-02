package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func createTestReposForPostClone() []cmd.Repo {
	return []cmd.Repo{
		{
			Name:  "test-repo1",
			Owner: cmd.Owner{Login: "testuser"},
		},
		{
			Name:  "test-repo2",
			Owner: cmd.Owner{Login: "testuser"},
		},
	}
}

func TestHandlePostClone(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

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

	repos := createTestReposForPostClone()

	t.Run("post-clone command enabled and available", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
			Integrations: cmd.IntegrationsConfig{
				PostClone: cmd.CommandConfig{
					Enabled: true,
					Command: "tea",
					Args:    []string{},
				},
			},
		})

		err := cmd.HandlePostClone(repos)
		if err != nil {
			t.Errorf("HandlePostClone() returned error: %v", err)
		}

		if len(executedCommands) != 2 {
			t.Errorf("Expected 2 commands (one per repo), got %d", len(executedCommands))
			return
		}

		for i, repo := range repos {
			if executedCommands[i][0] != "tea" {
				t.Errorf("Expected tea command for repo %d, got %v", i, executedCommands[i])
			}

			expectedPath := filepath.Join(env.tmpDir, "Projects", "testuser", repo.Name)
			if len(executedCommands[i]) < 2 || executedCommands[i][1] != expectedPath {
				t.Errorf("Expected path %s for repo %d, got %v", expectedPath, i, executedCommands[i])
			}
		}
	})

	t.Run("post-clone command disabled", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
			Integrations: cmd.IntegrationsConfig{
				PostClone: cmd.CommandConfig{
					Enabled: false,
					Command: "nvim",
					Args:    []string{},
				},
			},
		})

		err := cmd.HandlePostClone(repos)
		if err != nil {
			t.Errorf("HandlePostClone() returned error: %v", err)
		}

		if len(executedCommands) != 0 {
			t.Errorf("Expected no commands when disabled, got %d", len(executedCommands))
		}
	})

	t.Run("empty repos list", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Integrations: cmd.IntegrationsConfig{
				PostClone: cmd.CommandConfig{
					Enabled: true,
					Command: "tea",
					Args:    []string{},
				},
			},
		})

		err := cmd.HandlePostClone([]cmd.Repo{})
		if err != nil {
			t.Errorf("HandlePostClone() with empty repos returned error: %v", err)
		}

		if len(executedCommands) != 0 {
			t.Errorf("Expected no commands for empty repos, got %d", len(executedCommands))
		}
	})

	t.Run("command not available", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
			Integrations: cmd.IntegrationsConfig{
				PostClone: cmd.CommandConfig{
					Enabled: true,
					Command: "nonexistent-command",
					Args:    []string{},
				},
			},
		})

		err := cmd.HandlePostClone(repos)
		if err == nil {
			t.Error("Expected error when command not available")
		}

		if len(executedCommands) != 0 {
			t.Errorf("Expected no commands when command not available, got %d", len(executedCommands))
		}
	})
}

func TestOpenWithCommand(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

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

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
	})

	repos := createTestReposForPostClone()
	cmdConfig := cmd.CommandConfig{
		Enabled: true,
		Command: "tea",
		Args:    []string{},
	}

	err := cmd.OpenWithCommand(repos, cmdConfig)
	if err != nil {
		t.Errorf("OpenWithCommand() returned error: %v", err)
	}

	if len(executedCommands) != 2 {
		t.Errorf("Expected 2 commands (one per repo), got %d", len(executedCommands))
		return
	}

	for i, repo := range repos {
		if executedCommands[i][0] != "tea" {
			t.Errorf("Expected tea command for repo %d, got %v", i, executedCommands[i])
		}

		expectedPath := filepath.Join(env.tmpDir, "Projects", "testuser", repo.Name)
		if len(executedCommands[i]) < 2 || executedCommands[i][1] != expectedPath {
			t.Errorf("Expected path %s for repo %d, got %v", expectedPath, i, executedCommands[i])
		}
	}
}

func TestHandlePostCloneWithExistingRepos(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
		Integrations: cmd.IntegrationsConfig{
			PostClone: cmd.CommandConfig{
				Enabled: true,
				Command: "tea",
				Args:    []string{},
			},
		},
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
		err := os.MkdirAll(repoPath, 0o755)
		if err != nil {
			t.Fatalf("Failed to create repo directory %s: %v", repoPath, err)
		}
	}

	err := cmd.HandlePostClone(repos)
	if err != nil {
		t.Errorf("HandlePostClone() returned error: %v", err)
	}

	if len(executedCommands) != 2 {
		t.Errorf("Expected 2 commands (one per repo), got %d", len(executedCommands))
		return
	}

	for i, repo := range repos {
		if executedCommands[i][0] != "tea" {
			t.Errorf("Expected tea command for repo %d, got %v", i, executedCommands[i])
		}

		expectedPath := filepath.Join(env.tmpDir, "Projects", "testuser", repo.Name)
		if len(executedCommands[i]) < 2 || executedCommands[i][1] != expectedPath {
			t.Errorf("Expected path %s for repo %d, got %v", expectedPath, i, executedCommands[i])
		}
	}
}
