# AGENTS.md

## Project Overview
PriceNexus is a Go-based CLI tool for tracking product prices in Argentine online stores using a multi-agent architecture (LangChain Go).

## Setup & Prerequisites

1.  **Go**: Ensure Go 1.26+ is installed.
2.  **Environment**: Copy `.env.example` to `.env` and fill in required values.
    *   **OpenRouter**: Required for the Orchestrator agent (recommended model: `xiaomi/mimo-v2-flash`).
    *   **Ollama**: Required for local Web Searcher and Data Extractor agents (recommended model: `gemma4:e4b`).
3.  **Ollama Service**: If using local Ollama models, ensure the service is running:
    ```bash
    ollama serve
    ```
    Download the required model:
    ```bash
    ollama pull gemma4:e4b
    ```

## Essential Commands

Use the `Makefile` for standard operations:
-   **Build**: `make build` (creates `pricenexus` binary)
-   **Run**: `make run` (executes the CLI)
-   **Test**: `make test` (runs `go test -v ./...`)
-   **Clean**: `make clean`

Direct Go commands:
-   **Build**: `go build -o pricenexus ./cmd/cli`
-   **Test**: `go test ./...`
-   **Single Package Test**: `go test ./internal/agent/storage`

## Project Structure

```
PriceNexus/
├── cmd/cli/                  # CLI entrypoint (main.go)
├── internal/
│   ├── agent/                # Multi-agent system
│   │   ├── orchestrator/     # Main coordinator
│   │   ├── urlfinder/        # URL discovery agent
│   │   ├── pageloader/       # HTML fetching (no LLM)
│   │   ├── dataextractor/    # Data parsing agent
│   │   ├── validator/        # Result validation agent
│   │   └── storage/          # Database interaction
│   ├── scraper/              # Store-specific scrapers
│   └── db/                   # Database repository
├── docs/                     # Documentation (SETUP, ARCHITECTURE, etc.)
├── migrations/               # SQL migrations
└── prices.db                 # SQLite database (created on run)
```

## Configuration & Environment

**Key Environment Variables** (see `.env.example`):
-   `PRICE_NEXUS_ORCHESTRATOR_LLM`: Model for the orchestrator (e.g., `openrouter:xiaomi/mimo-v2-flash`).
-   `PRICE_NEXUS_WEBSEARCHER_LLM`: Model for the web searcher (e.g., `ollama:gemma4:e4b`).
-   `PRICE_NEXUS_DATAEXTRACTOR_LLM`: Model for the data extractor (e.g., `ollama:gemma4:e4b`).
-   `OPENROUTER_API_KEY`: API key for OpenRouter (mapped to `OPENAI_API_KEY` internally).

**Loading**: The app automatically loads `.env` if present in the root directory.

## Architecture & Agents

The system uses a multi-agent workflow:
1.  **Orchestrator**: Coordinates the workflow and delegates tasks.
2.  **URL Finder**: Searches the web for product URLs.
3.  **Page Loader**: Downloads HTML content (no LLM).
4.  **Data Extractor**: Extracts product data from HTML.
5.  **Validator**: Validates extracted data.
6.  **Storage**: Persists data to SQLite.

**LLM Dependencies**:
-   Orchestrator uses OpenRouter (remote API).
-   Web Searcher and Data Extractor default to Ollama (local).

## Testing

-   Run all tests: `make test` or `go test ./...`
-   Agent-specific tests: `go test ./internal/agent/...`
-   Ensure Ollama is running if testing local LLM integrations.

## Common Pitfalls / Gotchas

1.  **Ollama Not Running**: If using local models, `ollama serve` must be running in the background.
2.  **Missing API Key**: `OPENROUTER_API_KEY` is required for the orchestrator agent.
3.  **Database Locked**: If `prices.db` is locked, ensure no other process is using it.
4.  **Model Download**: Ensure `gemma4:e4b` (or configured model) is downloaded via `ollama pull`.

## References

-   [Documentation](docs/)
-   [Architecture Details](docs/ARCHITECTURE.md)
-   [LLM Configuration](docs/LLM_CONFIG.md)
