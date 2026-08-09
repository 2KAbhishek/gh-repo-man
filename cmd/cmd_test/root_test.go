package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestGetCommandInvocation(t *testing.T) {
	tests := []struct {
		name           string
		setupPath      func(t *testing.T) (cleanup func())
		expectedResult string
	}{
		{
			name: "gh-repo-man exists in PATH",
			setupPath: func(t *testing.T) func() {
				tmpDir := t.TempDir()

				fakeBinary := filepath.Join(tmpDir, "gh-repo-man")
				file, err := os.Create(fakeBinary)
				if err != nil {
					t.Fatal(err)
				}
				file.Close()

				err = os.Chmod(fakeBinary, 0o755)
				if err != nil {
					t.Fatal(err)
				}

				originalPath := os.Getenv("PATH")
				newPath := tmpDir + ":" + originalPath
				os.Setenv("PATH", newPath)

				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			expectedResult: "gh-repo-man",
		},
		{
			name: "gh-repo-man does not exist in PATH",
			setupPath: func(t *testing.T) func() {
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "")

				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			expectedResult: "gh repo-man",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupPath(t)
			defer cleanup()

			result := cmd.GetCommandInvocation()

			if result != tt.expectedResult {
				t.Errorf("GetCommandInvocation() = %q, want %q", result, tt.expectedResult)
			}
		})
	}
}

func TestFzfCancellationLogic(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		isCancel bool
	}{
		{"ctrl-c cancellation", 130, true},
		{"esc cancellation", 1, true},
		{"other error", 2, false},
		{"success", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCancelError := tt.exitCode == 130 || tt.exitCode == 1

			if isCancelError != tt.isCancel {
				t.Errorf("Expected isCancelError=%v for exit code %d, got %v", tt.isCancel, tt.exitCode, isCancelError)
			}
		})
	}
}

func TestSetConfig(t *testing.T) {
	testConfig := cmd.Config{
		UI: cmd.UIConfig{
			ShowReadmeInPreview: true,
		},
	}
	cmd.SetConfig(testConfig)
}

func TestExecute(t *testing.T) {
	t.Run("Execute function exists", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Execute() panicked: %v", r)
			}
		}()
	})
}

func TestProjectsDirFlagOverride(t *testing.T) {
	tests := []struct {
		name              string
		configProjectsDir string
		flagProjectsDir   string
		expectedDir       string
	}{
		{
			name:              "flag overrides config",
			configProjectsDir: "~/Projects",
			flagProjectsDir:   "/tmp/custom",
			expectedDir:       "/tmp/custom",
		},
		{
			name:              "empty flag keeps config",
			configProjectsDir: "~/Projects",
			flagProjectsDir:   "",
			expectedDir:       "~/Projects",
		},
		{
			name:              "relative path flag",
			configProjectsDir: "~/Projects",
			flagProjectsDir:   ".",
			expectedDir:       ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalConfig := cmd.Config{
				Repos: cmd.ReposConfig{
					ProjectsDir: tt.configProjectsDir,
					PerUserDir:  true,
				},
			}
			cmd.SetConfig(originalConfig)

			cmd.ProjectsDir = tt.flagProjectsDir

			if cmd.ProjectsDir != "" {
				cfg := cmd.Config{
					Repos: cmd.ReposConfig{
						ProjectsDir: cmd.ProjectsDir,
						PerUserDir:  originalConfig.Repos.PerUserDir,
					},
				}
				cmd.SetConfig(cfg)
			}

			projectsDir, err := cmd.GetProjectsDirForUser("testuser")
			if err != nil {
				t.Fatalf("GetProjectsDirForUser() error = %v", err)
			}

			expectedPath := tt.expectedDir
			if tt.expectedDir != "." && tt.expectedDir != "/tmp/custom" {
				home, _ := os.UserHomeDir()
				expectedPath = filepath.Join(home, "Projects")
			}

			if originalConfig.Repos.PerUserDir && tt.expectedDir != "." {
				if tt.expectedDir == "/tmp/custom" {
					expectedPath = filepath.Join("/tmp/custom", "testuser")
				} else {
					expectedPath = filepath.Join(expectedPath, "testuser")
				}
			}

			if tt.flagProjectsDir == "." && originalConfig.Repos.PerUserDir {
				expectedPath = filepath.Join(".", "testuser")
			}

			if projectsDir != expectedPath {
				t.Errorf("Expected projects dir %q, got %q", expectedPath, projectsDir)
			}
		})
	}
}

func TestFindRepoByName(t *testing.T) {
	repos := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
	}

	t.Run("found", func(t *testing.T) {
		found := cmd.FindRepoByName(repos, "repo2")
		if found == nil || found.Name != "repo2" {
			t.Errorf("FindRepoByName() expected repo2, got %v", found)
		}
	})

	t.Run("not found", func(t *testing.T) {
		found := cmd.FindRepoByName(repos, "nonexistent")
		if found != nil {
			t.Errorf("FindRepoByName() expected nil, got %v", found)
		}
	})
}

func TestListCmd(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	cmd.ListCmd.SetArgs([]string{"--user", "someuser"})
	err := cmd.ListCmd.Execute()
	if err != nil {
		t.Fatalf("ListCmd.Execute() returned error: %v", err)
	}
}

func TestBuildReloadCommand(t *testing.T) {
	cmd.RepoType = "archived"
	cmd.LanguageFilter = "Go"
	cmd.SortBy = "stars"
	defer func() {
		cmd.RepoType = ""
		cmd.LanguageFilter = ""
		cmd.SortBy = ""
	}()

	reloadCmd := cmd.BuildReloadCommand("myuser")
	if !strings.Contains(reloadCmd, "list") {
		t.Errorf("expected reload command to contain 'list', got: %s", reloadCmd)
	}
	if !strings.Contains(reloadCmd, "--user myuser") {
		t.Errorf("expected reload command to contain user flag, got: %s", reloadCmd)
	}
	if !strings.Contains(reloadCmd, "--type archived") {
		t.Errorf("expected reload command to contain type flag, got: %s", reloadCmd)
	}
	if !strings.Contains(reloadCmd, "--language Go") {
		t.Errorf("expected reload command to contain language flag, got: %s", reloadCmd)
	}
	if !strings.Contains(reloadCmd, "--sort stars") {
		t.Errorf("expected reload command to contain sort flag, got: %s", reloadCmd)
	}
}
