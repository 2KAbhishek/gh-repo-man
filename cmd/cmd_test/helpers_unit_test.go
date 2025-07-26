package cmd_test

import (
	"reflect"
	"testing"

	"github.com/2KAbhishek/gh-repo-manager/cmd"
)

func TestTopicNames(t *testing.T) {
	repo := cmd.Repo{
		Topics: []cmd.Topic{
			{Name: "go"},
			{Name: "cli"},
			{Name: "github"},
		},
	}

	expected := []string{"go", "cli", "github"}
	result := repo.TopicNames()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TopicNames() = %v, want %v", result, expected)
	}
}

func TestTopicNamesEmpty(t *testing.T) {
	repo := cmd.Repo{
		Topics: []cmd.Topic{},
	}

	expected := []string{}
	result := repo.TopicNames()

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TopicNames() with empty topics = %v, want %v", result, expected)
	}
}

func TestBuildRepoMap(t *testing.T) {
	repos := []cmd.Repo{
		{Name: "repo1", Description: "desc1"},
		{Name: "repo2", Description: "desc2"},
		{Name: "repo3", Description: "desc3"},
	}

	repoMap := cmd.BuildRepoMap(repos)

	if len(repoMap) != 3 {
		t.Errorf("BuildRepoMap() returned map with %d entries, want 3", len(repoMap))
	}

	if repoMap["repo1"].Description != "desc1" {
		t.Errorf("BuildRepoMap()['repo1'].Description = %q, want 'desc1'", repoMap["repo1"].Description)
	}

	if repoMap["repo2"].Description != "desc2" {
		t.Errorf("BuildRepoMap()['repo2'].Description = %q, want 'desc2'", repoMap["repo2"].Description)
	}

	if repoMap["repo3"].Description != "desc3" {
		t.Errorf("BuildRepoMap()['repo3'].Description = %q, want 'desc3'", repoMap["repo3"].Description)
	}
}

func TestSelectReposByNames(t *testing.T) {
	repoMap := map[string]cmd.Repo{
		"repo1": {Name: "repo1", Description: "desc1"},
		"repo2": {Name: "repo2", Description: "desc2"},
		"repo3": {Name: "repo3", Description: "desc3"},
	}

	selectedNames := []string{"repo1", "repo3", "", "nonexistent"}
	result := cmd.SelectReposByNames(repoMap, selectedNames)

	expected := []cmd.Repo{
		{Name: "repo1", Description: "desc1"},
		{Name: "repo3", Description: "desc3"},
	}

	if len(result) != 2 {
		t.Errorf("SelectReposByNames() returned %d repos, want 2", len(result))
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("SelectReposByNames() = %v, want %v", result, expected)
	}
}

func TestGetLanguageIcon(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"Go", "🐹"},
		{"Python", "🐍"},
		{"JavaScript", "📜"},
		{"Unknown Language", "📁"}, // default
		{"", "📁"},                 // default for empty
	}

	for _, test := range tests {
		result := cmd.GetLanguageIcon(test.language)
		if result != test.expected {
			t.Errorf("GetLanguageIcon(%q) = %q, want %q", test.language, result, test.expected)
		}
	}
}
