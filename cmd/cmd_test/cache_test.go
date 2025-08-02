package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestParseTTL(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     time.Duration
		wantErr  bool
	}{
		{"empty string", "", 24 * time.Hour, false},
		{"minutes", "30m", 30 * time.Minute, false},
		{"hours", "12h", 12 * time.Hour, false},
		{"days", "7d", 7 * 24 * time.Hour, false},
		{"invalid unit", "5x", 0, true},
		{"invalid value", "abch", 0, true},
		{"too short", "h", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmd.ParseTTL(tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTTL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheDirectoryCreation(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	cacheDir, err := cmd.GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir() failed: %v", err)
	}

	expectedCacheDir := filepath.Join(env.tmpDir, ".cache", "gh-repo-man")
	if cacheDir != expectedCacheDir {
		t.Errorf("Expected cache dir %s, got %s", expectedCacheDir, cacheDir)
	}

	readmeDir := filepath.Join(cacheDir, "readmes")
	if _, err := os.Stat(readmeDir); os.IsNotExist(err) {
		t.Error("README cache directory was not created")
	}
}

func TestReposCaching(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	testRepos := createTestRepos()

	err := cmd.SaveReposToCache("testuser", testRepos)
	if err != nil {
		t.Fatalf("SaveReposToCache() failed: %v", err)
	}

	loadedRepos, err := cmd.LoadReposFromCache("testuser")
	if err != nil {
		t.Fatalf("LoadReposFromCache() failed: %v", err)
	}

	if len(loadedRepos) != len(testRepos) {
		t.Errorf("Expected %d repos, got %d", len(testRepos), len(loadedRepos))
	}

	if loadedRepos[0].Name != testRepos[0].Name {
		t.Errorf("Expected repo name %s, got %s", testRepos[0].Name, loadedRepos[0].Name)
	}
}

func TestReadmeCaching(t *testing.T) {
	env := setupTempHome(t)
	defer env.cleanup()

	testContent := "# Test Repository\n\nThis is a test README."

	err := cmd.SaveReadmeToCache("testuser", "test-repo", testContent)
	if err != nil {
		t.Fatalf("SaveReadmeToCache() failed: %v", err)
	}

	loadedContent, err := cmd.LoadReadmeFromCache("testuser", "test-repo")
	if err != nil {
		t.Fatalf("LoadReadmeFromCache() failed: %v", err)
	}

	if loadedContent != testContent {
		t.Errorf("Expected content %q, got %q", testContent, loadedContent)
	}
}

func TestCacheValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	content := []byte(`{"test": "data"}`)
	err := os.WriteFile(testFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !cmd.IsCacheValid(testFile, 1*time.Hour) {
		t.Error("Recently created file should be valid")
	}

	if cmd.IsCacheValid(testFile, 1*time.Nanosecond) {
		t.Error("File should be invalid with very short TTL")
	}

	if cmd.IsCacheValid("/nonexistent/file", 1*time.Hour) {
		t.Error("Nonexistent file should be invalid")
	}
}

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
