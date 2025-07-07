package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

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
			fmt.Println("Error running fzf:", err)
			os.Exit(1)
		}

		selectedNames := strings.Split(strings.TrimSpace(out.String()), "\n")
		var selectedRepos []Repo
		for _, name := range selectedNames {
			if name == "" {
				continue
			}
			for _, repo := range repos {
				if repo.Name == name {
					selectedRepos = append(selectedRepos, repo)
					break
				}
			}
		}

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

		// Determine icon based on language
		icon := "Git"
		switch targetRepo.PrimaryLanguage.Name {
		case "Go":
			icon = "Go"
		case "Python":
			icon = ""
		case "JavaScript", "TypeScript":
			icon = ""
		case "Java":
			icon = "pr"
		case "C", "C++":
			icon = "do"
		case "Ruby":
			icon = ""
		case "PHP":
			icon = ""
		case "Rust":
			icon = ""
		case "Swift":
			icon = "swift"
		case "Kotlin":
			icon = ""
		case "Shell":
			icon = "sh"
		case "HTML":
			icon = "html"
		case "CSS":
			icon = "css"
		case "Lua":
			icon = "lua"
		}

		// Format repo info similar to the Lua template
		fmt.Printf("# %s\n\n%s Language: %s\n", targetRepo.Name, icon, targetRepo.PrimaryLanguage.Name)

		if targetRepo.Description != "" {
			fmt.Printf(" %s\n", targetRepo.Description)
		}

		fmt.Printf(" [Link](%s)\n\n", targetRepo.HTMLURL)
		fmt.Printf(" %d   %d   %d   %d\n",
			targetRepo.StargazerCount,
			targetRepo.ForkCount,
			targetRepo.WatchersCount,
			targetRepo.Issues.TotalCount,
		)
		fmt.Printf(" Owner: %s\n", targetRepo.Owner.Login)
		fmt.Printf(" Created At: %s\n", targetRepo.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf(" Last Updated: %s\n", targetRepo.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf(" Disk Usage: %d KB\n", targetRepo.DiskUsage)

		if targetRepo.HomepageURL != "" {
			fmt.Printf(" [Homepage](%s)\n", targetRepo.HomepageURL)
		}
		if targetRepo.IsFork {
			fmt.Println("\n>  Forked")
		}
		if targetRepo.IsArchived {
			fmt.Println("\n>  Archived")
		}
		if targetRepo.IsPrivate {
			fmt.Println("\n>  Private")
		}
		if targetRepo.IsTemplate {
			fmt.Println("\n>  Template")
		}
		if len(targetRepo.Topics) > 0 {
			fmt.Printf("\n Topics: %s\n", strings.Join(targetRepo.Topics, ", "))
		}

		fmt.Print("\n---\n") // Horizontal line

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
	},
}
