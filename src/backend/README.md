# Go Event & Betting Microservice

[![Go Report Card](https://goreportcard.com/badge/github.com/Arlan-Z/Golang-Event-Microservice)](https://goreportcard.com/report/github.com/Arlan-Z/Golang-Event-Microservice)
<!-- Добавьте другие значки по желанию (лицензия, сборка и т.д.) -->

This Go microservice manages betting events, allows users to place bets, processes payouts upon event finalization, and synchronizes event data from an external API source.

## Features

*   **Event Listing:** Displays currently active events available for betting (fetched from a local cache updated periodically).
*   **Bet Placement:** Allows users to place bets on active events, recording the odds at the time of the bet.
*   **Event Synchronization:** Periodically fetches event data (including status and results) from a configured external API and updates the local database.
*   **Automatic Finalization:** Automatically triggers bet calculation and payout notifications when an event's result (Win/Loss/Draw) is detected during synchronization.
*   **Manual Finalization:** Provides an API endpoint to manually trigger event finalization.
*   **Bet Cancellation:** Automatically cancels pending bets for events marked as "Canceled" by the external API source.
*   **Payout Notification:** Notifies a configured external payout service about winning bets.
*   **Health Checks:** Includes `/healthz` (liveness) and `/readyz` (readiness) probes.
*   **Structured Logging:** Uses `zap` for structured logging.
*   **Configuration:** Flexible configuration via `config.yaml` and environment variables.
*   **Database Migrations:** Manages database schema using `golang-migrate`.

## Prerequisites

*   **Go:** Version 1.18 or higher.
*   **Git:** For cloning the repository.
*   **SQLite3 Library:** Ensure the necessary build tools (like gcc/clang) are installed for the CGo dependency required by `mattn/go-sqlite3`. The library itself is managed by Go modules.
*   **golang-migrate CLI (Optional but Recommended):** For easier database schema management. [Installation Guide](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate).

## Getting Started

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/Arlan-Z/Golang-Event-Microservice.git
    cd Golang-Event-Microservice
    ```

2.  **Install Dependencies:**
    Fetch and install the required Go modules.
    ```bash
    go mod tidy
    # or
    go mod download
    ```

## Configuration

The service uses a `config.yaml` file in the project root and/or environment variables (environment variables override file values).

Create `config.yaml` in the root directory:

```yaml
# config.yaml
http_server:
  port: "8080"         # Port the service listens on (Env: HTTP_PORT)
  timeout: "5s"        # Request Read Timeout (Env: HTTP_TIMEOUT)

database:
  path: "./data/events.db"  # Path to the SQLite database file (Env: DB_PATH) - Ensure './data' directory exists!

payout_service:
  url: "http://localhost:8081" # Base URL of the external payout service (Env: PAYOUT_SVC_URL) - REQUIRED
  timeout: "3s"                # HTTP client timeout for the payout service (Env: PAYOUT_SVC_TIMEOUT)

event_source_api:              # External API for fetching events
  url: "https://arlan-api.azurewebsites.net" 
  timeout: "10s"               # HTTP client timeout for the event source API (Env: EVENT_SOURCE_TIMEOUT)
  sync_interval: "1m"          # How often to sync events (e.g., 1m, 5m, 30s) (Env: EVENT_SYNC_INTERVAL)
```

**Key Configuration Options & Environment Variables:**

*   `http_server.port` / `HTTP_PORT`: Port for the HTTP server.
*   `database.path` / `DB_PATH`: Filesystem path for the SQLite database. **The directory (`./data/` in the example) must exist.**
*   `payout_service.url` / `PAYOUT_SVC_URL`: **Required.** Base URL for the payout notification service.
*   `payout_service.timeout` / `PAYOUT_SVC_TIMEOUT`: Timeout for payout service requests.
*   `event_source_api.url` / `EVENT_SOURCE_URL`: **Required.** Base URL of the external API providing event data (Your C# service). **Remember to replace the default `http://localhost:5000`**.
*   `event_source_api.timeout` / `EVENT_SOURCE_TIMEOUT`: Timeout for event source API requests.
*   `event_source_api.sync_interval` / `EVENT_SYNC_INTERVAL`: Frequency of event synchronization.

## Database Migrations

Database schema changes are managed using `golang-migrate`. Migration files are in `./migrations`.

**Before running migrations:** Ensure the directory specified in `database.path` (e.g., `./data/`) exists.
```bash
mkdir -p ./data
```

**Run Migrations using one of the methods:**

**Method 1: `golang-migrate` CLI (Recommended)**
```bash
# Apply all pending 'up' migrations (use the DB path from your config):
migrate -database 'sqlite3://./data/events.db' -path ./migrations up

# Roll back the last applied migration:
# migrate -database 'sqlite3://./data/events.db' -path ./migrations down 1
```

**Method 2: Built-in Go Utility**
```bash
# Apply all pending 'up' migrations (use the DB path from your config):
go run ./cmd/migrate/main.go -dbpath ./data/events.db -path ./migrations -direction up

# Roll back all 'down' migrations (use with caution):
# go run ./cmd/migrate/main.go -dbpath ./data/events.db -path ./migrations -direction down
```

The database file (e.g., `./data/events.db`) will be created if it doesn't exist.

## Running the Service

1.  **Ensure `config.yaml` is present and correctly configured (especially `event_source_api.url`).**
2.  **Ensure the database directory exists and migrations have been applied.**
3.  **Ensure the external Event Source API (your C# service) and Payout Service (if testing payouts) are running and accessible.**
4.  **Run the application:**

    ```bash
    # Option 1: Build and run
    go build -o betting_service ./cmd/app/main.go
    ./betting_service

    # Option 2: Run directly
    go run ./cmd/app/main.go
    ```

The service will start, log initialization steps, begin synchronizing events from the configured source API, and listen for incoming HTTP requests on the configured port (e.g., `:8080`).

## API Interaction

All API endpoints are prefixed with `/api/v1`.

*   **`GET /api/v1/events`**
    *   **Description:** Retrieves a list of currently **active** events available for betting. Data comes from the local database cache, updated by the background syncer.
    *   **Response:** `200 OK` with a JSON array of `EventDTO` objects. Returns `[]` if no active events are found.
        ```json
        [
          {
            "id": "cefdce70-bd86-430d-836a-5fa6e072e13f",
            "eventName": "Golang VS C#",
            "homeTeam": "Golang",
            "awayTeam": "C#",
            "homeWinChance": 0.1309,
            "awayWinChance": 0.2849,
            "drawChance": 0.5841,
            "eventStartDate": "2025-04-10T15:47:00Z", // Time in UTC
            "eventEndDate": "2025-05-04T15:47:00Z",
            "eventResult": null, // null because the event is active
            "type": "Other"
          }
          // ... other active events
        ]
        ```

*   **`POST /api/v1/bets`**
    *   **Description:** Places a new bet on an **active** event. Records the current odds at the time the bet is placed.
    *   **Request Body (JSON):**
        ```json
        {
          "userId": "valid-uuid-string",   // User's unique identifier (UUID format)
          "eventId": "event-uuid-string", // ID of an ACTIVE event (UUID format)
          "amount": 10.50,                // Bet amount (must be > 0)
          "predictedOutcome": "HomeWin"   // "HomeWin", "AwayWin", or "Draw"
        }
        ```
    *   **Response:**
        *   `201 Created`: Bet successfully placed. Returns the created `BetDTO` object.
        *   `400 Bad Request`: Invalid request body (bad JSON, validation errors like non-UUIDs, amount <= 0, invalid outcome).
        *   `404 Not Found`: Event with the given `eventId` not found in the local database.
        *   `409 Conflict`: Event is not active (already finished, canceled, or not started depending on exact logic).
        *   `500 Internal Server Error`: Failure saving the bet to the database.

*   **`POST /api/v1/events/{eventID}/finalize`**
    *   **Description:** **Manually** triggers the finalization process for a specific event. Calculates pending bets and initiates payout notifications. *This is usually handled automatically by the syncer but can be used as a fallback or for testing.*
    *   **Path Parameter:** `{eventID}` - UUID of the event to finalize.
    *   **Request Body (JSON):**
        ```json
        {
          "result": "AwayWin" // The actual outcome: "HomeWin", "AwayWin", or "Draw"
        }
        ```
    *   **Response:**
        *   `200 OK`: Finalization process completed or successfully initiated (check logs for details, especially if bets failed).
        *   `400 Bad Request`: Invalid `result` value provided.
        *   `404 Not Found`: Event with the given ID not found.
        *   `409 Conflict`: Event was already finalized previously.
        *   `500 Internal Server Error`: Error during bet processing or payout notification.

*   **`GET /api/v1/healthz`**
    *   **Description:** Liveness probe. Indicates if the HTTP server is running.
    *   **Response:** `200 OK`

*   **`GET /api/v1/readyz`**
    *   **Description:** Readiness probe. Indicates if the service is ready to handle traffic (e.g., database connection is available).
    *   **Response:**
        *   `200 OK`: Service is ready.
        *   `503 Service Unavailable`: Service is not ready (e.g., DB ping failed).

## Automatic Event Processing (EventSyncer)

The service includes a background worker (`EventSyncer`) that runs periodically (defined by `event_source_api.sync_interval`):

1.  **Fetches All Events:** It calls `GET {event_source_api.url}/api/Events/all` (based on the C# controller).
2.  **Updates Local DB:** It uses `Upsert` to add new events or update existing event details (name, teams, odds, dates, status) in the local SQLite database.
3.  **Detects Finalization:** If the fetched data for an event includes a final result (`HomeWin`, `AwayWin`, `Draw`), the syncer automatically calls the internal `EventUseCase.FinalizeEvent` method. This triggers the calculation of winning/losing bets and sends payout notifications, just like the manual API call.
4.  **Detects Cancellation:** If the fetched data indicates an event is `Canceled`, the syncer marks the event as inactive locally and calls the internal `BetUseCase.CancelBetsForEvent` method to change the status of all pending bets for that event to `Canceled`.

This automation means you generally don't need to manually call the `/finalize` endpoint if your external event source API reliably updates event statuses and results.

## Project Structure

The project follows a standard Go layered architecture:

*   `cmd/`: Application entry points (`app` for the service, `migrate` for DB utility).
*   `internal/`: Private application code.
    *   `app/`: Core application setup (config, connections, startup, store).
    *   `data/`: Data structures (DB models, API DTOs).
    *   `deliveries/`: Adapters for input/output (HTTP handlers, external API clients).
    *   `pkg/`: Internal shared utility packages (validator).
    *   `repositories/`: Data access layer (SQLite implementations).
    *   `services/`: Service layer coordinating between delivery and use cases.
    *   `usecases/`: Core business logic.
    *   `sync/`: Background worker logic (EventSyncer).
*   `migrations/`: SQL database migration files.
*   `config.yaml`: Default configuration file.
*   `go.mod`, `go.sum`: Go module definition files.

## Testing

Unit and integration tests are included. Run tests using the standard Go command from the project root:

```bash
go test ./...
```

Tests cover:
*   Repository interactions with a test SQLite database.
*   Use case business logic using mocked dependencies.
*   HTTP handler responses using `httptest`.

---