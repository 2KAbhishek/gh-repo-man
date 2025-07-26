package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var config Config
var configPath string

// SetConfig allows tests to override the config
func SetConfig(cfg Config) {
	config = cfg
}

var User string

var rootCmd = &cobra.Command{
	Use:   "gh-repo-manager",
	Short: "A gh extension to manage your repositories.",
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := GetRepos(User)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var repoNames []string
		for _, repo := range repos {
			repoNames = append(repoNames, repo.Name)
		}

		fzfCmd := exec.Command("fzf", "--multi", "--ansi", "--preview", "gh-repo-manager preview {}")
		fzfCmd.Stdin = strings.NewReader(strings.Join(repoNames, "\n"))
		var out bytes.Buffer
		fzfCmd.Stdout = &out
		fzfCmd.Stderr = os.Stderr

		err = fzfCmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode := exitError.ExitCode()
				if exitCode == 130 || exitCode == 1 {
					fmt.Println("Selection cancelled.")
					return
				}
			}
			fmt.Println("Error running fzf:", err)
			os.Exit(1)
		}

		selectedNames := strings.Split(strings.TrimSpace(out.String()), "\n")
		repoMap := BuildRepoMap(repos)
		selectedRepos := SelectReposByNames(repoMap, selectedNames)

		if len(selectedRepos) > 0 {
			fmt.Println("Cloning selected repositories...")
			err = CloneRepos(selectedRepos)
			if err != nil {
				fmt.Println("Error during cloning:", err)
				os.Exit(1)
			}
			fmt.Println("Cloning complete.")
		} else {
			fmt.Println("No repositories selected.")
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&User, "user", "u", "", "The user to fetch repositories for.")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Path to configuration file.")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		config = LoadConfig(configPath)
	}

	config = LoadConfig(DefaultConfigPath)
	rootCmd.AddCommand(PreviewCmd)
}

var PreviewCmd = &cobra.Command{
	Use:    "preview [repo-name]",
	Short:  "Show details for a repository (used by fzf preview)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoName := args[0]

		repos, err := GetRepos(User)
		if err != nil {
			fmt.Println("Error fetching repos for preview:", err)
			return
		}

		var targetRepo Repo
		for _, repo := range repos {
			if repo.Name == repoName {
				targetRepo = repo
				break
			}
		}

		if targetRepo.Name == "" {
			fmt.Printf("Repository %s not found.\n", repoName)
			return
		}

		languageIcon := GetLanguageIcon(targetRepo.PrimaryLanguage.Name)
		fmt.Printf("# %s\n\n%s Language: %s\n", targetRepo.Name, languageIcon, targetRepo.PrimaryLanguage.Name)

		if targetRepo.Description != "" {
			fmt.Printf("%s %s\n", IconInfo, targetRepo.Description)
		}

		fmt.Printf("%s [Link](%s)\n\n", IconLink, targetRepo.HTMLURL)
		fmt.Printf("%s %d  %s %d  %s %d  %s %d\n",
			IconStar, targetRepo.StargazerCount,
			IconFork, targetRepo.ForkCount,
			IconWatch, targetRepo.Watchers.TotalCount,
			IconIssue, targetRepo.Issues.TotalCount,
		)
		fmt.Printf("%s Owner: %s\n", IconOwner, targetRepo.Owner.Login)
		fmt.Printf("%s Created At: %s\n", IconCalendar, targetRepo.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("%s Last Updated: %s\n", IconClock, targetRepo.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("%s Disk Usage: %d KB\n", IconDisk, targetRepo.DiskUsage)

		if targetRepo.HomepageURL != "" {
			fmt.Printf("%s [Homepage](%s)\n", IconHome, targetRepo.HomepageURL)
		}
		if targetRepo.IsFork {
			fmt.Printf("\n%s Forked\n", IconForked)
		}
		if targetRepo.IsArchived {
			fmt.Printf("\n%s Archived\n", IconArchived)
		}
		if targetRepo.IsPrivate {
			fmt.Printf("\n%s Private\n", IconPrivate)
		}
		if targetRepo.IsTemplate {
			fmt.Printf("\n%s Template\n", IconTemplate)
		}
		if len(targetRepo.Topics) > 0 {
			fmt.Printf("\n%s Topics: %s\n", IconTag, strings.Join(targetRepo.TopicNames(), ", "))
		}

		fmt.Print("\n---\n")

		if config.ShowReadmeInPreview {
			readmeContent, err := GetReadme(targetRepo.Owner.Login + "/" + targetRepo.Name)
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
	},
}
