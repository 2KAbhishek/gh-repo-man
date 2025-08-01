package cmd_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty username (current user)",
			username: "",
			wantErr:  false,
		},
		{
			name:     "valid simple username",
			username: "user123",
			wantErr:  false,
		},
		{
			name:     "valid username with hyphen",
			username: "test-user",
			wantErr:  false,
		},
		{
			name:     "valid username with underscore",
			username: "test_user",
			wantErr:  false,
		},
		{
			name:     "valid username with mixed characters",
			username: "test-user_123",
			wantErr:  false,
		},
		{
			name:     "single character username",
			username: "a",
			wantErr:  false,
		},
		{
			name:     "maximum length username",
			username: strings.Repeat("a", cmd.MaxUsernameLength),
			wantErr:  false,
		},

		{
			name:     "too long username",
			username: strings.Repeat("a", cmd.MaxUsernameLength+1),
			wantErr:  true,
			errMsg:   "username too long",
		},

		{
			name:     "username with semicolon",
			username: "user;rm-rf",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with pipe",
			username: "user|dangerous",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with ampersand",
			username: "user&command",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with dollar",
			username: "user$variable",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with backtick",
			username: "user`command`",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with parentheses",
			username: "user()",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with braces",
			username: "user{}",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with brackets",
			username: "user[]",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with angle brackets",
			username: "user<>",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with quotes",
			username: "user\"'",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},
		{
			name:     "username with backslash",
			username: "user\\escape",
			wantErr:  true,
			errMsg:   "contains invalid characters",
		},

		{
			name:     "username starting with hyphen",
			username: "-user",
			wantErr:  true,
			errMsg:   "format is invalid",
		},
		{
			name:     "username ending with hyphen",
			username: "user-",
			wantErr:  true,
			errMsg:   "format is invalid",
		},
		{
			name:     "username starting with underscore",
			username: "_user",
			wantErr:  true,
			errMsg:   "format is invalid",
		},
		{
			name:     "username ending with underscore",
			username: "user_",
			wantErr:  true,
			errMsg:   "format is invalid",
		},
		{
			name:     "username with space",
			username: "user name",
			wantErr:  true,
			errMsg:   "format is invalid",
		},
		{
			name:     "username with dot",
			username: "user.name",
			wantErr:  true,
			errMsg:   "format is invalid",
		},
		{
			name:     "username with multiple consecutive hyphens",
			username: "user--name",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.ValidateUsername(tt.username)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateUsername(%q) expected error, got nil", tt.username)
					return
				}

				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateUsername(%q) error = %q, want to contain %q", tt.username, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateUsername(%q) unexpected error = %q", tt.username, err.Error())
				}
			}
		})
	}
}

func TestGetReposWithValidation(t *testing.T) {
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

	// Don't actually execute commands in this test
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		return nil
	}

	_, err := cmd.GetRepos("user;rm-rf")
	if err == nil {
		t.Error("GetRepos with invalid username should return validation error")
	}

	if !strings.Contains(err.Error(), "invalid username") {
		t.Errorf("GetRepos validation error should mention 'invalid username', got: %q", err.Error())
	}
}
