package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-man/cmd"
	"gopkg.in/yaml.v3"
)

type testEnv struct {
	tmpDir       string
	originalHome string
}

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

func setupTempHome(t *testing.T) *testEnv {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	return &testEnv{
		tmpDir:       tmpDir,
		originalHome: originalHome,
	}
}

func (env *testEnv) cleanup() {
	os.Setenv("HOME", env.originalHome)
}

func createTestRepos() []cmd.Repo {
	return []cmd.Repo{
		{Name: "test-repo", Owner: cmd.Owner{Login: "testuser"}},
		{Name: "another-repo", Owner: cmd.Owner{Login: "testuser"}},
	}
}

func assertCacheOperationTime(t *testing.T, duration time.Duration, operation string) {
	maxDuration := 10 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("Cache %s took too long: %v (max: %v)", operation, duration, maxDuration)
	}
}

func createTempConfigFile(t *testing.T, config cmd.Config) string {
	f, err := os.CreateTemp("", "gh-repo-man-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}

	enc := yaml.NewEncoder(f)
	if err := enc.Encode(&config); err != nil {
		t.Fatal(err)
	}
	enc.Close()
	f.Close()

	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}
