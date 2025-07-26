package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
)

func TestCachePerformance(t *testing.T) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testRepos := []cmd.Repo{
		{Name: "test-repo", Owner: cmd.Owner{Login: "testuser"}},
	}

	// Test that saving and loading is fast
	start := time.Now()
	err := cmd.SaveReposToCache("testuser", testRepos)
	if err != nil {
		t.Fatalf("SaveReposToCache failed: %v", err)
	}
	saveTime := time.Since(start)

	start = time.Now()
	_, err = cmd.LoadReposFromCache("testuser")
	if err != nil {
		t.Fatalf("LoadReposFromCache failed: %v", err)
	}
	loadTime := time.Since(start)

	// Cache operations should be very fast (under 10ms)
	if saveTime > 10*time.Millisecond {
		t.Errorf("Cache save took too long: %v", saveTime)
	}
	if loadTime > 10*time.Millisecond {
		t.Errorf("Cache load took too long: %v", loadTime)
	}

	// Test cache file exists
	cachePath, err := cmd.GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir failed: %v", err)
	}

	repoFile := filepath.Join(cachePath, "testuser_repos.json")
	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		t.Error("Cache file was not created")
	}
}
