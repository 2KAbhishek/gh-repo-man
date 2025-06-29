package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	repos   []Repo
	cursor  int
	selected map[int]struct{}
}

func initialModel(repos []Repo) model {
	return model{
		repos:   repos,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.repos)-1 {
				m.cursor++
			}

		case "enter":
			var reposToClone []Repo
			for i := range m.selected {
				reposToClone = append(reposToClone, m.repos[i])
			}
			if len(reposToClone) > 0 {
				err := CloneRepos(reposToClone)
				if err != nil {
					fmt.Printf("Error cloning repositories: %v\n", err)
				}
			}
			return m, tea.Quit

		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Select repositories to clone:\n\n"

	for i, repo := range m.repos {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, repo.Name)
	}

	s += "\nPress enter to clone selected repositories, or q to quit.\n"
	return s
}