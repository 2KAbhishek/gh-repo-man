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

## Development

### Building from source

1.  Clone the repository:

    ```sh
    git clone https://github.com/2KAbhishek/gh-repo-manager.git
    cd gh-repo-manager
    ```

2.  Build the executable:

    ```sh
    go build -o gh-repo-manager main.go
    ```

3.  Install the extension locally:

    ```sh
    gh extension install .
    ```

### Running Tests

To run all tests:

```sh
go test ./...
```

## Todos

*   [x] Initialize Go module.
*   [x] Set up the basic project structure.
*   [x] Implement the `gh repo-manager` command.
*   [x] Fetch and display a user's repositories.
*   [x] Implement the interactive UI for browsing repositories.
*   [x] Add tests for fetching repositories.
*   [x] Implement the multi-clone functionality.
*   [x] Add tests for the cloning logic.
*   [x] Implement the `--user` flag to browse another user's repositories.
*   [x] Add tests for the `--user` flag.
*   [x] Refine the UI and add more repository details.
*   [x] Add error handling for API requests and cloning errors.
*   [x] Write comprehensive documentation.
*   [x] Create a release workflow.
