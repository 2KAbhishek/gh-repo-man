package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestCachePerformance(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	testRepos := createTestRepos()[:1]

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

	assertCacheOperationTime(t, saveTime, "save")
	assertCacheOperationTime(t, loadTime, "load")

	cachePath, err := cmd.GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir failed: %v", err)
	}

	repoFile := filepath.Join(cachePath, "testuser_repos.json")
	if _, err := os.Stat(repoFile); os.IsNotExist(err) {
		t.Error("Cache file was not created")
	}
}
