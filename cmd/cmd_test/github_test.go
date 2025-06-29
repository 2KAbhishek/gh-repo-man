package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

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
				fmt.Fprintf(os.Stdout, `[{"name":"userRepo1","description":"userDesc1","sshUrl":"git@github.com:user/userRepo1.git","stargazerCount":10,"forkCount":5}]`)
			} else { // No user provided
				fmt.Fprintf(os.Stdout, `[{"name":"repo1","description":"desc1","sshUrl":"git@github.com:user/repo1.git","stargazerCount":100,"forkCount":50},{"name":"repo2","description":"desc2","sshUrl":"git@github.com:user/repo2.git","stargazerCount":200,"forkCount":100}]`)
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
				fmt.Printf("Name: repo1\nDescription: desc1\nSSH URL: git@github.com:user/repo1.git\nStars: 100\nForks: 50\n")
			} else if repoName == "userRepo1" {
				fmt.Printf("Name: userRepo1\nDescription: userDesc1\nSSH URL: git@github.com:user/userRepo1.git\nStars: 10\nForks: 5\n")
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
		{Name: "repo1", Description: "desc1", Ssh_url: "git@github.com:user/repo1.git", StargazerCount: 100, ForkCount: 50},
		{Name: "repo2", Description: "desc2", Ssh_url: "git@github.com:user/repo2.git", StargazerCount: 200, ForkCount: 100},
	}

	if !reflect.DeepEqual(repos, expectedRepos) {
		t.Errorf("GetRepos() with empty user returned %+v, want %+v", repos, expectedRepos)
	}

	repos, err = cmd.GetRepos("someuser")
	if err != nil {
		t.Errorf("GetRepos() with a user returned an error: %v", err)
	}

	expectedUserRepos := []cmd.Repo{
		{Name: "userRepo1", Description: "userDesc1", Ssh_url: "git@github.com:user/userRepo1.git", StargazerCount: 10, ForkCount: 5},
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
		{Name: "repo1", Ssh_url: "git@github.com:user/repo1.git"},
		{Name: "repo2", Ssh_url: "git@github.com:user/repo2.git"},
	}

	err := cmd.CloneRepos(reposToClone)
	if err != nil {
		t.Errorf("CloneRepos() returned an error for successful cloning: %v", err)
	}

	reposToClone = []cmd.Repo{
		{Name: "fail_repo", Ssh_url: "fail_clone_url"},
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

	expectedOutput := "Name: repo1\nDescription: desc1\nSSH URL: git@github.com:user/repo1.git\nStars: 100\nForks: 50\n"
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Errorf("previewCmd output mismatch\nGot: %q\nWant: %q", buf.String(), expectedOutput)
	}
}
