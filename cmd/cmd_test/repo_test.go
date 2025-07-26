package cmd_test

import (
	"fmt"
	"github.com/2KAbhishek/gh-repo-manager/cmd"
	"reflect"
	"strings"
	"testing"
)

type repoData struct {
	name, language, description, url, owner, createdAt, updatedAt, homepage, readmeContent string
	stars, forks, watchers, issues, diskUsage                                              int
	topics                                                                                 []string
}

// buildExpectedPreviewOutput builds expected preview output using actual icon constants
func buildExpectedPreviewOutput(repoName, language, description, url string, stars, forks, watchers, issues int,
	owner, createdAt, updatedAt string, diskUsage int, homepage string, topics []string, readmeContent string) string {

	data := repoData{
		name: repoName, language: language, description: description, url: url,
		stars: stars, forks: forks, watchers: watchers, issues: issues,
		owner: owner, createdAt: createdAt, updatedAt: updatedAt,
		diskUsage: diskUsage, homepage: homepage, topics: topics, readmeContent: readmeContent,
	}

	return buildPreviewOutput(data)
}

func buildPreviewOutput(data repoData) string {
	languageIcon := cmd.GetLanguageIcon(data.language)
	output := fmt.Sprintf("# %s\n\n%s Language: %s\n", data.name, languageIcon, data.language)

	if data.description != "" {
		output += fmt.Sprintf("%s %s\n", cmd.IconInfo, data.description)
	}

	output += fmt.Sprintf("%s [Link](%s)\n\n", cmd.IconLink, data.url)
	output += fmt.Sprintf("%s %d  %s %d  %s %d  %s %d\n",
		cmd.IconStar, data.stars, cmd.IconFork, data.forks,
		cmd.IconWatch, data.watchers, cmd.IconIssue, data.issues)

	output += fmt.Sprintf("%s Owner: %s\n", cmd.IconOwner, data.owner)
	output += fmt.Sprintf("%s Created At: %s\n", cmd.IconCalendar, data.createdAt)
	output += fmt.Sprintf("%s Last Updated: %s\n", cmd.IconClock, data.updatedAt)
	output += fmt.Sprintf("%s Disk Usage: %d KB\n", cmd.IconDisk, data.diskUsage)

	if data.homepage != "" {
		output += fmt.Sprintf("%s [Homepage](%s)\n", cmd.IconHome, data.homepage)
	}

	if len(data.topics) > 0 {
		output += fmt.Sprintf("\n%s Topics: %s\n", cmd.IconTag, strings.Join(data.topics, ", "))
	}

	return output + "\n---\n" + data.readmeContent + "\n"
}

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
		{"Unknown Language", "📁"},
		{"", "📁"},
	}

	for _, test := range tests {
		result := cmd.GetLanguageIcon(test.language)
		if result != test.expected {
			t.Errorf("GetLanguageIcon(%q) = %q, want %q", test.language, result, test.expected)
		}
	}
}
