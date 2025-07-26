package cmd_test

import (
	"fmt"
	"github.com/2KAbhishek/gh-repo-manager/cmd"
	"strings"
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
