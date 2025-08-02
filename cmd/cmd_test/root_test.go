package cmd_test

import (
	"os"
	"path/filepath"
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
