<div align = "center">

<h1><a href="https://github.com/2kabhishek/gh-repo-manager">gh-repo-manager</a></h1>

<a href="https://github.com/2KAbhishek/gh-repo-manager/blob/main/LICENSE">
<img alt="License" src="https://img.shields.io/github/license/2kabhishek/gh-repo-manager?style=flat&color=eee&label="> </a>

<a href="https://github.com/2KAbhishek/gh-repo-manager/graphs/contributors">
<img alt="People" src="https://img.shields.io/github/contributors/2kabhishek/gh-repo-manager?style=flat&color=ffaaf2&label=People"> </a>

<a href="https://github.com/2KAbhishek/gh-repo-manager/stargazers">
<img alt="Stars" src="https://img.shields.io/github/stars/2kabhishek/gh-repo-manager?style=flat&color=98c379&label=Stars"></a>

<a href="https://github.com/2KAbhishek/gh-repo-manager/network/members">
<img alt="Forks" src="https://img.shields.io/github/forks/2kabhishek/gh-repo-manager?style=flat&color=66a8e0&label=Forks"> </a>

<a href="https://github.com/2KAbhishek/gh-repo-manager/watchers">
<img alt="Watches" src="https://img.shields.io/github/watchers/2kabhishek/gh-repo-manager?style=flat&color=f5d08b&label=Watches"> </a>

<a href="https://github.com/2KAbhishek/gh-repo-manager/pulse">
<img alt="Last Updated" src="https://img.shields.io/github/last-commit/2kabhishek/gh-repo-manager?style=flat&color=e06c75&label="> </a>

<h3>Manage GitHub Repositories with Ease 📦🚀</h3>

</div>

gh-repo-manager is a `gh CLI extension` that allows `developers` to `browse, clone, and manage their GitHub repositories interactively`.

## ✨ Features

- **Interactive UI (FZF):** Terminal-based UI using `fzf` to browse repositories with live preview
- **Multi-clone:** Clone multiple repositories at once with concurrent operations
- **Browse User Repos:** Browse public repositories of any GitHub user
- **Repository Details:** View comprehensive repo details including description, stars, forks, and README content
- **Context Support:** Proper cancellation and timeout handling for all operations
- **Security:** Input validation and command injection prevention
- **Performance:** Concurrent cloning with semaphore-based limiting

## ⚡ Setup

### ⚙️ Requirements

- `gh` CLI >= 2.0.0
- `fzf` for interactive browsing
- Go >= 1.19 (for building from source)

### 💻 Installation

#### Via GitHub CLI Extensions

```bash
gh extension install 2KAbhishek/gh-repo-manager
```

#### From Source

```bash
git clone https://github.com/2KAbhishek/gh-repo-manager
cd gh-repo-manager
go build -o gh-repo-manager main.go
gh extension install .
```

## 🚀 Usage

```bash
USAGE:
    gh repo-manager [--user USERNAME]

Arguments:
    --user, -u: Browse repositories for a specific user (optional)

Examples:
    # Browse your own repositories
    gh repo-manager

    # Browse another user's repositories
    gh repo-manager --user torvalds

    # Interactive selection with Tab/Shift+Tab, Enter to clone
```

### Navigation

- Use arrow keys to navigate through repositories
- Press `Tab` or `Shift+Tab` to select multiple repositories
- Press `Enter` to clone selected repositories
- View repository details in the preview pane

## 🏗️ What's Next

Planning to add repository management features like creating, archiving, and updating repositories.

### ✅ To-Do

- [x] Initialize Go module and project structure
- [x] Implement basic repository fetching and display
- [x] Add interactive FZF UI with preview
- [x] Implement multi-repository cloning
- [x] Add user flag for browsing other users' repos
- [x] Add comprehensive error handling and validation
- [x] Implement concurrent cloning with proper context support
- [x] Add extensive test coverage with mocking
- [x] Extract constants and improve code organization
- [x] Add Unicode icons and language mappings
- [ ] Add configuration file support
- [ ] Add caching support for better performance
- [ ] Add repository creation functionality
- [ ] Add repository archiving/unarchiving
- [ ] Add bulk repository operations

## 🧑‍💻 Behind The Code

### 🌈 Inspiration

gh-repo-manager was inspired by the need for a more efficient way to browse and clone multiple GitHub repositories without switching between browser and terminal.

### 💡 Challenges/Learnings

- The main challenges were implementing proper context cancellation for concurrent operations and handling GitHub API rate limits
- I learned about Go's context package, concurrent programming patterns, and effective CLI tool design

### 🧰 Tooling

- [dots2k](https://github.com/2kabhishek/dots2k) — Dev Environment
- [nvim2k](https://github.com/2kabhishek/nvim2k) — Personalized Editor
- [sway2k](https://github.com/2kabhishek/sway2k) — Desktop Environment
- [qute2k](https://github.com/2kabhishek/qute2k) — Personalized Browser

### 🔍 More Info

- [GitHub CLI](https://github.com/cli/cli) — GitHub's official CLI tool
- [fzf](https://github.com/junegunn/fzf) — Command-line fuzzy finder
- [Cobra](https://github.com/spf13/cobra) — CLI framework for Go

<hr>

<div align="center">

<strong>⭐ hit the star button if you found this useful ⭐</strong><br>

<a href="https://github.com/2KAbhishek/gh-repo-manager">Source</a>
| <a href="https://2kabhishek.github.io/blog" target="_blank">Blog </a>
| <a href="https://twitter.com/2kabhishek" target="_blank">Twitter </a>
| <a href="https://linkedin.com/in/2kabhishek" target="_blank">LinkedIn </a>
| <a href="https://2kabhishek.github.io/links" target="_blank">More Links </a>
| <a href="https://2kabhishek.github.io/projects" target="_blank">Other Projects </a>

</div>
