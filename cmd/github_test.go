package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestHelperProcess isn't a real test. It's a helper process that gets
// executed by the tests. It's used to mock the execution of the `gh` and `fzf` commands.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// The first argument is the test binary path, then -test.run=TestHelperProcess, then --, then the actual command and its args
	// So, the actual command starts at os.Args[3]

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
		// Mock fzf to return a selected repo
		fmt.Fprint(os.Stdout, "repo1\n")
	case "gh-repo-manager": // This is for the preview subcommand
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
	// We replace the real exec.Command with our mock version.
	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	// We restore the original exec.Command at the end of the test.
	defer func() { execCommand = exec.Command }()

	// Test case 1: Get repos for the current user
	repos, err := GetRepos("")
	if err != nil {
		t.Errorf("GetRepos() with empty user returned an error: %v", err)
	}

	expectedRepos := []Repo{
		{Name: "repo1", Description: "desc1", Ssh_url: "git@github.com:user/repo1.git", StargazerCount: 100, ForkCount: 50},
		{Name: "repo2", Description: "desc2", Ssh_url: "git@github.com:user/repo2.git", StargazerCount: 200, ForkCount: 100},
	}

	if !reflect.DeepEqual(repos, expectedRepos) {
		t.Errorf("GetRepos() with empty user returned %+v, want %+v", repos, expectedRepos)
	}

	// Test case 2: Get repos for a specific user
	repos, err = GetRepos("someuser")
	if err != nil {
		t.Errorf("GetRepos() with a user returned an error: %v", err)
	}

	expectedUserRepos := []Repo{
		{Name: "userRepo1", Description: "userDesc1", Ssh_url: "git@github.com:user/userRepo1.git", StargazerCount: 10, ForkCount: 5},
	}

	if !reflect.DeepEqual(repos, expectedUserRepos) {
		t.Errorf("GetRepos() with a user returned %+v, want %+v", repos, expectedUserRepos)
	}
}

func TestCloneRepos(t *testing.T) {
	// We replace the real exec.Command with our mock version.
	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	// We restore the original exec.Command at the end of the test.
	defer func() { execCommand = exec.Command }()

	// Test case 1: Successful cloning
	reposToClone := []Repo{
		{Name: "repo1", Ssh_url: "git@github.com:user/repo1.git"},
		{Name: "repo2", Ssh_url: "git@github.com:user/repo2.git"},
	}

	err := CloneRepos(reposToClone)
	if err != nil {
		t.Errorf("CloneRepos() returned an error for successful cloning: %v", err)
	}

	// Test case 2: Failed cloning
	reposToClone = []Repo{
		{Name: "fail_repo", Ssh_url: "fail_clone_url"},
	}

	err = CloneRepos(reposToClone)
	if err == nil {
		t.Error("CloneRepos() did not return an error for failed cloning")
	}
}

func TestFzfIntegration(t *testing.T) {
	// Mock exec.Command to simulate gh and fzf
	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
	defer func() { execCommand = exec.Command }()

	// Test previewCmd
	oldStdout := os.Stdout // Keep original stdout
	r, w, _ := os.Pipe()  // Create a pipe
	os.Stdout = w         // Redirect stdout to the pipe

	dummyCmd := &cobra.Command{}
	previewCmd.Run(dummyCmd, []string{"repo1"})

	w.Close()             // Close the write end of the pipe
	os.Stdout = oldStdout // Restore original stdout

	var buf bytes.Buffer // Declare buf here
	buf.ReadFrom(r) // Read from the pipe into the buffer

	expectedOutput := "Name: repo1\nDescription: desc1\nSSH URL: git@github.com:user/repo1.git\nStars: 100\nForks: 50\n"
	if !strings.Contains(buf.String(), expectedOutput) {
		t.Errorf("previewCmd output mismatch\nGot: %q\nWant: %q", buf.String(), expectedOutput)
	}
}