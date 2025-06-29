# gh repo-manager

A `gh` CLI extension to manage your GitHub repositories. Browse, clone, and manage your repos with ease.

## Features

*   **Interactive UI:** A terminal-based UI to browse your repositories.
*   **Multi-clone:** Clone multiple repositories at once.
*   **Browse User Repos:** Browse the public repositories of any user on GitHub.
*   **Repository Details:** View details of a repository, including its description, stars, and forks.

## Installation

1.  Install the `gh` CLI. See the official [installation guide](https://github.com/cli/cli#installation).
2.  Install the extension:

    ```sh
    gh extension install 2KAbhishek/gh-repo-manager
    ```

## Usage

### Browse Your Repositories

To browse your own repositories, simply run:

```sh
gh repo-manager
```

This will open an interactive terminal UI where you can browse your repositories, view their details, and select them for cloning.

### Browse Another User's Repositories

To browse the public repositories of another user, use the `--user` flag:

```sh
gh repo-manager --user <username>
```

For example:

```sh
gh repo-manager --user "torvalds"
```

### Clone Repositories

In the interactive UI, you can select multiple repositories to clone. Once you've made your selection, press `Enter` to clone them to your local machine.

## Todos

*   [ ] Initialize Go module.
*   [ ] Set up the basic project structure.
*   [ ] Implement the `gh repo-manager` command.
*   [ ] Fetch and display a user's repositories.
*   [ ] Implement the interactive UI for browsing repositories.
*   [ ] Add tests for fetching repositories.
*   [ ] Implement the multi-clone functionality.
*   [ ] Add tests for the cloning logic.
*   [ ] Implement the `--user` flag to browse another user's repositories.
*   [ ] Add tests for the `--user` flag.
*   [ ] Refine the UI and add more repository details.
*   [ ] Add error handling for API requests and cloning errors.
*   [ ] Write comprehensive documentation.
*   [ ] Create a release workflow.