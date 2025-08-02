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
