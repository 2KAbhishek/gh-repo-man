package cmd_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
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

		if command == "which" && len(args) == 1 && args[0] == "tea" {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "MOCK_TEA_AVAILABLE=true"}
			return cmd
		}

		if command == "tea" {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
			return cmd
		}

		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	repos := createTestReposForPostClone()

	t.Run("tea integration enabled and available", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
			Integrations: cmd.IntegrationsConfig{
				Tea: cmd.TeaConfig{
					Enabled:  true,
					AutoOpen: true,
				},
				Editor: cmd.EditorConfig{
					Command: "nvim",
				},
			},
		})

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
			filepath.Join(env.tmpDir, "Projects", "testuser", "test-repo1"),
			filepath.Join(env.tmpDir, "Projects", "testuser", "test-repo2"),
		}

		for i, expectedPath := range expectedPaths {
			if len(executedCommands[1]) <= i+1 || executedCommands[1][i+1] != expectedPath {
				t.Errorf("Expected path %s at position %d, got %s", expectedPath, i+1, executedCommands[1][i+1])
			}
		}
	})

	t.Run("tea integration disabled, fallback to editor", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
			Integrations: cmd.IntegrationsConfig{
				Tea: cmd.TeaConfig{
					Enabled:  false,
					AutoOpen: false,
				},
				Editor: cmd.EditorConfig{
					Command: "nvim",
				},
			},
		})

		err := cmd.HandlePostClone(repos)
		if err != nil {
			t.Errorf("HandlePostClone() returned error: %v", err)
		}

		if len(executedCommands) != 2 {
			t.Errorf("Expected 2 editor commands, got %d", len(executedCommands))
			return
		}

		for i, repo := range repos {
			if executedCommands[i][0] != "nvim" {
				t.Errorf("Expected nvim command, got %v", executedCommands[i])
			}

			expectedPath := filepath.Join(env.tmpDir, "Projects", "testuser", repo.Name)
			if executedCommands[i][1] != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, executedCommands[i][1])
			}
		}
	})

	t.Run("empty repos list", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Integrations: cmd.IntegrationsConfig{
				Tea: cmd.TeaConfig{
					Enabled:  true,
					AutoOpen: true,
				},
				Editor: cmd.EditorConfig{
					Command: "nvim",
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

	t.Run("no editor configured", func(t *testing.T) {
		executedCommands = nil
		cmd.SetConfig(cmd.Config{
			Repos: cmd.ReposConfig{
				ProjectsDir: "~/Projects",
				PerUserDir:  true,
			},
			Integrations: cmd.IntegrationsConfig{
				Tea: cmd.TeaConfig{
					Enabled:  false,
					AutoOpen: false,
				},
				Editor: cmd.EditorConfig{
					Command: "",
				},
			},
		})

		err := cmd.HandlePostClone(repos)
		if err != nil {
			t.Errorf("HandlePostClone() with no editor returned error: %v", err)
		}

		if len(executedCommands) != 0 {
			t.Errorf("Expected no commands when no editor configured, got %d", len(executedCommands))
		}
	})
}

func TestIsTeaAvailable(t *testing.T) {
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	t.Run("tea is available", func(t *testing.T) {
		cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "MOCK_TEA_AVAILABLE=true"}
			return cmd
		}

		available := cmd.IsTeaAvailable()
		if !available {
			t.Error("Expected tea to be available")
		}
	})

	t.Run("tea is not available", func(t *testing.T) {
		cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "MOCK_TEA_AVAILABLE=false"}
			return cmd
		}

		available := cmd.IsTeaAvailable()
		if available {
			t.Error("Expected tea to not be available")
		}
	})
}

func TestOpenWithTea(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	var executedCommand []string
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		executedCommand = append([]string{command}, args...)
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
	err := cmd.OpenWithTea(repos)
	if err != nil {
		t.Errorf("OpenWithTea() returned error: %v", err)
	}

	if executedCommand[0] != "tea" {
		t.Errorf("Expected 'tea', got %v", executedCommand[0])
	}

	expectedPaths := []string{
		filepath.Join(env.tmpDir, "Projects", "testuser", "test-repo1"),
		filepath.Join(env.tmpDir, "Projects", "testuser", "test-repo2"),
	}

	for i, expectedPath := range expectedPaths {
		if len(executedCommand) <= i+1 || executedCommand[i+1] != expectedPath {
			t.Errorf("Expected path %s at position %d, got %s", expectedPath, i+1, executedCommand[i+1])
		}
	}
}

func TestOpenWithEditor(t *testing.T) {
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
		Integrations: cmd.IntegrationsConfig{
			Editor: cmd.EditorConfig{
				Command: "nvim",
			},
		},
	})

	repos := createTestReposForPostClone()
	err := cmd.OpenWithEditor(repos)
	if err != nil {
		t.Errorf("OpenWithEditor() returned error: %v", err)
	}

	if len(executedCommands) != 2 {
		t.Errorf("Expected 2 editor commands, got %d", len(executedCommands))
		return
	}

	for i, repo := range repos {
		if executedCommands[i][0] != "nvim" {
			t.Errorf("Expected nvim command, got %v", executedCommands[i])
		}

		expectedPath := filepath.Join(env.tmpDir, "Projects", "testuser", repo.Name)
		if executedCommands[i][1] != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, executedCommands[i][1])
		}
	}
}
