package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	ProjectsDir    string
	RefreshCache   bool
)

var (
	previewUser string
	listUser    string
)

var rootCmd = &cobra.Command{
	Use:   "gh-repo-man",
	Short: "A gh extension to manage your repositories.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if ProjectsDir != "" {
			config.Repos.ProjectsDir = ProjectsDir
		}
	},
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

		targetRepo := FindRepoByName(repos, repoName)
		if targetRepo == nil {
			fmt.Printf("Repository %s not found.\n", repoName)
			return
		}

		fmt.Print(BuildRepoPreview(*targetRepo))
	},
}

var ListCmd = &cobra.Command{
	Use:    "list",
	Short:  "List repository names (used by fzf reload)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetUser := listUser
		if targetUser == "" {
			targetUser = User
		}

		oldRefresh := RefreshCache
		RefreshCache = true
		defer func() { RefreshCache = oldRefresh }()

		sortedRepos, err := processRepositories(targetUser)
		if err != nil {
			return err
		}

		for _, name := range extractRepoNames(sortedRepos) {
			fmt.Println(name)
		}
		return nil
	},
}

func GetCommandInvocation() string {
	if exe, err := os.Executable(); err == nil && exe != "" {
		base := filepath.Base(exe)
		if !strings.HasSuffix(base, ".test") && !strings.HasPrefix(base, "___") {
			if strings.Contains(exe, " ") {
				return `"` + exe + `"`
			}
			return exe
		}
	}
	if _, err := exec.LookPath("gh-repo-man"); err == nil {
		return "gh-repo-man"
	}
	return "gh repo-man"
}

func Execute() {
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&User, "user", "u", "", "The user to fetch repositories for.")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Path to configuration file.")
	rootCmd.Flags().StringVarP(&RepoType, "type", "t", "", "Filter by repository type (archived, forked, private, template)")
	rootCmd.Flags().StringVarP(&LanguageFilter, "language", "l", "", "Filter by primary language")
	rootCmd.Flags().StringVarP(&SortBy, "sort", "s", "", "Sort repositories by (created, forks, issues, language, name, pushed, size, stars, updated)")
	rootCmd.Flags().StringVarP(&ProjectsDir, "dir", "d", "", "Directory where repositories will be cloned (overrides config)")
	rootCmd.Flags().BoolVarP(&RefreshCache, "refresh", "r", false, "Force refresh repositories from GitHub, bypassing cache")

	PreviewCmd.Flags().StringVar(&previewUser, "user", "", "The user whose repositories to search for preview")
	rootCmd.AddCommand(PreviewCmd)

	ListCmd.Flags().StringVar(&listUser, "user", "", "The user whose repositories to list")
	ListCmd.Flags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Path to configuration file.")
	ListCmd.Flags().StringVarP(&RepoType, "type", "t", "", "Filter by repository type")
	ListCmd.Flags().StringVarP(&LanguageFilter, "language", "l", "", "Filter by primary language")
	ListCmd.Flags().StringVarP(&SortBy, "sort", "s", "", "Sort repositories by")
	rootCmd.AddCommand(ListCmd)
}

func runMain() error {
	sortedRepos, err := processRepositories(User)
	if err != nil {
		return err
	}

	if len(sortedRepos) == 0 {
		fmt.Println("No repositories found matching the criteria.")
		return nil
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

	finalRepos, err := processRepositories(User)
	if err != nil {
		finalRepos = sortedRepos
	}

	return handleRepoSelection(selectedNames, finalRepos)
}

func handleRepoSelection(selectedNames []string, sortedRepos []Repo) error {
	if len(selectedNames) == 0 {
		fmt.Println("No repositories selected.")
		return nil
	}

	repoMap := BuildRepoMap(sortedRepos)
	selectedRepos := SelectReposByNames(repoMap, selectedNames)

	if len(selectedRepos) > 0 {
		err := CloneRepos(selectedRepos)
		if err != nil {
			return fmt.Errorf("error during cloning: %w", err)
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
	var parts []string
	parts = append(parts, cmdInvocation, "preview", "{}")
	if configPath != "" && configPath != DefaultConfigPath {
		parts = append(parts, "--config", configPath)
	}
	if user != "" {
		parts = append(parts, "--user", user)
	}
	return strings.Join(parts, " ")
}

func BuildReloadCommand(user string) string {
	cmdInvocation := GetCommandInvocation()
	var parts []string
	parts = append(parts, cmdInvocation, "list")
	if configPath != "" && configPath != DefaultConfigPath {
		parts = append(parts, "--config", configPath)
	}
	if user != "" {
		parts = append(parts, "--user", user)
	}
	if RepoType != "" {
		parts = append(parts, "--type", RepoType)
	}
	if LanguageFilter != "" {
		parts = append(parts, "--language", LanguageFilter)
	}
	if SortBy != "" {
		parts = append(parts, "--sort", SortBy)
	}
	return strings.Join(parts, " ")
}

func runFzfSelection(repoNames []string, user string) ([]string, error) {
	previewCmd := buildPreviewCommand(user)
	reloadCmd := BuildReloadCommand(user)
	fzfArgs := []string{
		"--multi",
		"--preview", previewCmd,
		"--bind", "ctrl-r:reload(" + reloadCmd + ")",
		"--header", "Press Ctrl+r to refresh repositories",
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

func FindRepoByName(repos []Repo, name string) *Repo {
	for i := range repos {
		if repos[i].Name == name {
			return &repos[i]
		}
	}
	return nil
}
