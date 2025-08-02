package cmd_test

import (
	"bytes"
	"context"
	"fmt"
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

type mockTestSetup struct {
	env             *testEnv
	originalExecCmd func(string, ...string) *exec.Cmd
}

func setupMockTest(t *testing.T) *mockTestSetup {
	env := setupTempHome(t)

	cmd.SetConfig(cmd.Config{
		Repos: cmd.ReposConfig{
			ProjectsDir: "~/Projects",
			PerUserDir:  true,
		},
	})

	originalExecCmd := cmd.ExecCommand
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	return &mockTestSetup{
		env:             env,
		originalExecCmd: originalExecCmd,
	}
}

func (ts *mockTestSetup) cleanup() {
	ts.env.cleanup()
	cmd.ExecCommand = ts.originalExecCmd
}

func captureStdout(t *testing.T, fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		panic(fmt.Sprintf("Failed to read from pipe: %v", err))
	}
	return buf.String()
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	command := os.Args[3]
	switch command {
	case "gh":
		handleGhCommand()
	case "git":
		handleGitCommand()
	case "fzf":
		handleFzfCommand()
	case "gh-repo-man":
		handleGhRepoManagerCommand()
	case "which":
		handleWhichCommand()
	case "tea":
		handleTeaCommand()
	case "nvim", "vim", "code":
		handleEditorCommand()
	case "sleep":
		time.Sleep(15 * time.Second)
	}
	os.Exit(0)
}

func handleGhCommand() {
	if os.Args[4] == "repo" && os.Args[5] == "list" {
		hasUsername := false
		for i := 6; i < len(os.Args); i++ {
			arg := os.Args[i]
			if arg != "--limit" && arg != "1000" && arg != "--json" &&
				arg != "name,description,url,stargazerCount,forkCount,watchers,issues,owner,createdAt,updatedAt,diskUsage,homepageUrl,isFork,isArchived,isPrivate,isTemplate,repositoryTopics,primaryLanguage" {
				hasUsername = true
				break
			}
		}

		if hasUsername {
			fmt.Fprintf(os.Stdout, "[%s]", mockUserRepo1JSON)
		} else {
			fmt.Fprintf(os.Stdout, "[%s,%s]", mockRepo1JSON, mockRepo2JSON)
		}
	} else if os.Args[4] == "api" && os.Args[5] == "user" {
		fmt.Fprint(os.Stdout, `{"login":"testuser"}`)
	} else if os.Args[4] == "api" && strings.HasPrefix(os.Args[5], "repos/") && strings.HasSuffix(os.Args[5], "/readme") {
		repoFullName := strings.TrimSuffix(strings.TrimPrefix(os.Args[5], "repos/"), "/readme")
		switch repoFullName {
		case "user/repo1":
			fmt.Fprint(os.Stdout, "# Repo1 Readme\n\nThis is the readme content for repo1.")
		case "user/userRepo1":
			fmt.Fprint(os.Stdout, "# UserRepo1 Readme\n\nThis is the readme content for userRepo1.")
		default:
			fmt.Fprint(os.Stderr, "Not Found")
			os.Exit(1)
		}
	}
}

func handleGitCommand() {
	if os.Args[4] == "clone" {
		if os.Args[5] == "fail_clone_url" {
			fmt.Fprint(os.Stderr, "mock clone error")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "Cloning into '%s'...\n", os.Args[5])
	}
}

func handleFzfCommand() {
	if len(os.Args) > 7 {
		switch os.Args[7] {
		case "--test-cancel":
			os.Exit(130)
		case "--test-esc":
			os.Exit(1)
		}
	}
	fmt.Fprint(os.Stdout, "repo1\n")
}

func handleGhRepoManagerCommand() {
	if os.Args[4] == "preview" && len(os.Args) > 5 {
		repoName := os.Args[5]
		switch repoName {
		case "repo1":
			printMockPreview("repo1", "desc1", "https://github.com/user/repo1", 100, 50, 30, 20, "user", "2022-01-01 00:00:00", "2022-01-02 00:00:00", 1000, "https://user.github.io/repo1", "go, cli", "# Repo1 Readme\n\nThis is the readme content for repo1.")
		case "userRepo1":
			printMockPreview("userRepo1", "userDesc1", "https://github.com/user/userRepo1", 10, 5, 3, 2, "user", "2023-01-01 00:00:00", "2023-01-02 00:00:00", 100, "https://user.github.io/userRepo1", "go, cli", "# UserRepo1 Readme\n\nThis is the readme content for userRepo1.")
		default:
			fmt.Printf("Repository %s not found.\n", repoName)
		}
	}
}

