# Event & Betting Service

A Go microservice designed to manage betting events, allow users to place bets on these events, and process payouts upon event finalization by notifying an external payout service.

## Features

*   List active betting events.
*   Place bets on active events, recording the odds at the time of betting.
*   Finalize events based on a provided outcome.
*   Calculate winning bets based on recorded odds.
*   Notify an external payout service for winning bets.
*   Health checks (Liveness and Readiness probes).

## Prerequisites

*   **Go:** Version 1.18 or higher (due to generics usage, adjust if needed).
*   **Git:** For cloning the repository and fetching dependencies.
*   **SQLite3:** The service uses SQLite as its database. Ensure the `sqlite3` library is available if interacting with the database directly or running the `migrate` tool locally.
*   **golang-migrate CLI (Optional but Recommended):** For managing database schema changes. Install instructions: [golang-migrate installation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

## Getting Started

1.  **Clone the repository:**
    ```bash
    git clone <your-repository-url> # e.g., git clone git@gitlab.frhc.one:your-team/def-betting-api.git
    cd def-betting-api
    ```

2.  **Private Dependencies (If Applicable):**
    *This section is only necessary if the project uses private Go modules hosted on `gitlab.frhc.one`.*
    *   Ensure your SSH key is configured for passwordless access to GitLab:
        ```bash
        ssh -T git@gitlab.frhc.one
        # Should respond with "Welcome to GitLab, @your-username!" without asking for a password.
        ```
    *   Configure Go to use SSH for your private repository host:
        ```bash
        go env -w GOPRIVATE=gitlab.frhc.one/*
        ```
    *   Configure Git to rewrite HTTPS URLs to SSH for the private host. Add the following to your global `~/.gitconfig` file (or the project's `.git/config`):
        ```git
        [url "ssh://git@gitlab.frhc.one/"]
            insteadOf = https://gitlab.frhc.one/
        ```

3.  **Install Dependencies:**
    Fetch the required Go modules.
    ```bash
    go mod tidy
    # or
    go mod download
    ```

## Configuration

The service is configured using a `config.yaml` file located in the project root and/or environment variables. Environment variables override values from the config file.

Create a `config.yaml` file in the project root directory:

```yaml
# config.yaml
http_server:
  port: "8080"         # Port the service listens on
  timeout: "5s"        # Request Read Timeout

database:
  path: "./events.db"  # Path to the SQLite database file

payout_service:
  url: "http://localhost:8081" # Base URL of the external payout service
  timeout: "3s"                # HTTP client timeout for the payout service