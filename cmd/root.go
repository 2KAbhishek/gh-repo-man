package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	config     Config
	configPath string
)

// SetConfig allows tests to override the config
func SetConfig(cfg Config) {
	config = cfg
}

var (
	User           string
	RepoType       string
	LanguageFilter string
	SortBy         string
)
var previewUser string

var rootCmd = &cobra.Command{
	Use:   "gh-repo-man",
	Short: "A gh extension to manage your repositories.",
	Run: func(cmd *cobra.Command, args []string) {
		err := runMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

var PreviewCmd = &cobra.Command{
	Use:    "preview [repo-name]",
	Short:  "Show details for a repository (used by fzf preview)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoName := args[0]

		targetUser := previewUser
		if targetUser == "" {
			targetUser = User
		}

		repos, err := GetRepos(targetUser)
		if err != nil {
			fmt.Println("Error fetching repos for preview:", err)
			return
		}

		targetRepo := findRepoByName(repos, repoName)
		if targetRepo == nil {
			fmt.Printf("Repository %s not found.\n", repoName)
			return
		}

		displayRepoPreview(*targetRepo)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&User, "user", "u", "", "The user to fetch repositories for.")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Path to configuration file.")
	rootCmd.Flags().StringVarP(&RepoType, "type", "t", "", "Filter by repository type (archived, forked, private, template)")
	rootCmd.Flags().StringVarP(&LanguageFilter, "language", "l", "", "Filter by primary language")
	rootCmd.Flags().StringVarP(&SortBy, "sort", "s", "", "Sort repositories by (created, forks, issues, language, name, pushed, size, stars, updated)")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		SetConfigAndUpdateIcons(LoadConfig(configPath))

		if SortBy == "" {
			SortBy = config.Repos.SortBy
		}
		if RepoType == "" {
			RepoType = config.Repos.RepoType
		}
		if LanguageFilter == "" {
			LanguageFilter = config.Repos.Language
		}
	}

	SetConfigAndUpdateIcons(LoadConfig(DefaultConfigPath))

	PreviewCmd.Flags().StringVar(&previewUser, "user", "", "The user whose repositories to search for preview")
	rootCmd.AddCommand(PreviewCmd)
}

func runMain() error {
	sortedRepos, err := processRepositories(User)
	if err != nil {
		return err
	}

	repoNames := extractRepoNames(sortedRepos)

	selectedNames, err := runFzfSelection(repoNames, User)
	if err != nil {
		if err.Error() == "selection cancelled" {
			fmt.Println("Selection cancelled.")
			return nil
		}
		return err
	}

	return handleRepoSelection(selectedNames, sortedRepos)
}

func handleRepoSelection(selectedNames []string, sortedRepos []Repo) error {
	if len(selectedNames) == 0 {
		fmt.Println("No repositories selected.")
		return nil
	}

	repoMap := BuildRepoMap(sortedRepos)
	selectedRepos := SelectReposByNames(repoMap, selectedNames)

	if len(selectedRepos) > 0 {
		if config.UI.ProgressIndicators {
			fmt.Println("Cloning selected repositories...")
		}
		err := CloneRepos(selectedRepos)
		if err != nil {
			return fmt.Errorf("error during cloning: %w", err)
		}
		if config.UI.ProgressIndicators {
			fmt.Println("Cloning complete.")
		}

		err = HandlePostClone(selectedRepos)
		if err != nil {
			return fmt.Errorf("error during post-clone handling: %w", err)
		}
	} else {
		fmt.Println("No repositories selected.")
	}

	return nil
}

func buildPreviewCommand(user string) string {
	cmdInvocation := GetCommandInvocation()
	if user != "" {
		return fmt.Sprintf("%s preview {} --user %s", cmdInvocation, user)
	}
	return fmt.Sprintf("%s preview {}", cmdInvocation)
}

