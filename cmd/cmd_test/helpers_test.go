package cmd_test

import (
	"os"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
	"gopkg.in/yaml.v3"
)

type testEnv struct {
	tmpDir       string
	originalHome string
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
