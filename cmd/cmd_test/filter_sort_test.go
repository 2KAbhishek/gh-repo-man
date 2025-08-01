package cmd_test

import (
	"testing"
	"time"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func createTestReposForFilter() []cmd.Repo {
	return []cmd.Repo{
		{
			Name:            "repo1",
			StargazerCount:  100,
			ForkCount:       10,
			Issues:          cmd.Count{TotalCount: 5},
			DiskUsage:       1000,
			PrimaryLanguage: cmd.Language{Name: "Go"},
			CreatedAt:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			IsFork:          false,
			IsArchived:      false,
			IsPrivate:       false,
			IsTemplate:      false,
		},
		{
			Name:            "repo2",
			StargazerCount:  200,
			ForkCount:       20,
			Issues:          cmd.Count{TotalCount: 15},
			DiskUsage:       2000,
			PrimaryLanguage: cmd.Language{Name: "Python"},
			CreatedAt:       time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
			IsFork:          true,
			IsArchived:      false,
			IsPrivate:       false,
			IsTemplate:      false,
		},
		{
			Name:            "repo3",
			StargazerCount:  50,
			ForkCount:       5,
			Issues:          cmd.Count{TotalCount: 2},
			DiskUsage:       500,
			PrimaryLanguage: cmd.Language{Name: "JavaScript"},
			CreatedAt:       time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			IsFork:          false,
			IsArchived:      true,
			IsPrivate:       false,
			IsTemplate:      false,
		},
		{
			Name:            "repo4",
			StargazerCount:  300,
			ForkCount:       30,
			Issues:          cmd.Count{TotalCount: 25},
			DiskUsage:       3000,
			PrimaryLanguage: cmd.Language{Name: "Go"},
			CreatedAt:       time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
			IsFork:          false,
			IsArchived:      false,
			IsPrivate:       true,
			IsTemplate:      false,
		},
		{
			Name:            "repo5",
			StargazerCount:  150,
			ForkCount:       15,
			Issues:          cmd.Count{TotalCount: 8},
			DiskUsage:       1500,
			PrimaryLanguage: cmd.Language{Name: "TypeScript"},
			CreatedAt:       time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:       time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			IsFork:          false,
			IsArchived:      false,
			IsPrivate:       false,
			IsTemplate:      true,
		},
	}
}

func TestFilterRepositories(t *testing.T) {
	repos := createTestReposForFilter()

	t.Run("filter by forked type", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "forked", "")
		if len(filtered) != 1 || filtered[0].Name != "repo2" {
			t.Errorf("Expected 1 forked repo (repo2), got %d repos", len(filtered))
		}
	})

	t.Run("filter by archived type", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "archived", "")
		if len(filtered) != 1 || filtered[0].Name != "repo3" {
			t.Errorf("Expected 1 archived repo (repo3), got %d repos", len(filtered))
		}
	})

	t.Run("filter by private type", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "private", "")
		if len(filtered) != 1 || filtered[0].Name != "repo4" {
			t.Errorf("Expected 1 private repo (repo4), got %d repos", len(filtered))
		}
	})

	t.Run("filter by template type", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "template", "")
		if len(filtered) != 1 || filtered[0].Name != "repo5" {
			t.Errorf("Expected 1 template repo (repo5), got %d repos", len(filtered))
		}
	})

	t.Run("filter by language", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "", "Go")
		if len(filtered) != 2 {
			t.Errorf("Expected 2 Go repos, got %d repos", len(filtered))
		}
	})

	t.Run("filter by type and language", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "private", "Go")
		if len(filtered) != 1 || filtered[0].Name != "repo4" {
			t.Errorf("Expected 1 private Go repo (repo4), got %d repos", len(filtered))
		}
	})

	t.Run("no filters", func(t *testing.T) {
		filtered := cmd.FilterRepositories(repos, "", "")
		if len(filtered) != len(repos) {
			t.Errorf("Expected all %d repos, got %d repos", len(repos), len(filtered))
		}
	})
}

func TestSortRepositories(t *testing.T) {
	repos := createTestReposForFilter()

	t.Run("sort by name", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "name")
		expectedOrder := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}
		for i, expected := range expectedOrder {
			if sorted[i].Name != expected {
				t.Errorf("Expected repo at position %d to be %s, got %s", i, expected, sorted[i].Name)
			}
		}
	})

	t.Run("sort by stars", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "stars")
		if sorted[0].Name != "repo4" || sorted[0].StargazerCount != 300 {
			t.Errorf("Expected first repo to be repo4 with 300 stars")
		}
		if sorted[len(sorted)-1].Name != "repo3" || sorted[len(sorted)-1].StargazerCount != 50 {
			t.Errorf("Expected last repo to be repo3 with 50 stars")
		}
	})

	t.Run("sort by forks", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "forks")
		if sorted[0].Name != "repo4" || sorted[0].ForkCount != 30 {
			t.Errorf("Expected first repo to be repo4 with 30 forks")
		}
	})

	t.Run("sort by issues", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "issues")
		if sorted[0].Name != "repo4" || sorted[0].Issues.TotalCount != 25 {
			t.Errorf("Expected first repo to be repo4 with 25 issues")
		}
	})

	t.Run("sort by language", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "language")
		expectedOrder := []string{"repo1", "repo4", "repo3", "repo2", "repo5"} // Go, Go, JavaScript, Python, TypeScript
		for i, expected := range expectedOrder {
			if sorted[i].Name != expected {
				t.Errorf("Expected repo at position %d to be %s, got %s", i, expected, sorted[i].Name)
			}
		}
	})

	t.Run("sort by updated", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "updated")
		if sorted[0].Name != "repo4" { // Most recent update: 2023-08-01
			t.Errorf("Expected first repo to be repo4 (most recently updated)")
		}
	})

	t.Run("sort by created", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "created")
		if sorted[0].Name != "repo5" { // Most recent creation: 2023-05-01
			t.Errorf("Expected first repo to be repo5 (most recently created)")
		}
	})

	t.Run("sort by size", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "size")
		if sorted[0].Name != "repo4" || sorted[0].DiskUsage != 3000 {
			t.Errorf("Expected first repo to be repo4 with size 3000")
		}
	})

	t.Run("no sorting", func(t *testing.T) {
		sorted := cmd.SortRepositories(repos, "")
		if len(sorted) != len(repos) {
			t.Errorf("Expected same number of repos")
		}
		// Should maintain original order
		if sorted[0].Name != "repo1" {
			t.Errorf("Expected first repo to remain repo1")
		}
	})
}
