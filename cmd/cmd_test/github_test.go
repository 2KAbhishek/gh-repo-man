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

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	switch os.Args[3] { // This is the actual command being mocked (e.g., "gh", "git", "fzf", "gh-repo-manager")
	case "gh":
		if os.Args[4] == "repo" && os.Args[5] == "list" {
			// gh repo list [user] --limit 1000 --json ...
			// Check if a user is provided. The user would be at os.Args[6] if present, otherwise --limit is at os.Args[6]
			if len(os.Args) > 6 && os.Args[6] != "--limit" { // User is provided
				fmt.Fprintf(os.Stdout, `[{"name":"userRepo1","description":"userDesc1","url":"https://github.com/user/userRepo1","stargazerCount":10,"forkCount":5,"watchers":{"totalCount":3},"issues":{"totalCount":2},"owner":{"login":"user"},"createdAt":"2023-01-01T00:00:00Z","updatedAt":"2023-01-02T00:00:00Z","diskUsage":100,"homepageUrl":"https://user.github.io/userRepo1","isFork":false,"isArchived":false,"isPrivate":false,"isTemplate":false,"repositoryTopics":[{"name":"go"},{"name":"cli"}],"primaryLanguage":{"name":"Go"}}]`)
			} else { // No user provided
				fmt.Fprintf(os.Stdout, `[{"name":"repo1","description":"desc1","url":"https://github.com/user/repo1","stargazerCount":100,"forkCount":50,"watchers":{"totalCount":30},"issues":{"totalCount":20},"owner":{"login":"user"},"createdAt":"2022-01-01T00:00:00Z","updatedAt":"2022-01-02T00:00:00Z","diskUsage":1000,"homepageUrl":"https://user.github.io/repo1","isFork":false,"isArchived":false,"isPrivate":false,"isTemplate":false,"repositoryTopics":[{"name":"go"},{"name":"cli"}],"primaryLanguage":{"name":"Go"}},{"name":"repo2","description":"desc2","url":"https://github.com/user/repo2","stargazerCount":200,"forkCount":100,"watchers":{"totalCount":60},"issues":{"totalCount":40},"owner":{"login":"user"},"createdAt":"2022-03-01T00:00:00Z","updatedAt":"2022-03-02T00:00:00Z","diskUsage":2000,"homepageUrl":"","isFork":false,"isArchived":false,"isPrivate":false,"isTemplate":false,"repositoryTopics":[],"primaryLanguage":{"name":"Python"}}]`)
			}
		} else if os.Args[4] == "api" && strings.HasPrefix(os.Args[5], "repos/") && strings.HasSuffix(os.Args[5], "/readme") {
			repoFullName := strings.TrimSuffix(strings.TrimPrefix(os.Args[5], "repos/"), "/readme")
			if repoFullName == "user/repo1" {
				fmt.Fprint(os.Stdout, "# Repo1 Readme\n\nThis is the readme content for repo1.")
			} else if repoFullName == "user/userRepo1" {
				fmt.Fprint(os.Stdout, "# UserRepo1 Readme\n\nThis is the readme content for userRepo1.")
			} else {
				// Simulate 404 for other readmes
				fmt.Fprint(os.Stderr, "Not Found")
				os.Exit(1)
			}
		}
	case "git":
		if os.Args[4] == "clone" {
			if os.Args[5] == "fail_clone_url" {
				fmt.Fprint(os.Stderr, "mock clone error")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "Cloning into '%s'...\n", os.Args[5])
		}
	case "fzf":
		fmt.Fprint(os.Stdout, "repo1\n")
	case "gh-repo-manager":
		if os.Args[4] == "preview" && len(os.Args) > 5 { // os.Args[5] should be the repo name
			repoName := os.Args[5]
			if repoName == "repo1" {
				fmt.Printf("# %s\n\n%s Language: %s\n", repoName, "Git", "Go")
				fmt.Printf(" %s\n", "desc1")
				fmt.Printf(" [Link](%s)\n\n", "https://github.com/user/repo1")
				fmt.Printf(" %d   %d   %d   %d\n", 100, 50, 30, 20)
				fmt.Printf(" Owner: %s\n", "user")
				fmt.Printf(" Created At: %s\n", "2022-01-01 00:00:00")
				fmt.Printf(" Last Updated: %s\n", "2022-01-02 00:00:00")
				fmt.Printf(" Disk Usage: %d KB\n", 1000)
				fmt.Printf(" [Homepage](%s)\n", "https://user.github.io/repo1")
				fmt.Printf("\n Topics: %s\n", "go, cli")
				fmt.Print("\n---\n")
				fmt.Println("# Repo1 Readme\n\nThis is the readme content for repo1.")
			} else if repoName == "userRepo1" {
				fmt.Printf("# %s\n\n%s Language: %s\n", repoName, "Git", "Go")
				fmt.Printf(" %s\n", "userDesc1")
				fmt.Printf(" [Link](%s)\n\n", "https://github.com/user/userRepo1")
				fmt.Printf(" %d   %d   %d   %d\n", 10, 5, 3, 2)
				fmt.Printf(" Owner: %s\n", "user")
				fmt.Printf(" Created At: %s\n", "2023-01-01 00:00:00")
				fmt.Printf(" Last Updated: %s\n", "2023-01-02 00:00:00")
				fmt.Printf(" Disk Usage: %d KB\n", 100)
				fmt.Printf(" [Homepage](%s)\n", "https://user.github.io/userRepo1")
				fmt.Printf("\n Topics: %s\n", "go, cli")
				fmt.Print("\n---\n")
				fmt.Println("# UserRepo1 Readme\n\nThis is the readme content for userRepo1.")
			} else {
				fmt.Printf("Repository %s not found.\n", repoName)
			}
		}
	}
	os.Exit(0)
}

