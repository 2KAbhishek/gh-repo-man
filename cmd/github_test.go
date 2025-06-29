package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

// TestHelperProcess isn't a real test. It's a helper process that gets
// executed by the tests. It's used to mock the execution of the `gh` command.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// The first argument is the command name, so we need to check the second
	// argument to see what command we're mocking.
	if os.Args[4] == "repo" && os.Args[5] == "list" {
		// Check if a user is provided as an argument to gh repo list
		if len(os.Args) > 6 && os.Args[6] != "--json" {
			// This means a user was provided, so we can return a different set of repos
			fmt.Fprintf(os.Stdout, `[{"name":"userRepo1","description":"userDesc1","sshUrl":"git@github.com:user/userRepo1.git"}]`)
		} else {
			// No user provided, return default repos
			fmt.Fprintf(os.Stdout, `[{"name":"repo1","description":"desc1","sshUrl":"git@github.com:user/repo1.git"},{"name":"repo2","description":"desc2","sshUrl":"git@github.com:user/repo2.git"}]`)
		}
	} else if os.Args[4] == "clone" {
		if os.Args[5] == "fail_clone_url" {
			fmt.Fprint(os.Stderr, "mock clone error")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "Cloning into '%s'...\n", os.Args[5])
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
		{Name: "repo1", Description: "desc1", Ssh_url: "git@github.com:user/repo1.git"},
		{Name: "repo2", Description: "desc2", Ssh_url: "git@github.com:user/repo2.git"},
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
		{Name: "userRepo1", Description: "userDesc1", Ssh_url: "git@github.com:user/userRepo1.git"},
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