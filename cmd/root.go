package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var user string

var rootCmd = &cobra.Command{
	Use:   "gh-repo-manager",
	Short: "A gh extension to manage your repositories.",
	Run: func(cmd *cobra.Command, args []string) {
		repos, err := GetRepos(user)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		p := tea.NewProgram(initialModel(repos))
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
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
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "The user to fetch repositories for.")
}
