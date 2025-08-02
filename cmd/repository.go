package cmd

import (
	"fmt"
	"sort"
	"strings"
)

// BuildRepoMap creates a name-to-repo lookup map
func BuildRepoMap(repos []Repo) map[string]Repo {
	repoMap := make(map[string]Repo, len(repos))
	for _, repo := range repos {
		repoMap[repo.Name] = repo
	}
	return repoMap
}

// BuildRepoPreview creates a repository preview string
func BuildRepoPreview(repo Repo) string {
	var b strings.Builder

	languageIcon := GetLanguageIcon(repo.PrimaryLanguage.Name)
	b.WriteString(fmt.Sprintf("# %s\n\n%s Language: %s\n", repo.Name, languageIcon, repo.PrimaryLanguage.Name))

	if repo.Description != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", GetIcon("info"), repo.Description))
	}

	b.WriteString(fmt.Sprintf("%s [Link](%s)\n\n", GetIcon("link"), repo.HTMLURL))
	b.WriteString(fmt.Sprintf("%s %d  %s %d  %s %d  %s %d\n",
		GetIcon("star"), repo.StargazerCount,
		GetIcon("fork"), repo.ForkCount,
		GetIcon("watch"), repo.Watchers.TotalCount,
		GetIcon("issue"), repo.Issues.TotalCount,
	))
	b.WriteString(fmt.Sprintf("%s Owner: %s\n", GetIcon("owner"), repo.Owner.Login))
	b.WriteString(fmt.Sprintf("%s Created At: %s\n", GetIcon("calendar"), repo.CreatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("%s Last Updated: %s\n", GetIcon("clock"), repo.UpdatedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("%s Disk Usage: %d KB\n", GetIcon("disk"), repo.DiskUsage))

	if repo.HomepageURL != "" {
		b.WriteString(fmt.Sprintf("%s [Homepage](%s)\n", GetIcon("home"), repo.HomepageURL))
	}
	if repo.IsFork {
		b.WriteString(fmt.Sprintf("\n%s Forked\n", GetIcon("forked")))
	}
	if repo.IsArchived {
		b.WriteString(fmt.Sprintf("\n%s Archived\n", GetIcon("archived")))
	}
	if repo.IsPrivate {
		b.WriteString(fmt.Sprintf("\n%s Private\n", GetIcon("private")))
	}
	if repo.IsTemplate {
		b.WriteString(fmt.Sprintf("\n%s Template\n", GetIcon("template")))
	}
	if len(repo.Topics) > 0 {
		b.WriteString(fmt.Sprintf("\n%s Topics: %s\n", GetIcon("tag"), strings.Join(repo.TopicNames(), ", ")))
	}

	if config.UI.ShowReadmeInPreview {
		b.WriteString("\n---\n")
		readmeContent, err := GetReadme(repo.Owner.Login + "/" + repo.Name)
		if err != nil {
			b.WriteString(fmt.Sprintf("Error fetching README: %s\n", err))
		} else if readmeContent != "" {
			b.WriteString(readmeContent)
			b.WriteString("\n")
		} else {
			b.WriteString("No README found.\n")
		}
	}

	return b.String()
}

// SelectReposByNames filters repositories by name using map lookup
func SelectReposByNames(repoMap map[string]Repo, selectedNames []string) []Repo {
	var selectedRepos []Repo
	for _, name := range selectedNames {
		if name != "" {
			if repo, exists := repoMap[name]; exists {
				selectedRepos = append(selectedRepos, repo)
			}
		}
	}
	return selectedRepos
}

// FilterRepositories filters repositories based on type and language
func FilterRepositories(repos []Repo, repoType, language string) []Repo {
	if repoType == "" && language == "" {
		return repos
	}

	var filtered []Repo
	for _, repo := range repos {
		// Filter by type
		if repoType != "" {
			switch strings.ToLower(repoType) {
			case "archived":
				if !repo.IsArchived {
					continue
				}
			case "forked":
				if !repo.IsFork {
					continue
				}
			case "private":
				if !repo.IsPrivate {
					continue
				}
			case "template":
				if !repo.IsTemplate {
					continue
				}
			default:
			}
		}

		if language != "" && !strings.EqualFold(repo.PrimaryLanguage.Name, language) {
			continue
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// SortRepositories sorts repositories based on the specified criteria
func SortRepositories(repos []Repo, sortBy string) []Repo {
	if sortBy == "" {
		return repos
	}

	sorted := make([]Repo, len(repos))
	copy(sorted, repos)

	switch strings.ToLower(sortBy) {
	case "created":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
		})
	case "forks":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].ForkCount > sorted[j].ForkCount
		})
	case "issues":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Issues.TotalCount > sorted[j].Issues.TotalCount
		})
	case "language":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].PrimaryLanguage.Name) < strings.ToLower(sorted[j].PrimaryLanguage.Name)
		})
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		})
	case "pushed", "updated":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].UpdatedAt.After(sorted[j].UpdatedAt)
		})
	case "size":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].DiskUsage > sorted[j].DiskUsage
		})
	case "stars":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].StargazerCount > sorted[j].StargazerCount
		})
	}

	return sorted
}
