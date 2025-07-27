package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
	"github.com/spf13/cobra"
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
		ProjectsDir: "~/Projects",
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
	buf.ReadFrom(r)
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
	case "gh-repo-manager":
		handleGhRepoManagerCommand()
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

func TestCloneRepos(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	t.Run("successful cloning", func(t *testing.T) {
		reposToClone := []cmd.Repo{
			{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
			{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
		}

		err := cmd.CloneRepos(reposToClone)
		if err != nil {
			t.Errorf("CloneRepos() returned an error for successful cloning: %v", err)
		}
	})

	t.Run("failed cloning", func(t *testing.T) {
		reposToClone := []cmd.Repo{
			{Name: "fail_repo", HTMLURL: "fail_clone_url"},
		}

		err := cmd.CloneRepos(reposToClone)
		if err == nil {
			t.Error("CloneRepos() did not return an error for failed cloning")
		}
	})
}

func TestPreviewCmd(t *testing.T) {
	ts := setupMockTest(t)
	defer ts.cleanup()

	dummyCmd := &cobra.Command{}
	cmd.SetConfig(cmd.Config{ShowReadmeInPreview: true})

	t.Run("repo1 preview", func(t *testing.T) {
		output := captureStdout(t, func() {
			cmd.PreviewCmd.Run(dummyCmd, []string{"repo1"})
		})

		expected := buildExpectedPreviewOutput(
			"repo1", "Go", "desc1", "https://github.com/user/repo1",
			100, 50, 30, 20, "user", "2022-01-01 00:00:00", "2022-01-02 00:00:00",
			1000, "https://user.github.io/repo1", []string{"go", "cli"},
			"# Repo1 Readme\n\nThis is the readme content for repo1.",
		)

		if output != expected {
			t.Errorf("preview output mismatch\nGot: %q\nWant: %q", output, expected)
		}
	})

	t.Run("userRepo1 preview", func(t *testing.T) {
		oldUser := cmd.User
		cmd.User = "someuser"
		defer func() { cmd.User = oldUser }()

		output := captureStdout(t, func() {
			cmd.PreviewCmd.Run(dummyCmd, []string{"userRepo1"})
		})

		expected := buildExpectedPreviewOutput(
			"userRepo1", "Go", "userDesc1", "https://github.com/user/userRepo1",
			10, 5, 3, 2, "user", "2023-01-01 00:00:00", "2023-01-02 00:00:00",
			100, "https://user.github.io/userRepo1", []string{"go", "cli"},
			"# UserRepo1 Readme\n\nThis is the readme content for userRepo1.",
		)

		if output != expected {
			t.Errorf("preview output mismatch for user repo\nGot: %q\nWant: %q", output, expected)
		}
	})
}
