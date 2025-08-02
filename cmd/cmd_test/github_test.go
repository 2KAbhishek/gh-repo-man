package cmd_test

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

const (
	mockRepo1JSON     = `{"name":"repo1","description":"desc1","url":"https://github.com/user/repo1","stargazerCount":100,"forkCount":50,"watchers":{"totalCount":30},"issues":{"totalCount":20},"owner":{"login":"user"},"createdAt":"2022-01-01T00:00:00Z","updatedAt":"2022-01-02T00:00:00Z","diskUsage":1000,"homepageUrl":"https://user.github.io/repo1","isFork":false,"isArchived":false,"isPrivate":false,"isTemplate":false,"repositoryTopics":[{"name":"go"},{"name":"cli"}],"primaryLanguage":{"name":"Go"}}`
	mockRepo2JSON     = `{"name":"repo2","description":"desc2","url":"https://github.com/user/repo2","stargazerCount":200,"forkCount":100,"watchers":{"totalCount":60},"issues":{"totalCount":40},"owner":{"login":"user"},"createdAt":"2022-03-01T00:00:00Z","updatedAt":"2022-03-02T00:00:00Z","diskUsage":2000,"homepageUrl":"","isFork":false,"isArchived":false,"isPrivate":false,"isTemplate":false,"repositoryTopics":[],"primaryLanguage":{"name":"Python"}}`
	mockUserRepo1JSON = `{"name":"userRepo1","description":"userDesc1","url":"https://github.com/user/userRepo1","stargazerCount":10,"forkCount":5,"watchers":{"totalCount":3},"issues":{"totalCount":2},"owner":{"login":"user"},"createdAt":"2023-01-01T00:00:00Z","updatedAt":"2023-01-02T00:00:00Z","diskUsage":100,"homepageUrl":"https://user.github.io/userRepo1","isFork":false,"isArchived":false,"isPrivate":false,"isTemplate":false,"repositoryTopics":[{"name":"go"},{"name":"cli"}],"primaryLanguage":{"name":"Go"}}`
)

var (
	expectedRepo1 = cmd.Repo{
		Name: "repo1", Description: "desc1", HTMLURL: "https://github.com/user/repo1",
		StargazerCount: 100, ForkCount: 50, Watchers: cmd.Count{TotalCount: 30}, Issues: cmd.Count{TotalCount: 20},
		Owner: cmd.Owner{Login: "user"}, CreatedAt: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC), DiskUsage: 1000,
		HomepageURL: "https://user.github.io/repo1", IsFork: false, IsArchived: false, IsPrivate: false, IsTemplate: false,
		Topics: []cmd.Topic{{Name: "go"}, {Name: "cli"}}, PrimaryLanguage: cmd.Language{Name: "Go"},
	}

	expectedRepo2 = cmd.Repo{
		Name: "repo2", Description: "desc2", HTMLURL: "https://github.com/user/repo2",
		StargazerCount: 200, ForkCount: 100, Watchers: cmd.Count{TotalCount: 60}, Issues: cmd.Count{TotalCount: 40},
		Owner: cmd.Owner{Login: "user"}, CreatedAt: time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2022, 3, 2, 0, 0, 0, 0, time.UTC), DiskUsage: 2000, HomepageURL: "",
		IsFork: false, IsArchived: false, IsPrivate: false, IsTemplate: false,
		Topics: []cmd.Topic{}, PrimaryLanguage: cmd.Language{Name: "Python"},
	}

	expectedUserRepo1 = cmd.Repo{
		Name: "userRepo1", Description: "userDesc1", HTMLURL: "https://github.com/user/userRepo1",
		StargazerCount: 10, ForkCount: 5, Watchers: cmd.Count{TotalCount: 3}, Issues: cmd.Count{TotalCount: 2},
		Owner: cmd.Owner{Login: "user"}, CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), DiskUsage: 100,
		HomepageURL: "https://user.github.io/userRepo1", IsFork: false, IsArchived: false, IsPrivate: false, IsTemplate: false,
		Topics: []cmd.Topic{{Name: "go"}, {Name: "cli"}}, PrimaryLanguage: cmd.Language{Name: "Go"},
	}
)

func TestGetRepos(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	t.Run("empty user", func(t *testing.T) {
		repos, err := cmd.GetRepos("")
		if err != nil {
			t.Fatalf("GetRepos() with empty user returned an error: %v", err)
		}

		expected := []cmd.Repo{expectedRepo1, expectedRepo2}
		if !reflect.DeepEqual(repos, expected) {
			t.Errorf("GetRepos() with empty user returned %+v, want %+v", repos, expected)
		}
	})

	t.Run("specific user", func(t *testing.T) {
		repos, err := cmd.GetRepos("someuser")
		if err != nil {
			t.Fatalf("GetRepos() with a user returned an error: %v", err)
		}

		expected := []cmd.Repo{expectedUserRepo1}
		if !reflect.DeepEqual(repos, expected) {
			t.Errorf("GetRepos() with a user returned %+v, want %+v", repos, expected)
		}
	})
}

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

func TestGetReposWithValidation(t *testing.T) {
	originalExecCommand := cmd.ExecCommand
	defer func() { cmd.ExecCommand = originalExecCommand }()

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

func TestGetCurrentUsername(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	username, err := cmd.GetCurrentUsername()
	if err != nil {
		t.Errorf("GetCurrentUsername() returned error: %v", err)
	}

	if username != "testuser" {
		t.Errorf("GetCurrentUsername() = %q, want %q", username, "testuser")
	}
}

func TestGetReadme(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	t.Run("existing readme", func(t *testing.T) {
		content, err := cmd.GetReadme("user/repo1")
		if err != nil {
			t.Errorf("GetReadme() returned error: %v", err)
		}

		expected := "# Repo1 Readme\n\nThis is the readme content for repo1."
		if content != expected {
			t.Errorf("GetReadme() = %q, want %q", content, expected)
		}
	})

	t.Run("nonexistent readme", func(t *testing.T) {
		content, err := cmd.GetReadme("user/nonexistent")
		if err != nil {
			t.Errorf("GetReadme() for nonexistent repo should not error, got: %v", err)
		}

		if content != "" {
			t.Errorf("GetReadme() for nonexistent repo should return empty string, got: %q", content)
		}
	})

	t.Run("invalid repo format", func(t *testing.T) {
		_, err := cmd.GetReadme("invalid-format")
		if err == nil {
			t.Error("GetReadme() with invalid format should return error")
		}

		if !strings.Contains(err.Error(), "invalid repository name format") {
			t.Errorf("GetReadme() error should mention format, got: %q", err.Error())
		}
	})
}
