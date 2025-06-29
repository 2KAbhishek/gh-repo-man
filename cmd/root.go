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

		fmt.Printf("Name: %s\n", targetRepo.Name)
		fmt.Printf("Description: %s\n", targetRepo.Description)
		fmt.Printf("SSH URL: %s\n", targetRepo.Ssh_url)
		fmt.Printf("Stars: %d\n", targetRepo.StargazerCount)
		fmt.Printf("Forks: %d\n", targetRepo.ForkCount)

		fmt.Println("\n---\n") // Horizontal line

		readmeContent, err := GetReadme(targetRepo.Name)
		if err != nil {
			fmt.Println("Error fetching README:", err)
			return
		}
		fmt.Println(readmeContent)
	},
}