func TestGetRepos(t *testing.T) {
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { cmd.ExecCommand = exec.Command }()

	repos, err := cmd.GetRepos("")
	if err != nil {
		t.Errorf("GetRepos() with empty user returned an error: %v", err)
	}

	expectedRepos := []cmd.Repo{
		{Name: "repo1", Description: "desc1", HTMLURL: "https://github.com/user/repo1", StargazerCount: 100, ForkCount: 50, Watchers: cmd.Count{TotalCount: 30}, Issues: cmd.Count{TotalCount: 20}, Owner: cmd.Owner{Login: "user"}, CreatedAt: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC), DiskUsage: 1000, HomepageURL: "https://user.github.io/repo1", IsFork: false, IsArchived: false, IsPrivate: false, IsTemplate: false, Topics: []cmd.Topic{{Name: "go"}, {Name: "cli"}}, PrimaryLanguage: cmd.Language{Name: "Go"}},
		{Name: "repo2", Description: "desc2", HTMLURL: "https://github.com/user/repo2", StargazerCount: 200, ForkCount: 100, Watchers: cmd.Count{TotalCount: 60}, Issues: cmd.Count{TotalCount: 40}, Owner: cmd.Owner{Login: "user"}, CreatedAt: time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2022, 3, 2, 0, 0, 0, 0, time.UTC), DiskUsage: 2000, HomepageURL: "", IsFork: false, IsArchived: false, IsPrivate: false, IsTemplate: false, Topics: []cmd.Topic{}, PrimaryLanguage: cmd.Language{Name: "Python"}},
	}

	if !reflect.DeepEqual(repos, expectedRepos) {
		t.Errorf("GetRepos() with empty user returned %+v, want %+v", repos, expectedRepos)
	}

	repos, err = cmd.GetRepos("someuser")
	if err != nil {
		t.Errorf("GetRepos() with a user returned an error: %v", err)
	}

	expectedUserRepos := []cmd.Repo{
		{Name: "userRepo1", Description: "userDesc1", HTMLURL: "https://github.com/user/userRepo1", StargazerCount: 10, ForkCount: 5, Watchers: cmd.Count{TotalCount: 3}, Issues: cmd.Count{TotalCount: 2}, Owner: cmd.Owner{Login: "user"}, CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC), DiskUsage: 100, HomepageURL: "https://user.github.io/userRepo1", IsFork: false, IsArchived: false, IsPrivate: false, IsTemplate: false, Topics: []cmd.Topic{{Name: "go"}, {Name: "cli"}}, PrimaryLanguage: cmd.Language{Name: "Go"}},
	}

	if !reflect.DeepEqual(repos, expectedUserRepos) {
		t.Errorf("GetRepos() with a user returned %+v, want %+v", repos, expectedUserRepos)
	}
}

func TestCloneRepos(t *testing.T) {
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { cmd.ExecCommand = exec.Command }()

	reposToClone := []cmd.Repo{
		{Name: "repo1", HTMLURL: "https://github.com/user/repo1"},
		{Name: "repo2", HTMLURL: "https://github.com/user/repo2"},
	}

	err := cmd.CloneRepos(reposToClone)
	if err != nil {
		t.Errorf("CloneRepos() returned an error for successful cloning: %v", err)
	}

	reposToClone = []cmd.Repo{
		{Name: "fail_repo", HTMLURL: "fail_clone_url"},
	}

	err = cmd.CloneRepos(reposToClone)
	if err == nil {
		t.Error("CloneRepos() did not return an error for failed cloning")
	}
}

func TestFzfIntegration(t *testing.T) {
	cmd.ExecCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { cmd.ExecCommand = exec.Command }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	dummyCmd := &cobra.Command{}
	cmd.PreviewCmd.Run(dummyCmd, []string{"repo1"})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	expectedOutput := buildExpectedPreviewOutput(
		"repo1", "Go", "desc1", "https://github.com/user/repo1",
		100, 50, 30, 20, "user", "2022-01-01 00:00:00", "2022-01-02 00:00:00",
		1000, "https://user.github.io/repo1", []string{"go", "cli"},
		"# Repo1 Readme\n\nThis is the readme content for repo1.",
	)
	if buf.String() != expectedOutput {
		t.Errorf("previewCmd output mismatch\nGot: %q\nWant: %q", buf.String(), expectedOutput)
	}

	// Test with a user repo
	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	// Set the user for the GetRepos call within the previewCmd
	oldUser := cmd.User
	cmd.User = "someuser"

	cmd.PreviewCmd.Run(dummyCmd, []string{"userRepo1"})

	w.Close()
	os.Stdout = oldStdout
	cmd.User = oldUser // Restore user

	buf.Reset()
	buf.ReadFrom(r)

	expectedOutput = buildExpectedPreviewOutput(
		"userRepo1", "Go", "userDesc1", "https://github.com/user/userRepo1",
		10, 5, 3, 2, "user", "2023-01-01 00:00:00", "2023-01-02 00:00:00",
		100, "https://user.github.io/userRepo1", []string{"go", "cli"},
		"# UserRepo1 Readme\n\nThis is the readme content for userRepo1.",
	)
	if buf.String() != expectedOutput {
		t.Errorf("previewCmd output mismatch for user repo\nGot: %q\nWant: %q", buf.String(), expectedOutput)
	}
}