func printMockPreview(name, desc, url string, stars, forks, watchers, issues int, owner, createdAt, updatedAt string, diskUsage int, homepage, topics, readme string) {
	fmt.Printf("# %s\n\n%s Language: %s\n", name, "🐹", "Go")
	fmt.Printf("ℹ️ %s\n", desc)
	fmt.Printf("🔗 [Link](%s)\n\n", url)
	fmt.Printf("⭐ %d  🍴 %d  👁 %d  🐛 %d\n", stars, forks, watchers, issues)
	fmt.Printf("👤 Owner: %s\n", owner)
	fmt.Printf("📅 Created At: %s\n", createdAt)
	fmt.Printf("⏰ Last Updated: %s\n", updatedAt)
	fmt.Printf("💾 Disk Usage: %d KB\n", diskUsage)
	fmt.Printf("🏠 [Homepage](%s)\n", homepage)
	fmt.Printf("\n🏷 Topics: %s\n", topics)
	fmt.Print("\n---\n")
	fmt.Println(readme)
}

func handleWhichCommand() {
	if len(os.Args) > 4 && os.Args[4] == "tea" {
		if os.Getenv("MOCK_TEA_AVAILABLE") == "true" {
			os.Exit(0)
		}
		os.Exit(1)
	}
	os.Exit(1)
}

func handleTeaCommand() {
	os.Exit(0)
}

func handleEditorCommand() {
	os.Exit(0)
}

type repoData struct {
	name, language, description, url, owner, createdAt, updatedAt, homepage, readmeContent string
	stars, forks, watchers, issues, diskUsage                                              int
	topics                                                                                 []string
}

func buildExpectedPreviewOutput(repoName, language, description, url string, stars, forks, watchers, issues int,
	owner, createdAt, updatedAt string, diskUsage int, homepage string, topics []string, readmeContent string,
) string {
	data := repoData{
		name: repoName, language: language, description: description, url: url,
		stars: stars, forks: forks, watchers: watchers, issues: issues,
		owner: owner, createdAt: createdAt, updatedAt: updatedAt,
		diskUsage: diskUsage, homepage: homepage, topics: topics, readmeContent: readmeContent,
	}

	return buildPreviewOutput(data)
}

func buildPreviewOutput(data repoData) string {
	languageIcon := cmd.GetLanguageIcon(data.language)
	output := fmt.Sprintf("# %s\n\n%s Language: %s\n", data.name, languageIcon, data.language)

	if data.description != "" {
		output += fmt.Sprintf("%s %s\n", cmd.GetIcon("info"), data.description)
	}

	output += fmt.Sprintf("%s [Link](%s)\n\n", cmd.GetIcon("link"), data.url)
	output += fmt.Sprintf("%s %d  %s %d  %s %d  %s %d\n",
		cmd.GetIcon("star"), data.stars, cmd.GetIcon("fork"), data.forks,
		cmd.GetIcon("watch"), data.watchers, cmd.GetIcon("issue"), data.issues)

	output += fmt.Sprintf("%s Owner: %s\n", cmd.GetIcon("owner"), data.owner)
	output += fmt.Sprintf("%s Created At: %s\n", cmd.GetIcon("calendar"), data.createdAt)
	output += fmt.Sprintf("%s Last Updated: %s\n", cmd.GetIcon("clock"), data.updatedAt)
	output += fmt.Sprintf("%s Disk Usage: %d KB\n", cmd.GetIcon("disk"), data.diskUsage)

	if data.homepage != "" {
		output += fmt.Sprintf("%s [Homepage](%s)\n", cmd.GetIcon("home"), data.homepage)
	}

	if len(data.topics) > 0 {
		output += fmt.Sprintf("\n%s Topics: %s\n", cmd.GetIcon("tag"), strings.Join(data.topics, ", "))
	}

	return output + "\n---\n" + data.readmeContent + "\n"
}

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