func runFzfSelection(repoNames []string, user string) ([]string, error) {
	previewCmd := buildPreviewCommand(user)
	fzfArgs := []string{"--multi", "--preview", previewCmd}
	if config.UI.ColorOutput {
		fzfArgs = append(fzfArgs, "--ansi")
	}

	fzfCmd := exec.Command("fzf", fzfArgs...)
	fzfCmd.Stdin = strings.NewReader(strings.Join(repoNames, "\n"))
	var out bytes.Buffer
	fzfCmd.Stdout = &out
	fzfCmd.Stderr = os.Stderr

	err := fzfCmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			if exitCode == 130 || exitCode == 1 {
				return nil, fmt.Errorf("selection cancelled")
			}
		}
		return nil, fmt.Errorf("error running fzf: %w", err)
	}

	selectedNames := strings.Split(strings.TrimSpace(out.String()), "\n")
	return selectedNames, nil
}

func processRepositories(user string) ([]Repo, error) {
	repos, err := GetRepos(user)
	if err != nil {
		return nil, err
	}

	filteredRepos := FilterRepositories(repos, RepoType, LanguageFilter)
	sortedRepos := SortRepositories(filteredRepos, SortBy)

	return sortedRepos, nil
}

func extractRepoNames(repos []Repo) []string {
	var repoNames []string
	for _, repo := range repos {
		repoNames = append(repoNames, repo.Name)
	}
	return repoNames
}

func GetCommandInvocation() string {
	if _, err := exec.LookPath("gh-repo-man"); err == nil {
		return "gh-repo-man"
	}
	return "gh repo-man"
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func findRepoByName(repos []Repo, name string) *Repo {
	for _, repo := range repos {
		if repo.Name == name {
			return &repo
		}
	}
	return nil
}

func formatRepoHeader(repo Repo) {
	languageIcon := GetLanguageIcon(repo.PrimaryLanguage.Name)
	fmt.Printf("# %s\n\n%s Language: %s\n", repo.Name, languageIcon, repo.PrimaryLanguage.Name)

	if repo.Description != "" {
		fmt.Printf("%s %s\n", GetIcon("info"), repo.Description)
	}
}

func formatRepoStats(repo Repo) {
	fmt.Printf("%s [Link](%s)\n\n", GetIcon("link"), repo.HTMLURL)
	fmt.Printf("%s %d  %s %d  %s %d  %s %d\n",
		GetIcon("star"), repo.StargazerCount,
		GetIcon("fork"), repo.ForkCount,
		GetIcon("watch"), repo.Watchers.TotalCount,
		GetIcon("issue"), repo.Issues.TotalCount,
	)
	fmt.Printf("%s Owner: %s\n", GetIcon("owner"), repo.Owner.Login)
	fmt.Printf("%s Created At: %s\n", GetIcon("calendar"), repo.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("%s Last Updated: %s\n", GetIcon("clock"), repo.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("%s Disk Usage: %d KB\n", GetIcon("disk"), repo.DiskUsage)
}

func formatRepoFlags(repo Repo) {
	if repo.HomepageURL != "" {
		fmt.Printf("%s [Homepage](%s)\n", GetIcon("home"), repo.HomepageURL)
	}
	if repo.IsFork {
		fmt.Printf("\n%s Forked\n", GetIcon("forked"))
	}
	if repo.IsArchived {
		fmt.Printf("\n%s Archived\n", GetIcon("archived"))
	}
	if repo.IsPrivate {
		fmt.Printf("\n%s Private\n", GetIcon("private"))
	}
	if repo.IsTemplate {
		fmt.Printf("\n%s Template\n", GetIcon("template"))
	}
	if len(repo.Topics) > 0 {
		fmt.Printf("\n%s Topics: %s\n", GetIcon("tag"), strings.Join(repo.TopicNames(), ", "))
	}
}

func formatRepoReadme(repo Repo) {
	if !config.UI.ShowReadmeInPreview {
		return
	}

	fmt.Print("\n---\n")
	readmeContent, err := GetReadme(repo.Owner.Login + "/" + repo.Name)
	if err != nil {
		fmt.Println("Error fetching README:", err)
		return
	}
	if readmeContent != "" {
		fmt.Println(readmeContent)
	} else {
		fmt.Println("No README found.")
	}
}

func displayRepoPreview(repo Repo) {
	formatRepoHeader(repo)
	formatRepoStats(repo)
	formatRepoFlags(repo)
	formatRepoReadme(repo)
}
