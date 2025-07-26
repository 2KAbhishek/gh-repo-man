package cmd_test

import (
	"fmt"
	"github.com/2KAbhishek/gh-repo-manager/cmd"
)

// buildExpectedPreviewOutput builds the expected preview output using the actual icon constants
func buildExpectedPreviewOutput(repoName, language, description, url string, stars, forks, watchers, issues int, 
	owner, createdAt, updatedAt string, diskUsage int, homepage string, topics []string, readmeContent string) string {
	
	languageIcon := cmd.GetLanguageIcon(language)
	
	output := fmt.Sprintf("# %s\n\n%s Language: %s\n", repoName, languageIcon, language)
	
	if description != "" {
		output += fmt.Sprintf("%s %s\n", cmd.IconInfo, description)
	}
	
	output += fmt.Sprintf("%s [Link](%s)\n\n", cmd.IconLink, url)
	output += fmt.Sprintf("%s %d  %s %d  %s %d  %s %d\n",
		cmd.IconStar, stars,
		cmd.IconFork, forks,
		cmd.IconWatch, watchers,
		cmd.IconIssue, issues,
	)
	output += fmt.Sprintf("%s Owner: %s\n", cmd.IconOwner, owner)
	output += fmt.Sprintf("%s Created At: %s\n", cmd.IconCalendar, createdAt)
	output += fmt.Sprintf("%s Last Updated: %s\n", cmd.IconClock, updatedAt)
	output += fmt.Sprintf("%s Disk Usage: %d KB\n", cmd.IconDisk, diskUsage)
	
	if homepage != "" {
		output += fmt.Sprintf("%s [Homepage](%s)\n", cmd.IconHome, homepage)
	}
	
	if len(topics) > 0 {
		topicsStr := ""
		for i, topic := range topics {
			if i > 0 {
				topicsStr += ", "
			}
			topicsStr += topic
		}
		output += fmt.Sprintf("\n%s Topics: %s\n", cmd.IconTag, topicsStr)
	}
	
	output += "\n---\n"
	output += readmeContent + "\n"
	
	return output
}