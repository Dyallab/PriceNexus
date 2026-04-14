# AGENTS.md

## Project Overview
PriceNexus is a Go-based CLI tool for tracking product prices in Argentine online stores using a multi-agent architecture (LangChain Go). It uses OpenRouter's web_search server tool to dynamically discover products in Argentine stores (.com.ar and .ar domains only).

## Setup & Prerequisites

1.  **Go**: Ensure Go 1.21+ is installed.
2.  **Environment**: Copy `.env.example` to `.env` and fill in required values.
    *   **OpenRouter**: Required for Orchestrator and Web Searcher agents (recommended model: `xiaomi/mimo-v2-flash`).
        - Sign up at https://openrouter.ai
        - Generate an API key and set `OPENROUTER_API_KEY=your_key_here`
    *   **Ollama**: Required for local Data Extractor and Validator agents (recommended model: `gemma4:e4b`).
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
│   │   ├── websearcher/      # Web search via OpenRouter web_search tool
│   │   ├── pageloader/       # HTML fetching (no LLM)
│   │   ├── dataextractor/    # Data parsing agent (Ollama local)
│   │   ├── validator/        # Result validation agent (Ollama local)
│   │   ├── storage/          # Database interaction
│   │   ├── openrouter.go     # OpenRouter client with tool support
│   │   └── config.go         # LLM configuration
│   ├── scraper/              # Store-specific scrapers
│   └── db/                   # Database repository
├── docs/                     # Documentation (SETUP, ARCHITECTURE, LLM_CONFIG, etc)
├── migrations/               # SQL migrations
└── prices.db                 # SQLite database (created on run)
```

## Configuration & Environment

**Key Environment Variables** (see `.env.example`):
-   `OPENROUTER_API_KEY`: API key for OpenRouter (required for Orchestrator and Web Searcher).
-   `PRICE_NEXUS_ORCHESTRATOR_LLM`: Model for the orchestrator (default: `openrouter:xiaomi/mimo-v2-flash`).
-   `PRICE_NEXUS_WEBSEARCHER_LLM`: Model for the web searcher (default: `openrouter:xiaomi/mimo-v2-flash`). **Must be OpenRouter** - uses web_search tool.
-   `PRICE_NEXUS_DATAEXTRACTOR_LLM`: Model for the data extractor (default: `ollama:gemma4:e4b`).

**Loading**: The app automatically loads `.env` if present in the root directory.

## Architecture & Agents

The system uses a multi-agent workflow:

1.  **Orchestrator**: 
    - Coordinates the workflow and delegates tasks
    - LLM: OpenRouter (xiaomi/mimo-v2-flash recommended)

2.  **Web Searcher**: 
    - Searches the web for product URLs using OpenRouter's `openrouter:web_search` server tool
    - Automatically filters results to Argentine domains (.com.ar, .ar only)
    - LLM: OpenRouter (xiaomi/mimo-v2-flash recommended)
    - **No longer hardcodes store URLs** - discovers tiendas dynamically

3.  **Page Loader**: 
    - Downloads HTML content (no LLM required)

4.  **Data Extractor**: 
    - Extracts product data from HTML (prices, stock, shipping info)
    - LLM: Ollama local (gemma4:e4b recommended)

5.  **Validator**: 
    - Validates extracted data for accuracy
    - LLM: Ollama local (gemma4:e4b recommended)

6.  **Storage**: 
    - Persists data to SQLite
    - No LLM required

### Web Search with OpenRouter

The Web Searcher uses the **OpenRouter server tool** `openrouter:web_search`:

```json
{
  "type": "openrouter:web_search",
  "parameters": {
    "allowed_domains": [".com.ar", ".ar"],
    "max_results": 10
  }
}
```

**Benefits:**
- ✅ Dynamic search - discovers new Argentine stores automatically
- ✅ No hardcoded store list - flexible and future-proof
- ✅ Legal and official - no scraping, uses OpenRouter's official tools
- ✅ Intelligent - the model decides what to search for
- ✅ Precise - automatically filters to Argentine domains only
- ✅ Cost-effective - ~$0.04 per search ($4 per 1,000 results)

### Removed Components

The following have been **removed** in favor of dynamic search:

- ❌ **URLFinder Agent**: Previously hardcoded store URLs - no longer needed
- ❌ **pricetools/SearchTool**: Manual DuckDuckGo/Bing scraping - replaced by OpenRouter web_search

All URL discovery is now handled dynamically by the Web Searcher using OpenRouter's official tools.

## LLM Dependencies

-   **Orchestrator**: OpenRouter (remote API, requires `OPENROUTER_API_KEY`)
-   **Web Searcher**: OpenRouter (remote API, requires `OPENROUTER_API_KEY`, uses `web_search` tool)
-   **Data Extractor**: Ollama local (requires Ollama running locally)
-   **Validator**: Ollama local (requires Ollama running locally)

## Testing

-   Run all tests: `make test` or `go test ./...`
-   Agent-specific tests: `go test ./internal/agent/...`
-   Ensure Ollama is running if testing local LLM integrations.

## Common Pitfalls / Gotchas

1.  **Ollama Not Running**: If using local models, `ollama serve` must be running in the background.
2.  **Missing API Key**: `OPENROUTER_API_KEY` is required for the orchestrator and web searcher agents.
3.  **Web Searcher Must Use OpenRouter**: The web searcher requires OpenRouter because it uses the `openrouter:web_search` tool. You cannot use Ollama local for the web searcher.
4.  **Database Locked**: If `prices.db` is locked, ensure no other process is using it.
5.  **Model Download**: Ensure `gemma4:e4b` (or configured model) is downloaded via `ollama pull`.
6.  **Wrong Domain Filters**: The web search is configured to ONLY accept `.com.ar` and `.ar` domains - other domains are automatically filtered out.

## References

-   [Documentation](docs/)
-   [Architecture Details](docs/ARCHITECTURE.md)
-   [LLM Configuration](docs/LLM_CONFIG.md)
-   [Setup Guide](docs/SETUP.md)
-   [OpenRouter Web Search Docs](https://openrouter.ai/docs/guides/features/server-tools/web-search)